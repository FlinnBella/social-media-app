package services

import (
	"encoding/json"
	"fmt"
	"social-media-ai-video/config"
	"social-media-ai-video/models"
	video_models_reels_ffmpeg "social-media-ai-video/models/video_ffmpeg"

	//"time"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
)

type ContentGenerator struct {
	config *config.APIConfig
}

func NewContentGenerator(cfg *config.APIConfig) *ContentGenerator {
	return &ContentGenerator{config: cfg}
}

// GenerateShortVideo posts the prompt to N8N, extracts the returned video id,
// fetches the short video from the configured base URL, saves it under ./tmp,
// and returns the saved filename (basename).
// GenerateVideoMultipart streams prompt + images to the selected source webhook as multipart/form-data.
// It posts to the appropriate N8N URL based on the given VideoSource and returns the parsed response.
func (cg *ContentGenerator) GenerateVideoSchemaMultipart(videoRequest models.VideoGenerationRequest) (*video_models_reels_ffmpeg.VideoCompositionResponse, error) {
	var targetURL string
	switch videoRequest.Source {
	case models.VideoSourceReels:
		targetURL = cg.config.N8NREELSURL
	case models.VideoSourcePexels:
		targetURL = cg.config.N8NPLEXELSURL
	default:
		return nil, fmt.Errorf("unknown video source: %s", string(videoRequest.Source))
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("prompt", videoRequest.Prompt); err != nil {
		return nil, fmt.Errorf("failed to write prompt field: %v", err)
	}
	// Append each image as a repeated 'image' part and include matching 'image_name' fields
	for idx, img := range videoRequest.Images {
		name := fmt.Sprintf("image_%d.jpg", idx+1)
		if idx < len(videoRequest.ImageNames) && videoRequest.ImageNames[idx] != "" {
			name = videoRequest.ImageNames[idx]
		}
		if err := mw.WriteField("image_name", name); err != nil {
			return nil, fmt.Errorf("failed to write image_name field: %v", err)
		}
		part, err := mw.CreateFormFile("image", name)
		if err != nil {
			return nil, fmt.Errorf("failed to create file part: %v", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(img)); err != nil {
			return nil, fmt.Errorf("failed to copy file content: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest("POST", targetURL, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	//this right here seems to be assuming that the return data
	//is also multipart/form; could be source of error, need
	//to check with end of pipeline
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream error: %s - %s", resp.Status, string(respBytes))
	}

	var parsed video_models_reels_ffmpeg.VideoCompositionResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode upstream response: %v", err)
	}
	return &parsed, nil
}
