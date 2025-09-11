
import { z } from 'zod';

/*

Parent type

*/

export type TimelineCompositionResponse = {
    metadata: Metadata;
    theme: Theme;
    timeline: Timeline;
    music: Music;
}

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

export type Transition = {
    effect: string;
    easing: string;
}

export type TextStyle = {
    fontFamily: string;
    textStyle: string;
}

export type TextSegment = {
    text: string;
    startTime: number;
    duration: number;
    position: string;
    narrativeSource: string;
}

export type TextTimeline = {
    TextStyle: TextStyle;
    TextSegments: TextSegment[];
}

export type ImageTimeline = ImageTimelineSegment[];

export type ImageTimelineSegment = {
    ordering: number;
    startTime: number;
    duration: number;
    imageIndex?: number;
    transition?: Transition;
    description?: string;
    // UI-only field to attach a Blob for preview; not sent by backend
    imageUrl?: string;
    image?: File;
}

export type Timeline = {
    totalDuration: number;
    ImageTimeline: ImageTimeline;
    TextTimeline: TextTimeline;
}

export type Music = {
        enabled: boolean;
        genre: string;
        volume?: number;
}

// Zod schemas aligned to backend schema
export const ZTransition = z.object({
    effect: z.string(),
    easing: z.string(),
});

export const ZTextStyle = z.object({
    fontFamily: z.string(),
    textStyle: z.string(),
});

export const ZTextSegment = z.object({
    text: z.string(),
    startTime: z.number(),
    duration: z.number(),
    position: z.string(),
    narrativeSource: z.string(),
});

export const ZTextTimeline = z.object({
    TextStyle: ZTextStyle,
    TextSegments: z.array(ZTextSegment),
});


export const ZImageTimelineSegment = z.looseObject({
    ordering: z.number(),
    startTime: z.number(),
    duration: z.number(),
    imageIndex: z.number().optional(),
    transition: ZTransition.optional(),
    description: z.string().optional(),
    imageUrl: z.any().optional(),
    images: z.array(z.any()).optional(),
});


export const ZImageTimeline = z.array(ZImageTimelineSegment);

export const ZTimeline = z.object({
    totalDuration: z.number(),
    ImageTimeline: ZImageTimeline,
    TextTimeline: ZTextTimeline,
});

export const ZMetadata = z.object({
    totalDuration: z.number(),
    aspectRatio: z.string(),
    fps: z.string(),
    resolution: z.array(z.number()),
});

export const ZTheme = z.object({
    style: z.string(),
    grading: z.string(),
});

export const ZMusic = z.object({
        enabled: z.boolean(),
        genre: z.string(),
        volume: z.number().optional(),
    });

export const ZTimelineCompositionResponse = z.object({
    metadata: ZMetadata,
    theme: ZTheme,
    timeline: ZTimeline,
    music: ZMusic,
});