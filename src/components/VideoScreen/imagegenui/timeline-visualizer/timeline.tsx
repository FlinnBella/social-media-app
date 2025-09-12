// no local state needed
import { TimelineCard } from './timelinecard';
import type { ImageTimelineSegment, TextSegment } from '#types/timeline';
import { FFMpegRequestButton } from '@/components/api-request-buttons/FFMpegRequestButton';
import { VeoRequestButton } from '@/components/api-request-buttons/VeoRequestButton';
import { useSubmission } from '@/context/SubmissionContext';
import { MakeTypeFieldsRequired } from '@/utilites/typeutils';
import Xarrow, { Xwrapper } from 'react-xarrows';
import { Spinner } from '@/components/ui/shadcn-io/spinner';

interface TimelineProps {
    ImageSegments: ImageTimelineSegment[] | undefined;
    TextSegments?: TextSegment[];
    prompt?: string;
    images?: File[];
    previewUrls?: string[];
}

export const Timeline = ({ ImageSegments, TextSegments, prompt, images, previewUrls }: TimelineProps) => {
const { isLoading } = useSubmission();

/*
Logic kerfuffle 
I fucking hate coding; bunch of autistic shitheads in this field;
I hope to leave the industry soon :) 
*/

const MapImageSegments = (
    input: ImageTimelineSegment[] | undefined,
    TextSegments: TextSegment[] | undefined,
    images: File[] | undefined,
    previewUrls: string[] | undefined
) => {
    // Guard: coerce to array safely (accept wrapper object shape or direct array)
    if (!input) {
        console.error('No image segments found');
        return null;
    }
    const segmentsArray = input

    const sortedImageSegments = [...segmentsArray].sort((a, b) => a.ordering - b.ordering);
    // Ensure TextSegments exists before indexing
    if (!TextSegments) {
        console.error('No text segments found');
        return null;
    }
    return sortedImageSegments.map((segment, idx) => {
        const matchedScript = TextSegments.find(ts => ts && ts.startTime !== undefined) ? TextSegments[segment.ordering] : undefined;
        const fileForSegment = images?.[segment.ordering];
        const imageUrlForSegment = previewUrls?.[segment.ordering];
        const segmentWithFile: MakeTypeFieldsRequired<ImageTimelineSegment, 'image' | 'imageUrl'> = {
            ...segment,
            imageUrl: imageUrlForSegment || segment.imageUrl || '',
            image: fileForSegment as File
        };
        const aboveId = `segment-${segment.ordering}-above`;
        const cardId = `segment-${segment.ordering}`;
        const prev = sortedImageSegments[idx - 1];
        const prevCardId = prev ? `segment-${prev.ordering}` : undefined;
        return (
            <div key={segment.ordering} className="flex flex-col items-start gap-1">
                <div id={aboveId} className="h-2 w-8" />
                <div className="w-28 aspect-square" id={cardId}>
                    <TimelineCard key={segment.ordering} segment={segmentWithFile} script={matchedScript} />
                </div>
                {prevCardId && cardId&& <Xarrow lineColor={"black"} start={prevCardId} end={cardId} />}
            </div>
        )
    })
}

    return(
        <div>
            <div className={isLoading ? 'opacity-50' : 'bg-white rounded-lg p-4'}>
            <div>Visualize your shorts timeline!</div>
            <Xwrapper>
            <div className="flex flex-row flex-wrap gap-3"> {MapImageSegments(ImageSegments, TextSegments, images, previewUrls)} </div>
            </Xwrapper>
            <div className="flex flex-row items-center justify-center flex-wrap gap-3">
                {isLoading ? <p>Generating Video...</p> : <p> If everything looks good, click a button to generate your video!</p>}
                <div className="flex items-center gap-2">
                {isLoading ? (<> <Spinner variant="circle" /></>) :
                <>
                    <FFMpegRequestButton 
                        prompt={prompt || ''}
                        images={images || []}
                        disabled={isLoading}
                        className=""
                    />
                    <VeoRequestButton 
                        prompt={prompt || ''}
                        images={images || []}
                        disabled={isLoading}
                        className=""
                    />
                    </>
                }
                </div>
            </div>
            </div>
        </div>
    )
}


