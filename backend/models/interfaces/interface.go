package interfaces

import (
	"social-media-ai-video/config"
	"social-media-ai-video/models"
)

type VoiceOver interface {
	GenerateVoiceOver(text []string, cfg *config.APIConfig) (*models.FileOutput, error)
}

type MusicGeneration interface {
	GenerateMusic(genre string, cfg *config.APIConfig) (*models.FileOutput, error)
}

type VideoGeneration interface {
	GenerateVideo(prompt string, images []string, cfg *config.APIConfig) (*models.FileOutput, error)
}
