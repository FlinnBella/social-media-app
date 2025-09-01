package models

type TTSVoice struct {
	VoiceID   string  `json:"voiceId"`
	Speed     float64 `json:"speed"`
	Pitch     float64 `json:"pitch"`
	Stability float64 `json:"stability"`
}

type TTSSegment struct {
	Text     string `json:"text"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Emphasis string `json:"emphasis"`
}

type NarrationData struct {
	Script []TTSSegment `json:"script"`
	Voice  TTSVoice     `json:"voice"`
}

type NarrativeData struct {
	Hook  string   `json:"hook"`
	Story []string `json:"story"`
	Cta   string   `json:"cta"`
	Tone  string   `json:"tone"`
}

type TTSInput struct {
	Narrative NarrativeData `json:"narrative"`
	Narration NarrationData `json:"narration"`
}
