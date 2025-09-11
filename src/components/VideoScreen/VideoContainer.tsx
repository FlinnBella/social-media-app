import React, { useCallback, useState } from 'react';
import { Monitor, Smartphone } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Timeline as TimelineType } from '#types/timeline';
import { Timeline } from '@/components/VideoScreen/imagegenui/timeline-visualizer/timeline';
import { useSubmission } from '@/context/SubmissionContext';
import type { VideoResponse } from '#types/multipart';
// removed unused types

interface TimelineVideoContainerProps {
    timeline: TimelineType | null;
    isMobile: boolean;
    prompt?: string;
    images: File[];
    previewUrls: string[];
}

//Partial here as a workaround to signify 
//that the videourl is serpate form the timeline props
interface RenderVideoProps {
    videoUrl: string;
}

type Props = TimelineVideoContainerProps & Partial<RenderVideoProps>;

export const VideoContainer: React.FC<Props> = ({ timeline, isMobile, prompt, images, previewUrls } : Props) => {
    const { video, requestVideo } = useSubmission();
    //shoudl be binary being streamed to the client; 
    return (
        <div className={cn(
            "bg-gray-900 rounded-2xl p-4 shadow-2xl",
            isMobile ? "w-full max-w-sm" : "w-full max-w-4xl"
        )}>
            <div className={cn(
                "bg-black rounded-xl relative overflow-hidden",
                isMobile ? "aspect-[9/16] max-h-[600px]" : "aspect-video h-[400px]"
            )}>
                <div className="w-full h-full flex items-center justify-center bg-gray-900">

                    {/* Complex Check; perhaps just put the isVideo instead */}
                    {/* As a state type, to pass it on anyways */}
                    {/* As Javascript will eval null or truthy*/}
                    {(timeline && !video) ? (
                        <Timeline 
                            ImageSegments={timeline.ImageTimeline}
                            TextSegments={timeline.TextTimeline.TextSegments}
                            prompt={prompt}
                            images={images}
                            previewUrls={previewUrls}
                        />
                    ) : (video && video.videoUrl) ? (
                        <video
                            src={video.videoUrl}
                            controls
                            className="w-full h-full object-cover rounded-xl"
                            autoPlay
                            muted
                            loop
                        />
                    ) : (
                        <div className="text-center text-gray-400 p-8">
                            {isMobile ? (
                                <Smartphone className="w-16 h-16 mx-auto mb-4 opacity-50" />
                            ) : (
                                <Monitor className="w-16 h-16 mx-auto mb-4 opacity-50" />
                            )}
                            <p className="text-sm">Your AI-generated property video will appear here</p>
                            <p className="text-xs mt-2 opacity-75">
                                Optimized for {isMobile ? "mobile viewing" : "desktop presentation"}
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default VideoContainer;


