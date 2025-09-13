package handlers

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"net/http"
	"os"
	"path/filepath"
	"strings"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/models/interfaces"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
)

// VideoHandler uses compiler-first architecture
type VideoHandler struct {
	cfg           *config.APIConfig
	reelsCompiler services.VideoCompiler
	proCompiler   services.VideoCompiler
	// Services used directly by handlers (not through compilers)

	N8NService      *services.N8NService
	veo             interfaces.VideoGeneration
	voiceOver       interfaces.VoiceOver
	musicGeneration interfaces.MusicGeneration

	// SSE client management
	clients      map[string]chan gin.H
	clientsMutex sync.RWMutex
}

func NewVideoHandler(cfg *config.APIConfig, voiceOver interfaces.VoiceOver, musicGeneration interfaces.MusicGeneration, videoGeneration interfaces.VideoGeneration) *VideoHandler {

	return &VideoHandler{
		//api's
		cfg:             cfg,
		voiceOver:       voiceOver,
		musicGeneration: musicGeneration,

		// Compilers with their own builders and shared services
		reelsCompiler: services.NewReelsCompiler(musicGeneration, voiceOver),
		proCompiler:   services.NewProCompiler(musicGeneration, voiceOver),

		//Actually used to generate schema; should
		//eventually pass it n8n service as a method,
		//as it invokes n8n under the hood

		//n8n service to get the JSON, make network requests
		N8NService: services.NewN8NService(cfg),

		//veo to geenrate videos, fed into compilier
		veo: videoGeneration,

		// Initialize SSE client management
		clients: make(map[string]chan gin.H),
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

/*
Actully, there should really just be a universal timeline schema if I'm not mistaken?
*/

func (vh *VideoHandler) GenerateVideoTimeline(c *gin.Context) {
	//GUARDS for multipart/form-data
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"status": "error", "error": "Content-Type must be multipart/form-data"})
		return
	}
	// GUARDS END

	//sanity check
	//parsed from looked good to me
	//sanity check end

	targetURL := vh.cfg.N8BTIMELINEURL
	resp, err := vh.N8NService.Get(c, targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to get timeline: %v", err)})
		return
	}
	// Forward upstream status, headers, and JSON body directly to the client
	defer resp.Body.Close()
	for k, vals := range resp.Header {
		for _, v := range vals {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(resp.StatusCode)
	// Ensure JSON content type for the browser
	c.Writer.Header().Set("Content-Type", "application/json")
	// Stream body as-is
	if _, copyErr := io.Copy(c.Writer, resp.Body); copyErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to stream response: %v", copyErr)})
		return
	}

}

/*
Both of these are tools to generate the videos and return to the client
Should typically be invoked after the client has already generate a VideoTimelineSchema
*/

/*
GenerateProReels; Handler to invoke the pro compiler
*/
func (vh *VideoHandler) GenerateProReels(c *gin.Context) {
	// Enforce multipart/form-data only
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":  "Content-Type must be multipart/form-data",
			"status": "error",
		})
		return
	}

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

	imageTmpDir := filepath.Join(os.TempDir(), "pro_images")
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

	schemaValues := form.Value["schema"]
	if len(schemaValues) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "schema file is required"})
		return
	}
	schema := schemaValues[0]

	// Compile with AI schema blob and local image paths using PRO compiler
	// Compiler handles its own intermediate file cleanup and FFmpeg execution
	videoStream, err := vh.proCompiler.Compile([]byte(schema), localImagePaths, vh.cfg)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Clean up uploaded images after streaming
	defer func() {
		for _, file := range localImagePaths {
			os.Remove(file)
		}
	}()

	// Stream video directly to client
	c.Header("Content-Type", "video/mp4")
	c.Header("Cache-Control", "no-cache")

	if _, err := io.Copy(c.Writer, videoStream); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("failed to stream video: %v", err),
		})
		return
	}
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

	schemaValues := form.Value["schema"]
	if len(schemaValues) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "schema file is required"})
		return
	}
	schema := schemaValues[0]

	// Get client ID for SSE updates (optional)
	clientID := c.Query("client_id")

	// Send progress update if client is connected
	if clientID != "" {
		vh.SendToClient(clientID, "video_progress", gin.H{
			"stage":    "compiling",
			"message":  "Starting video compilation...",
			"progress": 10,
		})
	}

	// Compile with AI schema blob and local image paths using reels compiler
	// Compiler handles its own intermediate file cleanup and FFmpeg execution
	videoStream, err := vh.reelsCompiler.Compile([]byte(schema), localImagePaths, vh.cfg)
	if err != nil {
		// Send error update if client is connected
		if clientID != "" {
			vh.SendToClient(clientID, "video_error", gin.H{
				"error": err.Error(),
				"stage": "compilation_failed",
			})
		}
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": err.Error()})
		return
	}

	// Send progress update if client is connected
	if clientID != "" {
		vh.SendToClient(clientID, "video_progress", gin.H{
			"stage":    "streaming",
			"message":  "Video compilation complete, streaming to client...",
			"progress": 90,
		})
	}

	// Clean up uploaded images after streaming
	defer func() {
		for _, file := range localImagePaths {
			os.Remove(file)
		}
	}()

	// Stream video directly to client
	c.Header("Content-Type", "video/mp4")
	c.Header("Cache-Control", "no-cache")

	if _, err := io.Copy(c.Writer, videoStream); err != nil {
		// Send error update if client is connected
		if clientID != "" {
			vh.SendToClient(clientID, "video_error", gin.H{
				"error": fmt.Sprintf("failed to stream video: %v", err),
				"stage": "streaming_failed",
			})
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"error":  fmt.Sprintf("failed to stream video: %v", err),
		})
		return
	}

	// Send completion update if client is connected
	if clientID != "" {
		vh.SendToClient(clientID, "video_progress", gin.H{
			"stage":    "completed",
			"message":  "Video generation completed successfully!",
			"progress": 100,
		})
	}

}

// AddClient adds a new SSE client with the provided client ID
func (vh *VideoHandler) AddClient(clientID string) chan gin.H {
	clientChan := make(chan gin.H, 10) // Buffered channel for non-blocking sends

	vh.clientsMutex.Lock()
	vh.clients[clientID] = clientChan
	vh.clientsMutex.Unlock()

	return clientChan
}

// RemoveClient removes a client from the SSE client list
func (vh *VideoHandler) RemoveClient(clientID string) {
	vh.clientsMutex.Lock()
	if clientChan, exists := vh.clients[clientID]; exists {
		close(clientChan)
		delete(vh.clients, clientID)
	}
	vh.clientsMutex.Unlock()
}

// SendToClient sends a message to a specific client
func (vh *VideoHandler) SendToClient(clientID string, eventType string, data gin.H) {
	vh.clientsMutex.RLock()
	clientChan, exists := vh.clients[clientID]
	vh.clientsMutex.RUnlock()

	if exists {
		select {
		case clientChan <- gin.H{"type": eventType, "data": data}:
			// Message sent successfully
		default:
			// Channel is full, client might be slow or disconnected
			// Could log this or implement backpressure handling
		}
	}
}

/*
SSEStream to send small, event based updates to the client
Uses client ID from frontend request for client-specific messaging
*/
func (vh *VideoHandler) SSEVideoRequest(c *gin.Context) {
	// Get client ID from frontend request
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(400, gin.H{"error": "client_id query parameter is required"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Add this client to our client list
	clientChan := vh.AddClient(clientID)
	defer vh.RemoveClient(clientID)

	// Send initial connection event with client ID
	c.SSEvent("connected", gin.H{
		"status":    "connected",
		"client_id": clientID,
		"timestamp": time.Now().Unix(),
	})
	c.Writer.Flush()

	// Keep connection alive and send client-specific updates
	ticker := time.NewTicker(30 * time.Second) // Less frequent heartbeat
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			// Client disconnected
			return
		case <-ticker.C:
			// Send heartbeat
			c.SSEvent("heartbeat", gin.H{
				"timestamp": time.Now().Unix(),
				"client_id": clientID,
			})
			c.Writer.Flush()
		case message := <-clientChan:
			// Send client-specific message
			eventType := message["type"].(string)
			data := message["data"].(gin.H)
			c.SSEvent(eventType, data)
			c.Writer.Flush()
		}
	}
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
	/*
	   Pexels here needs to just upload it to the docker container for the ai-shorts-maker; nothing else
	*/
	//c.JSON(http.StatusOK, resp)
}
