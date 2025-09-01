package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"strings"
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

func (els *ElevenLabsService) GenerateSpeech(input models.TTSInput) error {
	// Concatenate narrative parts and narration script into one text
	var parts []string
	if input.Narrative.Hook != "" {
		parts = append(parts, input.Narrative.Hook)
	}
	if len(input.Narrative.Story) > 0 {
		parts = append(parts, strings.Join(input.Narrative.Story, " "))
	}
	if input.Narrative.Cta != "" {
		parts = append(parts, input.Narrative.Cta)
	}
	for _, seg := range input.Narration.Script {
		if strings.TrimSpace(seg.Text) != "" {
			parts = append(parts, seg.Text)
		}
	}
	text := strings.Join(parts, " ")

	payload := TTSRequest{
		Text:    text,
		ModelID: "eleven_monolingual_v1",
		VoiceSettings: map[string]interface{}{
			"stability":        input.Narration.Voice.Stability,
			"similarity_boost": 0.5,
			"speed":            input.Narration.Voice.Speed,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal TTS request: %v", err)
	}

	url := fmt.Sprintf("%s/text-to-speech/%s", els.config.ElevenLabsBaseURL, input.Narration.Voice.VoiceID)
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("TTS API returned status %d: %s", resp.StatusCode, string(body))
	}

	outputDir := filepath.Join(os.TempDir(), "tts_audio")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}
	file, err := os.CreateTemp(outputDir, "audio_*.mp3")
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

func MapCompositionToTTSInput(doc models.CompositionDocument) models.TTSInput {
	return models.TTSInput{
		Narrative: models.NarrativeData{
			Hook:  doc.Narrative.Hook,
			Story: doc.Narrative.Story,
			Cta:   doc.Narrative.Cta,
			Tone:  doc.Narrative.Tone,
		},
		Narration: models.NarrationData{
			Script: doc.Audio.Narration.Script,
			Voice:  doc.Audio.Narration.Voice,
		},
	}
}
