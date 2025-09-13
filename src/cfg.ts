/*
API ENDPOINTS
*/

export const API_ENDPOINTS = {
    generateVideoPexels: '/api/generate-video-pexels',
    generateVideoReels: '/api/generate-video-reels',
    generateVideoProReels: '/api/generate-video-pro-reels',
    generateVideoTimeline: '/api/generate-video-timeline',
} as const;

export type ApiEndpointKey = keyof typeof API_ENDPOINTS;
export type ApiEndpoint = (typeof API_ENDPOINTS)[ApiEndpointKey];



/*
SSE ENDPOINTS
*/

export const SSE_ENDPOINTS = {
    serverSSEUpdates: '/api/sse/video_update',
} as const;

export type SseEndpointKey = keyof typeof SSE_ENDPOINTS;
export type SseEndpoint = (typeof SSE_ENDPOINTS)[SseEndpointKey];