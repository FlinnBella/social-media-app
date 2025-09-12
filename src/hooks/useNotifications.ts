import { useMemo } from 'react';
import { useSSEContext } from '@/context/VideoProgressContext';

export const useNotifications = () => {
  const sse = useSSEContext();

  const notifications = useMemo(() => {
    const notificationEvents = sse.getEventsByType('notification');
    const lastNotification = sse.getLastEventByType('notification');
    
    return {
      notifications: notificationEvents,
      lastNotification: lastNotification?.data || null,
      hasNotifications: notificationEvents.length > 0,
      unreadCount: notificationEvents.filter(event => !event.data.read).length,
    };
  }, [sse.connection.events, sse.getEventsByType, sse.getLastEventByType]);

  return {
    ...notifications,
    clientId: sse.connection.clientId,
    isConnected: sse.connection.isConnected,
    connect: sse.connect,
    disconnect: sse.disconnect,
  };
};
