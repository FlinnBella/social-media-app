package services

import (
	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/services/internal"
	"social-media-ai-video/utils"

	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BackgroundMusic struct {
	cfg *config.APIConfig
}

// MusicFile is now replaced by models.FileOutput

func NewBackgroundMusic(cfg *config.APIConfig) *BackgroundMusic {
	return &BackgroundMusic{cfg: cfg}
}

// CreateBackgroundMusic selects a random track from the specified genre and trims it to 30 seconds
func (b *BackgroundMusic) GenerateMusic(genre string, cfg *config.APIConfig) (*models.FileOutput, error) {
	if genre == "" {
		return nil, fmt.Errorf("empty mood/genre")
	}

	// Get tracks for the specified genre
	tracks, exists := internal.MusicThemeMap[genre]
	if !exists {
		return nil, fmt.Errorf("unknown genre: %s", genre)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks available for genre: %s", genre)
	}

	// Select a random track
	rand.Seed(time.Now().UnixNano())
	selectedTrack := tracks[rand.Intn(len(tracks))]

	// Create unique request directory
	requestID := utils.GenerateUniqueID()
	tmpDir := filepath.Join("./tmp", "background_music", requestID)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}

	// Create trimmed version of the track (30 seconds)
	trimmedFilePath, err := b.trimMusicTo30Seconds(selectedTrack, tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir) // Clean up on error
		return nil, fmt.Errorf("failed to trim music: %v", err)
	}

	return models.NewFileOutput(trimmedFilePath, tmpDir), nil
}

// trimMusicTo30Seconds uses ffmpeg to trim a background music file to 30 seconds
func (b *BackgroundMusic) trimMusicTo30Seconds(inputFileName, tmpDir string) (string, error) {
	// Construct paths
	sourcePath := filepath.Join("services/internal/music", inputFileName)

	// Preserve original filename but add "bgm_30s_" prefix to indicate it's background music trimmed to 30 seconds
	// Remove .mp3 extension, add bgm_30s prefix, then add .mp3 back
	baseName := strings.TrimSuffix(inputFileName, ".mp3")
	trimmedFileName := "bgm_30s_" + baseName + ".mp3"
	outputPath := filepath.Join(tmpDir, trimmedFileName)

	// Check if source file exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source music file not found: %s", sourcePath)
	}

	// Use ffmpeg to trim to 30 seconds
	cmd := exec.Command("ffmpeg",
		"-i", sourcePath,
		"-t", "30", // Duration: 30 seconds
		"-c", "copy", // Copy without re-encoding for speed
		"-y", // Overwrite output file
		outputPath,
	)

	// Run ffmpeg command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed to trim music: %v", err)
	}

	// Verify output file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("trimmed music file was not created")
	}

	return outputPath, nil
}
