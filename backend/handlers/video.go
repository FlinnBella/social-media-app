package handlers

import (
	"fmt"
	"net/http"
	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	contentGenerator *services.ContentGenerator
	videoProcessor   *services.VideoProcessor
	elevenLabs       *services.ElevenLabsService
}

func NewVideoHandler(cfg *config.APIConfig) *VideoHandler {
	return &VideoHandler{
		contentGenerator: services.NewContentGenerator(cfg),
		videoProcessor:   services.NewVideoProcessor(),
		elevenLabs:       services.NewElevenLabsService(cfg),
	}
}

func (vh *VideoHandler) GenerateVideo(c *gin.Context) {
	var req models.VideoGenerationRequest
	
	// Handle both JSON and form data
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.VideoGenerationResponse{
				Error:  "Invalid request format",
				Status: "error",
			})
			return
		}
	} else {
		// Handle form data
		req.Prompt = c.PostForm("prompt")
		if req.Prompt == "" {
			c.JSON(http.StatusBadRequest, models.VideoGenerationResponse{
				Error:  "Prompt is required",
				Status: "error",
			})
			return
		}
		
		// Handle file upload if present
		file, fileHeader, err := c.Request.FormFile("file")
		if err == nil && file != nil {
			defer file.Close()
			// For demo purposes, we'll just log the file info
			// In production, you'd process the uploaded file as reference
			c.Header("X-Uploaded-File", fileHeader.Filename)
		}
	}

	// New flow: trigger short video generation + download
	filename, err := vh.contentGenerator.GenerateShortVideo(req.Prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.VideoGenerationResponse{
			Error:  fmt.Sprintf("Failed to generate short video: %v", err),
			Status: "error",
		})
		return
	}

	// Build absolute URL to backend static file
	proto := c.Request.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	host := c.Request.Host
	videoURL := fmt.Sprintf("%s://%s/static/%s", proto, host, filename)

	c.JSON(http.StatusOK, models.VideoGenerationResponse{
		VideoURL: videoURL,
		Status:   "success",
	})
}

// GetComposition returns the video composition structure for debugging
func (vh *VideoHandler) GetComposition(c *gin.Context) {
	prompt := c.Query("prompt")
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt parameter is required"})
		return
	}

	composition, err := vh.contentGenerator.GenerateVideoComposition(prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, composition)
}