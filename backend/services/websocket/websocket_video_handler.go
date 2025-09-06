package websocket

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"social-media-ai-video/config"
	"social-media-ai-video/services"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

// WebSocketVideoHandler handles all WebSocket-based video generation operations
type WebSocketVideoHandler struct {
	cfg              *config.APIConfig
	contentGenerator *services.ContentGenerator
	elevenLabs       *services.ElevenLabsService
	backgroundMusic  *services.BackgroundMusic
	ffmpegCompiler   *services.CompositionCompiler
	
	// Session management
	realtorSessions map[string]*RealtorSession
	realtorMutex    sync.RWMutex
	upgrader        websocket.Upgrader
}

// RealtorSession represents a realtor upload session
type RealtorSession struct {
	ID         string
	Connection *websocket.Conn
	Status     string
	Progress   int
	CreatedAt  time.Time
}

// PropertyPhoto represents an uploaded property image
type PropertyPhoto struct {
	Index    int    `json:"index"`
	Filename string `json:"filename"`
	Data     []byte `json:"-"`
	MimeType string `json:"mime_type"`
}

// PropertyData represents property information
type PropertyData struct {
	Address       string   `json:"address"`
	Price         float64  `json:"price,omitempty"`
	Bedrooms      int      `json:"bedrooms,omitempty"`
	Bathrooms     float64  `json:"bathrooms,omitempty"`
	SquareFootage int      `json:"square_footage,omitempty"`
	PropertyType  string   `json:"property_type,omitempty"`
	Features      []string `json:"features,omitempty"`
}

// ImageAnalysisTask represents a task for concurrent image processing
type ImageAnalysisTask struct {
	SessionID string
	Photo     PropertyPhoto
	Index     int
}

// ImageAnalysisResult represents the result of image analysis
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

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	SessionID string
	Status    string
	Progress  int
	Message   string
	Data      interface{}
}

// NewWebSocketVideoHandler creates a new WebSocket video handler
func NewWebSocketVideoHandler(cfg *config.APIConfig) *WebSocketVideoHandler {
	return &WebSocketVideoHandler{
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

// InitiateSession creates a new realtor session
func (ws *WebSocketVideoHandler) InitiateSession() (string, string) {
	sessionID := uuid.New().String()
	websocketURL := fmt.Sprintf("ws://localhost:8080/api/realtor-ws?session=%s", sessionID)
	return sessionID, websocketURL
}

// HandleWebSocketConnection upgrades HTTP connection to WebSocket and manages session
func (ws *WebSocketVideoHandler) HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, sessionID string) error {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("WebSocket upgrade failed: %v", err)
	}
	defer conn.Close()

	// Create and register session
	ws.realtorMutex.Lock()
	session := &RealtorSession{
		ID:         sessionID,
		Connection: conn,
		Status:     "connected",
		Progress:   0,
		CreatedAt:  time.Now(),
	}
	ws.realtorSessions[sessionID] = session
	ws.realtorMutex.Unlock()

	// Send initial status
	ws.sendUpdate(sessionID, "connected", 0, "WebSocket connected, ready for upload")

	// Handle incoming messages
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Process the upload message
		ws.processRealtorUpload(sessionID, message)
	}

	// Cleanup session
	ws.cleanupSession(sessionID)
	return nil
}

// CancelSession cancels a realtor session
func (ws *WebSocketVideoHandler) CancelSession(sessionID string) error {
	ws.realtorMutex.Lock()
	session, exists := ws.realtorSessions[sessionID]
	if exists {
		if session.Connection != nil {
			session.Connection.Close()
		}
		delete(ws.realtorSessions, sessionID)
	}
	ws.realtorMutex.Unlock()
	
	if !exists {
		return fmt.Errorf("session not found")
	}
	
	return nil
}

// processRealtorUpload processes incoming property upload data
func (ws *WebSocketVideoHandler) processRealtorUpload(sessionID string, message map[string]interface{}) {
	ws.sendUpdate(sessionID, "processing", 10, "Processing property photos...")
	
	propertyData, photos, err := ws.extractUploadData(message)
	if err != nil {
		ws.sendUpdate(sessionID, "error", 0, fmt.Sprintf("Upload processing failed: %v", err))
		return
	}

	ws.sendUpdate(sessionID, "analyzing", 20, fmt.Sprintf("Analyzing %d property photos with AI...", len(photos)))
	
	// Start concurrent image analysis with streaming
	go ws.analyzePhotosStreaming(sessionID, propertyData, photos)
}

// analyzePhotosStreaming handles concurrent image analysis with streaming updates
func (ws *WebSocketVideoHandler) analyzePhotosStreaming(sessionID string, propertyData PropertyData, photos []PropertyPhoto) {
	totalPhotos := len(photos)
	
	// Create channels for concurrent processing
	imageQueue := make(chan ImageAnalysisTask, totalPhotos)
	resultsQueue := make(chan ImageAnalysisResult, totalPhotos)
	progressQueue := make(chan ProgressUpdate, 100)
	
	// Start progress streamer
	go ws.streamProgressUpdates(progressQueue)
	
	// Start worker pool for concurrent image analysis
	workerCount := 5
	if totalPhotos < 5 {
		workerCount = totalPhotos
	}
	
	var workerWg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go ws.imageAnalysisWorker(imageQueue, resultsQueue, progressQueue, &workerWg)
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
		workerWg.Wait()
		close(resultsQueue)
	}()
	
	// Collect results and generate final schema
	ws.collectResultsAndGenerateSchema(sessionID, propertyData, resultsQueue, progressQueue, totalPhotos)
}

// imageAnalysisWorker processes images concurrently using Gemini Vision API
func (ws *WebSocketVideoHandler) imageAnalysisWorker(
	imageQueue <-chan ImageAnalysisTask,
	resultsQueue chan<- ImageAnalysisResult,
	progressQueue chan<- ProgressUpdate,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	
	// Initialize Gemini client for this worker
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: ws.cfg.GoogleVeoAPIKey,
	})
	if err != nil {
		log.Printf("Failed to initialize Gemini client: %v", err)
		return
	}
	defer client.Close()
	
	for task := range imageQueue {
		result := ws.analyzeImageWithGemini(client, task)
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
func (ws *WebSocketVideoHandler) analyzeImageWithGemini(client *genai.Client, task ImageAnalysisTask) ImageAnalysisResult {
	result := ImageAnalysisResult{
		Index:    task.Index,
		Filename: task.Photo.Filename,
		Error:    nil,
	}
	
	// Use mock implementation for now (can be switched to real Gemini calls)
	result = ws.mockImageAnalysis(task.Photo)
	
	// Uncomment for real Gemini implementation:
	/*
	image := &genai.Image{
		ImageBytes: task.Photo.Data,
		MIMEType:   task.Photo.MimeType,
	}
	
	prompt := `Analyze this property photo for real estate marketing. Provide JSON response with:
	{
		"room_type": "kitchen|living_room|bedroom|bathroom|exterior|dining_room|etc",
		"features": ["feature1", "feature2"],
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
	
	result = ws.parseGeminiImageResponse(response, task.Photo)
	*/
	
	return result
}

// mockImageAnalysis provides realistic mock analysis for development
func (ws *WebSocketVideoHandler) mockImageAnalysis(photo PropertyPhoto) ImageAnalysisResult {
	filename := strings.ToLower(photo.Filename)
	
	// Simulate processing delay
	time.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond)
	
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

// sendUpdate sends a progress update to the client
func (ws *WebSocketVideoHandler) sendUpdate(sessionID, status string, progress int, message string) {
	ws.realtorMutex.RLock()
	session, exists := ws.realtorSessions[sessionID]
	ws.realtorMutex.RUnlock()
	
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

// cleanupSession removes a session from memory
func (ws *WebSocketVideoHandler) cleanupSession(sessionID string) {
	ws.realtorMutex.Lock()
	delete(ws.realtorSessions, sessionID)
	ws.realtorMutex.Unlock()
}

// extractUploadData parses WebSocket message for property data and photos
func (ws *WebSocketVideoHandler) extractUploadData(message map[string]interface{}) (PropertyData, []PropertyPhoto, error) {
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