import React, { useState, useCallback, useRef, useEffect } from 'react';
import type { VideoProgressData, VideoErrorData } from '../components/VideoProgressBar';

interface UseVideoProgressOptions {
  sseUrl: string;
  onComplete?: () => void;
  onError?: (error: VideoErrorData) => void;
}

export const useVideoProgress = ({ sseUrl, onComplete, onError }: UseVideoProgressOptions) => {
  const [progress, setProgress] = useState<VideoProgressData | null>(null);
  const [error, setError] = useState<VideoErrorData | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const [clientId, setClientId] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  
  const eventSourceRef = useRef<EventSource | null>(null);
  const progressTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const connectSSE = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const newClientId = `client_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    setClientId(newClientId);
    
    const es = new EventSource(`${sseUrl}?client_id=${newClientId}`);
    eventSourceRef.current = es;

    const onConnected = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      setClientId(data.client_id);
      setIsConnected(true);
    };

    const onVideoProgress = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      setProgress(data);
      setError(null);
      
      // Auto-hide progress bar after completion
      if (data.stage === 'completed') {
        if (progressTimeoutRef.current) {
          clearTimeout(progressTimeoutRef.current);
        }
        progressTimeoutRef.current = setTimeout(() => {
          setIsVisible(false);
          onComplete?.();
        }, 3000); // Hide after 3 seconds
      }
    };

    const onVideoError = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      setError(data);
      setProgress(null);
      onError?.(data);
    };

    const onError = (event: Event) => {
      console.error("SSE Error:", event);
      setError({
        error: 'Connection lost. Please try again.',
        stage: 'connection_failed'
      });
      setIsConnected(false);
    };

    es.addEventListener("connected", onConnected as EventListener);
    es.addEventListener("video_progress", onVideoProgress as EventListener);
    es.addEventListener("video_error", onVideoError as EventListener);
    es.addEventListener("error", onError as EventListener);

    return () => {
      es.removeEventListener("connected", onConnected as EventListener);
      es.removeEventListener("video_progress", onVideoProgress as EventListener);
      es.removeEventListener("video_error", onVideoError as EventListener);
      es.removeEventListener("error", onError as EventListener);
      es.close();
      eventSourceRef.current = null;
      setIsConnected(false);
    };
  }, [sseUrl, onComplete, onError]);

  const startProgress = useCallback(() => {
    setProgress({ stage: 'processing', message: 'Starting video generation...', progress: 0 });
    setError(null);
    setIsVisible(true);
    
    // Connect to SSE
    connectSSE();
  }, [connectSSE]);

  const stopProgress = useCallback(() => {
    setIsVisible(false);
    setProgress(null);
    setError(null);
    setClientId(null);
    
    // Disconnect from SSE
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const resetProgress = useCallback(() => {
    setProgress(null);
    setError(null);
    setIsVisible(false);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      if (progressTimeoutRef.current) {
        clearTimeout(progressTimeoutRef.current);
      }
    };
  }, []);

  return {
    progress,
    error,
    isVisible,
    clientId,
    startProgress,
    stopProgress,
    resetProgress,
    isConnected,
    isLoading: false // We manage loading state internally
  };
};
