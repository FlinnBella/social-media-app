package models

type VideoCompositionResponse struct {
	Metadata Metadata `json:"metadata"`
	Theme    struct {
		Style   string `json:"style"`
		Mood    string `json:"mood"`
		Grading string `json:"grading"`
	} `json:"theme"`

	Timeline Timeline `json:"timeline"`

	Audio struct {
		Narration struct {
			Voice TTSVoice `json:"voice"`
		} `json:"narration"`
		Music struct {
			Enabled bool     `json:"enabled"`
			TrackID string   `json:"trackId"`
			Genre   string   `json:"genre"`
			Mood    string   `json:"mood"`
			Volume  float64  `json:"volume"`
			FadeIn  *float64 `json:"fadeIn"`
			FadeOut *float64 `json:"fadeOut"`
		} `json:"music"`
	} `json:"audio"`
}

type Metadata struct {
	Resolution    []int  `json:"resolution"`
	TotalDuration int    `json:"totalDuration"`
	AspectRatio   string `json:"aspectRatio"`
	Fps           string `json:"fps"`
}

// New: item-level type for timeline array
type Timeline struct {
	TotalDuration int           `json:"totalDuration"`
	ImageTimeline ImageTimeline `json:"ImageTimeline"`
	TextTimeline  TextTimeline  `json:"TextTimeline"`
}

type ImageTimeline struct {
	ImageSegments []ImageSegment `json:"ImageSegments"`
}

type ImageSegment struct {
	Ordering   int                    `json:"ordering"`
	ImageIndex int                    `json:"imageIndex"`
	StartTime  int                    `json:"startTime"`
	Duration   int                    `json:"duration"`
	Transition TransitionTimelineItem `json:"Transition"`
}

type TransitionTimelineItem struct {
	Effect string `json:"effect"`
	Easing string `json:"easing"`
}

type TextTimeline struct {
	TextStyle    TextStyle     `json:"TextStyle"`
	TextSegments []TextSegment `json:"TextSegments"`
}

type TextSegment struct {
	ID              int     `json:"id"`
	Text            string  `json:"text"`
	StartTime       int     `json:"startTime"`
	Duration        int     `json:"duration"`
	Position        string  `json:"position"`
	NarrativeSource string  `json:"narrativeSource"`
	ImageRef        *string `json:"imageRef,omitempty"`
}

type TextStyle struct {
	FontFamily string `json:"fontFamily"`
	TextStyle  string `json:"textStyle"`
}

type TTSVoice struct {
	VoiceID   string  `json:"voiceId"`
	Emphasis  string  `json:"emphasis"`
	Speed     float64 `json:"speed"`
	Pitch     float64 `json:"pitch"`
	Stability float64 `json:"stability"`
}

// ===============================================
/*
OLD MODELS BELOW HERE

*/
// ===============================================

// New: narration script item to match JSON array
type NarrationScriptItem struct {
	Text   string `json:"text"`
	Timing *struct {
		Start int `json:"start"`
		End   int `json:"end"`
	} `json:"timing"`
	Emphasis string `json:"emphasis"`
}

type TextOverlay struct {
	Text     string `json:"text"`
	Position string `json:"position"`
	Theme    string `json:"theme"`
}
