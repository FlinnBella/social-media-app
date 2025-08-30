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