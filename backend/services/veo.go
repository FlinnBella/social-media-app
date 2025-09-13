package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"

	"social-media-ai-video/config"
	"social-media-ai-video/models"
	"social-media-ai-video/models/interfaces"
	"social-media-ai-video/utils"

	//errgroup for concurrency
	"golang.org/x/sync/errgroup"
)

type VeoService struct {
	cfg *config.APIConfig
}

func NewVeoService(cfg *config.APIConfig) *VeoService {
	return &VeoService{
		cfg: cfg,
	}
}

// Interface-compliant method for single video generation
func (vs *VeoService) GenerateVideo(prompt string, images []string, cfg *config.APIConfig) (*models.FileOutput, error) {
	// For now, generate the first image as a single video
	if len(images) == 0 {
		return nil, fmt.Errorf("no images provided")
	}

	// Create unique request directory
	requestID := utils.GenerateUniqueID()
	outputDir := filepath.Join("./tmp", "veo", requestID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Generate video using the first image
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  vs.cfg.GoogleVeoAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %v", err)
	}

	// Read and process the first image
	imageData, err := os.ReadFile(images[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %v", err)
	}

	// Convert to base64
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(imageData)))
	base64.StdEncoding.Encode(dst, imageData)

	img := &genai.Image{ImageBytes: dst, MIMEType: "image/jpeg"}

	// Generate video
	operation, err := client.Models.GenerateVideos(
		ctx,
		"veo-3.0-generate-001",
		prompt,
		img,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start video generation: %v", err)
	}

	// Poll until complete
	for !operation.Done {
		log.Println("Waiting for video generation to complete...")
		time.Sleep(10 * time.Second)
		operation, err = client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to check operation status: %v", err)
		}
	}

	// Download the generated video
	video := operation.Response.GeneratedVideos[0]
	client.Files.Download(ctx, video.Video, nil)

	if video.Video.MIMEType != "video/mp4" {
		return nil, fmt.Errorf("invalid video mime type: %s", video.Video.MIMEType)
	}

	// Write to disk
	filename := fmt.Sprintf("veo_%s.mp4", requestID)
	filepath := filepath.Join(outputDir, filename)

	if err := os.WriteFile(filepath, video.Video.VideoBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write video file: %v", err)
	}

	// Return FileOutput
	return models.NewFileOutput(filepath, outputDir), nil
}

//pass this to the ffmpeg composer & compilier
//can use ducktyping to make it idenitcal to the other 'pure' ffmpeg implementation

func (vs *VeoService) GenerateVideoMultipart(c *gin.Context, boundary string) ([]string, [][]byte, error) {
	promptFound := false
	//Multipart checks done before; stream to generate multiple videos

	//Each image detected triggers a new video generation,
	//stream it into veo and send off when done.

	mr := multipart.NewReader(c.Request.Body, boundary)

	// Read prompt and the first image part into memory (SDK requires []byte). Limit image size to 25MB.
	var img []*genai.Image
	req_prompt := ""
	//eventually turn this into a goroutine, and hook channels to it
	//right now this dosen't handle multple images at all;
	//it's going to loop over the images, and just return the last one
	func() {
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read part error: %v", err)})
				return
			}

			// so, client labels the formName parameter for images with the field name "image"
			if part.FileName() != "" && part.FormName() == "image" {
				lr := io.LimitReader(part, 25<<20)
				b, rerr := io.ReadAll(lr)
				part.Close()
				if rerr != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read image error: %v", rerr)})
					return
				}

				mimeType := part.Header.Get("Content-Type")
				if strings.HasPrefix(mimeType, "image/png") || strings.HasPrefix(mimeType, "image/jpeg") {
					//have to convert to b64
					dst := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
					base64.StdEncoding.Encode(dst, b)
					img = append(img, &genai.Image{ImageBytes: dst, MIMEType: mimeType})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("invalid image mime type: %s", mimeType)})
					return
				} // do not break; continue scanning parts to also capture prompt
			} else if part.FileName() == "" && part.FormName() == "prompt" && !promptFound {
				b, rerr := io.ReadAll(part)
				part.Close()
				if rerr != nil {
					c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": fmt.Sprintf("read prompt error: %v", rerr)})
					return
				}
				req_prompt = string(b)
				promptFound = true
			}
			part.Close()
		}
	}()

	/*

		GOOGLE VEO SDK LOGIC

	*/

	// Create unique request directory for this multipart request
	requestID := utils.GenerateUniqueID()
	outputDir := filepath.Join("./tmp", "veo", requestID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  vs.cfg.GoogleVeoAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create genai client: %v", err)
	}

	prompt := req_prompt

	//need to implement a goroutine here, writer and reader impl
	//each goroutine
	g := new(errgroup.Group)
	filenames := make([]string, len(img))
	videos := make([][]byte, len(img))

	//goroutine spawns
	for i, im := range img {
		i, im := i, im // capture loop variables for the closure

		g.Go(func() error {
			operation, _ := client.Models.GenerateVideos(
				ctx,
				"veo-3.0-generate-001",
				prompt,
				im,
				nil,
			)

			// Poll the operation status until the video is ready.
			for !operation.Done {
				log.Println("Waiting for video generation to complete...")
				time.Sleep(10 * time.Second)
				operation, _ = client.Operations.GetVideosOperation(ctx, operation, nil)
			}

			// Download the generated video.
			video := operation.Response.GeneratedVideos[0]
			client.Files.Download(ctx, video.Video, nil)

			if video.Video.MIMEType != "video/mp4" {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("invalid video mime type: %s", video.Video.MIMEType)})
				return nil
			}
			filename := fmt.Sprintf("veo_%s_%d.mp4", requestID, i)
			filepath := filepath.Join(outputDir, filename)

			// Write video to disk
			if err := os.WriteFile(filepath, video.Video.VideoBytes, 0644); err != nil {
				return fmt.Errorf("failed to write video file %s: %v", filename, err)
			}

			log.Printf("Generated video saved to %s\n", filepath)

			// preserve order by assigning by index
			filenames[i] = filepath
			videos[i] = video.Video.VideoBytes
			return nil
		})
	}

	// wait for all goroutines to finish before returning
	if err := g.Wait(); err != nil {
		return nil, nil, fmt.Errorf("video generation failed: %v", err)
	}

	return filenames, videos, nil
}

// Interface compliance check
var _ interfaces.VideoGeneration = (*VeoService)(nil)
