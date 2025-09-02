package config

import "os"

type APIConfig struct {
	Environment       string
	ElevenLabsAPIKey  string
	ElevenLabsBaseURL string
	N8NPLEXELSURL     string
	N8NREELSURL       string
	N8NAPIKey         string
	ShortVideoBaseURL string
	Port              string
}

func LoadAPIConfig() *APIConfig {
	// Determine environment (default to development)
	env := getEnvOrDefault("APP_ENV", "development")

	// Base URLs (can be overridden below)
	N8NPLEXELSURL := getEnvOrDefault("N8N_PLEXELS_URL", "https://evandickinson.app.n8n.cloud/webhook/pexels-workflow")
	N8NREELSURL := getEnvOrDefault("N8N_REELS_URL", "https://evandickinson.app.n8n.cloud/webhook/reels-workflow")
	if env == "development" {
		N8NPLEXELSURL = getEnvOrDefault("N8N_PLEXELS_URL", "https://evandickinson.app.n8n.cloud/webhook-test/pexels-workflow")
		N8NREELSURL = getEnvOrDefault("N8N_REELS_URL", "https://evandickinson.app.n8n.cloud/webhook-test/reels-workflow")
	}

	return &APIConfig{
		Environment: env,
		//using the mary voice id: spanish, young BITCH!
		ElevenLabsAPIKey:  getEnvOrDefault("ELEVENLABS_API_KEY", "sk_c78e929e9d436804555d72838e56b279d994faf76151a3b6"),
		ElevenLabsBaseURL: getEnvOrDefault("ELEVENLABS_BASE_URL", "https://api.elevenlabs.io/v1/text-to-speech/:tvWD4i07Hg5L4uEvbxYV"),
		N8NPLEXELSURL:     N8NPLEXELSURL,
		N8NREELSURL:       N8NREELSURL,
		N8NAPIKey:         getEnvOrDefault("N8N_API_KEY", "n8n_api_key_here"),
		ShortVideoBaseURL: getEnvOrDefault("SHORT_VIDEO_BASE_URL", "http://34.66.33.115:3123"),
		Port:              getEnvOrDefault("PORT", "8080"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
