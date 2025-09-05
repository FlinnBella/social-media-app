package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	models "social-media-ai-video/models"
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

type Metadata_FFmpeg struct {
	TotalDuration int
	AspectRatio   string
	FPS           int
	Width         int
	Height        int
}

// CommandBuildInput contains everything needed to construct ffmpeg args

type CommandBuildInput struct {
	Metadata_FFmpeg Metadata_FFmpeg
	Timeline        models.Timeline
	// Images referenced by index in timeline (ImageIndex)
	ImagePaths []string
	// Audio assets
	Audio AudioConfig
	// Output file path (absolute or working-directory relative)
	OutputPath string
}

// FFmpegCommandBuilder converts a high-level composition into a single ffmpeg command

type FFmpegCommandBuilder struct{}

func NewFFmpegCommandBuilder() *FFmpegCommandBuilder { return &FFmpegCommandBuilder{} }

// High-level compiler orchestrating schema -> assets -> args
// Accepts the ai-generated composition JSON and services to resolve audio assets.

type CompositionCompiler struct {
	builder      *FFmpegCommandBuilder
	bgMusic      *BackgroundMusic
	voiceService *ElevenLabsService
}

//Can see the compiler takes the music and voice services; all-in-one stop

func NewCompositionCompiler(builder *FFmpegCommandBuilder, bg *BackgroundMusic, els *ElevenLabsService) *CompositionCompiler {
	return &CompositionCompiler{builder: builder, bgMusic: bg, voiceService: els}
}

type Compilier interface {
	Compile(jsonAISchemaBlob []byte, imagePaths []string) ([]string, []string, string, error)
}

// Compile takes the AI JSON blob and image paths (ordered by index) and returns ffmpeg args and resolved output paths used.
func (cc *CompositionCompiler) Compile(jsonAISchemaBlob []byte, imagePaths []string) ([]string, []string, string, error) {
	//schema object
	var vc models.VideoCompositionResponse

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
	meta := Metadata_FFmpeg{
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
	if cc.voiceService != nil {
		// Ensure a tmp dir for TTS
		ttsDir := filepath.Join(os.TempDir(), "tts_audio")
		if err := os.MkdirAll(ttsDir, 0o755); err != nil {
			return nil, nil, "", fmt.Errorf("failed to create tts tmp dir: %v", err)
		}
		filenames, fileoutputmap, err := cc.voiceService.GenerateSpeechToTmp(ttsInput, ttsDir)
		if err != nil {
			return nil, nil, "", fmt.Errorf("tts generation failed: %v", err)
		}

		ttsNarrationPathsMap = fileoutputmap
		ttsNarrationPaths = filenames

	}

	// Resolve music if enabled
	musicPath := ""
	musicName := ""

	if vc.Audio.Music.Enabled && cc.bgMusic != nil {
		mf, err := cc.bgMusic.CreateBackgroundMusic(vc.Audio.Music.Mood, vc.Audio.Music.Genre)
		if err != nil {
			return nil, nil, "", fmt.Errorf("bgm download failed: %v", err)
		}
		musicPath = mf.FilePath
		musicName = mf.FileName
	}

	// Auto-generate an output path under the OS temp directory
	autoOutput := filepath.Join(os.TempDir(), fmt.Sprintf("short_%d.mp4", time.Now().UnixNano()))

	args, err := cc.builder.Build(CommandBuildInput{
		Metadata_FFmpeg: meta,
		Timeline:        vc.Timeline,
		ImagePaths:      imagePaths,
		Audio: AudioConfig{
			ttsNarrationPaths: ttsNarartionFiles{FilePath: ttsNarrationPathsMap, FileName: ttsNarrationPaths},
			MusicEnabled:      vc.Audio.Music.Enabled,
			MusicPath:         MusicFiles{MusicPath: musicPath, MusicName: musicName},
			MusicVolume:       vc.Audio.Music.Volume,
			NarrationVolume:   1.0,
		},
		OutputPath: autoOutput,
	})
	if err != nil {
		return nil, []string{}, "", err
	}
	return args, ttsNarrationPaths, autoOutput, nil
}

func (b *FFmpegCommandBuilder) Build(in CommandBuildInput) ([]string, error) {
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
	sorted := make([]models.ImageSegment, len(in.Timeline.ImageTimeline.ImageSegments))
	copy(sorted, in.Timeline.ImageTimeline.ImageSegments)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StartTime < sorted[j].StartTime })

	// Input list: images + audio(s)
	args := []string{"-y"}

	// Image inputs (each once); we will reference by indices
	for _, p := range in.ImagePaths {
		// Images as looping inputs turned into video in filters using loop filter
		args = append(args, "-i", p)
	}

	// Audio inputs appended at the end so index math is predictable
	numImageInputs := len(in.ImagePaths)
	audioInputStart := numImageInputs
	var narrationPath []string
	for i := 0; i < len(in.Audio.ttsNarrationPaths.FileName); i++ {
		narrationPath = append(narrationPath, in.Audio.ttsNarrationPaths.FilePath[in.Audio.ttsNarrationPaths.FileName[i]])
		args = append(args, "-i", narrationPath[i])
	}
	musicIdx := -1
	narrIdx := -1
	if len(narrationPath) > 0 {
		narrIdx = audioInputStart
		if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" {
			musicIdx = audioInputStart + 1
		}
	} else if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" {
		// Only music
		args = append(args, "-i", in.Audio.MusicPath.MusicPath)
		musicIdx = audioInputStart
	}
	if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" && musicIdx == -1 {
		// Music not yet added (narration present handled earlier). Add now.
		args = append(args, "-i", in.Audio.MusicPath.MusicPath)
		if narrIdx >= 0 {
			musicIdx = narrIdx + 1
		} else {
			musicIdx = audioInputStart
		}
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
		// We'll use: scale=W:H:force_original_aspect_ratio=decrease, pad=W:H:(ow-iw)/2:(oh-ih)/2,format=yuv420p
		filter += fmt.Sprintf("%s scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=yuv420p,fps=%d,trim=duration=%f,setpts=PTS-STARTPTS %s;",
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
		xy := positionXY(textsegments[textIdx].Position, in.Metadata_FFmpeg.Width, in.Metadata_FFmpeg.Height)
		bg := ""
		/*
			if textsegments[textIdx].BackgroundColor != "" {
				bg = fmt.Sprintf(":box=1:boxcolor=%s@0.6", textsegments[textIdx].BackgroundColor)
			}
		*/
		labelOut := fmt.Sprintf("[vtx%d]", textIdx)
		enable := fmt.Sprintf("enable='between(t,%.3f,%.3f)'", float64(t.StartTime), float64(t.StartTime+t.Duration))
		filter += fmt.Sprintf("%s drawtext=text=%s:fontcolor=%s:fontsize=%d:x=%s:y=%s%s:%s %s;",
			videoLabel, text, colorOrDefault("FFFFFF"), maxInt(0, 12), xy[0], xy[1], bg, enable, labelOut)
		videoLabel = labelOut
		textIdx++
	}

	// Label final video stream
	finalVideoLabel := videoLabel
	if finalVideoLabel == "[basev]" {
		filter += "[basev]copy[vout];"
		finalVideoLabel = "[vout]"
	}

	// Audio mixing
	audioMap := ""
	if narrIdx >= 0 && musicIdx >= 0 {
		mv := clamp01(in.Audio.MusicVolume)
		nv := in.Audio.NarrationVolume
		if nv <= 0 {
			nv = 1.0
		}
		filter += fmt.Sprintf("[%d:a]volume=%0.2f[na];", narrIdx, nv)
		filter += fmt.Sprintf("[%d:a]volume=%0.2f[ma];", musicIdx, mv)
		filter += "[na][ma]amix=inputs=2:duration=first:dropout_transition=2[aout];"
		audioMap = "[aout]"
	} else if narrIdx >= 0 {
		filter += fmt.Sprintf("[%d:a]volume=%0.2f[aout];", narrIdx, maxFloat(in.Audio.NarrationVolume, 1.0))
		audioMap = "[aout]"
	} else if musicIdx >= 0 {
		filter += fmt.Sprintf("[%d:a]volume=%0.2f[aout];", musicIdx, clamp01(in.Audio.MusicVolume))
		audioMap = "[aout]"
	}

	args = append(args,
		"-filter_complex", filter,
		"-map", finalVideoLabel,
	)
	if audioMap != "" {
		args = append(args, "-map", audioMap)
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
	if audioMap != "" {
		args = append(args, "-c:a", "aac")
	}

	// Ensure directory exists is caller's job; we only reference the path
	args = append(args, filepath.Clean(in.OutputPath))
	return args, nil
}

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
	default: // center
		return [2]string{"(w-tw)/2", "(h-th)/2"}
	}
}
