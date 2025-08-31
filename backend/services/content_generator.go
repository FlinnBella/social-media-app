package services

import (
	"encoding/json"
	"fmt"
	"social-media-ai-video/models"
	"social-media-ai-video/config"
	//"time"
	"net/http"
	"bytes"
)

type ContentGenerator struct{
	config *config.APIConfig
}

func NewContentGenerator(cfg *config.APIConfig) *ContentGenerator {
	return &ContentGenerator{config: cfg}
}

// GenerateVideoComposition creates a structured video composition based on user prompt
func (cg *ContentGenerator) GenerateVideoComposition(prompt string) (*models.VideoCompositionRequest, error) {
	//start with a netowrk
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt: %v", err)
	}
	resp, err := http.Post(cg.config.N8NURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to N8N: %v", err)
	}
	defer resp.Body.Close()

	var composition models.VideoCompositionRequest
	if err := json.NewDecoder(resp.Body).Decode(&composition); err != nil {
		return nil, fmt.Errorf("failed to decode n8n response: %v", err)
	}
	return &composition, nil
	
	// For demo purposes, we'll generate a mock composition
	// In production, this would call your N8N webhook or AI service
	
	// Sample Pexels video URLs (these are real URLs for demo)
	//sampleVideos := []string{
	//	"https://videos.pexels.com/video-files/3571264/3571264-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/2278095/2278095-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/1851190/1851190-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/2792043/2792043-uhd_2560_1440_30fps.mp4",
	//}

	// Generate random composition based on prompt
	//rand.Seed(time.Now().UnixNano())
	//numSegments := rand.Intn(3) + 2 // 2-4 segments
	//totalDuration := float64(15 + rand.Intn(16)) // 15-30 seconds

}