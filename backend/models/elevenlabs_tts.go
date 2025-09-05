package models

type TTSSegment struct {
	Text     string `json:"text"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Emphasis string `json:"emphasis"`
}

type TTSInput struct {
	TextInput     []TextSegment `json:"textInput"`
	VoiceSettings TTSVoice      `json:"voiceSettings"`
}
