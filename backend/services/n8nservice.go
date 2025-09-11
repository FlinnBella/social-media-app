package services

import (
	"fmt"
	"mime"
	"net/http"
	"time"

	"social-media-ai-video/config"

	"github.com/gin-gonic/gin"
)

type N8NService struct {
	cfg *config.APIConfig
}

func NewN8NService(cfg *config.APIConfig) *N8NService {
	return &N8NService{cfg: cfg}
}

/*

Abstraction to return all the http responses from n8n

*/

func (ns *N8NService) Get(c *gin.Context, targetURL string) (*http.Response, error) {
	//Guard
	switch targetURL {
	//Eventually should just guard for everthing in the service here
	case ns.cfg.N8BTIMELINEURL:

		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		n8nReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to create request: %v", err)})
			return nil, err
		}

		n8nReq.Header.Set("Content-Type", c.GetHeader("Content-Type"))

		resp, err := client.Do(n8nReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": fmt.Sprintf("failed to send request: %v", err)})
			return nil, err
		}

		if resp == nil || resp.ContentLength == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "No content returned"})
			return nil, fmt.Errorf("no content returned")
		}

		//Response guard
		ct := resp.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(ct)
		if mediaType != "application/json" {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": "Content-Type must be application/json"})
			return nil, fmt.Errorf("unexpected content type: %s", resp.Header.Get("Content-Type"))
		}
		//GUARDS END

		return resp, nil

	case ns.cfg.N8NREELSURL:
		return nil, fmt.Errorf("unsupported target URL (Need to implement later): %s", targetURL)

	case ns.cfg.N8NPLEXELSURL:
		return nil, fmt.Errorf("unsupported target URL (Need to implement later): %s", targetURL)

	default:
		return nil, fmt.Errorf("unknown target URL: %s", targetURL)
	}
}
