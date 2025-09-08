import type { ImageSegment } from '@/hooks/useMultiPartFormData';

export const TimelineCard = ({ segment }: { segment: ImageSegment }) => {
    return (
        <div>
            <div>{segment.image}</div>
            <div>{segment.script}</div>
        </div>
    )
}


