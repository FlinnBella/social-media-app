package services

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"social-media-ai-video/config"

	"fmt"
	"time"
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

// ResolveAndDownload treats trackIdOrURL as a direct URL for now.
// Future: look up track by trackId/genre/mood via a catalog/service.
func (b *BackgroundMusic) CreateBackgroundMusic(mood string, genre string) (*MusicFile, error) {
	if mood == "" || genre == "" {
		return nil, fmt.Errorf("empty mood/genre")
	}
	resp, err := http.Get(fmt.Sprintf("https://api.soundcloud.com/tracks?client_id=YOUR_CLIENT_ID&genre=%s&mood=%s", genre, mood))
	if err != nil {
		return nil, fmt.Errorf("failed to GET music: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("music download failed: %s", resp.Status)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate filename: %w", err)
	}
	file_name := fmt.Sprintf("background_music_%d.mp3", time.Now().UnixNano())
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
