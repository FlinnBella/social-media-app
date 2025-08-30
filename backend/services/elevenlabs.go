package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"social-media-ai-video/config"
	"social-media-ai-video/models"
)

type ElevenLabsService struct {
	config *config.APIConfig
}

type TTSRequest struct {
	Text          string                 `json:"text"`
	ModelID       string                 `json:"model_id"`
	VoiceSettings map[string]interface{} `json:"voice_settings"`
}

func NewElevenLabsService(cfg *config.APIConfig) *ElevenLabsService {
	return &ElevenLabsService{config: cfg}
}

func (els *ElevenLabsService) GenerateSpeech(ttsConfig models.TTSConfig, outputPath string) error {
	// Prepare request payload
	payload := TTSRequest{
		Text:    ttsConfig.Script,
		ModelID: "eleven_monolingual_v1",
		VoiceSettings: map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.5,
			"speed":           ttsConfig.Speed,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal TTS request: %v", err)
	}

	// Make API request
	url := fmt.Sprintf("%s/text-to-speech/%s", els.config.ElevenLabsBaseURL, ttsConfig.VoiceID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create TTS request: %v", err)
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", els.config.ElevenLabsAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make TTS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TTS API returned status %d", resp.StatusCode)
	}

	// Save audio file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create audio file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save audio file: %v", err)
	}

	return nil
}