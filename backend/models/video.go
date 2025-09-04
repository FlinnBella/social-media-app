package models

// VideoSource selects which generator pipeline to use
type VideoSource string

const (
	VideoSourceReels  VideoSource = "reels"
	VideoSourcePexels VideoSource = "pexels"
)

// VideoGenerationRequest carries prompt and images for schema generation
type VideoGenerationRequest struct {
	Prompt     string      `form:"prompt"`
	Images     [][]byte    `form:"image"`
	ImageNames []string    `form:"image_name"`
	Source     VideoSource `form:"source"`
}

type VideoSegment struct {
	ID             string     `json:"id"`
	PexelsVideoURL string     `json:"pexels_video_url"`
	StartTime      float64    `json:"start_time"`
	Duration       float64    `json:"duration"`
	Transition     Transition `json:"transition"`
	Effects        Effects    `json:"effects"`
}

type Transition struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

type Effects struct {
	Crop  *CropEffect `json:"crop,omitempty"`
	Zoom  float64     `json:"zoom"`
	Speed float64     `json:"speed"`
}

type CropEffect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}
