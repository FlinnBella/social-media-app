package services

import (
	"encoding/json"
	"fmt"
	"social-media-ai-video/config"
	"social-media-ai-video/models"

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

// GenerateVideoComposition creates a structured video composition based on user prompt
func (cg *ContentGenerator) GenerateVideoComposition(prompt string) (*models.VideoCompositionRequest, error) {
	//start with a netowrk
	jsonData, err := json.Marshal(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt: %v", err)
	}
	resp, err := http.Post(cg.config.N8NPLEXELSURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to N8N: %v", err)
	}
	defer resp.Body.Close()

	// Debug: print N8N response status and body for verification
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read n8n response: %v", readErr)
	}
	fmt.Printf("N8N POST %s -> status: %s\n", cg.config.N8NPLEXELSURL, resp.Status)
	fmt.Printf("N8N response body: %s\n", string(bodyBytes))
	// Reset body for subsequent JSON decoding
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var composition models.VideoCompositionRequest
	if err := json.NewDecoder(resp.Body).Decode(&composition); err != nil {
		return nil, fmt.Errorf("failed to decode n8n response: %v", err)
	}
	return &composition, nil

}

// GenerateShortVideo posts the prompt to N8N, extracts the returned video id,
// fetches the short video from the configured base URL, saves it under ./tmp,
// and returns the saved filename (basename).
// GenerateVideoMultipart streams prompt + images to the selected source webhook as multipart/form-data.
// It posts to the appropriate N8N URL based on the given VideoSource and returns the parsed response.
func (cg *ContentGenerator) GenerateVideoSchemaMultipart(videoRequest models.VideoGenerationRequest) (*models.VideoCompositionResponse, error) {
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

	var parsed models.VideoCompositionResponse
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode upstream response: %v", err)
	}
	return &parsed, nil
}
