package services

import (
	"context"
	"encoding/base64"
	"io"
	"log"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"

	"social-media-ai-video/config"

	"fmt"
	"net/http"
	"strings"
)

type VeoService struct {
	cfg *config.APIConfig
}

func NewVeoService(cfg *config.APIConfig) *VeoService {
	return &VeoService{
		cfg: cfg,
	}
}

func (vs *VeoService) GenerateVideoMultipart(c *gin.Context, boundary string) (string, []byte) {
	promptFound := false
	//Multipart checks done before; stream to generate multiple videos

	//Each image detected triggers a new video generation,
	//stream it into veo and send off when done.

	mr := multipart.NewReader(c.Request.Body, boundary)

	// Read prompt and the first image part into memory (SDK requires []byte). Limit image size to 25MB.
	var img *genai.Image
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
					img = &genai.Image{ImageBytes: dst, MIMEType: mimeType}
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

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	prompt := req_prompt

	//need to implement a goroutine here, writer and reader impl
	//each goroutine

	operation, _ := client.Models.GenerateVideos(
		ctx,
		"veo-3.0-generate-001",
		prompt,
		img,
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
		return "", nil
	}
	fname := fmt.Sprintf("veo_%d.mp4", time.Now().UnixNano())

	log.Printf("Generated video saved to %s\n", fname)
	return fname, video.Video.VideoBytes
	//avoid os file writes for now
	//_ = os.WriteFile(fname, video.Video.VideoBytes, 0644)
}
