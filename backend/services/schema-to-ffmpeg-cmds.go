package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"social-media-ai-video/models"
	video_models "social-media-ai-video/models/timeline"
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
	FilePath []string
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
	Timeline        video_models.TimelineComposition
	// Images referenced by index in timeline (ImageIndex)
	ImageorVideoPaths []string
	// Audio assets
	Audio AudioConfig
	// Output path for the final video
	FinalVideoPath   string
	FinalVideoTmpDir string
}

// VideoCompiler defines the interface for video compilation strategies
type VideoCompiler interface {
	Compile(jsonSchemaBlob []byte, InputFilePaths []string) (io.Reader, error)
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

// Compile takes the AI JSON blob and image paths and returns a video stream
// Handles its own cleanup of intermediate files
func (rc *ReelsCompiler) Compile(jsonSchemaBlob []byte, InputImagePaths []string) (io.Reader, error) {
	var reelsTimeline video_models.TimelineComposition
	err := json.Unmarshal(jsonSchemaBlob, &reelsTimeline)
	if err != nil {
		return nil, fmt.Errorf("invalid composition json: %v. Given json: %s", err, string(jsonSchemaBlob))
	}
	// Track intermediate files for cleanup
	var intermediateFiles []string

	// Clean up intermediate files when function exits
	defer func() {
		for _, file := range intermediateFiles {
			os.Remove(file)
		}
	}()

	// Map Properties.Metadata.Properties
	if len(reelsTimeline.Metadata.Resolution) != 2 {
		return nil, fmt.Errorf("invalid resolution resolution array %v", reelsTimeline.Metadata.Resolution)
	}

	fps := 30
	if reelsTimeline.Metadata.Fps != "" {
		if parsed, err := strconv.Atoi(reelsTimeline.Metadata.Fps); err == nil {
			fps = parsed
		}
	}
	meta := Metadata_Universal_FFmpeg{
		TotalDuration: reelsTimeline.Metadata.TotalDuration,
		AspectRatio:   reelsTimeline.Metadata.AspectRatio,
		FPS:           fps,
		Width:         reelsTimeline.Metadata.Resolution[0],
		Height:        reelsTimeline.Metadata.Resolution[1],
	}

	// Resolve narration via ElevenLabs
	// Extract all text from TextSegments
	textSegments := reelsTimeline.Timeline.TextTimeline.TextSegments
	textInputs := make([]string, len(textSegments)) // Pre-allocate with exact size
	for i, segment := range textSegments {
		textInputs[i] = segment.Text
	}

	ttsInput := models.TTSInput{
		TextInput: textInputs,
	}

	var elevenlabsFileData *models.FileOutput

	//Generate tts narration elevenlabs
	if rc.voiceService.config != nil {
		data, err := rc.voiceService.GenerateSpeechToTmp(ttsInput.TextInput)
		if err != nil {
			return nil, fmt.Errorf("tts generation failed: %v", err)
		}
		//extend scope of data
		elevenlabsFileData = data
		// Add TTS files to cleanup list
		intermediateFiles = append(intermediateFiles, data.FilePath)
	}

	// Resolve music if enabled
	musicPath := ""
	musicName := ""

	if reelsTimeline.Music.Enabled && rc.bgMusic.cfg != nil {
		mf, err := rc.bgMusic.CreateBackgroundMusic(reelsTimeline.Music.Genre)
		if err != nil {
			return nil, fmt.Errorf("bgm download failed: %v", err)
		}
		musicPath = mf.FilePath
		musicName = mf.FileName

		// Add music file to cleanup list
		intermediateFiles = append(intermediateFiles, mf.FilePath)
	}

	// Auto-generate an output path under the OS temp directory
	tmpDir := filepath.Join(os.TempDir(), "reels_video")
	autoOutput := filepath.Join(tmpDir, fmt.Sprintf("short_%d.mp4", time.Now().UnixNano()))

	// Execute FFmpeg and return video stream
	return rc.Build(CommandBuildInput{
		Metadata_FFmpeg:   meta,
		Timeline:          reelsTimeline,
		ImageorVideoPaths: InputImagePaths,
		Audio: AudioConfig{
			ttsNarrationPaths: ttsNarartionFiles{FilePath: []string{elevenlabsFileData.FilePath}, FileName: []string{elevenlabsFileData.FileName}},
			MusicEnabled:      reelsTimeline.Music.Enabled,
			MusicPath:         MusicFiles{MusicPath: musicPath, MusicName: musicName},
			MusicVolume:       reelsTimeline.Music.Volume,
			NarrationVolume:   1.0,
		},
		FinalVideoPath:   autoOutput,
		FinalVideoTmpDir: tmpDir,
	})
}

// Compile implements VideoCompiler interface for ProCompiler
// Pro-specific compilation: high quality, videos, professional processing
// Compilier actually works the same exact way as reels; but we're avoiding abstractions for now
func (pc *ProCompiler) Compile(jsonSchemaBlob []byte, InputVideoPaths []string) (io.Reader, error) {
	var proTimeline video_models.TimelineComposition
	err := json.Unmarshal(jsonSchemaBlob, &proTimeline)
	if err != nil {
		return nil, fmt.Errorf("invalid composition json: %v. Given json: %s", err, string(jsonSchemaBlob))
	}
	// Track intermediate files for cleanup
	var intermediateFiles []string

	// Clean up intermediate files when function exits
	defer func() {
		for _, file := range intermediateFiles {
			os.Remove(file)
		}
	}()

	// Map Properties.Metadata.Properties
	if len(proTimeline.Metadata.Resolution) != 2 {
		return nil, fmt.Errorf("invalid resolution resolution array %v", proTimeline.Metadata.Resolution)
	}

	fps := 30
	if proTimeline.Metadata.Fps != "" {
		if parsed, err := strconv.Atoi(proTimeline.Metadata.Fps); err == nil {
			fps = parsed
		}
	}
	meta := Metadata_Universal_FFmpeg{
		TotalDuration: proTimeline.Metadata.TotalDuration,
		AspectRatio:   proTimeline.Metadata.AspectRatio,
		FPS:           fps,
		Width:         proTimeline.Metadata.Resolution[0],
		Height:        proTimeline.Metadata.Resolution[1],
	}

	// Resolve narration via ElevenLabs
	// Extract all text from TextSegments
	textSegments := proTimeline.Timeline.TextTimeline.TextSegments
	textInputs := make([]string, len(textSegments)) // Pre-allocate with exact size
	for i, segment := range textSegments {
		textInputs[i] = segment.Text
	}

	ttsInput := models.TTSInput{
		TextInput: textInputs,
	}

	var elevenlabsFileData *models.FileOutput

	//Generate tts narration elevenlabs
	if pc.voiceService.config != nil {
		data, err := pc.voiceService.GenerateSpeechToTmp(ttsInput.TextInput)
		if err != nil {
			return nil, fmt.Errorf("tts generation failed: %v", err)
		}
		//extend scope of data
		elevenlabsFileData = data
		// Add TTS files to cleanup list
		intermediateFiles = append(intermediateFiles, data.FilePath)
	}

	// Resolve music if enabled
	musicPath := ""
	musicName := ""

	if proTimeline.Music.Enabled && pc.bgMusic.cfg != nil {
		mf, err := pc.bgMusic.CreateBackgroundMusic(proTimeline.Music.Genre)
		if err != nil {
			return nil, fmt.Errorf("bgm download failed: %v", err)
		}
		musicPath = mf.FilePath
		musicName = mf.FileName

		// Add music file to cleanup list
		intermediateFiles = append(intermediateFiles, mf.FilePath)
	}

	// Auto-generate an output path under the OS temp directory
	tmpDir := filepath.Join(os.TempDir(), "reels_video")
	autoOutput := filepath.Join(tmpDir, fmt.Sprintf("short_%d.mp4", time.Now().UnixNano()))

	// Execute FFmpeg and return video stream
	return pc.Build(CommandBuildInput{
		Metadata_FFmpeg:   meta,
		Timeline:          proTimeline,
		ImageorVideoPaths: InputVideoPaths,
		Audio: AudioConfig{
			ttsNarrationPaths: ttsNarartionFiles{FilePath: []string{elevenlabsFileData.FilePath}, FileName: []string{elevenlabsFileData.FileName}},
			MusicEnabled:      proTimeline.Music.Enabled,
			MusicPath:         MusicFiles{MusicPath: musicPath, MusicName: musicName},
			MusicVolume:       proTimeline.Music.Volume,
			NarrationVolume:   1.0,
		},
		FinalVideoPath:   autoOutput,
		FinalVideoTmpDir: tmpDir,
	})
}

func (rc *ReelsCompiler) Build(in CommandBuildInput) (io.Reader, error) {
	if in.Metadata_FFmpeg.Width <= 0 || in.Metadata_FFmpeg.Height <= 0 || in.Metadata_FFmpeg.FPS <= 0 {
		return nil, fmt.Errorf("invalid metadata: width/height/fps must be > 0")
	}
	if in.FinalVideoPath == "" {
		return nil, fmt.Errorf("missing output path")
	}

	// Validate image input paths exist
	for _, p := range in.ImageorVideoPaths {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("missing input file: %s: %v", p, err)
		}
	}

	// Validate image indices
	for _, t := range in.Timeline.Timeline.ImageTimeline.ImageSegments {
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImageorVideoPaths) {
			return nil, fmt.Errorf("image item %d references invalid image index", t.ImageIndex)
		}

	}

	// Sort timeline by start time to build proper concat order
	sorted := make([]video_models.ImageSegment, len(in.Timeline.Timeline.ImageTimeline.ImageSegments))
	copy(sorted, in.Timeline.Timeline.ImageTimeline.ImageSegments)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].StartTime < sorted[j].StartTime })

	// Input list: images + audio(s)
	/*
	   # FFMPEG CMD STARTS HERE
	*/
	args := []string{"-y"}

	// Image inputs (each once); we will reference by indices
	for _, p := range in.ImageorVideoPaths {
		// APPEND IMAGES AS INPUTS INTO FFMPEG - FIRST INPUT APPEND
		args = append(args, "-i", p)
	}

	// Audio inputs appended at the end so index math is predictable
	numImageInputs := len(in.ImageorVideoPaths)
	audioInputStart := numImageInputs
	// Validate narration and music file paths before adding as inputs

	musicIdx := -1
	narrIdx := -1

	// guard against invalid tts directory, or filename
	for i := 0; i < len(in.Audio.ttsNarrationPaths.FileName); i++ {
		fn := in.Audio.ttsNarrationPaths.FileName[i]
		p := in.Audio.ttsNarrationPaths.FilePath[i]
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
		args = append(args, "-i", in.Audio.ttsNarrationPaths.FilePath[i])

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
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImageorVideoPaths) {
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
		if t.ImageIndex >= 0 && t.ImageIndex < len(in.ImageorVideoPaths) {
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
	textsegments := in.Timeline.Timeline.TextTimeline.TextSegments
	for _, t := range sorted {
		if t.ImageIndex < 0 || t.ImageIndex >= len(in.ImageorVideoPaths) {
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

	// Execute FFmpeg and return stdout as Reader
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	// Return the stdout pipe as a Reader
	return stdout, nil
}

// Reels Compiler and Compile both have
// Compile implements VideoCompiler interface for ReelsCompiler

func (pc *ProCompiler) Build(in CommandBuildInput) (io.Reader, error) {
	var finalVideo []byte
	// TODO: Implement Pro-specific FFmpeg args and execution
	// For now, delegate to ReelsCompiler logic

	//puuting as input image/video paths, tts narration, and music
	//perhaps need to arrange inputs in specific order? Regardless need
	//to keep track of my inpuuts indices in the args
	args := []string{"-y"}
	for _, p := range in.ImageorVideoPaths {
		args = append(args, "-i", p)
	}
	for _, a := range in.Audio.ttsNarrationPaths.FilePath {
		args = append(args, "-i", a)
	}
	if in.Audio.MusicEnabled && in.Audio.MusicPath.MusicPath != "" {
		args = append(args, "-i", in.Audio.MusicPath.MusicPath)
	}

	//filter complex logic (don't fully understand what this is)
	args = append(args, "-filter_complex", filter, "-map", finalVideoLabel)
	if len(audioMaps) > 0 {
		for _, am := range audioMaps {
			args = append(args, "-map", am)
		}
	}

	// demuxing so I can compose multiple videos based on indicies

	//where do I combine the audio and video streams? When do I apply the 'filters'?
	//what other logic can I add to this
	//I know this is stupid, but let's go with the safe option for now
	err := os.WriteFile(in.FinalVideoPath, finalVideo, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write final video: %v", err)
	}
	sameVideoFileReadFromDisk, err := os.ReadFile(in.FinalVideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read final video: %v", err)
	}
	return bytes.NewReader(sameVideoFileReadFromDisk), nil

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
