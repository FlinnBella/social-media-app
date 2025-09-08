import type { ImageSegment } from '#types/multipart';

export const TimelineCard = ({ segment }: { segment: ImageSegment }) => {
    return (
        <div>
            <div>{segment.image}</div>
            <div>{segment.script}</div>
        </div>
    )
}