import { useEffect, useRef, useState } from "react";

export const useSSE = (serverUrl: string) => {
    const [data, setData] = useState<unknown>(null);
    const [error, setError] = useState<unknown>(null);
    const [loading, setLoading] = useState<boolean>(false);
    const eventSourceRef = useRef<EventSource | null>(null);

    useEffect(() => {
        if (!serverUrl) return;

        setLoading(true);
        const es = new EventSource(serverUrl);
        eventSourceRef.current = es;

        const onImageTimeline = (event: MessageEvent) => {
            setData({ type: "imageTimeline", payload: event.data });
        };
        const onProgressUpdate = (event: MessageEvent) => {
            setData({ type: "progressUpdate", payload: event.data });
        };
        const onBatchComplete = (event: MessageEvent) => {
            setData({ type: "batchComplete", payload: event.data });
            setLoading(false);
        };
        const onError = (event: Event) => {
            setError(event);
            setLoading(false);
        };

        es.addEventListener("imageTimeline", onImageTimeline as EventListener);
        es.addEventListener("progressUpdate", onProgressUpdate as EventListener);
        es.addEventListener("batchComplete", onBatchComplete as EventListener);
        es.addEventListener("error", onError as EventListener);

        return () => {
            es.removeEventListener("imageTimeline", onImageTimeline as EventListener);
            es.removeEventListener("progressUpdate", onProgressUpdate as EventListener);
            es.removeEventListener("batchComplete", onBatchComplete as EventListener);
            es.removeEventListener("error", onError as EventListener);
            es.close();
            eventSourceRef.current = null;
        };
    }, [serverUrl]);

    return { data, error, loading };
};