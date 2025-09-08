package models

//this is going down a bad road
import "social-media-ai-video/models/video_ffmpeg"

type TTSSegment struct {
	Text     string `json:"text"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Emphasis string `json:"emphasis"`
}

type TTSInput struct {
	TextInput     []video_ffmpeg.TextSegment `json:"textInput"`
	VoiceSettings video_ffmpeg.TTSVoice      `json:"voiceSettings"`
}
