import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { toast } from 'sonner';
import type {  FinalVideoResponse } from '#types/multipart';
import { submitVideoRequest } from '@/hooks/submitVideoRequest';
import type { ApiEndpointKey } from '@/cfg';
import type { TimelineCompositionResponse, TimelineSegments } from '#types/timeline';

/*
Sources of errors perhaps emerge with the types here;
need to ensure that the response objects and interfaces are compatible, 
and guard to ensure that it can fit the schema of the interfaces
*/

export interface Message {
  id: string;
  type: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  videoUrl?: string;
  socialLinks?: {
    instagram: string;
    tiktok: string;
    twitter: string;
    facebook: string;
  };
}

interface SubmissionContextValue {
  isLoading: boolean;
  messages: Message[];
  timelineSegments: TimelineSegments[] | null;
  submit: (apiKey: ApiEndpointKey, prompt: string, images: File[]) => Promise<TimelineCompositionResponse | FinalVideoResponse | Error>;
  setTimelineSegments: React.Dispatch<React.SetStateAction<TimelineSegments[] | null>>;
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>;
}

const SubmissionContext = createContext<SubmissionContextValue | undefined>(undefined);

export const SubmissionProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [isLoading, setIsLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [timelineSegments, setTimelineSegments] = useState<TimelineSegments[] | null>(null);

  const submit = useCallback(async (apiKey: ApiEndpointKey, prompt: string, images: File[]) => {
    if (!prompt || !prompt.trim()) {
      const message = 'Please enter a property description';
      toast.error(message);
      return { name: 'No prompt entered', message: message } as Error;
    }
    if (!images || images.length === 0) {
      const message = 'Please upload at least one image';
      toast.error(message);
      return { name: 'No images uploaded', message: message } as Error;
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: prompt || 'Uploaded property photos',
      timestamp: new Date(),
    };
    setMessages(prev => [...prev, userMessage]);
    setIsLoading(true);
    try {
      const resp = await submitVideoRequest({ prompt, images, apiKey });
      if ('timeline' in resp && resp.timeline && Array.isArray(resp.timeline) && resp.timeline.length > 0) {
        setTimelineSegments(resp.timeline);
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: 'Please review the generated timeline below. Approve to start video rendering.',
          timestamp: new Date(),
        };
        setMessages(prev => [...prev, assistantMessage]);
      } else if (!resp) {
        toast.error( 'Failed to submit images');
      }
      return resp;
    } catch (e: any) {
      const err: Error = e instanceof Error ? e : new Error(String(e));
      toast.error(err.message);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: `Sorry, there was an error generating your property video: ${err.message}`,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const value = useMemo<SubmissionContextValue>(() => ({
    isLoading,
    messages,
    timelineSegments,
    submit,
    setTimelineSegments,
    setMessages,
  }), [isLoading, messages, timelineSegments, submit]);

  return (
    <SubmissionContext.Provider value={value}>
      {children}
    </SubmissionContext.Provider>
  );
};

export function useSubmission() {
  const ctx = useContext(SubmissionContext);
  if (!ctx) throw new Error('useSubmission must be used within a SubmissionProvider');
  return ctx;
}


