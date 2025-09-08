package models

/*


These are the models used for the timeline composition

Partially re-used from the video composition for ffmpeg

*/

type TimelineComposition struct {
	Metadata Metadata `json:"metadata"`
	Theme    Theme    `json:"theme"`
	Timeline Timeline `json:"timeline"`
	Music    Music    `json:"music"`
}

type Metadata struct {
	Resolution    []int  `json:"resolution"`
	TotalDuration int    `json:"totalDuration"`
	AspectRatio   string `json:"aspectRatio"`
	Fps           string `json:"fps"`
}

type Theme struct {
	Style   string `json:"style"`
	Grading string `json:"grading"`
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
	Text            string `json:"text"`
	StartTime       int    `json:"startTime"`
	Duration        int    `json:"duration"`
	Position        string `json:"position"`
	NarrativeSource string `json:"narrativeSource"`
}

type TextStyle struct {
	FontFamily string `json:"fontFamily"`
	TextStyle  string `json:"textStyle"`
}

type Music struct {
	Enabled bool    `json:"enabled"`
	Genre   string  `json:"genre"`
	Volume  float64 `json:"volume"`
}
