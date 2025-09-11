package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"social-media-ai-video/models"
	"social-media-ai-video/models/video_ffmpeg"
	"sort"
	"strconv"
	"time"
)

// Command builder for generating a single ffmpeg invocation from a high-level composition
// This purposefully produces ffmpeg args ([]string) so the caller can run:
//   exec.Command("ffmpeg", args...)
// The builder is pure and does not perform any IO.

// CompositionProperties.Metadata.Properties captures global video settings
// Width/Height must match the selected aspect ratio
// FPS currently limited to 24 or 30 in the schema

// going to need to revamp entire struct models; schema is just for ai, not for mapping

// AudioConfig holds prepared audio assets

/*
CODEBASE IS INSANELY CLUTTERED,
NEED TO IMPLEMENT METHODS TO CLEAN IT UP IMMEDIATELY

STEPS:
	- TRY TO CUT DOWN ON NUMBER OF STRUCTS
	- TRY TO CLEAN UP INTERFACES AND NEWFUNCTIONS AND THINGS
	- PORT THE COMPILERS INTO SOMETHING MORE USEABLE, AND BREAK IT UP
	- IN IMPLS FOR THE PRO SERVICE AND THE REELS SERVICE

*/

type AudioConfig struct {
	ttsNarrationPaths ttsNarartionFiles // path to narration (mp3/m4a)
	MusicEnabled      bool
	MusicPath         MusicFiles
	MusicVolume       float64 // 0..1
	NarrationVolume   float64 // 0..1, if 0 treat as 1.0
}

type ttsNarartionFiles struct {
	FilePath map[string]string
	FileName []string
}

type MusicFiles struct {
	MusicPath string
	MusicName string
}

//AudioConfig end

// ffmpeg
type Metadata_Universal_FFmpeg struct {
	TotalDuration int
	AspectRatio   string
	FPS           int
	Width         int
	Height        int
}

// CommandBuildInput contains everything needed to construct ffmpeg args

type CommandBuildInput struct {
	Metadata_FFmpeg Metadata_Universal_FFmpeg
	Timeline        video_ffmpeg.Timeline
	// Images referenced by index in timeline (ImageIndex)
	ImagePaths []string
	// Audio assets
	Audio AudioConfig
	// Output file path (absolute or working-directory relative)
	OutputPath string
}

// VideoCompiler defines the interface for video compilation strategies
type VideoCompiler interface {
	Compile(jsonAISchemaBlob []byte, imagePaths []string) ([]string, []string, string, error)
	Build(in CommandBuildInput) ([]string, error)
}

//Compiler structs

// ReelsCompiler - standard video compilation for social media reels
type ReelsCompiler struct {
	bgMusic      BackgroundMusic
	voiceService ElevenLabsService
}

func NewReelsCompiler(bg *BackgroundMusic, els *ElevenLabsService) *ReelsCompiler {
	return &ReelsCompiler{bgMusic: *bg, voiceService: *els}
}

// ProCompiler - high-quality video compilation with different ffmpeg strategy
type ProCompiler struct {
	bgMusic      BackgroundMusic
	voiceService ElevenLabsService
}

func NewProCompiler(bg *BackgroundMusic, els *ElevenLabsService) *ProCompiler {
	return &ProCompiler{bgMusic: *bg, voiceService: *els}
}

// Compile takes the AI JSON blob and image paths (ordered by index) and returns ffmpeg args and resolved output paths used.
func (rc *ReelsCompiler) Compile(jsonAISchemaBlob []byte, imagePaths []string) ([]string, []string, string, error) {
	//schema object
	var vc video_ffmpeg.VideoCompositionResponse

	// unwrap optional top-level {"output": ...} wrapper if present
	var outer struct {
		Output json.RawMessage `json:"output"`
	}
	if err := json.Unmarshal(jsonAISchemaBlob, &outer); err == nil && len(outer.Output) > 0 {
		jsonAISchemaBlob = outer.Output
	}

	//jsonAISchemaBlob should conform to schema, place in vc
	if err := json.Unmarshal(jsonAISchemaBlob, &vc); err != nil {
		return nil, []string{}, "", fmt.Errorf("invalid composition json: %v. Given json: %s", err, string(jsonAISchemaBlob))
	}

	// Map Properties.Metadata.Properties
	if len(vc.Metadata.Resolution) != 2 {
		return nil, []string{}, "", fmt.Errorf("invalid resolution resolution array %v", vc.Metadata.Resolution)
	}
	fps := 30
	if vc.Metadata.Fps != "" {
		if parsed, err := strconv.Atoi(vc.Metadata.Fps); err == nil {
			fps = parsed
		}
	}
	meta := Metadata_Universal_FFmpeg{
		TotalDuration: vc.Metadata.TotalDuration,
		AspectRatio:   vc.Metadata.AspectRatio,
		FPS:           fps,
		Width:         vc.Metadata.Resolution[0],
		Height:        vc.Metadata.Resolution[1],
	}

	// Resolve narration via ElevenLabs
	ttsInput := models.TTSInput{
		TextInput:     vc.Timeline.TextTimeline.TextSegments,
		VoiceSettings: vc.Audio.Narration.Voice,
	}

	ttsNarrationPathsMap := map[string]string{}
	ttsNarrationPaths := []string{}

	//Generate tts narration elevenlabs
	if &rc.voiceService != nil {
		// Ensure a tmp dir for TTS
		ttsDir := filepath.Join(os.TempDir(), "tts_audio")
		if err := os.MkdirAll(ttsDir, 0o755); err != nil {
			return nil, nil, "", fmt.Errorf("failed to create tts tmp dir: %v", err)
		}
		filenames, fileoutputmap, err := rc.voiceService.GenerateSpeechToTmp(ttsInput, ttsDir)
		if err != nil {
			return nil, nil, "", fmt.Errorf("tts generation failed: %v", err)
		}

		ttsNarrationPathsMap = fileoutputmap
		ttsNarrationPaths = filenames

	}

	// Resolve music if enabled
	musicPath := ""
	musicName := ""

	if vc.Audio.Music.Enabled && &rc.bgMusic != nil {
		mf, err := rc.bgMusic.CreateBackgroundMusic(vc.Audio.Music.Mood, vc.Audio.Music.Genre)
		if err != nil {
			return nil, nil, "", fmt.Errorf("bgm download failed: %v", err)
		}
		musicPath = mf.FilePath
		musicName = mf.FileName
	}

	// Auto-generate an output path under the OS temp directory
	autoOutput := filepath.Join(os.TempDir(), fmt.Sprintf("short_%d.mp4", time.Now().UnixNano()))

	//seperate imps for pro and reels builders
	args, err := rc.Build(in)

	return args, ttsNarrationPaths, autoOutput, nil
}

func (rc *ReelsCompiler) Build(in CommandBuildInput) ([]string, error) {
	if in.Metadata_FFmpeg.Width <= 0 || in.Metadata_FFmpeg.Height <= 0 || in.Metadata_FFmpeg.FPS <= 0 {
		return nil, fmt.Errorf("invalid metadata: width/height/fps must be > 0")
	}
	if in.OutputPath == "" {
		return nil, fmt.Errorf("missing output path")
	}

	// Validate image input paths exist
	for _, p := range in.ImagePaths {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("missing input file: %s: %v", p, err)
		}
	}

	// Validate image indices
	for _, t := range in.Timeline.ImageTimeline.ImageSegments {
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImagePaths) {
			return nil, fmt.Errorf("image item %d references invalid image index", t.ImageIndex)
		}

	}

	// Sort timeline by start time to build proper concat order
	sorted := make([]video_ffmpeg.ImageSegment, len(in.Timeline.ImageTimeline.ImageSegments))
	copy(sorted, in.Timeline.ImageTimeline.ImageSegments)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StartTime < sorted[j].StartTime })

	// Input list: images + audio(s)
	/*
	   # FFMPEG CMD STARTS HERE
	*/
	args := []string{"-y"}

	// Image inputs (each once); we will reference by indices
	for _, p := range in.ImagePaths {
		// APPEND IMAGES AS INPUTS INTO FFMPEG - FIRST INPUT APPEND
		args = append(args, "-i", p)
	}

	// Audio inputs appended at the end so index math is predictable
	numImageInputs := len(in.ImagePaths)
	audioInputStart := numImageInputs
	// Validate narration and music file paths before adding as inputs

	musicIdx := -1
	narrIdx := -1

	// guard against invalid tts directory, or filename
	for i := 0; i < len(in.Audio.ttsNarrationPaths.FileName); i++ {
		fn := in.Audio.ttsNarrationPaths.FileName[i]
		p := in.Audio.ttsNarrationPaths.FilePath[fn]
		if p == "" {
			return nil, fmt.Errorf("missing narration path for %s", fn)
		}
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("missing narration file: %s: %v", p, err)
		}
	}
	//end guard

	// guard against invalid music file path, or directory
	if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" {
		if _, err := os.Stat(in.Audio.MusicPath.MusicPath); err != nil {
			return nil, fmt.Errorf("missing music file: %s: %v", in.Audio.MusicPath.MusicPath, err)
		}
	}
	for i := 0; i < len(in.Audio.ttsNarrationPaths.FileName); i++ {
		args = append(args, "-i", in.Audio.ttsNarrationPaths.FilePath[in.Audio.ttsNarrationPaths.FileName[i]])

		//init elevenlabs tts file(s)
		narrIdx = audioInputStart
	}
	if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" && musicIdx == -1 {
		// Music not yet added (narration present handled earlier). Add now.
		args = append(args, "-i", in.Audio.MusicPath.MusicPath)
	} else {
		musicIdx = audioInputStart + len(in.Audio.ttsNarrationPaths.FileName) - 1
	}

	// Build filter_complex
	filter := ""

	// For each image timeline item, construct a stream that lasts its duration
	// We map image input index -> variable label like [imgN]
	imageStreamCount := 0
	for idx, t := range sorted {
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImagePaths) {
			continue
		}
		imgInputIdx := t.ImageIndex
		labelIn := fmt.Sprintf("[%d:v]", imgInputIdx)
		labelOut := fmt.Sprintf("[seg%d]", idx)
		// scale to canvas, pad/crop, set fps, set duration
		// Use: scale, pad (if needed), fps, tpad=stop_mode=clone:stop_duration=duration, setpts=N/(FPS*TB)
		// Use loop filter to make image into frames: loop=loop=FPS*duration:size=1:start=0, fps=FPS
		// Use tpad to clone last frame to desired duration for still images, then normalize PTS
		filter += fmt.Sprintf("%s scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=yuv420p,fps=%d,tpad=stop_mode=clone:stop_duration=%f,setpts=PTS-STARTPTS %s;",
			labelIn, in.Metadata_FFmpeg.Width, in.Metadata_FFmpeg.Height, in.Metadata_FFmpeg.Width, in.Metadata_FFmpeg.Height, in.Metadata_FFmpeg.FPS, float64(t.Duration), labelOut)
		imageStreamCount++
	}

	// Concatenate all video segments in order
	concatInputs := ""
	for idx, t := range sorted {
		if t.ImageIndex >= 0 && t.ImageIndex < len(in.ImagePaths) {
			concatInputs += fmt.Sprintf("[seg%d]", idx)
		}
	}
	if concatInputs == "" {
		return nil, fmt.Errorf("no visual segments present")
	}
	filter += fmt.Sprintf("%s concat=n=%d:v=1:a=0[basev];", concatInputs, imageStreamCount)

	// Apply text overlays with enable between(t, start, end)
	videoLabel := "[basev]"
	textIdx := 0
	textsegments := in.Timeline.TextTimeline.TextSegments
	for _, t := range sorted {
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImagePaths) {
			continue
		}
		if textIdx >= len(textsegments) {
			break
		}
		text := escapeDrawtext(textsegments[textIdx].Text)
		//textsegments[textIdx].Position
		xy := positionXY("center-bottom", in.Metadata_FFmpeg.Width, in.Metadata_FFmpeg.Height)
		bg := ""
		/*
			if textsegments[textIdx].BackgroundColor != "" {
				bg = fmt.Sprintf(":box=1:boxcolor=%s@0.6", textsegments[textIdx].BackgroundColor)
			}
		*/
		labelOut := fmt.Sprintf("[vtx%d]", textIdx)
		enable := fmt.Sprintf("enable='between(t,%.3f,%.3f)'", float64(t.StartTime), float64(t.StartTime+t.Duration))
		filter += fmt.Sprintf("%s drawtext=text=%s:fontcolor=white:borderw=%d:bordercolor=%s:fontsize=%d:x=%s:y=%s%s:%s %s;",
			videoLabel, text, 2, "black", maxInt(0, 24), xy[0], xy[1], bg, enable, labelOut)
		videoLabel = labelOut
		textIdx++
	}

	// Label final video stream
	finalVideoLabel := videoLabel
	if finalVideoLabel == "[basev]" {
		filter += "[basev]copy[vout];"
		finalVideoLabel = "[vout]"
	}

	// Audio processing: build modular stems [ma] (music) and [na] (narration), then create mixed [mixa]
	// Music stem
	if musicIdx >= 0 {
		mv := clamp01(in.Audio.MusicVolume)
		filter += fmt.Sprintf("[%d:a]volume=%0.2f,atrim=0:%f,asetpts=PTS-STARTPTS[ma];", musicIdx, mv, float64(in.Metadata_FFmpeg.TotalDuration))
	}
	// Narration stem
	nv := in.Audio.NarrationVolume
	if nv <= 0 {
		nv = 1.0
	}
	filter += fmt.Sprintf("[%d:a]volume=%0.2f,apad,atrim=0:%f,asetpts=PTS-STARTPTS[na];", narrIdx, nv, float64(in.Metadata_FFmpeg.TotalDuration))
	// Mixed overlay track
	var audioMaps []string
	if musicIdx >= 0 {
		filter += "[na][ma]amix=inputs=2:duration=longest:dropout_transition=2[mixa];"
		audioMaps = []string{"[mixa]"}
	} else {
		// No music present; fall back to narration only
		audioMaps = []string{"[na]"}
	}

	args = append(args,
		"-filter_complex", filter,
		"-map", finalVideoLabel,
	)
	if len(audioMaps) > 0 {
		for _, am := range audioMaps {
			args = append(args, "-map", am)
		}
	} else {
		args = append(args, "-an")
	}

	// Output settings
	args = append(args,
		"-r", fmt.Sprintf("%d", in.Metadata_FFmpeg.FPS),
		"-s", fmt.Sprintf("%dx%d", in.Metadata_FFmpeg.Width, in.Metadata_FFmpeg.Height),
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-preset", "fast",
		"-crf", "23",
	)
	if len(audioMaps) > 0 {
		args = append(args, "-c:a", "aac")
	}

	// Ensure directory exists is caller's job; we only reference the path
	args = append(args, filepath.Clean(in.OutputPath))
	return args, nil
}

// Reels Compiler and Compile both have
// Compile implements VideoCompiler interface for ReelsCompiler

// Compile implements VideoCompiler interface for ProCompiler
// Pro-specific compilation: high quality, videos, professional processing
func (pc *ProCompiler) Compile(jsonAISchemaBlob []byte, imagePaths []string) ([]string, []string, string, error) {
	// Pro-specific schema processing
	// TODO: Parse Pro-specific JSON schema
	// TODO: Process videos for high-quality output
	// TODO: Generate Pro-specific FFmpeg args (different codecs, quality settings)

	// This actually does the building here
	//seperate imps for pro and reels builders
	args, err := pc.Build(in)

	//this returns the stuff back
	return args, nil, "", err
}

func (pc *ProCompiler) Build(in CommandBuildInput) ([]string, error) {
	return pc.Build(in)
}

/*
FFMPEG HELPER FUNCTIONS
*/
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func maxInt(v int, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func maxFloat(v float64, def float64) float64 {
	if v <= 0 {
		return def
	}
	return v
}

func colorOrDefault(hex string) string {
	if hex == "" {
		return "white"
	}
	return hex
}

// escapeDrawtext escapes characters for ffmpeg drawtext
func escapeDrawtext(s string) string {
	// Basic escaping: colon, backslash, quotes
	esc := s
	esc = replaceAll(esc, "\\", "\\\\")
	esc = replaceAll(esc, ":", "\\:")
	// Avoid using \' which some linters flag; build as "\\" + "'"
	esc = replaceAll(esc, "'", "\\"+"'")
	return esc
}

func replaceAll(s, old, new string) string {
	for {
		idx := -1
		for i := 0; i+len(old) <= len(s); i++ {
			if s[i:i+len(old)] == old {
				idx = i
				break
			}
		}
		if idx < 0 {
			return s
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
}

// positionXY maps semantic positions to x/y expressions usable by ffmpeg drawtext
func positionXY(pos string, w, h int) [2]string {
	switch pos {
	case "center-left":
		return [2]string{"(w-tw)/4", "(h-th)/2"}
	case "center-right":
		return [2]string{"3*(w-tw)/4", "(h-th)/2"}
	case "center-bottom":
		return [2]string{"(w-tw)/2", "h-th-20"}
	default: // center
		return [2]string{"(w-tw)/2", "(h-th)/2"}
	}
}
