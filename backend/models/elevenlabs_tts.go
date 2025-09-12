package models

//this is going down a bad road

type TTSSegment struct {
	Text     string `json:"text"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Emphasis string `json:"emphasis"`
}

type TTSInput struct {
	TextInput     []string `json:"textInput"`
	VoiceSettings *struct {
		VoiceID   string  `json:"voiceId"`
		Emphasis  string  `json:"emphasis"`
		Speed     float64 `json:"speed"`
		Pitch     float64 `json:"pitch"`
		Stability float64 `json:"stability"`
	} `json:"voiceSettings"`
}
