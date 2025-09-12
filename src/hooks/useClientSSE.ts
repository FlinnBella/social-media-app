import { useEffect, useRef, useState, useCallback } from "react";

export interface VideoProgress {
    stage: string;
    message: string;
    progress: number;
}

export interface VideoError {
    error: string;
    stage: string;
}

export interface SSEData {
    type: string;
    data: any;
}

export const useClientSSE = (baseUrl: string) => {
    const [data, setData] = useState<SSEData | null>(null);
    const [error, setError] = useState<unknown>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const [clientId, setClientId] = useState<string | null>(null);
    const [connected, setConnected] = useState<boolean>(false);
    const eventSourceRef = useRef<EventSource | null>(null);

    const connect = useCallback(() => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        setLoading(true);
        setError(null);
        
        // Generate a unique client ID for this session
        const newClientId = `client_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        setClientId(newClientId);
        
        const es = new EventSource(`${baseUrl}?client_id=${newClientId}`);
        eventSourceRef.current = es;

        const onConnected = (event: MessageEvent) => {
            const data = JSON.parse(event.data);
            setClientId(data.client_id);
            setConnected(true);
            setLoading(false);
            setData({ type: "connected", data });
        };

        const onHeartbeat = (event: MessageEvent) => {
            const data = JSON.parse(event.data);
            setData({ type: "heartbeat", data });
        };

        const onVideoProgress = (event: MessageEvent) => {
            const data = JSON.parse(event.data);
            setData({ type: "video_progress", data });
        };

        const onVideoError = (event: MessageEvent) => {
            const data = JSON.parse(event.data);
            setData({ type: "video_error", data });
        };

        const onError = (event: Event) => {
            console.error("SSE Error:", event);
            setError(event);
            setLoading(false);
            setConnected(false);
        };

        es.addEventListener("connected", onConnected as EventListener);
        es.addEventListener("heartbeat", onHeartbeat as EventListener);
        es.addEventListener("video_progress", onVideoProgress as EventListener);
        es.addEventListener("video_error", onVideoError as EventListener);
        es.addEventListener("error", onError as EventListener);

        return () => {
            es.removeEventListener("connected", onConnected as EventListener);
            es.removeEventListener("heartbeat", onHeartbeat as EventListener);
            es.removeEventListener("video_progress", onVideoProgress as EventListener);
            es.removeEventListener("video_error", onVideoError as EventListener);
            es.removeEventListener("error", onError as EventListener);
            es.close();
            eventSourceRef.current = null;
            setConnected(false);
        };
    }, [baseUrl]);

    const disconnect = useCallback(() => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
            eventSourceRef.current = null;
        }
        setConnected(false);
        setClientId(null);
    }, []);

    useEffect(() => {
        return () => {
            if (eventSourceRef.current) {
                eventSourceRef.current.close();
            }
        };
    }, []);

    return { 
        data, 
        error, 
        loading, 
        connected, 
        clientId, 
        connect, 
        disconnect 
    };
};
