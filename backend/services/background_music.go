package services

import (
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

	filePath, fileName, err := func(mood string, genre string) (string, string, error) {
		//set this up later, but just invoke function for now; no actual logic
		filePath, fileName, err := SelectMusicHelper(mood, genre)
		if err != nil {
			return "", "", err
		}
		fileName = "Aurora%20on%20the%20Boulevard%20-%20National%20Sweetheart.mp3"
		filePath = "backend/music/Aurora%20on%20the%20Boulevard%20-%20National%20Sweetheart.mp3"

		return filePath, fileName, nil

	}(mood, genre)
	if err != nil {
		fmt.Errorf("failed to select music: %v", err)
		panic(err)
	}

	return &MusicFile{FilePath: filePath, FileName: fileName}, nil
}

/*
Future logic to select specific mp3; for now just pull a particular one from library
*/
func SelectMusicHelper(mood string, genre string) (string, string, error) {
	return "", "", nil
}
