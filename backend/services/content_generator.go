package services

import (
	"encoding/json"
	"fmt"
	"social-media-ai-video/models"
	"social-media-ai-video/config"
	//"time"
	"net/http"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ContentGenerator struct{
	config *config.APIConfig
}

func NewContentGenerator(cfg *config.APIConfig) *ContentGenerator {
	return &ContentGenerator{config: cfg}
}

// GenerateVideoComposition creates a structured video composition based on user prompt
func (cg *ContentGenerator) GenerateVideoComposition(prompt string) (*models.VideoCompositionRequest, error) {
	//start with a netowrk
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt: %v", err)
	}
	resp, err := http.Post(cg.config.N8NURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to N8N: %v", err)
	}
	defer resp.Body.Close()

	// Debug: print N8N response status and body for verification
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read n8n response: %v", readErr)
	}
	fmt.Printf("N8N POST %s -> status: %s\n", cg.config.N8NURL, resp.Status)
	fmt.Printf("N8N response body: %s\n", string(bodyBytes))
	// Reset body for subsequent JSON decoding
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var composition models.VideoCompositionRequest
	if err := json.NewDecoder(resp.Body).Decode(&composition); err != nil {
		return nil, fmt.Errorf("failed to decode n8n response: %v", err)
	}
	return &composition, nil
	
	// For demo purposes, we'll generate a mock composition
	// In production, this would call your N8N webhook or AI service
	
	// Sample Pexels video URLs (these are real URLs for demo)
	//sampleVideos := []string{
	//	"https://videos.pexels.com/video-files/3571264/3571264-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/2278095/2278095-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/1851190/1851190-uhd_2560_1440_30fps.mp4",
	//	"https://videos.pexels.com/video-files/2792043/2792043-uhd_2560_1440_30fps.mp4",
	//}

	// Generate random composition based on prompt
	//rand.Seed(time.Now().UnixNano())
	//numSegments := rand.Intn(3) + 2 // 2-4 segments
	//totalDuration := float64(15 + rand.Intn(16)) // 15-30 seconds

}

// GenerateShortVideo posts the prompt to N8N, extracts the returned video id,
// fetches the short video from the configured base URL, saves it under ./tmp,
// and returns the saved filename (basename).
func (cg *ContentGenerator) GenerateShortVideo(prompt string) (string, error) {
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to marshal prompt: %v", err)
	}
	resp, err := http.Post(cg.config.N8NURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to send request to N8N: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read n8n response: %v", readErr)
	}
	fmt.Printf("N8N short-video POST %s -> status: %s\n", cg.config.N8NURL, resp.Status)
	fmt.Printf("N8N short-video response body: %s\n", string(bodyBytes))

	// Decode according to provided schema: { videoId, videoTitle }
	type n8nResp struct {
		VideoID    string `json:"videoId"`
		VideoTitle string `json:"videoTitle"`
		TotalResponseTime string `json:"total-response-time"`
	}
	var nr n8nResp
	if err := json.Unmarshal(bodyBytes, &nr); err != nil {
		// Fallback: treat body as plain string id
		trim := strings.TrimSpace(string(bodyBytes))
		trim = strings.Trim(trim, `"`)
		if trim == "" {
			return "", fmt.Errorf("failed to decode n8n response: %v", err)
		}
		nr.VideoID = trim
	}
	videoID := nr.VideoID
	if videoID == "" {
		return "", fmt.Errorf("n8n response missing videoId")
	}

	base := strings.TrimRight(cg.config.ShortVideoBaseURL, "/")
	getURL := fmt.Sprintf("%s/api/short-video/%s", base, videoID)
	fmt.Printf("Fetching short video from: %s\n", getURL)

	// Wait 5 minutes before first GET, then poll every 40 seconds (5 intervals)
	const initialDelay = 5 * time.Minute
	const perAttemptDelay = 120 * time.Second
	time.Sleep(initialDelay)

	var getResp *http.Response
	const maxAttempts = 6 // 5 waits => up to 6 attempts
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		start := time.Now()
		getResp, err = http.Get(getURL)
		elapsed := time.Since(start)
		fmt.Printf("Attempt %d/%d: GET took %s\n", attempt, maxAttempts, elapsed)
		if err == nil && getResp.StatusCode == http.StatusOK {
			break
		}
		if err != nil {
			fmt.Printf("Attempt %d/%d: error: %v\n", attempt, maxAttempts, err)
		}
		if getResp != nil {
			b, _ := io.ReadAll(getResp.Body)
			getResp.Body.Close()
			fmt.Printf("Attempt %d/%d: GET failed: %s - %s\n", attempt, maxAttempts, getResp.Status, string(b))
		}
		if attempt < maxAttempts {
			time.Sleep(perAttemptDelay)
		}
	}
	if err != nil {
		return "", fmt.Errorf("failed to GET short video: %v", err)
	}
	if getResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(getResp.Body)
		getResp.Body.Close()
		return "", fmt.Errorf("short video GET failed: %s - %s", getResp.Status, string(b))
	}
	defer getResp.Body.Close()

	if err := os.MkdirAll("./tmp", 0755); err != nil {
		return "", fmt.Errorf("failed to create tmp dir: %v", err)
	}

	ext := ".mp4"
	if ct := getResp.Header.Get("Content-Type"); strings.Contains(ct, "webm") {
		ext = ".webm"
	} else if strings.Contains(ct, "ogg") {
		ext = ".ogg"
	} else if strings.Contains(ct, "quicktime") || strings.Contains(ct, "mov") {
		ext = ".mov"
	}
	filename := fmt.Sprintf("short_%s_%d%s", videoID, time.Now().Unix(), ext)
	outPath := filepath.Join("./tmp", filename)

	outFile, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, getResp.Body); err != nil {
		return "", fmt.Errorf("failed to save video: %v", err)
	}

	return filename, nil
}