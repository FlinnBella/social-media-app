import type { ImageSegment, TextSegment } from '#types/timeline';

export const TimelineCard = ({ segment, script }: { segment: ImageSegment, script: TextSegment }) => {

    const imageUrl = segment.imageUrl ? URL.createObjectURL(segment.imageUrl) : undefined;
    return (
        <div>
            <div><img src={imageUrl} alt="Image composition generated for video" /></div>
            <div>{script.text}</div>
        </div>
    )
}