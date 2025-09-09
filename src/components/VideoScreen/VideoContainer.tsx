import React from 'react';
import { Monitor, Smartphone } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { ImageSegment, TextSegment } from '#types/timeline';
import { Timeline } from '@/components/VideoScreen/imagegenui/timeline-visualizer/timeline';
// removed unused types

interface VideoContainerProps {
    timelineSegments: ImageSegment[] | null;
    videoUrl?: string | null;
    isMobile: boolean;
    prompt?: string;
    images?: File[];
}

export const VideoContainer: React.FC<VideoContainerProps> = ({ timelineSegments, videoUrl, isMobile, prompt, images }) => {
    //shoudl be binary being streamed to the client; 
    //need to convert it to a blob url for the src to render
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
                    {timelineSegments ? (
                        <Timeline 
                            ImageSegments={timelineSegments}
                            prompt={prompt}
                            images={images}
                        />
                    ) : videoUrl ? (
                        <video
                            src={videoUrl}
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


