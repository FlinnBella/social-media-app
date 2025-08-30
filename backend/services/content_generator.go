package services

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"social-media-ai-video/models"
	"time"
)

type ContentGenerator struct{}

func NewContentGenerator() *ContentGenerator {
	return &ContentGenerator{}
}

// GenerateVideoComposition creates a structured video composition based on user prompt
func (cg *ContentGenerator) GenerateVideoComposition(prompt string) (*models.VideoCompositionRequest, error) {
	// For demo purposes, we'll generate a mock composition
	// In production, this would call your N8N webhook or AI service
	
	// Sample Pexels video URLs (these are real URLs for demo)
	sampleVideos := []string{
		"https://videos.pexels.com/video-files/3571264/3571264-uhd_2560_1440_30fps.mp4",
		"https://videos.pexels.com/video-files/2278095/2278095-uhd_2560_1440_30fps.mp4",
		"https://videos.pexels.com/video-files/1851190/1851190-uhd_2560_1440_30fps.mp4",
		"https://videos.pexels.com/video-files/2792043/2792043-uhd_2560_1440_30fps.mp4",
	}

	// Generate random composition based on prompt
	rand.Seed(time.Now().UnixNano())
	numSegments := rand.Intn(3) + 2 // 2-4 segments
	totalDuration := float64(15 + rand.Intn(16)) // 15-30 seconds

	segments := make([]models.VideoSegment, numSegments)
	segmentDuration := totalDuration / float64(numSegments)
	
	for i := 0; i < numSegments; i++ {
		videoURL := sampleVideos[rand.Intn(len(sampleVideos))]
		transitionType := "cut"
		if i > 0 && rand.Float32() > 0.5 {
			transitionType = "fade"
		}

		segments[i] = models.VideoSegment{
			ID:             fmt.Sprintf("segment_%d", i),
			PexelsVideoURL: videoURL,
			StartTime:      float64(i) * segmentDuration,
			Duration:       segmentDuration,
			Transition: models.Transition{
				Type:     transitionType,
				Duration: 0.5,
			},
			Effects: models.Effects{
				Zoom:  1.0,
				Speed: 1.0,
			},
		}
	}

	// Generate script based on prompt
	script := cg.generateScript(prompt)

	composition := &models.VideoCompositionRequest{
		VideoLength: totalDuration,
		AspectRatio: "9:16",
		Resolution: models.Resolution{
			Width:  1080,
			Height: 1920,
		},
		BackgroundMusic: models.BackgroundMusic{
			Enabled: true,
			Volume:  0.3,
			Genre:   "upbeat",
		},
		TTSConfig: models.TTSConfig{
			Script:  script,
			VoiceID: "21m00Tcm4TlvDq8ikWAM", // Default ElevenLabs voice
			Speed:   1.0,
			Volume:  0.8,
		},
		VideoSegments: segments,
	}

	return composition, nil
}

func (cg *ContentGenerator) generateScript(prompt string) string {
	// Simple script generation based on prompt
	// In production, this would use AI to generate appropriate narration
	
	scripts := []string{
		fmt.Sprintf("Check out this amazing content about %s! This is exactly what you've been looking for.", prompt),
		fmt.Sprintf("Here's something incredible about %s that will blow your mind!", prompt),
		fmt.Sprintf("You won't believe what we discovered about %s. Watch until the end!", prompt),
		fmt.Sprintf("The ultimate guide to %s is here. Don't miss out on this!", prompt),
	}
	
	rand.Seed(time.Now().UnixNano())
	return scripts[rand.Intn(len(scripts))]
}