// no local state needed
import { TimelineCard } from './timelinecard';
import type { ImageSegment } from '#types/multipart';
import { FFMpegRequestButton } from '@/components/api-request-buttons/FFMpegRequestButton';
import { VeoRequestButton } from '@/components/api-request-buttons/VeoRequestButton';
import { useSubmission } from '@/context/SubmissionContext';

interface TimelineProps {
    ImageSegments: ImageSegment[];
    prompt?: string;
    images?: File[];
}

export const Timeline = ({ ImageSegments, prompt, images }: TimelineProps) => {
const { isLoading } = useSubmission();

const MapImageSegments = (ImageSegments: ImageSegment[]) => {
    const sortedImageSegments = [...ImageSegments].sort((a, b) => a.ordering - b.ordering);
    return sortedImageSegments.map((segment) => {
        return <TimelineCard key={segment.id} segment={segment} />
    })
}

    return(
        <div>
            <div className={isLoading ? 'opacity-50' : ''}>
            <div>Visualize your shorts timeline!</div>

            <div> {MapImageSegments(ImageSegments)} </div>

            <div>
                <p> If everything looks good, click a button to generate your video!</p>
                <div className="flex items-center gap-2">
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
                </div>
            </div>
            </div>
            {isLoading && <div>Loading...</div>}

        </div>
    )
}


