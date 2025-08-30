# Social Media AI Video Generator

A beautiful web application that generates AI videos for social media platforms using Veo3 API.

## Features

- iPhone 15-inspired UI design
- AI video generation with text prompts and reference images
- Support for multiple social media formats (Instagram, TikTok, Twitter, Facebook)
- Real-time chat interface
- File upload and camera capture
- Responsive design with pink gradient background

## Getting Started

### Frontend (React + Vite)
```bash
npm install
npm run dev
```

### Backend (Go + Gin)
```bash
cd backend
go mod tidy
go run main.go
```

The frontend will be available at `http://localhost:5173` and the backend at `http://localhost:8080`.

## Configuration

Update the `VEO3_API_KEY` in `backend/main.go` with your actual Veo3 API key.

## Usage

1. Enter a text description of the video you want to create
2. Optionally upload an image or video as reference
3. Click send to generate your AI video
4. Once generated, use the social media links to share your content

## Project Structure

```
├── src/
│   ├── App.tsx                 # Main React component with iPhone UI
│   ├── main.tsx               # React app entry point
│   └── index.css              # Tailwind CSS imports
├── backend/
│   ├── main.go                # Go server entry point with Gin router
│   ├── go.mod                 # Go module dependencies
│   ├── config/
│   │   └── api.go             # API configuration for external services
│   ├── models/
│   │   └── video.go           # Data models for video generation
│   ├── handlers/
│   │   └── video.go           # HTTP handlers for video endpoints
│   ├── services/
│   │   ├── video_processor.go # FFmpeg video processing service
│   │   ├── elevenlabs.go      # ElevenLabs TTS integration
│   │   └── content_generator.go # Video composition generation
│   └── schema/
│       └── video_request.json # JSON schema for video generation requests
├── package.json               # Frontend dependencies
├── tailwind.config.js         # Tailwind CSS configuration
├── vite.config.ts            # Vite build configuration
└── README.md                 # Project documentation
```

## Architecture Overview

### Frontend (React + TypeScript)
- **iPhone 15 Frame**: Displays generated videos in an authentic iPhone design
- **Chat Interface**: Simple input form for text prompts and file uploads
- **Social Media Integration**: Direct links to platform upload pages
- **Responsive Design**: Works across all device sizes with pink gradient background

### Backend (Go + Gin Framework)
- **RESTful API**: Clean endpoints for video generation and health checks
- **Video Processing Pipeline**: Complete workflow from prompt to finished video
- **External API Integration**: Connects to Veo3, ElevenLabs, and Pexels
- **Temporary File Management**: Downloads, processes, and cleans up video files

## Video Processing Workflow

1. **User Input**: Text prompt and optional reference image/video
2. **Content Generation**: Creates structured video composition plan
3. **Asset Download**: Fetches Pexels videos based on composition
4. **Audio Generation**: Creates TTS narration using ElevenLabs
5. **Video Assembly**: Uses FFmpeg to combine videos with transitions
6. **Audio Mixing**: Overlays TTS and background music
7. **Format Optimization**: Outputs 9:16 aspect ratio for social media
8. **Cleanup**: Removes temporary files after serving

## FFmpeg Video Processing

The backend uses `ffmpeg-go` to handle sophisticated video operations:

- **Video Concatenation**: Combines multiple Pexels video segments
- **Transition Effects**: Supports cuts and fade transitions between clips
- **Audio Mixing**: Blends TTS narration with background music
- **Format Conversion**: Ensures compatibility across social platforms
- **Resolution Scaling**: Automatically crops and resizes for Instagram/TikTok
- **Temporary Storage**: Downloads and processes videos in `/tmp` directory

## JSON Schema Structure

The video generation follows a structured JSON format defined in `backend/schema/video_request.json`:

- **Video Specifications**: Length, aspect ratio, resolution settings
- **Audio Configuration**: TTS script, voice settings, background music
- **Video Segments**: Array of Pexels videos with timing and effects
- **Transition Settings**: Cut/fade effects between video segments
- **Processing Effects**: Crop, zoom, and speed adjustments

## API Endpoints

### POST `/api/v1/generate-video`
Generates a social media video from user prompt and optional reference files.

**Request**: Form data with `prompt` (string) and optional `file` (image/video)
**Response**: JSON with `videoUrl`, `status`, and optional `error`

### GET `/api/v1/composition`
Returns the video composition structure for debugging purposes.

**Query Parameters**: `prompt` (string)
**Response**: JSON with complete video composition plan

### GET `/health`
Health check endpoint for service monitoring.

## External Services

- **Veo3 API**: AI video generation (placeholder integration)
- **ElevenLabs**: Text-to-speech conversion for video narration
- **Pexels**: Stock video content for video segments
- **FFmpeg**: Video processing and audio mixing

## Development Notes

- Frontend runs on port 5173 (Vite default)
- Backend runs on port 8080 (configurable via environment)
- CORS configured for localhost development
- No authentication required for demo purposes
- Temporary files automatically cleaned up after processing
- IBM Plex Sans font used throughout the interface