import React, { useState, useCallback, useRef, useEffect } from 'react';
import type { VideoErrorData } from '../components/VideoProgressBar';
import { SSE_ENDPOINTS } from '@/cfg';


interface UseVideoProgressOptions {
  sseUrl: string;
}

export const useVideoProgress = ({ sseUrl = SSE_ENDPOINTS.serverSSEUpdates}: UseVideoProgressOptions) => {
  const [progressEvents, setProgressEvents] = useState<{stage: string, progress: number}>({stage: 'processing', progress: 0});
  const [error, setError] = useState<VideoErrorData | null>(null);
  const [isVisible, setIsVisible] = useState(false);
  const [clientId, setClientId] = useState<string | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  
  const eventSourceRef = useRef<EventSource | null>(null);
  const progressTimeoutRef = useRef<number | null>(null);

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
      //guard to verify client id is preserved
      if (data.client_id !== clientId) {
        console.error("Client ID mismatch:", data.client_id, clientId);
        setError({
          error: 'Client ID mismatch. Please try again.',
          stage: 'client_id_mismatch'
        });
        setIsConnected(false);
        return;
      }
      setIsConnected(true);
    };

    const onVideoProgress = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      setProgressEvents(data);
      setError(null);
      
      // Auto-hide progress bar after completion
      if (data.stage === 'completed') {
        if (progressTimeoutRef.current) {
          clearTimeout(progressTimeoutRef.current);
        }
        progressTimeoutRef.current = setTimeout(() => {
          setIsVisible(false);
        }, 3000); // Hide after 3 seconds
      }
    };

    const onVideoError = (event: MessageEvent) => {
      const data = JSON.parse(event.data);
      setError(data);
      setProgressEvents({stage: 'error', progress: 0});
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
  }, [sseUrl]);

  const startProgress = useCallback(() => {
    setProgressEvents({stage: 'processing', progress: 0});
    setError(null);
    setIsVisible(true);
    
    // Connect to SSE
    connectSSE();
  }, [connectSSE]);

  const stopProgress = useCallback(() => {
    setIsVisible(false);
    setProgressEvents({stage: 'processing', progress: 0});
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
    setProgressEvents({stage: 'processing', progress: 0});
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
    progressEvents,
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
