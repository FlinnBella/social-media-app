import { toast } from "sonner";
import type { TimelineStageResponse, FinalVideoResponse, MultiPartAction } from "#types/multipart";

export const MULTIPART_ACTIONS = {
    SendImageTimeline: 'SendImageTimeline',
    imageTimeline: 'imageTimeline',
    finalVideo: 'finalVideo',
} as const;

export type MultiPartActionsMap = typeof MULTIPART_ACTIONS;

// take in the actual formdata, what action is occuring
// then perform that keeping the server url in mind
export const useMultiPartFormData = async (formData: any, currentAction: MultiPartAction, serverUrl: string): Promise<TimelineStageResponse | FinalVideoResponse> => {

    // use this to handle all form data operations
    switch (currentAction) {
        case MULTIPART_ACTIONS.SendImageTimeline:
            return sendImageTimelineHandler(formData, serverUrl);
        case MULTIPART_ACTIONS.imageTimeline:
            return imageTimelineHandler(formData, serverUrl);
        case MULTIPART_ACTIONS.finalVideo:
            return finalVideoHandler(serverUrl);
        default:
            console.error('Invalid form data type');
            toast.error('Invalid form data type');
            return { ok: false, error: 'Invalid form data type' };
    }

    // these will handle sending and receiving the form-data; 
    // put all the network logic in some function calls

    async function sendImageTimelineHandler(fd: any, _serverUrl: string): Promise<TimelineStageResponse> {
        try {
            if (!(fd instanceof FormData)) {
                const message = 'Expected FormData payload for SendImageTimeline';
                console.error(message);
                toast.error(message);
                return { ok: false, error: message, timeline: [] };
            }
            const response = await fetch(_serverUrl, {
                method: 'POST',
                body: fd,
            });
            const contentType = response.headers.get('content-type') || '';

            if (!response.ok) {
                const errText = contentType.includes('application/json') ? JSON.stringify(await response.json()) : await response.text();
                throw new Error(errText || 'Failed to submit images');
            }

            if (contentType.includes('application/json')) {
                // Expecting intermediary timeline stage
                const result = await response.json();
                const out: TimelineStageResponse = {
                    ok: true,
                    timeline: result?.timeline,
                    batchId: result?.batchId,
                };
                return out;
            } else {
                return { ok: false, error: 'Invalid content type', timeline: [] };
            }

        } catch (error: any) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            console.error(message);
            toast.error(message);
            return { ok: false, error: message, timeline: [] };
        }
    }

    async function imageTimelineHandler(payload: any, _serverUrl: string): Promise<TimelineStageResponse> {
        try {
            const isFormData = payload instanceof FormData;
            const body = isFormData ? payload : JSON.stringify(payload ?? {});
            const headers = isFormData ? {} : { 'Content-Type': 'application/json' } as HeadersInit;
            const response = await fetch(_serverUrl, {
                method: 'POST',
                headers,
                body,
            });
            if (!response.ok) {
                const ct = response.headers.get('content-type') || '';
                const errText = ct.includes('application/json') ? JSON.stringify(await response.json()) : await response.text();
                throw new Error(errText || 'Failed to submit timeline approval');
            }
            const maybeJson = (() => {
                try { return response.json(); } catch { return null; }
            })();
            // do not await twice if it's a promise
            let result: any = null;
            if (maybeJson && typeof (maybeJson as any).then === 'function') {
                result = await (maybeJson as Promise<any>);
            }
            return { ok: true, batchId: result?.batchId };
        } catch (error: any) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            console.error(message);
            toast.error(message);
            return { ok: false, error: message };
        }
    }

    async function finalVideoHandler(sseUrl: string): Promise<FinalVideoResponse> {
        return { ok: true, videoUrl: sseUrl };
    }
};
