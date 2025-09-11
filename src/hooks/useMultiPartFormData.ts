import { toast } from "sonner";
import type {  VideoResponse, MultiPartAction, VideoRequest } from "#types/multipart";
import type { TimelineCompositionResponse } from "#types/timeline";
import { ZTimelineCompositionResponse } from "#types/timeline";
import {unWrapN8N} from "@/hooks/unWrapN8N";
import {API_ENDPOINTS} from "@/cfg";
import { ZVideoResponseUniversal } from "#types/multipart";

export const MULTIPART_ACTIONS = {
    SendImageTimeline: 'SendImageTimeline',
    imageTimeline: 'imageTimeline',
    finalVideo: 'finalVideo',
} as const;

export type MultiPartActionsMap = typeof MULTIPART_ACTIONS;

// take in the actual formdata, what action is occuring
// then perform that keeping the server url in mind
export const useMultiPartFormData = async (formData: any, currentAction: MultiPartAction, ...cfg: any): Promise<TimelineCompositionResponse | VideoResponse | Error | null > => {

    // use this to handle all form data operations
    switch (currentAction) {
        case MULTIPART_ACTIONS.SendImageTimeline:
            return sendImageTimelineHandler(formData);
        case MULTIPART_ACTIONS.imageTimeline:
            return imageTimelineHandler(formData);
        case MULTIPART_ACTIONS.finalVideo:
            return finalVideoHandler({formData, apiPath: cfg[0]});
        default:
            console.error('Invalid form data type');
            toast.error('Invalid form data type');
            return { name: "Invalid Form Data", message: 'Invalid form data type' } as Error;
    }

    // these will handle sending and receiving the form-data; 
    // put all the network logic in some function calls

    async function sendImageTimelineHandler(fd: any): Promise<TimelineCompositionResponse | Error> {
        try {
            if (!(fd instanceof FormData)) {
                const message = 'Expected FormData payload for SendImageTimeline';
                console.error(message);
                toast.error(message);
                return { name: 'Expected FormData payload for SendImageTimeline', message: message } as Error;
            }
            const response = await fetch(API_ENDPOINTS.generateVideoTimeline, {
                method: 'POST',
                body: fd,
            });
            const contentType = response.headers.get('content-type') || '';

            if (!response.ok) {
                const errText = contentType.includes('application/json') ? JSON.stringify(await response.json()) : await response.text();
                throw new Error(errText || 'Failed to submit images');
            }

            if (response.status >= 400) {
                throw new Error('Error submitting images'); 
            }

            if (contentType.includes('application/json')) {
                // Expecting intermediary timeline stage

                const result = await response.json();


                // Validate and transform upstream schema into UI shape
                const parsed = ZTimelineCompositionResponse.safeParse(unWrapN8N(result));
                if (!parsed.success) {
                    const details = parsed.error.issues.map((i) => `${i.path.join('.')}: ${i.message}`).join('; ');
                    const message = `Invalid timeline response: ${details || 'failed validation'}`;
                    toast.error(message);
                    throw new Error(message);
                }

                return unWrapN8N(result) as TimelineCompositionResponse;
            } else {
                throw new Error('Invalid content type');
            }

        } catch (error: any) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            console.error(message);
            toast.error(message);
            return { name: 'Unknown error', message: message } as Error;
        }
    }

    async function imageTimelineHandler(payload: any): Promise<TimelineCompositionResponse | Error> {
        try {
            const isFormData = payload instanceof FormData;
            const body = isFormData ? payload : JSON.stringify(payload ?? {});
            const headers = isFormData ? {} : { 'Content-Type': 'application/json' } as HeadersInit;
            const response = await fetch(API_ENDPOINTS.generateVideoTimeline, {
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
            if (maybeJson && typeof (maybeJson as any).then === 'function') {
                await (maybeJson as Promise<any>);
            }
            return { name: 'Unknown error', message: "Unknown error occured! No timeline returned!" } as Error;
        } catch (error: any) {
            const message = error instanceof Error ? error.message : 'Unknown error';
            console.error(message);
            toast.error(message);
            return { name: 'Unknown error', message: message } as Error;
        }
    }

/*
Shoot images and prompt up to the server
*/
    async function finalVideoHandler(VideoRequest: VideoRequest): Promise<VideoResponse> {
        switch (VideoRequest.apiPath) {
            case 'generateVideoReels': {
                const res = await fetch(API_ENDPOINTS.generateVideoReels, {
                    method: 'POST',
                    body: VideoRequest.formData,
                });
                if (!res.ok) {
                    throw new Error(`generateVideoReels failed: ${res.status} ${res.statusText}`);
                }
                const blob = await res.blob();
                const objectUrl = URL.createObjectURL(blob);
                const result = { videoUrl: objectUrl } as const;
                const parsed = ZVideoResponseUniversal.safeParse(result);
                if (!parsed.success) {
                    URL.revokeObjectURL(objectUrl);
                    throw new Error('Invalid response');
                }
                return parsed.data;
            }
            case 'generateVideoProReels': {
                const res = await fetch(API_ENDPOINTS.generateVideoProReels, {
                    method: 'POST',
                    body: VideoRequest.formData,
                });
                if (!res.ok) {
                    throw new Error(`generateVideoProReels failed: ${res.status} ${res.statusText}`);
                }
                const blob = await res.blob();
                const objectUrl = URL.createObjectURL(blob);
                const result = { videoUrl: objectUrl } as const;
                const parsed = ZVideoResponseUniversal.safeParse(result);
                if (!parsed.success) {
                    URL.revokeObjectURL(objectUrl);
                    throw new Error('Invalid response');
                }
                return parsed.data;
            }
            default:
                throw new Error('Invalid API path');
        }
    }
};
