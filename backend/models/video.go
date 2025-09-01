package models

// VideoSource selects which generator pipeline to use
type VideoSource string

const (
	VideoSourceReels  VideoSource = "reels"
	VideoSourcePexels VideoSource = "pexels"
)

type VideoGenerationRequest struct {
	Prompt     string      `form:"prompt"`
	Images     [][]byte    `form:"image"`
	ImageNames []string    `form:"image_name"`
	Source     VideoSource `form:"source"`
}

type VideoGenerationResponse struct {
	VideoURL string `json:"videoUrl,omitempty"`
	Error    string `json:"error,omitempty"`
	Status   string `json:"status"`
}

type VideoCompositionRequest struct {
	VideoLength     float64         `json:"video_length"`
	AspectRatio     string          `json:"aspect_ratio"`
	Resolution      Resolution      `json:"resolution"`
	BackgroundMusic BackgroundMusic `json:"background_music"`
	TTSConfig       TTSConfig       `json:"tts_config"`
	VideoSegments   []VideoSegment  `json:"video_segments"`
}

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type BackgroundMusic struct {
	Enabled bool    `json:"enabled"`
	Volume  float64 `json:"volume"`
	Genre   string  `json:"genre"`
}

type TTSConfig struct {
	Script  string  `json:"script"`
	VoiceID string  `json:"voice_id"`
	Speed   float64 `json:"speed"`
	Volume  float64 `json:"volume"`
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
