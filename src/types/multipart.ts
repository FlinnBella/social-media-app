import { z } from "zod"

export type MultiPartAction = 'SendImageTimeline' | 'imageTimeline' | 'finalVideo';

export type VideoRequest = {
    formData: FormData;
    apiKey: 'generateVideoReels' | 'generateVideoProReels';
    clientId?: string;
}


export type VideoResponse = {
    videoUrl: string;
}

export const ZVideoResponseUniversal = z.object({
    videoUrl: z.string(),
});


