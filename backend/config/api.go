package config

import "os"

type APIConfig struct {
	ElevenLabsAPIKey string
	ElevenLabsBaseURL string
	PexelsAPIKey     string
	PexelsBaseURL    string
	Port             string
}

func LoadAPIConfig() *APIConfig {
	return &APIConfig{
		ElevenLabsAPIKey:  getEnvOrDefault("ELEVENLABS_API_KEY", "your-elevenlabs-api-key-here"),
		ElevenLabsBaseURL: getEnvOrDefault("ELEVENLABS_BASE_URL", "https://api.elevenlabs.io/v1"),
		PexelsAPIKey:      getEnvOrDefault("PEXELS_API_KEY", "your-pexels-api-key-here"),
		PexelsBaseURL:     getEnvOrDefault("PEXELS_BASE_URL", "https://api.pexels.com/v1"),
		Port:              getEnvOrDefault("PORT", "8080"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}