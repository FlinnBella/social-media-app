package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"social-media-ai-video/config"

	"google.golang.org/genai"
)

// VeoItem represents a stored base64 video and its MIME type
type VeoItem struct {
	MIMEType string
	Base64   string // raw base64 without data URI prefix
}

// VeoService manages Veo generation and an in-memory base64 store
type VeoService struct {
	cfg   *config.APIConfig
	mu    sync.RWMutex
	store map[string]*VeoItem
}

func NewVeoService(cfg *config.APIConfig) *VeoService {
	return &VeoService{cfg: cfg, store: make(map[string]*VeoItem)}
}

// Store operations
func (s *VeoService) Put(id string, item *VeoItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = item
}

func (s *VeoService) Get(id string) (*VeoItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	it, ok := s.store[id]
	return it, ok
}

func (s *VeoService) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, id)
}

// GenerateID creates a random hex id
func (s *VeoService) GenerateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// DecodedLen returns the exact decoded length for a base64 string
func (s *VeoService) DecodedLen(b64 string) int64 {
	l := len(b64)
	if l == 0 {
		return 0
	}
	pad := 0
	if l >= 1 && b64[l-1] == '=' {
		pad++
	}
	if l >= 2 && b64[l-2] == '=' {
		pad++
	}
	groups := l / 4
	return int64(groups*3 - pad)
}

// DecodeRange decodes only the minimal base64 segment that covers [start,end]
// and returns the exact bytes and the actual end offset served.
func (s *VeoService) DecodeRange(b64 string, start, end int64) ([]byte, int64, error) {
	if start < 0 {
		start = 0
	}
	total := s.DecodedLen(b64)
	if end >= total {
		end = total - 1
	}
	if start > end || start >= total {
		return nil, -1, fmt.Errorf("invalid range")
	}
	b64Len := len(b64)
	groupStart := start / 3
	groupEnd := end / 3
	b64Start := int(groupStart * 4)
	b64End := int((groupEnd + 1) * 4)
	if b64Start < 0 {
		b64Start = 0
	}
	if b64End > b64Len {
		b64End = b64Len
	}
	segment := b64[b64Start:b64End]
	dec, err := base64.StdEncoding.DecodeString(segment)
	if err != nil {
		return nil, -1, err
	}
	segOffset := int(start - (groupStart * 3))
	need := int(end-start) + 1
	if segOffset < 0 {
		segOffset = 0
	}
	if segOffset > len(dec) {
		segOffset = len(dec)
	}
	to := segOffset + need
	if to > len(dec) {
		to = len(dec)
	}
	payload := dec[segOffset:to]
	actualEnd := start + int64(len(payload)) - 1
	return payload, actualEnd, nil
}

// GenerateFirstVideoBase64 calls Veo and returns either base64+mime for the first video,
// or if base64 is unavailable, the raw bytes+mime via Files.Download.
func (s *VeoService) GenerateFirstVideoBase64(ctx context.Context, prompt string, img *genai.Image) (b64, mime string, fallbackBytes []byte, fallbackMime string, err error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: s.cfg.GoogleVeoAPIKey})
	if err != nil {
		return "", "", nil, "", fmt.Errorf("init veo client: %w", err)
	}
	op, err := client.Models.GenerateVideos(ctx, "veo-3.0-generate-preview", prompt, img, nil)
	if err != nil {
		return "", "", nil, "", fmt.Errorf("generate videos: %w", err)
	}
	for !op.Done {
		op, err = client.Operations.GetVideosOperation(ctx, op, nil)
		if err != nil {
			return "", "", nil, "", fmt.Errorf("poll operation: %w", err)
		}
	}
	if op.Response == nil || len(op.Response.GeneratedVideos) == 0 || op.Response.GeneratedVideos[0] == nil {
		return "", "", nil, "", fmt.Errorf("no video generated")
	}
	// Try to extract base64 via JSON (REST shape)
	raw, _ := json.Marshal(op)
	var generic map[string]any
	_ = json.Unmarshal(raw, &generic)
	// Capitalized path (SDK-internal)
	if resp, ok := generic["Response"].(map[string]any); ok {
		if vids, ok := resp["GeneratedVideos"].([]any); ok && len(vids) > 0 {
			if vid0, ok := vids[0].(map[string]any); ok {
				if videoObj, ok := vid0["Video"].(map[string]any); ok {
					if s, ok := videoObj["BytesBase64Encoded"].(string); ok {
						b64 = s
					}
					if mt, ok := videoObj["MIMEType"].(string); ok {
						mime = mt
					}
				}
			}
		}
	}
	// Lowercase REST path
	if b64 == "" {
		if resp, ok := generic["response"].(map[string]any); ok {
			if vids, ok := resp["videos"].([]any); ok && len(vids) > 0 {
				if vid0, ok := vids[0].(map[string]any); ok {
					if s, ok := vid0["bytesBase64Encoded"].(string); ok {
						b64 = s
					}
					if mt, ok := vid0["mimeType"].(string); ok {
						mime = mt
					}
				}
			}
		}
	}
	if b64 != "" {
		if strings.HasPrefix(b64, "data:") {
			if comma := strings.Index(b64, ","); comma != -1 {
				b64 = b64[comma+1:]
			}
		}
		if strings.TrimSpace(mime) == "" {
			mime = "video/mp4"
		}
		return b64, mime, nil, "", nil
	}
	// Fallback to download via Files API
	v := op.Response.GeneratedVideos[0]
	bytes, ferr := client.Files.Download(ctx, genai.NewDownloadURIFromGeneratedVideo(v), nil)
	if ferr != nil {
		return "", "", nil, "", fmt.Errorf("download failed: %w", ferr)
	}
	fmime := "video/mp4"
	if v.Video != nil && strings.TrimSpace(v.Video.MIMEType) != "" {
		fmime = v.Video.MIMEType
	}
	return "", "", bytes, fmime, nil
}
