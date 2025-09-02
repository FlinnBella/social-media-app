package models

import "encoding/json"

type CompositionMetadataInstance struct {
	TotalDuration float64 `json:"totalDuration"`
	AspectRatio   string  `json:"aspectRatio"`
	FPS           int     `json:"fps"`
	Resolution    []int   `json:"resolution"` // [width, height]
}

type ImageSegmentInstance struct {
	ImageIndex int `json:"imageIndex"`
	// Animation and Crop are currently ignored for MVP synthesis
}

type TextStyleInstance struct {
	FontSize        float64 `json:"fontSize"`
	Color           string  `json:"color"`
	BackgroundColor string  `json:"backgroundColor"`
	Animation       string  `json:"animation"`
}

type TextOverlayInstance struct {
	Text     string             `json:"text"`
	Position string             `json:"position"`
	Style    *TextStyleInstance `json:"style"`
}

type TransitionInstance struct {
	Effect string `json:"effect"`
	Easing string `json:"easing"`
}

type TimelineItemInstance struct {
	ID        string          `json:"id"`
	StartTime float64         `json:"startTime"`
	Duration  float64         `json:"duration"`
	Type      string          `json:"type"`
	Content   json.RawMessage `json:"content"`
}

type NarrationVoiceInstance struct {
	VoiceID   string  `json:"voiceId"`
	Speed     float64 `json:"speed"`
	Pitch     float64 `json:"pitch"`
	Stability float64 `json:"stability"`
}

type NarrationScriptTiming struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type NarrationScriptItem struct {
	Text     string                 `json:"text"`
	Timing   *NarrationScriptTiming `json:"timing"`
	Emphasis string                 `json:"emphasis"`
}

type NarrationInstanceFull struct {
	Script []NarrationScriptItem  `json:"script"`
	Voice  NarrationVoiceInstance `json:"voice"`
}

type MusicInstance struct {
	Enabled bool    `json:"enabled"`
	TrackID string  `json:"trackId"`
	Genre   string  `json:"genre"`
	Mood    string  `json:"mood"`
	Volume  float64 `json:"volume"`
}

type AudioInstanceFull struct {
	Narration NarrationInstanceFull `json:"narration"`
	Music     MusicInstance         `json:"music"`
}

type NarrativeInstanceFull struct {
	Hook  string   `json:"hook"`
	Story []string `json:"story"`
	Cta   string   `json:"cta"`
	Tone  string   `json:"tone"`
}

type VideoCompositionInstance struct {
	Metadata  CompositionMetadataInstance `json:"metadata"`
	Theme     map[string]any              `json:"theme"`
	Narrative NarrativeInstanceFull       `json:"narrative"`
	Timeline  []TimelineItemInstance      `json:"timeline"`
	Audio     AudioInstanceFull           `json:"audio"`
}
