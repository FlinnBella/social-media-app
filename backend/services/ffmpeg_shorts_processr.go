package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"social-media-ai-video/models"
	"time"
)

type VideoProcessor struct {
	tempDir string
}

func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{
		tempDir: filepath.Join(os.TempDir(), "video_processing"),
	}
}

// need a new schema here that takes in the file dirs, filenames, and specs for ffmpeg
func (vp *VideoProcessor) ProcessVideo(request *models.VideoCompositionRequest) (string, error) {
	// Create temporary directory
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	sessionDir := filepath.Join(vp.tempDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %v", err)
	}
	defer os.RemoveAll(sessionDir) // Clean up after processing

	// Download all video segments
	videoFiles := make([]string, len(request.VideoSegments))
	for i, segment := range request.VideoSegments {
		videoPath := filepath.Join(sessionDir, fmt.Sprintf("segment_%d.mp4", i))
		if err := vp.downloadVideo(segment.PexelsVideoURL, videoPath); err != nil {
			return "", fmt.Errorf("failed to download video segment %d: %v", i, err)
		}
		videoFiles[i] = videoPath
	}

	// Generate TTS audio
	audioPath := filepath.Join(sessionDir, "tts_audio.mp3")
	if err := vp.generateTTS(request.TTSConfig, audioPath); err != nil {
		return "", fmt.Errorf("failed to generate TTS: %v", err)
	}

	// Download background music
	musicPath := filepath.Join(sessionDir, "background_music.mp3")
	if request.BackgroundMusic.Enabled {
		if err := vp.downloadBackgroundMusic(request.BackgroundMusic.Genre, musicPath); err != nil {
			return "", fmt.Errorf("failed to download background music: %v", err)
		}
	}

	// Process and combine videos
	outputPath := filepath.Join(sessionDir, "final_video.mp4")
	if err := vp.combineVideos(request, videoFiles, audioPath, musicPath, outputPath); err != nil {
		return "", fmt.Errorf("failed to combine videos: %v", err)
	}

	// Read final video for serving
	videoData, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read final video: %v", err)
	}

	// For demo purposes, we'll return a placeholder URL
	// In production, you'd upload to cloud storage and return the URL
	finalVideoURL := fmt.Sprintf("data:video/mp4;base64,%s", videoData)

	return finalVideoURL, nil
}

func (vp *VideoProcessor) downloadVideo(url, outputPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func (vp *VideoProcessor) generateTTS(config models.TTSConfig, outputPath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-t", "10",
		"-c:a", "mp3",
		"-y", // overwrite output if exists
		outputPath,
	)
	return cmd.Run()
}

func (vp *VideoProcessor) downloadBackgroundMusic(genre, outputPath string) error {
	cmd := exec.Command(
		"ffmpeg",
		"-f", "lavfi",
		"-i", "anullsrc=channel_layout=stereo:sample_rate=48000",
		"-t", "30",
		"-c:a", "mp3",
		"-y",
		outputPath,
	)
	return cmd.Run()
}

func (vp *VideoProcessor) combineVideos(
	request *models.VideoCompositionRequest,
	videoFiles []string,
	audioPath, musicPath, outputPath string,
) error {

	args := []string{}

	// Input videos
	for _, v := range videoFiles {
		args = append(args, "-i", v)
	}

	// Input TTS audio
	args = append(args, "-i", audioPath)

	// Input background music if enabled
	if request.BackgroundMusic.Enabled {
		args = append(args, "-i", musicPath)
	}

	// Build filter_complex string
	filterComplex := vp.buildFilterComplex(request, len(videoFiles))
	var audioFilter string
	if request.BackgroundMusic.Enabled {
		audioFilter = fmt.Sprintf("[%d:a][%d:a]amix=inputs=2:duration=first:dropout_transition=2,volume=%.2f[audio]",
			len(videoFiles), len(videoFiles)+1, request.TTSConfig.Volume)
	} else {
		audioFilter = fmt.Sprintf("[%d:a]volume=%.2f[audio]", len(videoFiles), request.TTSConfig.Volume)
	}

	fullFilter := filterComplex + ";" + audioFilter

	args = append(args,
		"-filter_complex", fullFilter,
		"-map", "[video]",
		"-map", "[audio]",
		"-c:v", "libx264",
		"-c:a", "aac",
		"-preset", "fast",
		"-crf", "23",
		"-r", "30",
		"-s", fmt.Sprintf("%dx%d", request.Resolution.Width, request.Resolution.Height),
		"-y", // overwrite
		outputPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v, output: %s", err, string(output))
	}
	return nil
}

func (vp *VideoProcessor) buildFilterComplex(request *models.VideoCompositionRequest, numVideos int) string {
	var filter string

	// Scale and crop each video to target resolution
	for i := 0; i < numVideos; i++ {
		filter += fmt.Sprintf("[%d:v]scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d[v%d];",
			i, request.Resolution.Width, request.Resolution.Height, request.Resolution.Width, request.Resolution.Height, i)
	}

	// Concatenate videos with transitions
	if numVideos == 1 {
		filter += "[v0]copy[video]"
	} else {
		// Build concatenation with transitions
		concatInputs := ""
		for i := 0; i < numVideos; i++ {
			concatInputs += fmt.Sprintf("[v%d]", i)
		}
		filter += fmt.Sprintf("%sconcat=n=%d:v=1:a=0[video]", concatInputs, numVideos)
	}

	return filter
}
