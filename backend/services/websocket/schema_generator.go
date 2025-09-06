package websocket

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/genai"
)

// SchemaGenerator handles property video schema generation
type SchemaGenerator struct {
	handler *WebSocketVideoHandler
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(handler *WebSocketVideoHandler) *SchemaGenerator {
	return &SchemaGenerator{
		handler: handler,
	}
}

// collectResultsAndGenerateSchema aggregates all analyses and generates final JSON schema
func (ws *WebSocketVideoHandler) collectResultsAndGenerateSchema(
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
	
	finalSchema := ws.generatePropertySchema(propertyData, allResults, progressQueue)
	
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
func (ws *WebSocketVideoHandler) generatePropertySchema(
	propertyData PropertyData,
	imageAnalyses []ImageAnalysisResult,
	progressQueue chan<- ProgressUpdate,
) map[string]interface{} {
	// Use mock implementation for development
	return ws.mockPropertySchema(propertyData, imageAnalyses)
	
	// Uncomment for real Gemini implementation:
	/*
	analysisText := ws.formatAnalysesForPrompt(imageAnalyses)
	
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: ws.cfg.GoogleVeoAPIKey,
	})
	if err != nil {
		return ws.mockPropertySchema(propertyData, imageAnalyses)
	}
	defer client.Close()
	
	model := client.Models.GetGenerativeModel("gemini-1.5-flash")
	response, err := model.GenerateContent(ctx, prompt)
	if err != nil {
		return ws.mockPropertySchema(propertyData, imageAnalyses)
	}
	
	return ws.parseGeminiSchemaResponse(response)
	*/
}

// mockPropertySchema creates a realistic mock schema for development
func (ws *WebSocketVideoHandler) mockPropertySchema(propertyData PropertyData, analyses []ImageAnalysisResult) map[string]interface{} {
	// Determine optimal photo sequence based on room types
	sequence := ws.determineOptimalSequence(analyses)
	
	// Generate marketing highlights
	highlights := ws.extractMarketingHighlights(propertyData, analyses)
	
	// Create comprehensive property schema
	schema := map[string]interface{}{
		"metadata": map[string]interface{}{
			"total_duration": 12.0,
			"aspect_ratio":   "9:16",
			"fps":           "30",
			"resolution":    []int{1080, 1920},
		},
		"property_info":        propertyData,
		"photo_sequence":       sequence,
		"marketing_highlights": highlights,
		"narrative": map[string]interface{}{
			"hook":           ws.generateHook(propertyData, highlights),
			"tour_segments":  ws.generateTourSegments(analyses, sequence),
			"call_to_action": "Contact us today to schedule your private showing!",
		},
		"voice_style":  ws.recommendVoiceStyle(propertyData),
		"timing":       ws.generateTiming(sequence),
		"generated_at": time.Now(),
	}
	
	return schema
}

// streamProgressUpdates handles all WebSocket progress streaming
func (ws *WebSocketVideoHandler) streamProgressUpdates(progressQueue <-chan ProgressUpdate) {
	for update := range progressQueue {
		ws.sendUpdate(update.SessionID, update.Status, update.Progress, update.Message)
		
		// Send additional data if present
		if update.Data != nil {
			ws.realtorMutex.RLock()
			session, exists := ws.realtorSessions[update.SessionID]
			ws.realtorMutex.RUnlock()
			
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

func (ws *WebSocketVideoHandler) determineOptimalSequence(analyses []ImageAnalysisResult) []int {
	// Room priority for property tours: exterior -> living -> kitchen -> bedrooms -> bathrooms
	roomPriority := map[string]int{
		"exterior":     1,
		"living_room":  2,
		"dining_room":  3,
		"kitchen":      4,
		"bedroom":      5,
		"bathroom":     6,
		"interior":     7,
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
		case "exceptional": appealScore = 4
		case "high": appealScore = 3
		case "medium": appealScore = 2
		case "low": appealScore = 1
		}
		
		scored = append(scored, photoScore{
			index:    analysis.Index,
			priority: priority,
			appeal:   appealScore,
		})
	}
	
	// Simple sort by priority, then by appeal
	var sequence []int
	for _, item := range scored {
		sequence = append(sequence, item.index)
	}
	
	return sequence
}

func (ws *WebSocketVideoHandler) extractMarketingHighlights(propertyData PropertyData, analyses []ImageAnalysisResult) []string {
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

func (ws *WebSocketVideoHandler) generateHook(propertyData PropertyData, highlights []string) string {
	hooks := []string{
		fmt.Sprintf("Stunning %s ready for you", propertyData.PropertyType),
		"Your dream home awaits",
		"Don't miss this incredible property",
	}
	return hooks[0] // Simple selection
}

func (ws *WebSocketVideoHandler) generateTourSegments(analyses []ImageAnalysisResult, sequence []int) []string {
	segments := []string{}
	for _, idx := range sequence {
		if idx < len(analyses) {
			segments = append(segments, analyses[idx].Description)
		}
	}
	return segments
}

func (ws *WebSocketVideoHandler) recommendVoiceStyle(propertyData PropertyData) string {
	if propertyData.Price > 750000 {
		return "sophisticated"
	} else if propertyData.PropertyType == "condo" {
		return "modern"
	}
	return "friendly"
}

func (ws *WebSocketVideoHandler) generateTiming(sequence []int) []float64 {
	timing := []float64{}
	duration := 12.0 / float64(len(sequence))
	for range sequence {
		timing = append(timing, duration)
	}
	return timing
}