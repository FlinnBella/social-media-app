export type MultiPartAction = 'SendImageTimeline' | 'imageTimeline' | 'finalVideo';


export interface FinalVideoResponse {
    ok: boolean;
    error?: string;
    videoUrl?: string;
}


