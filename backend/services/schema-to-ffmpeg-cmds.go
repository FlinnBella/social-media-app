package services

import (
	"fmt"
	"path/filepath"
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

// CompositionMetadata captures global video settings
// Width/Height must match the selected aspect ratio
// FPS currently limited to 24 or 30 in the schema

type CompositionMetadata struct {
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

// TimelineItem is a single instruction on the timeline

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
	Metadata CompositionMetadata
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

// Build assembles ffmpeg args for a single-pass render
// Strategy:
// - Provide all images as inputs (-loop 1 -t duration per item not needed if we use filter timeline)
// - We use filter_complex to scale to canvas WxH, pad as needed, compose overlays, and concatenate segments
// - For simplicity, we convert each image item into a video stream with fps, scale, and setpts to absolute timeline using trim + setpts
// - Text overlays are applied on top of the base track within the time window using enable='between(t, start, end)'
// - Transitions currently support fade and cut; others map to cut for MVP
// - Audio: mix narration and BGM (if enabled) with volumes, map to output
func (b *FFmpegCommandBuilder) Build(in CommandBuildInput) ([]string, error) {
	if in.Metadata.Width <= 0 || in.Metadata.Height <= 0 || in.Metadata.FPS <= 0 {
		return nil, fmt.Errorf("invalid metadata: width/height/fps must be > 0")
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
			labelIn, in.Metadata.Width, in.Metadata.Height, in.Metadata.Width, in.Metadata.Height, in.Metadata.FPS, t.Duration, labelOut)
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
		xy := positionXY(t.Text.Position, in.Metadata.Width, in.Metadata.Height)
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
		"-r", fmt.Sprintf("%d", in.Metadata.FPS),
		"-s", fmt.Sprintf("%dx%d", in.Metadata.Width, in.Metadata.Height),
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
