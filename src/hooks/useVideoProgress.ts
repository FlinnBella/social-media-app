import { useMemo } from 'react';
import { useSSEContext } from '@/context/VideoProgressContext';
import type { VideoProgressData, VideoErrorData } from '@/features/video-progress/components/VideoProgressBar';

export const useVideoProgress = () => {
  const sse = useSSEContext();

  const videoProgress = useMemo(() => {
    const progressEvents = sse.getEventsByType('video_progress');
    const errorEvents = sse.getEventsByType('video_error');
    
    return {
      progress: progressEvents.length > 0 ? progressEvents[progressEvents.length - 1].data as VideoProgressData : null,
      error: errorEvents.length > 0 ? errorEvents[errorEvents.length - 1].data as VideoErrorData : null,
      isVisible: sse.connection.isVisible && (progressEvents.length > 0 || errorEvents.length > 0),
    };
  }, [sse.connection.events, sse.connection.isVisible, sse.getEventsByType]);

  return {
    ...videoProgress,
    clientId: sse.connection.clientId,
    isConnected: sse.connection.isConnected,
    connect: sse.connect,
    disconnect: sse.disconnect,
  };
};
