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

func NewElevenLabsService(cfg *config.APIConfig) *ElevenLabsService {
	return &ElevenLabsService{config: cfg}
}

type TTSRequest struct {
	Text          string                  `json:"text"`
	ModelID       string                  `json:"model_id"`
	VoiceSettings *map[string]interface{} `json:"voice_settings"`
}

const (
	femaleVoiceId = "Dslrhjl3ZpzrctukrQSN"
)

// GenerateSpeechToTmp generates TTS audio and writes it under tmpDir.
// Returns a FileOutput containing the file path, filename, and temp directory.
// generates a set of audio files, used for concatenated in ffmpeg
func (els *ElevenLabsService) GenerateSpeechToTmp(text []string) (*models.FileOutput, error) {
	// Buffer for text-to-speech parts

	//exact outpath paths into TmpDir
	//exact filenames; prefixed with elevenlabs_
	one_shot_script := strings.Join(text, " ")

	payload := TTSRequest{
		Text:          one_shot_script,
		ModelID:       "eleven_multilingual_v2",
		VoiceSettings: nil,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal TTS request: %v", err)
	}

	// For now, hardcode the voice ID

	url := fmt.Sprintf("%s/text-to-speech/%s", els.config.ElevenLabsBaseURL, femaleVoiceId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS request: %v", err)
	}

	req.Header.Set("Accept", "audio/mpeg")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", els.config.ElevenLabsAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make TTS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS API returned status %d: %s", resp.StatusCode, resp.Body)
	}
	/*
		Creating TmpDir and adding file
	*/
	tmpDir, err := os.MkdirTemp("", "elevenlabs_voice-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	file := filepath.Join(tmpDir, "elevenlabs_voice.mp3")

	// Create the file and stream directly from response body
	fileHandle, err := os.Create(file)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to create audio file: %v", err)
	}
	defer fileHandle.Close()

	// Stream directly from HTTP response to file
	_, err = io.Copy(fileHandle, resp.Body)
	if err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to write audio file: %v", err)
	}
	//set up mapping structure; want to split voice files up eventually... perhaps?

	// Optional: Schedule cleanup after 1 hour (uncomment if you want automatic cleanup)
	// go func() {
	//     time.Sleep(1 * time.Hour)
	//     os.RemoveAll(tmpDir)
	// }()

	return models.NewFileOutput(file, tmpDir), nil
}
