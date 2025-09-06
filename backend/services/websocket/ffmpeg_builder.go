package websocket

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"social-media-ai-video/services"
)

// FFmpegBuilder provides a builder pattern for constructing FFmpeg commands
type FFmpegBuilder struct {
	inputs          []string
	filters         []string
	outputOptions   []string
	outputPath      string
	backgroundMusic *services.BackgroundMusic
	elevenLabs      *services.ElevenLabsService
}

// FFmpegCompiler compiles video schemas into executable FFmpeg commands
type FFmpegCompiler struct {
	builder         *FFmpegBuilder
	backgroundMusic *services.BackgroundMusic
	elevenLabs      *services.ElevenLabsService
}

// VideoSchema represents the structure of a video composition schema
type VideoSchema struct {
	Metadata            SchemaMetadata  `json:"metadata"`
	PropertyInfo        PropertyData    `json:"property_info"`
	PhotoSequence       []int           `json:"photo_sequence"`
	MarketingHighlights []string        `json:"marketing_highlights"`
	Narrative           SchemaNarrative `json:"narrative"`
	VoiceStyle          string          `json:"voice_style"`
	Timing              []float64       `json:"timing"`
}

// SchemaMetadata represents video metadata
type SchemaMetadata struct {
	TotalDuration float64 `json:"total_duration"`
	AspectRatio   string  `json:"aspect_ratio"`
	FPS           string  `json:"fps"`
	Resolution    []int   `json:"resolution"`
}

// SchemaNarrative represents the video narrative structure
type SchemaNarrative struct {
	Hook         string   `json:"hook"`
	TourSegments []string `json:"tour_segments"`
	CallToAction string   `json:"call_to_action"`
}

// NewFFmpegBuilder creates a new FFmpeg command builder
func NewFFmpegBuilder() *FFmpegBuilder {
	return &FFmpegBuilder{
		inputs:        []string{},
		filters:       []string{},
		outputOptions: []string{},
	}
}

// NewFFmpegCompiler creates a new FFmpeg compiler with builder pattern
func NewFFmpegCompiler(backgroundMusic *services.BackgroundMusic, elevenLabs *services.ElevenLabsService) *FFmpegCompiler {
	return &FFmpegCompiler{
		builder:         NewFFmpegBuilder(),
		backgroundMusic: backgroundMusic,
		elevenLabs:      elevenLabs,
	}
}

// AddInput adds an input file to the FFmpeg command
func (fb *FFmpegBuilder) AddInput(inputPath string) *FFmpegBuilder {
	fb.inputs = append(fb.inputs, "-i", inputPath)
	return fb
}

// AddFilter adds a filter to the FFmpeg command
func (fb *FFmpegBuilder) AddFilter(filter string) *FFmpegBuilder {
	fb.filters = append(fb.filters, filter)
	return fb
}

// SetOutput sets the output path and options
func (fb *FFmpegBuilder) SetOutput(outputPath string, options ...string) *FFmpegBuilder {
	fb.outputPath = outputPath
	fb.outputOptions = append(fb.outputOptions, options...)
	return fb
}

// Build constructs the complete FFmpeg command arguments
func (fb *FFmpegBuilder) Build() []string {
	args := []string{}

	// Add all inputs
	args = append(args, fb.inputs...)

	// Add filters if any
	if len(fb.filters) > 0 {
		args = append(args, "-filter_complex", strings.Join(fb.filters, ";"))
	}

	// Add output options
	args = append(args, fb.outputOptions...)

	// Add output path
	if fb.outputPath != "" {
		args = append(args, fb.outputPath)
	}

	return args
}

// Reset clears the builder for reuse
func (fb *FFmpegBuilder) Reset() *FFmpegBuilder {
	fb.inputs = []string{}
	fb.filters = []string{}
	fb.outputOptions = []string{}
	fb.outputPath = ""
	return fb
}

// Compile processes a video schema and generates FFmpeg command
func (fc *FFmpegCompiler) Compile(schemaBytes []byte, imagePaths []string) ([]string, *VideoSchema, string, error) {
	// Parse the schema
	var schema VideoSchema
	if err := json.Unmarshal(schemaBytes, &schema); err != nil {
		return nil, nil, "", fmt.Errorf("failed to parse video schema: %v", err)
	}

	// Create temporary output file
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("property_video_%d.mp4", time.Now().Unix()))

	// Reset builder for fresh command
	fc.builder.Reset()

	// Build FFmpeg command based on schema
	args, err := fc.buildFromSchema(&schema, imagePaths, outputPath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to build FFmpeg command: %v", err)
	}

	return args, &schema, outputPath, nil
}

// CompileAndExecute processes schema and executes FFmpeg command
func (fc *FFmpegCompiler) CompileAndExecute(schemaBytes []byte, imagePaths []string) (string, error) {
	args, _, outputPath, err := fc.Compile(schemaBytes, imagePaths)
	if err != nil {
		return "", err
	}

	// Execute FFmpeg command
	cmd := exec.Command("ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("ffmpeg execution failed: %v, output: %s", err, string(output))
	}

	// Verify output file exists and is not empty
	if fi, err := os.Stat(outputPath); err != nil || fi.Size() == 0 {
		return "", fmt.Errorf("output file missing or empty")
	}

	return outputPath, nil
}

// buildFromSchema constructs FFmpeg command from video schema
func (fc *FFmpegCompiler) buildFromSchema(schema *VideoSchema, imagePaths []string, outputPath string) ([]string, error) {
	if len(imagePaths) == 0 {
		return nil, fmt.Errorf("no image paths provided")
	}

	// Add all image inputs
	for _, imgPath := range imagePaths {
		fc.builder.AddInput(imgPath)
	}

	// Generate video filters based on schema
	filters := fc.generateVideoFilters(schema, len(imagePaths))
	for _, filter := range filters {
		fc.builder.AddFilter(filter)
	}

	// Set output options based on metadata
	outputOptions := fc.generateOutputOptions(schema)
	fc.builder.SetOutput(outputPath, outputOptions...)

	return fc.builder.Build(), nil
}

// generateVideoFilters creates filter chain based on schema
func (fc *FFmpegCompiler) generateVideoFilters(schema *VideoSchema, imageCount int) []string {
	filters := []string{}

	// Scale and format all images
	for i := 0; i < imageCount; i++ {
		// Scale each image to target resolution
		scaleFilter := fmt.Sprintf("[%d:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black,setdar=%s,fps=%s[img%d]",
			i,
			schema.Metadata.Resolution[0], schema.Metadata.Resolution[1],
			schema.Metadata.Resolution[0], schema.Metadata.Resolution[1],
			schema.Metadata.AspectRatio,
			schema.Metadata.FPS,
			i)
		filters = append(filters, scaleFilter)
	}

	// Create timeline based on timing
	if len(schema.Timing) > 0 {
		timeline := fc.generateTimeline(schema, imageCount)
		filters = append(filters, timeline)
	}

	// Add fade transitions
	transitionFilter := fc.generateTransitions(imageCount)
	if transitionFilter != "" {
		filters = append(filters, transitionFilter)
	}

	return filters
}

// generateTimeline creates a timeline filter for image sequencing
func (fc *FFmpegCompiler) generateTimeline(schema *VideoSchema, imageCount int) string {
	if len(schema.PhotoSequence) == 0 || len(schema.Timing) == 0 {
		// Default timeline
		return fmt.Sprintf("concat=n=%d:v=1:a=0[v]", imageCount)
	}

	// Build complex timeline with custom durations
	timelineInputs := []string{}
	for i, _ := range schema.Timing {
		if i < len(schema.PhotoSequence) && schema.PhotoSequence[i] < imageCount {
			imgIndex := schema.PhotoSequence[i]
			timelineInputs = append(timelineInputs, fmt.Sprintf("[img%d]", imgIndex))
		}
	}

	return fmt.Sprintf("%s concat=n=%d:v=1:a=0[v]", strings.Join(timelineInputs, ""), len(timelineInputs))
}

// generateTransitions creates fade transition effects
func (fc *FFmpegCompiler) generateTransitions(imageCount int) string {
	if imageCount <= 1 {
		return ""
	}

	// Simple fade transitions between images
	return "[v]fade=t=in:st=0:d=0.5,fade=t=out:st=11.5:d=0.5[v_faded]"
}

// generateOutputOptions creates output encoding options based on metadata
func (fc *FFmpegCompiler) generateOutputOptions(schema *VideoSchema) []string {
	options := []string{
		"-c:v", "libx264", // Video codec
		"-preset", "medium", // Encoding speed vs quality
		"-crf", "23", // Quality setting
		"-pix_fmt", "yuv420p", // Pixel format for compatibility
		"-r", schema.Metadata.FPS, // Frame rate
		"-t", fmt.Sprintf("%.1f", schema.Metadata.TotalDuration), // Duration
	}

	// Add aspect ratio if specified
	if schema.Metadata.AspectRatio != "" {
		options = append(options, "-aspect", schema.Metadata.AspectRatio)
	}

	return options
}

// Enhanced builder methods for specific video operations

// AddPropertyOverlay adds property information overlay
func (fb *FFmpegBuilder) AddPropertyOverlay(propertyData PropertyData) *FFmpegBuilder {
	// Add text overlay with property information
	text := fmt.Sprintf("%s\\n$%.0f", propertyData.Address, propertyData.Price)
	overlay := fmt.Sprintf("drawtext=text='%s':fontfile=/System/Library/Fonts/Arial.ttf:fontsize=24:fontcolor=white:x=10:y=10", text)
	return fb.AddFilter(overlay)
}

// AddLogoOverlay adds realtor logo overlay
func (fb *FFmpegBuilder) AddLogoOverlay(logoPath string) *FFmpegBuilder {
	fb.AddInput(logoPath)
	// Overlay logo in bottom right corner
	overlay := "[v][1:v]overlay=main_w-overlay_w-10:main_h-overlay_h-10[v]"
	return fb.AddFilter(overlay)
}

// AddBackgroundMusic adds background music to the video
func (fb *FFmpegBuilder) AddBackgroundMusic(musicPath string, volume float64) *FFmpegBuilder {
	fb.AddInput(musicPath)
	// Mix audio with specified volume
	audioMix := fmt.Sprintf("[1:a]volume=%.2f[music]", volume)
	return fb.AddFilter(audioMix)
}

// AddVoiceover adds AI-generated voiceover
func (fb *FFmpegBuilder) AddVoiceover(voiceoverPath string) *FFmpegBuilder {
	fb.AddInput(voiceoverPath)
	return fb
}

// QuickBuild provides a simplified interface for common video operations
func (fc *FFmpegCompiler) QuickBuild(config VideoCompilerConfig) ([]string, string, error) {
	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("quick_video_%d.mp4", time.Now().Unix()))

	fc.builder.Reset()

	// Add inputs
	for _, imgPath := range config.ImagePaths {
		fc.builder.AddInput(imgPath)
	}

	// Add basic video processing
	fc.builder.AddFilter(fmt.Sprintf("scale=%d:%d", config.Width, config.Height))

	// Set output
	fc.builder.SetOutput(outputPath, "-c:v", "libx264", "-crf", "23", "-t", fmt.Sprintf("%.1f", config.Duration))

	return fc.builder.Build(), outputPath, nil
}

// VideoCompilerConfig provides configuration for quick builds
type VideoCompilerConfig struct {
	ImagePaths []string
	Width      int
	Height     int
	Duration   float64
	FPS        string
}

// DefaultConfig returns a default configuration for property videos
func DefaultConfig() VideoCompilerConfig {
	return VideoCompilerConfig{
		Width:    1080,
		Height:   1920,
		Duration: 12.0,
		FPS:      "30",
	}
}
