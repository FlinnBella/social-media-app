package models

type NarrativeInstance struct {
	Hook  string   `json:"hook"`
	Story []string `json:"story"`
	Cta   string   `json:"cta"`
	Tone  string   `json:"tone"`
}

type NarrationInstance struct {
	Script []TTSSegment `json:"script"`
	Voice  TTSVoice     `json:"voice"`
}

type CompositionAudio struct {
	Narration NarrationInstance `json:"narration"`
}

type CompositionDocument struct {
	Narrative NarrativeInstance `json:"narrative"`
	Audio     CompositionAudio  `json:"audio"`
}
