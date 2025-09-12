import React, { createContext, useContext, useState, useCallback, useRef, useEffect } from 'react';

// Generic SSE event types
export interface SSEEvent {
  type: string;
  data: any;
  timestamp?: number;
}

export interface SSEConnection {
  clientId: string | null;
  isConnected: boolean;
  isVisible: boolean;
  events: SSEEvent[];
  lastEvent: SSEEvent | null;
}

interface SSEContextValue {
  connection: SSEConnection;
  connect: (sseUrl: string, clientId?: string) => void;
  disconnect: () => void;
  clearEvents: () => void;
  // Event filtering helpers
  getEventsByType: (type: string) => SSEEvent[];
  getLastEventByType: (type: string) => SSEEvent | null;
}

const SSEContext = createContext<SSEContextValue | undefined>(undefined);

export const SSEProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [connection, setConnection] = useState<SSEConnection>({
    clientId: null,
    isConnected: false,
    isVisible: false,
    events: [],
    lastEvent: null,
  });
  
  const eventSourceRef = useRef<EventSource | null>(null);

  const connectSSE = useCallback((sseUrl: string, providedClientId?: string) => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const clientId = providedClientId || `client_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    const es = new EventSource(`${sseUrl}?client_id=${clientId}`);
    eventSourceRef.current = es;

    const onMessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data);
        const sseEvent: SSEEvent = {
          type: event.type || 'message',
          data,
          timestamp: Date.now(),
        };

        setConnection(prev => ({
          ...prev,
          lastEvent: sseEvent,
          events: [...prev.events, sseEvent],
        }));
      } catch (error) {
        console.error('Failed to parse SSE message:', error);
      }
    };

    const onOpen = () => {
      setConnection(prev => ({
        ...prev,
        clientId,
        isConnected: true,
        isVisible: true,
      }));
    };

    const onError = (event: Event) => {
      console.error("SSE Error:", event);
      setConnection(prev => ({
        ...prev,
        isConnected: false,
        lastEvent: {
          type: 'error',
          data: { error: 'Connection lost. Please try again.' },
          timestamp: Date.now(),
        },
      }));
    };

    es.addEventListener("open", onOpen);
    es.addEventListener("message", onMessage);
    es.addEventListener("error", onError);

    // Add listeners for specific event types
    const eventTypes = ['connected', 'video_progress', 'video_error', 'heartbeat', 'timeline_update', 'notification'];
    eventTypes.forEach(eventType => {
      es.addEventListener(eventType, onMessage);
    });

    return () => {
      es.removeEventListener("open", onOpen);
      es.removeEventListener("message", onMessage);
      es.removeEventListener("error", onError);
      eventTypes.forEach(eventType => {
        es.removeEventListener(eventType, onMessage);
      });
      es.close();
      eventSourceRef.current = null;
      setConnection(prev => ({
        ...prev,
        isConnected: false,
      }));
    };
  }, []);

  const connect = useCallback((sseUrl: string, clientId?: string) => {
    connectSSE(sseUrl, clientId);
  }, [connectSSE]);

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setConnection(prev => ({
      ...prev,
      isConnected: false,
      isVisible: false,
    }));
  }, []);

  const clearEvents = useCallback(() => {
    setConnection(prev => ({
      ...prev,
      events: [],
      lastEvent: null,
    }));
  }, []);

  const getEventsByType = useCallback((type: string) => {
    return connection.events.filter(event => event.type === type);
  }, [connection.events]);

  const getLastEventByType = useCallback((type: string) => {
    const events = connection.events.filter(event => event.type === type);
    return events.length > 0 ? events[events.length - 1] : null;
  }, [connection.events]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, []);

  const value: SSEContextValue = {
    connection,
    connect,
    disconnect,
    clearEvents,
    getEventsByType,
    getLastEventByType,
  };

  return (
    <SSEContext.Provider value={value}>
      {children}
    </SSEContext.Provider>
  );
};

export const useSSEContext = () => {
  const context = useContext(SSEContext);
  if (context === undefined) {
    throw new Error('useSSEContext must be used within a SSEProvider');
  }
  return context;
};

// Backward compatibility
export const useVideoProgressContext = useSSEContext;
export const VideoProgressProvider = SSEProvider;
