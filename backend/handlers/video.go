package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

type VideoHandler struct {
	cfg              *config.APIConfig
	contentGenerator *services.ContentGenerator
	elevenLabs       *services.ElevenLabsService
	backgroundMusic  *services.BackgroundMusic
	ffmpegCompiler   *services.CompositionCompiler
	veo              *services.VeoService
}

func NewVideoHandler(cfg *config.APIConfig) *VideoHandler {
	return &VideoHandler{
		cfg:              cfg,
		contentGenerator: services.NewContentGenerator(cfg),
		elevenLabs:       services.NewElevenLabsService(cfg),
		backgroundMusic:  services.NewBackgroundMusic(cfg),
		ffmpegCompiler:   services.NewCompositionCompiler(services.NewFFmpegCommandBuilder(), services.NewBackgroundMusic(cfg), services.NewElevenLabsService(cfg)),
		veo:              services.NewVeoService(cfg),
	}
}

// ServeVeoVideo streams a base64-encoded video with HTTP Range support
func (vh *VideoHandler) ServeVeoVideo(c *gin.Context) {
	id := c.Param("id")
	vi, ok := vh.veo.Get(id)
	if !ok || vi == nil || vi.Base64 == "" {
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "error": "video not found"})
		return
	}

	total := vh.veo.DecodedLen(vi.Base64)
	if total <= 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "invalid video data"})
		return
	}

	ct := vi.MIMEType
	if strings.TrimSpace(ct) == "" {
		ct = "video/mp4"
	}

	rangeHeader := c.GetHeader("Range")
	c.Header("Accept-Ranges", "bytes")
	if rangeHeader == "" {
		// Full content
		c.Header("Content-Type", ct)
		c.Header("Content-Length", fmt.Sprintf("%d", total))
		c.Status(http.StatusOK)

		enc := base64.NewDecoder(base64.StdEncoding, strings.NewReader(vi.Base64))
		buf := make([]byte, 64*1024)
		for {
			n, err := enc.Read(buf)
			if n > 0 {
				if _, werr := c.Writer.Write(buf[:n]); werr != nil {
					return
				}
				c.Writer.Flush()
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}
		}
		vh.veo.Delete(id)
		return
	}

	// Single-range support: bytes=start-end
	var start, end int64
	start = 0
	end = total - 1
	if strings.HasPrefix(rangeHeader, "bytes=") {
		r := strings.TrimPrefix(rangeHeader, "bytes=")
		parts := strings.SplitN(r, "-", 2)
		if len(parts) == 2 {
			if parts[0] != "" {
				if s, err := parseInt64(parts[0]); err == nil {
					start = s
				}
			}
			if parts[1] != "" {
				if e, err := parseInt64(parts[1]); err == nil {
					end = e
				}
			}
		}
	}
	if start < 0 {
		start = 0
	}
	if end >= total {
		end = total - 1
	}
	if start > end || start >= total {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", total))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	payload, actualEnd, derr := vh.veo.DecodeRange(vi.Base64, start, end)
	if derr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "decode error"})
		return
	}

	c.Status(http.StatusPartialContent)
	c.Header("Content-Type", ct)
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, actualEnd, total))
	c.Header("Content-Length", fmt.Sprintf("%d", len(payload)))
	_, _ = c.Writer.Write(payload)
	if start == 0 && actualEnd == total-1 {
		vh.veo.Delete(id)
	}
	return
}

func parseInt64(s string) (int64, error) {
	var x int64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid number")
		}
		x = x*10 + int64(ch-'0')
	}
	return x, nil
}

func (vh *VideoHandler) GenerateVideoReels(c *gin.Context) {
	// Enforce multipart/form-data only
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":  "Content-Type must be multipart/form-data",
			"status": "error",
		})
		return
	}

	// Buffer the original request body so we can both parse and forward it
	origBody, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("failed to read request body: %v", readErr)})
		return
	}

	//reset request body
	c.Request.Body = io.NopCloser(bytes.NewReader(origBody))

	// Parse incoming multipart form to extract and save images locally
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid multipart form"})
		return
	}
	files := form.File["image"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "at least one image is required (field name: image)"})
		return
	}

	imageTmpDir := filepath.Join(os.TempDir(), "reels_images")
	if err := os.MkdirAll(imageTmpDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to create temp dir: %v", err)})
		return
	}

	var localImagePaths []string
	for idx, fh := range files {
		src, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to open uploaded file: %v", err)})
			return
		}
		defer src.Close()

		basename := fmt.Sprintf("%03d_%s", idx, fh.Filename)
		localPath := filepath.Join(imageTmpDir, basename)
		out, err := os.Create(localPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to create temp image file: %v", err)})
			return
		}
		if _, err := io.Copy(out, src); err != nil {
			out.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to write temp image file: %v", err)})
			return
		}
		out.Close()
		localImagePaths = append(localImagePaths, localPath)
	}

	// Forward the original multipart body to N8N Reels webhook without rebuilding
	targetURL := vh.cfg.N8NREELSURL
	if targetURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "N8N Reels URL not configured"})
		return
	}

	req, err := http.NewRequest("POST", targetURL, bytes.NewReader(origBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to create upstream request: %v", err)})
		return
	}
	// Preserve the original Content-Type with boundary
	req.Header.Set("Content-Type", ct)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": fmt.Sprintf("upstream request failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		upstreamBody, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusBadGateway, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("upstream %s: %s", resp.Status, string(upstreamBody)),
		})
		return
	}

	// Read upstream JSON response body
	respBytes, readUpErr := io.ReadAll(resp.Body)
	if readUpErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": fmt.Sprintf("failed to read upstream response: %v", readUpErr)})
		return
	}

	// Compile with AI schema blob and local image paths
	args, _, outputPath, err := vh.ffmpegCompiler.Compile(respBytes, localImagePaths)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}
	defer os.Remove(outputPath)

	cmd := exec.Command("ffmpeg", args...)
	// Run ffmpeg and capture output for diagnostics
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("ffmpeg args: %v\n", args)
		fmt.Printf("ffmpeg error: %v\n", err)
		fmt.Printf("ffmpeg output: %s\n", string(output))
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"error":   fmt.Sprintf("ffmpeg failed: %v", err),
			"details": string(output),
		})
		return
	}

	// Ensure output file exists and is non-empty before serving
	if fi, statErr := os.Stat(outputPath); statErr != nil || fi.Size() == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("output file missing or empty: %v", statErr),
		})
		return
	}

	// After ffmpeg finishes and you have outputPath
	c.Header("Content-Type", "video/mp4")
	c.File(outputPath) // streams via http.ServeFile; supports Range (seek/scrub)

}

func (vh *VideoHandler) GenerateProReels(c *gin.Context) {
	// Enforce multipart/form-data only
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"status": "error", "error": "Content-Type must be multipart/form-data"})
		return
	}

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid multipart content-type"})
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "missing multipart boundary"})
		return
	}
	// Reader over incoming multipart body
	mr := multipart.NewReader(c.Request.Body, boundary)

	// Read prompt and the first image part into memory (SDK requires []byte). Limit image size to 25MB.
	var img *genai.Image
	prompt := ""
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read part error: %v", err)})
			return
		}
		if part.FileName() != "" && part.FormName() == "image" {
			lr := io.LimitReader(part, 25<<20)
			b, rerr := io.ReadAll(lr)
			part.Close()
			if rerr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read image error: %v", rerr)})
				return
			}
			mimeType := part.Header.Get("Content-Type")
			if mimeType == "" {
				mimeType = mime.TypeByExtension(filepath.Ext(part.FileName()))
				if mimeType == "" {
					mimeType = "image/jpeg"
				}
			}
			img = &genai.Image{ImageBytes: b, MIMEType: mimeType}
			// do not break; continue scanning parts to also capture prompt
		} else if part.FileName() == "" && part.FormName() == "prompt" {
			b, rerr := io.ReadAll(part)
			part.Close()
			if rerr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read prompt error: %v", rerr)})
				return
			}
			prompt = string(b)
		}
		part.Close()
	}

	if strings.TrimSpace(prompt) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "prompt is required"})
		return
	}

	ctx := context.Background()
	b64, mimeOut, fallbackBytes, fallbackMime, gerr := vh.veo.GenerateFirstVideoBase64(ctx, prompt, img)
	if gerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": gerr.Error()})
		return
	}

	if b64 != "" {
		if strings.TrimSpace(mimeOut) == "" {
			mimeOut = "video/mp4"
		}
		id := vh.veo.GenerateID()
		vh.veo.Put(id, &services.VeoItem{MIMEType: mimeOut, Base64: b64})
		c.JSON(http.StatusOK, gin.H{"ok": true, "videoUrl": fmt.Sprintf("/api/veo/video/%s", id)})
		return
	}

	// Fallback: we have bytes; write to temp file and stream via Range
	tmpFile, err := os.CreateTemp("", "veo3-*.mp4")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to create temp file: %v", err)})
		return
	}
	defer func(name string) {
		_ = tmpFile.Close()
		_ = os.Remove(name)
	}(tmpFile.Name())
	if _, err := tmpFile.Write(fallbackBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to write temp video: %v", err)})
		return
	}
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to rewind temp video: %v", err)})
		return
	}
	ctOut := fallbackMime
	if ctOut == "" {
		ctOut = "video/mp4"
	}
	c.Header("Content-Type", ctOut)
	c.File(tmpFile.Name())
	return
}

// this function is currently broken; fix later
func (vh *VideoHandler) GenerateVideoPexels(c *gin.Context) {
	// Enforce multipart/form-data only
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":  "Improper Data Format: Content-Type must be multipart/form-data",
			"status": "error",
		})
		return
	}

	prompt := c.PostForm("prompt")
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Prompt is required",
			"status": "error",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "invalid multipart form"})
		return
	}
	files := form.File["image"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "at least one image is required (field name: image)"})
		return
	}

	vr := models.VideoGenerationRequest{Prompt: prompt, Source: models.VideoSourcePexels}
	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to open uploaded file: %v", err)})
			return
		}
		b, err := io.ReadAll(src)
		src.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to read uploaded file: %v", err)})
			return
		}
		vr.Images = append(vr.Images, b)
		vr.ImageNames = append(vr.ImageNames, fh.Filename)
	}

	resp, svcErr := vh.contentGenerator.GenerateVideoSchemaMultipart(vr)
	if svcErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": svcErr.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (vh *VideoHandler) SSEStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.SSEvent("error", gin.H{"status": "error", "error": "SSE not implemented"})
	c.Writer.Flush()
	return
}
