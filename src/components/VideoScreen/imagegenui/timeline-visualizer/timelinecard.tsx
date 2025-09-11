import type { ImageTimelineSegment, TextSegment } from '#types/timeline';
import { MakeTypeFieldsRequired } from '@/utilites/typeutils';
import { toast } from 'sonner';


export const TimelineCard = ({ segment, script }: { segment: MakeTypeFieldsRequired<ImageTimelineSegment, 'imageUrl' | 'image'>, script?: TextSegment }) => {
    const src = segment.imageUrl;

    if (!src) {
        const err = 'Image URL is required';
        toast.error(err);
        return (<div>Error: {err}</div>)
    }

    return (
        <div className="flex flex-col gap-2">
            <div className="w-full max-w-[200px] aspect-square rounded-lg overflow-hidden bg-black/5 dark:bg-white/5">
                <img
                    src={src}
                    alt="Image composition generated for video"
                    className="w-full h-full object-cover"
                />
            </div>
            {script && <div className="text-sm text-black dark:text-white/80">{script.text}</div>}
        </div>
    )
}


