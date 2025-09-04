package handlers

import (
	"bytes"
	"fmt"
	"strings"

	"io"
	"net/http"

	"os"
	"path/filepath"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	cfg              *config.APIConfig
	contentGenerator *services.ContentGenerator
	elevenLabs       *services.ElevenLabsService
	backgroundMusic  *services.BackgroundMusic
	ffmpegCompiler   *services.CompositionCompiler
}

func NewVideoHandler(cfg *config.APIConfig) *VideoHandler {
	return &VideoHandler{
		cfg:              cfg,
		contentGenerator: services.NewContentGenerator(cfg),
		elevenLabs:       services.NewElevenLabsService(cfg),
		backgroundMusic:  services.NewBackgroundMusic(cfg),
		ffmpegCompiler:   services.NewCompositionCompiler(services.NewFFmpegCommandBuilder(), services.NewBackgroundMusic(cfg), services.NewElevenLabsService(cfg)),
	}
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
	args, paths, outputPath, err := vh.ffmpegCompiler.Compile(respBytes, localImagePaths)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"args": args, "paths": paths, "outputPath": outputPath})
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
