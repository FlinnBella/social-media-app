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
	"time"
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

// GenerateSpeechToTmp generates TTS audio and writes it under tmpDir.
// Returns the absolute output path and the filename.
// generates a set of audio files, used for concatenated in ffmpeg
func (els *ElevenLabsService) GenerateSpeechToTmp(input models.TTSInput, tmpDir string) (filenames []string, fileoutputmap map[string]string, err error) {
	// Buffer for text-to-speech parts
	var parts []string
	//exact outpath paths into TmpDir
	//exact filenames; prefixed with elevenlabs_
	var flnames []string
	var filetomap map[string]string
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
		return []string{}, map[string]string{}, fmt.Errorf("failed to marshal TTS request: %v", err)
	}

	url := fmt.Sprintf("%s/text-to-speech/%s", els.config.ElevenLabsBaseURL, input.Narration.Voice.VoiceID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return []string{}, map[string]string{}, fmt.Errorf("failed to create TTS request: %v", err)
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", els.config.ElevenLabsAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, map[string]string{}, fmt.Errorf("failed to make TTS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return []string{}, map[string]string{}, fmt.Errorf("TTS API returned status %d: %s", resp.StatusCode, string(body))
	}

	if tmpDir == "" {
		tmpDir = filepath.Join(os.TempDir(), "tts_audio")
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return []string{}, map[string]string{}, fmt.Errorf("failed to create temp dir: %v", err)
	}
	filename := fmt.Sprintf("audio_%d.mp3", time.Now().UnixNano())
	outputPath := filepath.Join(tmpDir, filename)
	file, err := os.Create(outputPath)
	if err != nil {
		return []string{}, map[string]string{}, fmt.Errorf("failed to create audio file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return []string{}, map[string]string{}, fmt.Errorf("failed to save audio file: %v", err)
	}
	filetomap[filename] = outputPath

	return flnames, filetomap, nil
}
