
/*

Parent type

*/

export type TimelineCompositionResponse = {
    metadata: Metadata;
    theme: Theme;
    timelineSegments: TimelineSegments;
    music: Music;
}

/*
Child types; but types allow for some extensibility
*/

export type Metadata = {
    totalDuration: number;
    aspectRatio: string;
    fps: string;
    resolution: number[];
}

export type Theme = {
    style: string;
    grading: string;
}

export type TimelineSegments = {
    imageSegments: ImageSegment[];
    textSegments: TextSegment[];
}

export type ImageSegment = {
    ordering: number;
    startTime: number;
    duration: number;
    imageUrl?: Blob;
}

export type TextSegment = {
    text: string;
    startTime: number;
    duration: number;
}

export type Music = {
    enabled: boolean;
    genre: string;
}