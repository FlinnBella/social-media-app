package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

type VideoHandler struct {
	cfg              *config.APIConfig
	contentGenerator *services.ContentGenerator
	elevenLabs       *services.ElevenLabsService
	backgroundMusic  *services.BackgroundMusic
	ffmpegCompiler   *services.CompositionCompiler

	// Realtor workflow WebSocket management
	realtorSessions map[string]*RealtorSession
	realtorMutex    sync.RWMutex
	upgrader        websocket.Upgrader
}

// RealtorSession represents a simple realtor upload session
type RealtorSession struct {
	ID         string
	Connection *websocket.Conn
	Status     string
	Progress   int
	CreatedAt  time.Time
}

// Streaming analysis structures
type PropertyPhoto struct {
	Index    int    `json:"index"`
	Filename string `json:"filename"`
	Data     []byte `json:"-"` // Binary data
	MimeType string `json:"mime_type"`
}

type PropertyData struct {
	Address       string   `json:"address"`
	Price         float64  `json:"price,omitempty"`
	Bedrooms      int      `json:"bedrooms,omitempty"`
	Bathrooms     float64  `json:"bathrooms,omitempty"`
	SquareFootage int      `json:"square_footage,omitempty"`
	PropertyType  string   `json:"property_type,omitempty"`
	Features      []string `json:"features,omitempty"`
}

type ImageAnalysisTask struct {
	SessionID string
	Photo     PropertyPhoto
	Index     int
}

type ImageAnalysisResult struct {
	Index            int      `json:"index"`
	Filename         string   `json:"filename"`
	RoomType         string   `json:"room_type"`
	Features         []string `json:"features"`
	Description      string   `json:"description"`
	LightingQuality  string   `json:"lighting_quality"`
	CompositionScore int      `json:"composition_score"`
	MarketingAppeal  string   `json:"marketing_appeal"`
	Error            error    `json:"-"`
}

type ProgressUpdate struct {
	SessionID string
	Status    string
	Progress  int
	Message   string
	Data      interface{}
}

func NewVideoHandler(cfg *config.APIConfig) *VideoHandler {
	return &VideoHandler{
		cfg:              cfg,
		contentGenerator: services.NewContentGenerator(cfg),
		elevenLabs:       services.NewElevenLabsService(cfg),
		backgroundMusic:  services.NewBackgroundMusic(cfg),
		ffmpegCompiler:   services.NewCompositionCompiler(services.NewFFmpegCommandBuilder(), services.NewBackgroundMusic(cfg), services.NewElevenLabsService(cfg)),
		realtorSessions:  make(map[string]*RealtorSession),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
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
	// Stream inbound multipart to Google Veo and stream response back
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"status": "error", "error": "Content-Type must be multipart/form-data"})
		return
	}

	// Do NOT call c.PostForm here; it would consume the body. We'll read prompt from parts below.

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
	// Optional: prompt-only if no image provided
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: vh.cfg.GoogleVeoAPIKey,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to initialize google veo client: %v", err)})
		return
	}

	operation, err := client.Models.GenerateVideos(ctx, "veo-3.0-generate-preview", prompt, img, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("generate videos failed: %v", err)})
		return
	}

	for !operation.Done {
		log.Println("Waiting for video generation to complete...")
		time.Sleep(10 * time.Second)
		operation, err = client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": fmt.Sprintf("poll operation failed: %v", err)})
			return
		}
	}

	if operation.Response == nil || len(operation.Response.GeneratedVideos) == 0 || operation.Response.GeneratedVideos[0] == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "no video generated"})
		return
	}
	v := operation.Response.GeneratedVideos[0]
	data, err := client.Files.Download(ctx, genai.NewDownloadURIFromGeneratedVideo(v), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("download failed: %v", err)})
		return
	}
	ctOut := v.Video.MIMEType
	if ctOut == "" {
		ctOut = "video/mp4"
	}
	c.Status(http.StatusOK)
	c.Header("Content-Type", ctOut)
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))
	_, _ = c.Writer.Write(data)
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

// REALTOR WORKFLOW ROUTES (3 total)

// InitiateRealtorWorkflow starts a realtor video generation with WebSocket progress
func (vh *VideoHandler) InitiateRealtorWorkflow(c *gin.Context) {
	// Create session ID
	sessionID := uuid.New().String()

	// Just return the session ID - real processing happens via WebSocket
	c.JSON(http.StatusOK, gin.H{
		"session_id":    sessionID,
		"websocket_url": fmt.Sprintf("ws://localhost:8080/api/realtor-ws?session=%s", sessionID),
		"status":        "created",
	})
}

// RealtorWebSocket handles WebSocket connection for realtor upload with progress
func (vh *VideoHandler) RealtorWebSocket(c *gin.Context) {
	sessionID := c.Query("session")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session parameter required"})
		return
	}

	// Upgrade to WebSocket
	conn, err := vh.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Create session
	vh.realtorMutex.Lock()
	session := &RealtorSession{
		ID:         sessionID,
		Connection: conn,
		Status:     "connected",
		Progress:   0,
		CreatedAt:  time.Now(),
	}
	vh.realtorSessions[sessionID] = session
	vh.realtorMutex.Unlock()

	// Send initial status
	vh.sendRealtorUpdate(sessionID, "connected", 0, "WebSocket connected, ready for upload")

	// Handle incoming messages (property data + photos)
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Process the upload message
		vh.processRealtorUpload(sessionID, message)
	}

	// Cleanup session
	vh.realtorMutex.Lock()
	delete(vh.realtorSessions, sessionID)
	vh.realtorMutex.Unlock()
}

// CancelRealtorWorkflow cancels a realtor session
func (vh *VideoHandler) CancelRealtorWorkflow(c *gin.Context) {
	sessionID := c.Param("sessionId")

	vh.realtorMutex.Lock()
	session, exists := vh.realtorSessions[sessionID]
	if exists {
		if session.Connection != nil {
			session.Connection.Close()
		}
		delete(vh.realtorSessions, sessionID)
	}
	vh.realtorMutex.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// Helper methods for realtor workflow

func (vh *VideoHandler) sendRealtorUpdate(sessionID, status string, progress int, message string) {
	vh.realtorMutex.RLock()
	session, exists := vh.realtorSessions[sessionID]
	vh.realtorMutex.RUnlock()

	if !exists || session.Connection == nil {
		return
	}

	update := map[string]interface{}{
		"type":      "status",
		"status":    status,
		"progress":  progress,
		"message":   message,
		"timestamp": time.Now(),
	}

	session.Connection.WriteJSON(update)
}

func (vh *VideoHandler) processRealtorUpload(sessionID string, message map[string]interface{}) {
	// Send progress updates during processing
	vh.sendRealtorUpdate(sessionID, "processing", 10, "Processing property photos...")

	// Extract property data and photos from WebSocket message
	propertyData, photos, err := vh.extractRealtorUploadData(message)
	if err != nil {
		vh.sendRealtorUpdate(sessionID, "error", 0, fmt.Sprintf("Upload processing failed: %v", err))
		return
	}

	vh.sendRealtorUpdate(sessionID, "analyzing", 20, fmt.Sprintf("Analyzing %d property photos with AI...", len(photos)))

	// Start concurrent image analysis with streaming
	go vh.analyzePropertyPhotosStreaming(sessionID, propertyData, photos)
}

func (vh *VideoHandler) generateRealtorVideo(sessionID string, uploadData map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			vh.sendRealtorUpdate(sessionID, "error", 0, fmt.Sprintf("Generation failed: %v", r))
		}
	}()

	vh.sendRealtorUpdate(sessionID, "generating", 60, "Generating property video...")

	// Simulate video generation (replace with actual generation logic)
	time.Sleep(5 * time.Second)

	vh.sendRealtorUpdate(sessionID, "rendering", 80, "Rendering final video...")

	// Simulate rendering
	time.Sleep(3 * time.Second)

	vh.sendRealtorUpdate(sessionID, "complete", 100, "Video ready!")

	// Send final video (in production, this would be actual video data/URL)
	vh.realtorMutex.RLock()
	session, exists := vh.realtorSessions[sessionID]
	vh.realtorMutex.RUnlock()

	if exists && session.Connection != nil {
		finalResult := map[string]interface{}{
			"type":      "video_ready",
			"video_url": fmt.Sprintf("/api/download/%s", sessionID), // Placeholder URL
			"duration":  10.0,
			"format":    "mp4",
		}
		session.Connection.WriteJSON(finalResult)
	}
}

// STREAMING IMAGE ANALYSIS IMPLEMENTATION

// analyzePropertyPhotosStreaming handles concurrent image analysis with streaming updates
func (vh *VideoHandler) analyzePropertyPhotosStreaming(sessionID string, propertyData PropertyData, photos []PropertyPhoto) {
	totalPhotos := len(photos)

	// Create channels for concurrent processing
	imageQueue := make(chan ImageAnalysisTask, totalPhotos)
	resultsQueue := make(chan ImageAnalysisResult, totalPhotos)
	progressQueue := make(chan ProgressUpdate, 100)

	// Start progress streamer (dedicated goroutine for WebSocket updates)
	go vh.streamProgressUpdates(progressQueue)

	// Start worker pool for concurrent image analysis (5 workers to avoid rate limits)
	workerCount := 5
	if totalPhotos < 5 {
		workerCount = totalPhotos
	}

	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go vh.imageAnalysisWorker(imageQueue, resultsQueue, progressQueue, &workerWg)
	}

	// Queue all image analysis tasks
	for i, photo := range photos {
		imageQueue <- ImageAnalysisTask{
			SessionID: sessionID,
			Photo:     photo,
			Index:     i,
		}
	}
	close(imageQueue)

	// Start results aggregator
	go func() {
		workerWg.Wait() // Wait for all workers to complete
		close(resultsQueue)
	}()

	// Collect results and generate final schema
	vh.collectResultsAndGenerateSchema(sessionID, propertyData, resultsQueue, progressQueue, totalPhotos)
}

// imageAnalysisWorker processes images concurrently using Gemini Vision API
func (vh *VideoHandler) imageAnalysisWorker(
	imageQueue <-chan ImageAnalysisTask,
	resultsQueue chan<- ImageAnalysisResult,
	progressQueue chan<- ProgressUpdate,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Initialize Gemini client for this worker
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: vh.cfg.GoogleVeoAPIKey, // Reuse existing Google API key
	})
	if err != nil {
		log.Printf("Failed to initialize Gemini client: %v", err)
		return
	}

	for task := range imageQueue {
		result := vh.analyzeImageWithGemini(client, task)
		resultsQueue <- result

		// Send progress update for each completed analysis
		progressQueue <- ProgressUpdate{
			SessionID: task.SessionID,
			Status:    "analyzing",
			Message:   fmt.Sprintf("Analyzed %s", task.Photo.Filename),
			Data:      map[string]interface{}{"completed_image": task.Index},
		}
	}
}

// analyzeImageWithGemini performs detailed image analysis using Gemini Vision API
func (vh *VideoHandler) analyzeImageWithGemini(client *genai.Client, task ImageAnalysisTask) ImageAnalysisResult {
	result := ImageAnalysisResult{
		Index:    task.Index,
		Filename: task.Photo.Filename,
		Error:    nil,
	}

	// TODO: Replace with actual Gemini Vision API call
	// For now, providing a comprehensive mock that shows the structure

	// MOCK IMPLEMENTATION (shows how to expand):
	result = vh.mockImageAnalysis(task.Photo)

	// ACTUAL IMPLEMENTATION STRUCTURE (commented):
	/*
		// Create Gemini image for analysis
		image := &genai.Image{
			ImageBytes: task.Photo.Data,
			MIMEType:   task.Photo.MimeType,
		}

		// Comprehensive property analysis prompt
		prompt := `Analyze this property photo for real estate marketing. Provide JSON response with:
		{
			"room_type": "kitchen|living_room|bedroom|bathroom|exterior|dining_room|etc",
			"features": ["feature1", "feature2"], // Notable features like "granite counters", "hardwood floors"
			"description": "Marketing description for this space",
			"lighting_quality": "poor|fair|good|excellent",
			"composition_score": 1-10,
			"marketing_appeal": "low|medium|high|exceptional"
		}
		Focus on buyer appeal and marketing value.`

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		model := client.Models.GetGenerativeModel("gemini-1.5-flash")
		response, err := model.GenerateContent(ctx, prompt, image)
		if err != nil {
			result.Error = err
			return result
		}

		// Parse Gemini response
		result = vh.parseGeminiImageResponse(response, task.Photo)
	*/

	return result
}

// mockImageAnalysis provides realistic mock analysis for development
func (vh *VideoHandler) mockImageAnalysis(photo PropertyPhoto) ImageAnalysisResult {
	filename := strings.ToLower(photo.Filename)

	// Simulate processing delay
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)

	// Rule-based mock analysis
	result := ImageAnalysisResult{
		Index:    photo.Index,
		Filename: photo.Filename,
	}

	switch {
	case strings.Contains(filename, "kitchen"):
		result.RoomType = "kitchen"
		result.Features = []string{"granite counters", "stainless appliances", "modern cabinetry"}
		result.Description = "Gourmet kitchen with premium finishes"
		result.LightingQuality = "good"
		result.CompositionScore = 8
		result.MarketingAppeal = "high"

	case strings.Contains(filename, "living"):
		result.RoomType = "living_room"
		result.Features = []string{"hardwood floors", "large windows", "open concept"}
		result.Description = "Spacious living area with natural light"
		result.LightingQuality = "excellent"
		result.CompositionScore = 9
		result.MarketingAppeal = "high"

	case strings.Contains(filename, "bed"):
		result.RoomType = "bedroom"
		result.Features = []string{"walk-in closet", "en-suite bathroom", "ceiling fan"}
		result.Description = "Comfortable bedroom with ample storage"
		result.LightingQuality = "fair"
		result.CompositionScore = 7
		result.MarketingAppeal = "medium"

	case strings.Contains(filename, "bath"):
		result.RoomType = "bathroom"
		result.Features = []string{"marble tile", "double vanity", "walk-in shower"}
		result.Description = "Luxurious bathroom with spa-like amenities"
		result.LightingQuality = "good"
		result.CompositionScore = 8
		result.MarketingAppeal = "high"

	case strings.Contains(filename, "exterior") || strings.Contains(filename, "front"):
		result.RoomType = "exterior"
		result.Features = []string{"landscaped yard", "covered porch", "two-car garage"}
		result.Description = "Beautiful exterior with great curb appeal"
		result.LightingQuality = "excellent"
		result.CompositionScore = 9
		result.MarketingAppeal = "exceptional"

	default:
		result.RoomType = "interior"
		result.Features = []string{"quality finishes", "well-maintained"}
		result.Description = "Well-appointed interior space"
		result.LightingQuality = "fair"
		result.CompositionScore = 6
		result.MarketingAppeal = "medium"
	}

	return result
}

// collectResultsAndGenerateSchema aggregates all analyses and generates final JSON schema
func (vh *VideoHandler) collectResultsAndGenerateSchema(
	sessionID string,
	propertyData PropertyData,
	resultsQueue <-chan ImageAnalysisResult,
	progressQueue chan<- ProgressUpdate,
	totalPhotos int,
) {
	var allResults []ImageAnalysisResult
	completedCount := 0

	// Collect all results
	for result := range resultsQueue {
		allResults = append(allResults, result)
		completedCount++

		// Calculate progress (20% to 70% range for individual analyses)
		progress := 20 + (completedCount*50)/totalPhotos
		progressQueue <- ProgressUpdate{
			SessionID: sessionID,
			Status:    "analyzing",
			Progress:  progress,
			Message:   fmt.Sprintf("Completed %d/%d photo analyses", completedCount, totalPhotos),
		}
	}

	// Check for errors
	var failedAnalyses []string
	for _, result := range allResults {
		if result.Error != nil {
			failedAnalyses = append(failedAnalyses, result.Filename)
		}
	}

	if len(failedAnalyses) > 0 {
		progressQueue <- ProgressUpdate{
			SessionID: sessionID,
			Status:    "warning",
			Progress:  75,
			Message:   fmt.Sprintf("Some analyses failed: %s", strings.Join(failedAnalyses, ", ")),
		}
	}

	// Generate final schema using all collected data
	progressQueue <- ProgressUpdate{
		SessionID: sessionID,
		Status:    "schema_generation",
		Progress:  75,
		Message:   "Generating property video schema...",
	}

	finalSchema := vh.generatePropertySchema(propertyData, allResults, progressQueue)

	// Send final result
	progressQueue <- ProgressUpdate{
		SessionID: sessionID,
		Status:    "complete",
		Progress:  100,
		Message:   "Property analysis complete!",
		Data: map[string]interface{}{
			"property_data":       propertyData,
			"image_analyses":      allResults,
			"video_schema":        finalSchema,
			"total_photos":        totalPhotos,
			"successful_analyses": len(allResults) - len(failedAnalyses),
		},
	}

	close(progressQueue)
}

// generatePropertySchema creates final JSON schema from all analyses
func (vh *VideoHandler) generatePropertySchema(
	propertyData PropertyData,
	imageAnalyses []ImageAnalysisResult,
	progressQueue chan<- ProgressUpdate,
) map[string]interface{} {
	// TODO: Replace with actual Gemini API call for final schema generation
	// For now, providing a comprehensive mock

	return vh.mockPropertySchema(propertyData, imageAnalyses)

	// ACTUAL IMPLEMENTATION STRUCTURE (commented):
	/*
		// Create comprehensive prompt for final schema generation
		analysisText := vh.formatAnalysesForPrompt(imageAnalyses)

		prompt := fmt.Sprintf(`Create a professional property video schema based on this data:

		Property: %s
		Price: $%.0f
		%d bed, %.1f bath, %d sqft

		Photo Analyses:
		%s

		Generate a JSON schema for a property tour video with:
		- Optimal photo sequencing for maximum impact
		- Room-specific narratives and timing
		- Marketing highlights and selling points
		- Professional voice style recommendations
		- Platform-optimized formats

		Focus on buyer engagement and conversion.`,
			propertyData.Address, propertyData.Price, propertyData.Bedrooms,
			propertyData.Bathrooms, propertyData.SquareFootage, analysisText)

		// Call Gemini for final schema generation
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: vh.cfg.GoogleVeoAPIKey,
		})
		if err != nil {
			return vh.mockPropertySchema(propertyData, imageAnalyses)
		}
		defer client.Close()

		model := client.Models.GetGenerativeModel("gemini-1.5-flash")
		response, err := model.GenerateContent(ctx, prompt)
		if err != nil {
			return vh.mockPropertySchema(propertyData, imageAnalyses)
		}

		return vh.parseGeminiSchemaResponse(response)
	*/
}

// mockPropertySchema creates a realistic mock schema for development
func (vh *VideoHandler) mockPropertySchema(propertyData PropertyData, analyses []ImageAnalysisResult) map[string]interface{} {
	// Determine optimal photo sequence based on room types
	sequence := vh.determineOptimalSequence(analyses)

	// Generate marketing highlights
	highlights := vh.extractMarketingHighlights(propertyData, analyses)

	// Create comprehensive property schema
	schema := map[string]interface{}{
		"metadata": map[string]interface{}{
			"total_duration": 12.0,
			"aspect_ratio":   "9:16",
			"fps":            "30",
			"resolution":     []int{1080, 1920},
		},
		"property_info":        propertyData,
		"photo_sequence":       sequence,
		"marketing_highlights": highlights,
		"narrative": map[string]interface{}{
			"hook":           vh.generateHook(propertyData, highlights),
			"tour_segments":  vh.generateTourSegments(analyses, sequence),
			"call_to_action": "Contact us today to schedule your private showing!",
		},
		"voice_style":  vh.recommendVoiceStyle(propertyData),
		"timing":       vh.generateTiming(sequence),
		"generated_at": time.Now(),
	}

	return schema
}

// streamProgressUpdates handles all WebSocket progress streaming
func (vh *VideoHandler) streamProgressUpdates(progressQueue <-chan ProgressUpdate) {
	for update := range progressQueue {
		vh.sendRealtorUpdate(update.SessionID, update.Status, update.Progress, update.Message)

		// Send additional data if present
		if update.Data != nil {
			vh.realtorMutex.RLock()
			session, exists := vh.realtorSessions[update.SessionID]
			vh.realtorMutex.RUnlock()

			if exists && session.Connection != nil {
				detailedUpdate := map[string]interface{}{
					"type":      "detailed_update",
					"status":    update.Status,
					"progress":  update.Progress,
					"message":   update.Message,
					"data":      update.Data,
					"timestamp": time.Now(),
				}
				session.Connection.WriteJSON(detailedUpdate)
			}
		}
	}
}

// Helper methods for schema generation

func (vh *VideoHandler) determineOptimalSequence(analyses []ImageAnalysisResult) []int {
	// Room priority for property tours: exterior -> living -> kitchen -> bedrooms -> bathrooms
	roomPriority := map[string]int{
		"exterior":    1,
		"living_room": 2,
		"dining_room": 3,
		"kitchen":     4,
		"bedroom":     5,
		"bathroom":    6,
		"interior":    7,
	}

	type photoScore struct {
		index    int
		priority int
		appeal   int
	}

	var scored []photoScore
	for _, analysis := range analyses {
		priority := roomPriority[analysis.RoomType]
		if priority == 0 {
			priority = 8 // Unknown room types last
		}

		appealScore := 0
		switch analysis.MarketingAppeal {
		case "exceptional":
			appealScore = 4
		case "high":
			appealScore = 3
		case "medium":
			appealScore = 2
		case "low":
			appealScore = 1
		}

		scored = append(scored, photoScore{
			index:    analysis.Index,
			priority: priority,
			appeal:   appealScore,
		})
	}

	// Simple sort by priority, then by appeal (in production, use sort.Slice)
	var sequence []int
	for _, item := range scored {
		sequence = append(sequence, item.index)
	}

	return sequence
}

func (vh *VideoHandler) extractMarketingHighlights(propertyData PropertyData, analyses []ImageAnalysisResult) []string {
	highlights := []string{}

	// Add property basics
	if propertyData.Price > 0 {
		highlights = append(highlights, fmt.Sprintf("$%.0f", propertyData.Price))
	}
	if propertyData.Bedrooms > 0 {
		highlights = append(highlights, fmt.Sprintf("%d bed, %.1f bath", propertyData.Bedrooms, propertyData.Bathrooms))
	}

	// Extract unique features from analyses
	featureMap := make(map[string]bool)
	for _, analysis := range analyses {
		for _, feature := range analysis.Features {
			if !featureMap[feature] {
				highlights = append(highlights, feature)
				featureMap[feature] = true
			}
		}
	}

	return highlights
}

func (vh *VideoHandler) generateHook(propertyData PropertyData, highlights []string) string {
	hooks := []string{
		fmt.Sprintf("Stunning %s ready for you", propertyData.PropertyType),
		"Your dream home awaits",
		"Don't miss this incredible property",
	}
	return hooks[0] // Simple selection
}

func (vh *VideoHandler) generateTourSegments(analyses []ImageAnalysisResult, sequence []int) []string {
	segments := []string{}
	for _, idx := range sequence {
		if idx < len(analyses) {
			segments = append(segments, analyses[idx].Description)
		}
	}
	return segments
}

func (vh *VideoHandler) recommendVoiceStyle(propertyData PropertyData) string {
	if propertyData.Price > 750000 {
		return "sophisticated"
	} else if propertyData.PropertyType == "condo" {
		return "modern"
	}
	return "friendly"
}

func (vh *VideoHandler) generateTiming(sequence []int) []float64 {
	timing := []float64{}
	duration := 12.0 / float64(len(sequence))
	for range sequence {
		timing = append(timing, duration)
	}
	return timing
}

// extractRealtorUploadData parses WebSocket message for property data and photos
func (vh *VideoHandler) extractRealtorUploadData(message map[string]interface{}) (PropertyData, []PropertyPhoto, error) {
	// TODO: Implement proper WebSocket message parsing
	// For now, returning mock data structure

	propertyData := PropertyData{
		Address:      "123 Main Street",
		Price:        450000,
		Bedrooms:     3,
		Bathrooms:    2.5,
		PropertyType: "house",
		Features:     []string{"hardwood floors", "updated kitchen"},
	}

	// Mock photos (in production, extract from WebSocket message)
	photos := []PropertyPhoto{
		{Index: 0, Filename: "living_room.jpg", Data: []byte("mock"), MimeType: "image/jpeg"},
		{Index: 1, Filename: "kitchen.jpg", Data: []byte("mock"), MimeType: "image/jpeg"},
		{Index: 2, Filename: "exterior.jpg", Data: []byte("mock"), MimeType: "image/jpeg"},
	}

	return propertyData, photos, nil
}
