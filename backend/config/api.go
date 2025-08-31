package config

import "os"

type APIConfig struct {
	ElevenLabsAPIKey string
	ElevenLabsBaseURL string
	PexelsAPIKey     string
	PexelsBaseURL    string
	N8NURL           string
	N8NAPIKey        string
	Port             string
}

func LoadAPIConfig() *APIConfig {
	return &APIConfig{
		ElevenLabsAPIKey:  getEnvOrDefault("ELEVENLABS_API_KEY", "your-elevenlabs-api-key-here"),
		ElevenLabsBaseURL: getEnvOrDefault("ELEVENLABS_BASE_URL", "https://api.elevenlabs.io/v1"),
		PexelsAPIKey:      getEnvOrDefault("PEXELS_API_KEY", "your-pexels-api-key-here"),
		PexelsBaseURL:     getEnvOrDefault("PEXELS_BASE_URL", "https://api.pexels.com/v1"),
		N8NURL:        	   getEnvOrDefault("N8N_URL", "https://evandickinson.app.n8n.cloud/webhook-test/5ca39975-fffb-4405-96c1-2be4c5eb5dbe"),
		N8NAPIKey:         getEnvOrDefault("N8N_API_KEY", "n8n_38761e841376469eaa017d3c898b52c2"),
		AISHORTS_BaseURL:  getEnvOrDefault("AISHORTS_GCPURL", "34.66.33.115:3123"),
		Port:              getEnvOrDefault("PORT", "8080"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}