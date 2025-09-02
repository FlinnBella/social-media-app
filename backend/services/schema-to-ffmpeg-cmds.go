package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"social-media-ai-video/models"
	"sort"
)

// Command builder for generating a single ffmpeg invocation from a high-level composition
// This purposefully produces ffmpeg args ([]string) so the caller can run:
//   exec.Command("ffmpeg", args...)
// The builder is pure and does not perform any IO.

type AspectRatio string

const (
	AspectRatio9x16 AspectRatio = "9:16"
	AspectRatio1x1  AspectRatio = "1:1"
)

// CompositionProperties.Metadata.Properties captures global video settings
// Width/Height must match the selected aspect ratio
// FPS currently limited to 24 or 30 in the schema

type Metadata struct {
	TotalDuration float64
	AspectRatio   AspectRatio
	FPS           int
	Width         int
	Height        int
}

// TimelineItemType discriminates the content kind per timeline entry

type TimelineItemType string

const (
	TimelineItemImage       TimelineItemType = "image"
	TimelineItemTextOverlay TimelineItemType = "text_overlay"
	TimelineItemTransition  TimelineItemType = "transition"
)

// ImageSegmentContent references an input image by index and optional transforms

type ImageSegmentContent struct {
	ImageIndex int
	// Animation and Crop omitted for now in ffmpeg synthesis; can be added later
}

// TextOverlayContent defines drawtext overlay

type TextOverlayContent struct {
	Text string
	// position: center-left | center-right | center
	Position string
	// style
	FontSize        int
	ColorHex        string // #RRGGBB
	BackgroundColor string // #RRGGBB
	Animation       string // fade_in | slide_in | type_on | bounce | none
}

// TransitionContent defines inter-clip transition

type TransitionContent struct {
	Effect string // fade | dissolve | slide | zoom | cut
	Easing string // linear | ease-in | ease-out | ease-in-out
}

// TimelineItem is a single vcruction on the timeline

type TimelineItem struct {
	ID        string
	StartTime float64
	Duration  float64
	Type      TimelineItemType
	Image     *ImageSegmentContent
	Text      *TextOverlayContent
	Trans     *TransitionContent
}

// AudioConfig holds prepared audio assets

type AudioConfig struct {
	NarrationPath   string // path to narration (mp3/m4a)
	MusicEnabled    bool
	MusicPath       string
	MusicVolume     float64 // 0..1
	NarrationVolume float64 // 0..1, if 0 treat as 1.0
}

// CommandBuildInput contains everything needed to construct ffmpeg args

type CommandBuildInput struct {
	Metadata Metadata
	Timeline []TimelineItem
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

// Compile takes the AI JSON blob, image paths (ordered by index), a tmpDir and returns ffmpeg args and resolved output paths used.
func (cc *CompositionCompiler) Compile(jsonBlob []byte, imagePaths []string, tmpDir string, outputPath string) ([]string, string, string, error) {
	var vc models.VideoCompositionResponseSchema

	if err := json.Unmarshal(jsonBlob, &vc); err != nil {
		return nil, "", "", fmt.Errorf("invalid composition json: %v", err)
	}

	// Map Properties.Metadata.Properties
	if len(vc.Properties.Metadata.Properties.Resolution.OneOf[0].Enum) != 2 {
		return nil, "", "", fmt.Errorf("invalid resolution in Properties.Metadata.Properties")
	}
	meta := Metadata{
		TotalDuration: vc.Properties.Metadata.Properties.TotalDuration,
		AspectRatio:   AspectRatio(vc.Properties.Metadata.Properties.AspectRatio),
		FPS:           vc.Properties.Metadata.Properties.FPS,
		Width:         vc.Properties.Metadata.Properties.Resolution[0],
		Height:        vc.Properties.Metadata.Properties.Resolution[1],
	}

	// Resolve narration via ElevenLabs
	ttsInput := models.TTSInput{
		Narrative: models.NarrativeData{
			Hook:  vc.Properties.Narrative.Properties.Hook,
			Story: vc.Properties.Narrative.Properties.Story,
			Cta:   vc.Properties.Narrative.Properties.Cta,
			Tone:  vc.Properties.Narrative.Properties.Tone,
		},
		Narration: models.NarrationData{
			//nah I don't like this shit
			Script: mapNarrationScript(vc.Properties.Audio.Properties.Narration.Script),
			Voice: models.TTSVoice{
				VoiceID:   vc.Audio.Narration.Voice.VoiceID,
				Speed:     vc.Audio.Narration.Voice.Speed,
				Pitch:     vc.Audio.Narration.Voice.Pitch,
				Stability: vc.Audio.Narration.Voice.Stability,
			},
		},
	}
	ttsNarrationPaths := []string{}
	if cc.voiceService != nil {
		paths, _, err := cc.voiceService.GenerateSpeechToTmp(ttsInput, filepath.Join(tmpDir, "tts_audio"))
		if err != nil {
			return nil, "", "", fmt.Errorf("tts generation failed: %v", err)
		}
		for i := range paths {
			ttsNarrationPaths = append(ttsNarrationPaths, paths[i])
		}
	}

	// Resolve music if enabled
	musicPath := ""
	if vc.Properties.Audio.Properties.Music.Properties.Enabled && cc.bgMusic != nil {
		//this needs to change; it shouldn't be posting a trackID just a music theming
		mf, err := cc.bgMusic.ResolveAndDownload(vc.Audio.Music.TrackID)
		if err != nil {
			return nil, "", "", fmt.Errorf("bgm download failed: %v", err)
		}
		musicPath = mf.FilePath
	}

	// Build timeline
	//TODO: Still need to make the buisness logic
	timeline, err := func(tl vc.Properties.Timeline) ([]TimelineItem, error) {

		switch tl.Type {
		case "image":

		case "text-overlay":

		case "video":

		case "transition"

		case "audio"

		default:
			return nil, fmt.Errorf("unknown timeline type: %s", timeline.Type)

		}
	}
	if err != nil {
		return nil, "", "", err
	}

	args, err := cc.builder.Build(CommandBuildInput{
		Metadata:   meta,
		Timeline:   timeline,
		ImagePaths: imagePaths,
		Audio: AudioConfig{
			NarrationPath:   narrationPath,
			MusicEnabled:    vc.Properties.Audio.Properties.Music.Properties.Enabled,
			MusicPath:       musicPath,
			MusicVolume:     vc.Properties.Audio.Properties.Music.Properties.Volume,
			NarrationVolume: 1.0,
		},
		OutputPath: outputPath,
	})
	if err != nil {
		return nil, "", "", err
	}
	return args, narrationPath, musicPath, nil
}

func mapNarrationScript(in []models.NarrationScriptItem) []models.TTSSegment {
	out := make([]models.TTSSegment, 0, len(in))
	for _, s := range in {
		seg := models.TTSSegment{Text: s.Text, Emphasis: s.Emphasis}
		if s.Timing != nil {
			seg.Start = int(s.Timing.Start)
			seg.End = int(s.Timing.End)
		}
		out = append(out, seg)
	}
	return out
}

func mapTimeline(in []models.TimelineItemvcance) ([]TimelineItem, error) {
	items := make([]TimelineItem, 0, len(in))
	for _, it := range in {
		item := TimelineItem{ID: it.ID, StartTime: it.StartTime, Duration: it.Duration}
		switch it.Type {
		case "image":
			var img models.ImageSegmentvcance
			if err := json.Unmarshal(it.Content, &img); err != nil {
				return nil, fmt.Errorf("invalid image content in %s: %v", it.ID, err)
			}
			item.Type = TimelineItemImage
			item.Image = &ImageSegmentContent{ImageIndex: img.ImageIndex}
		case "text_overlay":
			var tx models.TextOverlayvcance
			if err := json.Unmarshal(it.Content, &tx); err != nil {
				return nil, fmt.Errorf("invalid text content in %s: %v", it.ID, err)
			}
			item.Type = TimelineItemTextOverlay
			style := tx.Style
			fontsize := 12
			color := ""
			bg := ""
			anim := ""
			if style != nil {
				fontsize = int(style.FontSize)
				color = style.Color
				bg = style.BackgroundColor
				anim = style.Animation
			}
			item.Text = &TextOverlayContent{Text: tx.Text, Position: tx.Position, FontSize: fontsize, ColorHex: color, BackgroundColor: bg, Animation: anim}
		case "transition":
			var tr models.Transitionvcance
			if err := json.Unmarshal(it.Content, &tr); err != nil {
				return nil, fmt.Errorf("invalid transition content in %s: %v", it.ID, err)
			}
			item.Type = TimelineItemTransition
			item.Trans = &TransitionContent{Effect: tr.Effect, Easing: tr.Easing}
		default:
			return nil, fmt.Errorf("unknown timeline type: %s", it.Type)
		}
		items = append(items, item)
	}
	return items, nil
}

// Build assembles ffmpeg args for a single-pass render
// Strategy:
// - Provide all images as inputs (-loop 1 -t duration per item not needed if we use filter timeline)
// - We use filter_complex to scale to canvas WxH, pad as needed, compose overlays, and concatenate segments
// - For simplicity, we convert each image item into a video stream with fps, scale, and setpts to absolute timeline using trim + setpts
// - Text overlays are applied on top of the base track within the time window using enable='between(t, start, end)'
// - Transitions currently support fade and cut; others map to cut for MVP
// - Audio: mix narration and BGM (if enabled) with volumes, map to output
func (b *FFmpegCommandBuilder) Build(in CommandBuildInput) ([]string, error) {
	if in.Properties.Metadata.Properties.Width <= 0 || in.Properties.Metadata.Properties.Height <= 0 || in.Properties.Metadata.Properties.FPS <= 0 {
		return nil, fmt.Errorf("invalid Properties.Metadata.Properties: width/height/fps must be > 0")
	}
	if in.OutputPath == "" {
		return nil, fmt.Errorf("missing output path")
	}

	// Validate image indices
	for _, t := range in.Timeline {
		if t.Type == TimelineItemImage {
			if t.Image == nil || t.Image.ImageIndex < 0 || t.Image.ImageIndex >= len(in.ImagePaths) {
				return nil, fmt.Errorf("timeline %s references invalid image index", t.ID)
			}
		}
	}

	// Sort timeline by start time to build proper concat order
	sorted := make([]TimelineItem, len(in.Timeline))
	copy(sorted, in.Timeline)
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
	if in.Audio.NarrationPath != "" {
		args = append(args, "-i", in.Audio.NarrationPath)
	}
	musicIdx := -1
	narrIdx := -1
	if in.Audio.NarrationPath != "" {
		narrIdx = audioInputStart
		if in.Audio.MusicEnabled && in.Audio.MusicPath != "" {
			musicIdx = audioInputStart + 1
		}
	} else if in.Audio.MusicEnabled && in.Audio.MusicPath != "" {
		// Only music
		args = append(args, "-i", in.Audio.MusicPath)
		musicIdx = audioInputStart
	}
	if in.Audio.MusicEnabled && in.Audio.MusicPath != "" && musicIdx == -1 {
		// Music not yet added (narration present handled earlier). Add now.
		args = append(args, "-i", in.Audio.MusicPath)
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
		if t.Type != TimelineItemImage || t.Image == nil {
			continue
		}
		imgInputIdx := t.Image.ImageIndex
		labelIn := fmt.Sprintf("[%d:v]", imgInputIdx)
		labelOut := fmt.Sprintf("[seg%d]", idx)
		// scale to canvas, pad/crop, set fps, set duration
		// Use: scale, pad (if needed), fps, tpad=stop_mode=clone:stop_duration=duration, setpts=N/(FPS*TB)
		// Use loop filter to make image into frames: loop=loop=FPS*duration:size=1:start=0, fps=FPS
		// We'll use: scale=W:H:force_original_aspect_ratio=decrease, pad=W:H:(ow-iw)/2:(oh-ih)/2,format=yuv420p
		filter += fmt.Sprintf("%s scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,format=yuv420p,fps=%d,trim=duration=%f,setpts=PTS-STARTPTS %s;",
			labelIn, in.Properties.Metadata.Properties.Width, in.Properties.Metadata.Properties.Height, in.Properties.Metadata.Properties.Width, in.Properties.Metadata.Properties.Height, in.Properties.Metadata.Properties.FPS, t.Duration, labelOut)
		imageStreamCount++
	}

	// Concatenate all video segments in order
	concatInputs := ""
	for idx, t := range sorted {
		if t.Type == TimelineItemImage {
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
	for _, t := range sorted {
		if t.Type != TimelineItemTextOverlay || t.Text == nil {
			continue
		}
		text := escapeDrawtext(t.Text.Text)
		xy := positionXY(t.Text.Position, in.Properties.Metadata.Properties.Width, in.Properties.Metadata.Properties.Height)
		bg := ""
		if t.Text.BackgroundColor != "" {
			bg = fmt.Sprintf(":box=1:boxcolor=%s@0.6", t.Text.BackgroundColor)
		}
		labelOut := fmt.Sprintf("[vtx%d]", textIdx)
		enable := fmt.Sprintf("enable='between(t,%.3f,%.3f)'", t.StartTime, t.StartTime+t.Duration)
		filter += fmt.Sprintf("%s drawtext=text=%s:fontcolor=%s:fontsize=%d:x=%s:y=%s%s:%s %s;",
			videoLabel, text, colorOrDefault(t.Text.ColorHex), maxInt(t.Text.FontSize, 12), xy[0], xy[1], bg, enable, labelOut)
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
		"-r", fmt.Sprintf("%d", in.Properties.Metadata.Properties.FPS),
		"-s", fmt.Sprintf("%dx%d", in.Properties.Metadata.Properties.Width, in.Properties.Metadata.Properties.Height),
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
