export type MultiPartAction = 'SendImageTimeline' | 'imageTimeline' | 'finalVideo';

export interface ImageSegment {
    id: string;
    ordering: number;
    image: string;
    script: string;
}

export interface TimelineStageResponse {
    ok: boolean;
    error?: string;
    timeline?: ImageSegment[];
    batchId?: string;
}

export interface FinalVideoResponse {
    ok: boolean;
    error?: string;
    videoUrl?: string;
}


