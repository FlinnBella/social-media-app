import React, { createContext, useCallback, useContext, useMemo, useState } from 'react';
import { toast } from 'sonner';
import type {  VideoResponse, VideoRequest } from '#types/multipart';
import { useMultiPartFormData, MULTIPART_ACTIONS } from '@/hooks/useMultiPartFormData';
import type { ApiEndpointKey } from '@/cfg';
import { API_ENDPOINTS } from '@/cfg';
import type { TimelineCompositionResponse, Timeline as TimelineType } from '#types/timeline';

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
  submitTimeline: (prompt: string, images: File[]) => Promise<TimelineCompositionResponse | Error>;
  requestVideo: (apiKey: ApiEndpointKey, prompt: string, images: File[]) => Promise<VideoResponse | Error>;
  setIsLoading: React.Dispatch<React.SetStateAction<boolean>>;
  setTimeline: React.Dispatch<React.SetStateAction<TimelineType | null>>;
  setMessages: React.Dispatch<React.SetStateAction<Message[]>>;
}

const SubmissionContext = createContext<SubmissionContextValue | undefined>(undefined);

export const SubmissionProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const [video, setVideo] = useState<VideoResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [timeline, setTimeline] = useState<TimelineType | null>(null);

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
      const url = import.meta.env.PROD ? path : `http://localhost:8080${path}`;
      const dev_url = `http://localhost:8080${path}`;
      const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.SendImageTimeline, dev_url);
      // When the timeline API responds, update timeline state and assistant message
      // Adapt to the actual response shape when backend finalizes

      //guard
      if (resp instanceof Error || resp === null) {
        console.error(resp?.message);
        return resp as Error;
      }
      //guard end


      //testing
      console.log(resp);
      //testing end

      const hasTimeline = (resp as TimelineCompositionResponse)?.timeline;
      if (hasTimeline) {
        setTimeline(hasTimeline);
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: 'Please review the generated timeline below. Approve to start video rendering.',
          timestamp: new Date(),
        };
        setMessages(prev => [...prev, assistantMessage]);
      }
      return resp as TimelineCompositionResponse | Error;

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

  const requestVideo = useCallback(async (apiKey: ApiEndpointKey, prompt: string, images: File[]) => {

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: prompt || 'Uploaded property photos',
      timestamp: new Date(),
    };
    setMessages(prev => [...prev, userMessage]);

    setIsLoading(true);
    try {

      //create a new form data to shoot up.
      const formData = new FormData();
      formData.append('prompt', prompt);
      for (const file of images) {
        formData.append('image', file, file.name);
      }
      const path = API_ENDPOINTS[apiKey];
      const url = import.meta.env.PROD ? path : `http://localhost:8080${path}`;
/*
Want to dynamically choose eventually 'what' path;
free ffmpeg or veo3, based on the button params; we'll give it a second however,
for now
*/
      const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.finalVideo, API_ENDPOINTS.generateVideoProReels);
      // Final video request may return a URL or success boolean
      if ((resp as VideoResponse)) {
        const { videoUrl } = resp as VideoResponse;
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: videoUrl ? 'Your video is ready!'
            : 'Your video request has been submitted. You will be notified when it is ready.',
          timestamp: new Date(),
          videoUrl,
        };
        setVideo(resp as VideoResponse);
        setMessages(prev => [...prev, assistantMessage]);
      }

      return resp as VideoResponse | Error;
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
    }
  }, []);

  const value = useMemo<SubmissionContextValue>(() => ({
    isLoading,
    video,
    messages,
    timeline,
    submitTimeline,
    requestVideo,
    setIsLoading,
    setTimeline,
    setMessages,
  }), [isLoading, messages, timeline, video, submitTimeline, requestVideo]);

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


