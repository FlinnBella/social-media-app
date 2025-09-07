import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { TimelineCard } from './timelinecard';
// when it's invoked, it's going to be done so via a fetch call
//fetch call is going to fill it with a bunch of data,
//data gets decomposed to the image segments
//and all from the multipart form data
export const Timeline = (ImageSegments: ImageSegment[]) => {
const [isLoading, setIsLoading] = useState(false);

const MapImageSegments = (ImageSegments: ImageSegment[]) => {
    //want it to order the divs based off of the ordering property
    const sortedImageSegments = ImageSegments.sort((a, b) => a.ordering - b.ordering);
    return sortedImageSegments.map((ImageSegment) => {
        return <TimelineCard ImageSegment={ImageSegment} />
    })
}

const GenerateVideo = async () => {
    setIsLoading(true);

    //method here needs to be changed somehow
    //to reflect that it should just be 
    //a websocket call

    const response = await fetch('/api/generate-video', {
        method: 'POST',
        body: FormData
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