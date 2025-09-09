package handlers

import (
	"bytes"
	"fmt"
	"io"

	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
)

// these are all tools the function can tap into
type VideoHandler struct {
	cfg              *config.APIConfig
	contentGenerator *services.ContentGenerator
	elevenLabs       *services.ElevenLabsService
	backgroundMusic  *services.BackgroundMusic
	ffmpegCompiler   *services.CompositionCompiler
	N8NService       *services.N8NService
	veo              *services.VeoService
}

func NewVideoHandler(cfg *config.APIConfig) *VideoHandler {
	return &VideoHandler{
		cfg:              cfg,
		contentGenerator: services.NewContentGenerator(cfg),
		elevenLabs:       services.NewElevenLabsService(cfg),
		backgroundMusic:  services.NewBackgroundMusic(cfg),
		ffmpegCompiler:   services.NewCompositionCompiler(services.NewFFmpegCommandBuilder(), services.NewBackgroundMusic(cfg), services.NewElevenLabsService(cfg)),
		N8NService:       services.NewN8NService(cfg),
		veo:              services.NewVeoService(cfg),
	}
}

/*
Some tool to parse a string into an int64(???)
*/
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

/*
GenerateVideoTimeline is function to generate a video timeline, so they can
somewhat visualize what the video will look like in the final composition

Wrapper around the N8N Service which directs the data to the correct
N8N endpoint
*/

func (vh *VideoHandler) GenerateVideoTimeline(c *gin.Context) {
	//GUARDS for multipart/form-data
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"status": "error", "error": "Content-Type must be multipart/form-data"})
		return
	}
	// GUARDS END

	targetURL := vh.cfg.N8BTIMELINEURL
	resp, err := vh.N8NService.Get(c, targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to get timeline: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": resp.Body})
	defer resp.Body.Close()

}

/*
this tool will be an api endpoint, and will be called to actually generate a video for the client
should take in a modified prompt, from the timeline schema
make a seprate api request, but pass modified data to it
*/

/*
This method has been completely FUCKED by Cursor AI; need to fix it
*/
func (vh *VideoHandler) GenerateProReels(c *gin.Context) {
	// Enforce multipart/form-data only
	//GUARDS
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
	// GUARDS END

	/*
		Everything below here is going to just be put in the veo function handler
		returns name and bytes, which we c.DataFromReader straight to the client as bytes
	*/
	mp4_name, video_bytes := vh.veo.GenerateVideoMultipart(c, boundary)

	//content disposition mapping
	ContentDisposition := map[string]string{"Content-Disposition": fmt.Sprintf("attachment; filename=\"%s\"", mp4_name)}

	c.DataFromReader(http.StatusOK, int64(len(video_bytes)), "video/mp4", bytes.NewReader(video_bytes), ContentDisposition)
}

/*
SSEStream to send small, event based updates to the client
*/
func (vh *VideoHandler) SSEStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.SSEvent("error", gin.H{"status": "error", "error": "SSE not implemented"})
	c.Writer.Flush()
	return
}

/*
FFMPEG implementation; not as good as the google veo for sure
*/
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

/*
Previous implementation
*/

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
