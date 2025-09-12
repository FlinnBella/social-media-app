import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { toast } from 'sonner';
import type {  VideoResponse } from '#types/multipart';
import { ZVideoResponseUniversal } from '#types/multipart';
import { useMultiPartFormData, MULTIPART_ACTIONS } from '@/hooks/useMultiPartFormData';
import type { ApiEndpointKey } from '@/cfg';
import { API_ENDPOINTS } from '@/cfg';
import type { TimelineCompositionResponse, Timeline as TimelineType } from '#types/timeline';
import { ZTimelineCompositionResponse } from '#types/timeline';
import { useSSEContext } from './VideoProgressContext';

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
  video: VideoResponse | null;
  messages: Message[];
  timeline: TimelineType | null;
  timelineComposition: TimelineCompositionResponse | null;
  submitTimeline: (prompt: string, images: File[]) => Promise<TimelineCompositionResponse | Error>;
  requestVideo: (apiKey: ApiEndpointKey, prompt: string, images: File[], timelineComposition: TimelineCompositionResponse, clientId?: string) => Promise<VideoResponse | Error>;
  setIsLoading: React.Dispatch<React.SetStateAction<boolean>>;
  setTimeline: React.Dispatch<React.SetStateAction<TimelineType | null>>;
  setTimelineComposition: React.Dispatch<React.SetStateAction<TimelineCompositionResponse | null>>;
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>;
}

const SubmissionContext = createContext<SubmissionContextValue | undefined>(undefined);

export const SubmissionProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [video, setVideo] = useState<VideoResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [timeline, setTimeline] = useState<TimelineType | null>(null);
  const [timelineComposition, setTimelineComposition] = useState<TimelineCompositionResponse | null>(null);
  
  // SSE context for real-time updates
  const sse = useSSEContext();

  const submitTimeline = useCallback(async (prompt: string, images: File[]) => {
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
      const formData = new FormData();
      formData.append('prompt', prompt);
      for (const file of images) {
        formData.append('image', file, file.name);
      }
      const path = API_ENDPOINTS['generateVideoTimeline'];
      const dev_url = `http://localhost:8080${path}`;
      const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.SendImageTimeline, dev_url);
      // When the timeline API responds, update timeline state and assistant message
      // Adapt to the actual response shape when backend finalizes

      // Guard: Check if response is an error or null
      if (resp instanceof Error || resp === null) {
        console.error(resp?.message);
        return resp as Error;
      }

      // Additional Zod validation: Double-check the response structure
      const parsedResponse = ZTimelineCompositionResponse.safeParse(resp);
      if (!parsedResponse.success) {
        const details = parsedResponse.error.issues.map((i) => `${i.path.join('.')}: ${i.message}`).join('; ');
        const message = `Invalid timeline response: ${details || 'failed validation'}`;
        console.error(message);
        toast.error(message);
        return { name: 'Invalid Timeline Response', message: message } as Error;
      }

      // Response is now validated and type-safe
      const validatedResponse = parsedResponse.data;
      const hasTimeline = validatedResponse?.timeline;
      
      if (hasTimeline) {
        setTimeline(hasTimeline);
        // Persist the full timeline composition for later use in video request
        setTimelineComposition(validatedResponse);
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: 'Please review the generated timeline below. Approve to start video rendering.',
          timestamp: new Date(),
        };
        setMessages(prev => [...prev, assistantMessage]);
      }
      
      return validatedResponse;

      //error guard
    } catch (e: any) {
      const err: Error = e instanceof Error ? e : new Error(String(e));
      toast.error(err.message);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: `Sorry, there was an error generating the timeline: ${err.message}`,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const requestVideo = useCallback(async (apiKey: ApiEndpointKey, prompt: string, images: File[], timelineComposition?: TimelineCompositionResponse, clientId?: string) => {

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: prompt || 'Uploaded property photos',
      timestamp: new Date(),
    };
    setMessages(prev => [...prev, userMessage]);

    setIsLoading(true);
    
    // Start SSE connection for progress tracking
    sse.connect('http://localhost:8080/api/sse');
    
    try {

      //create a new form data to shoot up.
      const formData = new FormData();
      formData.append('prompt', prompt);
      for (const file of images) {
        formData.append('image', file, file.name);
      }
      
      // Package timeline composition as JSON and add to FormData
      if (timelineComposition) {
        const timelineJson = JSON.stringify(timelineComposition);
        formData.append('timeline', timelineJson);
        
        // Alternative: You could also create a Blob for better content type handling
        // const timelineBlob = new Blob([timelineJson], { type: 'application/json' });
        // formData.append('timeline', timelineBlob, 'timeline.json');
      }
      // const path = API_ENDPOINTS[apiKey];
/*
Want to dynamically choose eventually 'what' path;
free ffmpeg or veo3, based on the button params; we'll give it a second however,
for now
*/
      // Use clientId from SSE context if not provided
      const effectiveClientId = clientId || sse.connection.clientId;
      const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.finalVideo, apiKey, effectiveClientId);
      
      // Guard: Check if response is an error or null
      if (resp instanceof Error || resp === null) {
        console.error('Video request failed:', resp?.message);
        return resp as Error;
      }

      // Zod validation: Safely parse and validate the response
      const parsedResponse = ZVideoResponseUniversal.safeParse(resp);
      if (!parsedResponse.success) {
        const details = parsedResponse.error.issues.map((i) => `${i.path.join('.')}: ${i.message}`).join('; ');
        const message = `Invalid video response: ${details || 'failed validation'}`;
        console.error(message);
        toast.error(message);
        return { name: 'Invalid Video Response', message: message } as Error;
      }

      // Response is now validated and type-safe
      const validatedResponse = parsedResponse.data;
      const { videoUrl } = validatedResponse;
      
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: videoUrl ? 'Your video is ready!'
          : 'Your video request has been submitted. You will be notified when it is ready.',
        timestamp: new Date(),
        videoUrl,
      };
      setVideo(validatedResponse);
      setMessages(prev => [...prev, assistantMessage]);

      return validatedResponse;
    } catch (e: any) {
      const err: Error = e instanceof Error ? e : new Error(String(e));
      toast.error(err.message);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: `Sorry, there was an error requesting your video: ${err.message}`,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
      throw err;
    } finally {
      setIsLoading(false);
      // Stop SSE connection
      sse.disconnect();
    }
  }, [sse]);

  const value = useMemo<SubmissionContextValue>(() => ({
    isLoading,
    video,
    messages,
    timeline,
    timelineComposition,
    submitTimeline,
    requestVideo,
    setIsLoading,
    setTimeline,
    setTimelineComposition,
    setMessages,
  }), [isLoading, messages, timeline, timelineComposition, video, submitTimeline, requestVideo]);

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


