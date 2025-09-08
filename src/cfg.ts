/*
API ENDPOINTS
*/

export const API_ENDPOINTS = {
    generateVideoPexels: '/api/generate-video-pexels',
    generateVideoReels: '/api/generate-video-reels',
    generateVideoProReels: '/api/generate-video-pro-reels',
} as const;

export type ApiEndpointKey = keyof typeof API_ENDPOINTS;
export type ApiEndpoint = (typeof API_ENDPOINTS)[ApiEndpointKey];



/*
SSE ENDPOINTS
*/

export const SSE_ENDPOINTS = {
    serverSSE: '/api/sse',
} as const;

export type SseEndpointKey = keyof typeof SSE_ENDPOINTS;
export type SseEndpoint = (typeof SSE_ENDPOINTS)[SseEndpointKey];