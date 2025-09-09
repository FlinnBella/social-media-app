import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { TimelineCard } from './timelinecard';
import type { ImageSegment } from '#types/timeline';
// when it's invoked, it's going to be done so via a fetch call
//fetch call is going to fill it with a bunch of data,
//data gets decomposed to the image segments
//and all from the multipart form data
export const Timeline = ({ ImageSegments }: { ImageSegments: ImageSegment[] }) => {
const [isLoading, setIsLoading] = useState(false);

const MapImageSegments = (ImageSegments: ImageSegment[]) => {
    //want it to order the divs based off of the ordering property
    const sortedImageSegments = [...ImageSegments].sort((a, b) => a.ordering - b.ordering);
    return sortedImageSegments.map((segment) => {
        return <TimelineCard key={segment.imageIndex} segment={segment} />
    })
}

const GenerateVideo = async () => {
    setIsLoading(true);
    await fetch('/api/generate-video', {
        method: 'POST',
        body: new FormData()
    });
    setIsLoading(false);
}

    return(
        <div>
            <div className={isLoading ? 'opacity-50' : ''}>
            <div>Visualize your shorts timeline!</div>

            {/* Going to need to ensure these arranged properly */}
            <div> {MapImageSegments(ImageSegments)} </div>

            <div> <p> If everything looks good, click the button to generate your video!</p>
            <div> <Button onClick={GenerateVideo}>Generate Video</Button> </div> </div>
            </div>
            {isLoading && <div>Loading...</div>}

        </div>
    )
}