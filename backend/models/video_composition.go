package models

type VideoCompositionResponse struct {
	Properties struct {
		Metadata struct {
			Resolution    []int  `json:"resolution"`
			TotalDuration int    `json:"totalDuration"`
			AspectRatio   string `json:"aspectRatio"`
			Fps           string `json:"fps"`
		} `json:"properties"`
	} `json:"metadata"`
	Theme struct {
		Style        string            `json:"style"`
		Mood         string            `json:"mood"`
		ColorPalette map[string]string `json:"colorPalette"`
		Typography   map[string]string `json:"typography"`
	} `json:"theme"`
	Narrative struct {
		Hook  string   `json:"hook"`
		Story []string `json:"story"`
		Cta   string   `json:"cta"`
		Tone  string   `json:"tone"`
	} `json:"narrative"`

	Timeline []TimelineItem `json:"timeline"`

	Audio struct {
		Narration struct {
			Script   []NarrationScriptItem `json:"script"`
			Emphasis string                `json:"emphasis"`
			Voice    TTSVoice              `json:"voice"`
		} `json:"narration"`
		Music struct {
			Enabled bool    `json:"enabled"`
			TrackID string  `json:"trackId"`
			Genre   string  `json:"genre"`
			Mood    string  `json:"mood"`
			Volume  float64 `json:"volume"`
			FadeIn  float64 `json:"fadeIn"`
			FadeOut float64 `json:"fadeOut"`
		} `json:"music"`
	} `json:"audio"`
}

type ImageSegment struct {
	ImageIndex int    `json:"imageIndex"`
	Theme      string `json:"theme"`
}

type TextOverlay struct {
	Text     string `json:"text"`
	Position string `json:"position"`
	Theme    string `json:"theme"`
}

type TransitionTimelineItem struct {
	Effect string `json:"effect"`
	Easing string `json:"easing"`
}

// New: item-level type for timeline array
type TimelineItem struct {
	ID        string `json:"id"`
	StartTime int    `json:"startTime"`
	Duration  int    `json:"duration"`
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`

	ImageSegment *ImageSegment           `json:"imageSegment,omitempty"`
	TextOverlay  *TextOverlay            `json:"textOverlay,omitempty"`
	Transition   *TransitionTimelineItem `json:"transition,omitempty"`
}

// New: narration script item to match JSON array
type NarrationScriptItem struct {
	Text   string `json:"text"`
	Timing *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"timing"`
	Emphasis string `json:"emphasis"`
}
