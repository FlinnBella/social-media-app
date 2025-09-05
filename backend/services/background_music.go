package services

import (
	"path/filepath"

	"social-media-ai-video/config"

	"fmt"
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

	filePath, err := SelectMusicHelper(mood, genre)
	if err != nil {
		return nil, err
	}

	fileName := ""
	if filePath != "" {
		fileName = filepath.Base(filePath)
	}

	return &MusicFile{FilePath: filePath, FileName: fileName}, nil
}

/*
Future logic to select specific mp3; for now just pull a particular one from library
*/
func SelectMusicHelper(mood string, genre string) (string, error) {
	return "", nil
}
