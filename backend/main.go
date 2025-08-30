package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type VideoRequest struct {
	Prompt string `json:"prompt"`
}

type VideoResponse struct {
	VideoURL string `json:"videoUrl"`
	Error    string `json:"error,omitempty"`
}

type Veo3Request struct {
	Prompt     string `json:"prompt"`
	Duration   int    `json:"duration"`
	AspectRatio string `json:"aspect_ratio"`
	ImageURL   string `json:"image_url,omitempty"`
}

type Veo3Response struct {
	VideoURL string `json:"video_url"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

const VEO3_API_KEY = "your-veo3-api-key-here"
const VEO3_API_URL = "https://api.veo3.ai/v1/generate"

func main() {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	r.POST("/api/generate-video", generateVideo)

	fmt.Println("Server starting on :8080")
	r.Run(":8080")
}

func generateVideo(c *gin.Context) {
	prompt := c.PostForm("prompt")
	file, fileHeader, err := c.Request.FormFile("file")

	if prompt == "" && err != nil {
		c.JSON(http.StatusBadRequest, VideoResponse{
			Error: "Either prompt or file is required",
		})
		return
	}

	var imageURL string
	
	// Handle file upload if present
	if err == nil && file != nil {
		defer file.Close()
		
		// For demo purposes, we'll simulate uploading to a temporary storage
		// In production, you'd upload to cloud storage and get a URL
		imageURL = fmt.Sprintf("temp://uploaded-file-%d", time.Now().Unix())
		fmt.Printf("Received file: %s, size: %d bytes\n", fileHeader.Filename, fileHeader.Size)
	}

	// Prepare request to Veo3 API
	veo3Req := Veo3Request{
		Prompt:      prompt,
		Duration:    25, // 25 seconds as middle ground
		AspectRatio: "9:16", // Default to vertical for social media
		ImageURL:    imageURL,
	}

	// Call Veo3 API
	videoURL, err := callVeo3API(veo3Req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, VideoResponse{
			Error: fmt.Sprintf("Failed to generate video: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, VideoResponse{
		VideoURL: videoURL,
	})
}

func callVeo3API(req Veo3Request) (string, error) {
	// Simulate API call for demo purposes
	// In production, replace this with actual Veo3 API integration
	
	fmt.Printf("Calling Veo3 API with prompt: %s\n", req.Prompt)
	
	// Simulate processing time
	time.Sleep(3 * time.Second)
	
	// Return a demo video URL (you would replace this with actual Veo3 response)
	demoVideoURL := "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
	
	/* 
	// Actual Veo3 API call would look like this:
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", VEO3_API_URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+VEO3_API_KEY)

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var veo3Resp Veo3Response
	if err := json.NewDecoder(resp.Body).Decode(&veo3Resp); err != nil {
		return "", err
	}

	if veo3Resp.Error != "" {
		return "", fmt.Errorf("Veo3 API error: %s", veo3Resp.Error)
	}

	return veo3Resp.VideoURL, nil
	*/
	
	return demoVideoURL, nil
}