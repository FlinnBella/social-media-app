package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"social-media-ai-video/config"
)

type BackgroundMusic struct {
	cfg *config.APIConfig
}

type MusicFile struct {
	FilePath string
	FileName string
}

func NewBackgroundMusic(cfg *config.APIConfig) *BackgroundMusic {
	return &BackgroundMusic{cfg: cfg}
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ResolveAndDownload treats trackIdOrURL as a direct URL for now.
// Future: look up track by trackId/genre/mood via a catalog/service.
func (b *BackgroundMusic) ResolveAndDownload(trackIdOrURL string) (*MusicFile, error) {
	if trackIdOrURL == "" {
		return nil, fmt.Errorf("empty track id/url")
	}
	resp, err := http.Get(trackIdOrURL)
	if err != nil {
		return nil, fmt.Errorf("failed to GET music: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("music download failed: %s", resp.Status)
	}

	id, err := randomHex(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate filename: %w", err)
	}
	file_name := id + ".mp3"
	outputDir := filepath.Join(os.TempDir(), "background_music")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create tmp dir: %w", err)
	}
	file_path := filepath.Join(outputDir, file_name)

	out, err := os.Create(file_path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &MusicFile{FilePath: file_path, FileName: file_name}, nil
}
