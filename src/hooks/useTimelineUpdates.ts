import { useMemo } from 'react';
import { useSSEContext } from '@/context/SSEContext';

export const useTimelineUpdates = () => {
  const sse = useSSEContext();

  const timelineUpdates = useMemo(() => {
    const timelineEvents = sse.getEventsByType('timeline_update');
    const lastTimelineEvent = sse.getLastEventByType('timeline_update');
    
    return {
      updates: timelineEvents,
      lastUpdate: lastTimelineEvent?.data || null,
      hasUpdates: timelineEvents.length > 0,
    };
  }, [sse.connection.events, sse.getEventsByType, sse.getLastEventByType]);

  return {
    ...timelineUpdates,
    clientId: sse.connection.clientId,
    isConnected: sse.connection.isConnected,
    connect: sse.connect,
    disconnect: sse.disconnect,
  };
};
