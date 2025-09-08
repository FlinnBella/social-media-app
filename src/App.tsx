import React, { useState, useRef, useEffect } from 'react';
import { toast } from 'sonner';
import { Upload, Camera } from 'lucide-react';
import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';
import { useAutoResizeTextarea } from '@/hooks/use-auto-resize-textarea';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
 
import SocialSharePanel from '@/components/SocialSharePanel';
import { useMultiPartFormData, MULTIPART_ACTIONS } from '@/hooks/useMultiPartFormData';
import type { TimelineStageResponse, FinalVideoResponse, ImageSegment } from '@/hooks/useMultiPartFormData';
import VideoContainer from '@/components/VideoScreen/VideoContainer';

interface Message {
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


function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [previewUrls, setPreviewUrls] = useState<string[]>([]);
  const [timelineSegments, setTimelineSegments] = useState<ImageSegment[] | null>(null);
  const [pendingApproval, setPendingApproval] = useState(false);
  const [approveUrl, setApproveUrl] = useState<string | null>(null);
  const [sseUrl, setSseUrl] = useState<string | null>(null);
  
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);
  const previewUrlsRef = useRef<string[]>([]);
  
  const { textareaRef, adjustHeight } = useAutoResizeTextarea({
    minHeight: 72,
    maxHeight: 300,
  });

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  useEffect(() => {
    // Revoke previously created URLs before creating new ones
    previewUrlsRef.current.forEach((url) => {
      try { URL.revokeObjectURL(url); } catch {}
    });
    const urls = selectedFiles.map((file) => URL.createObjectURL(file));
    previewUrlsRef.current = urls;
    setPreviewUrls(urls);
    return () => {
      urls.forEach((url) => {
        try { URL.revokeObjectURL(url); } catch {}
      });
    };
  }, [selectedFiles]);

  const removeSelectedFile = (index: number) => {
    setSelectedFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const clearAllSelected = () => {
    setSelectedFiles([]);
  };

  
  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files) return;
    const imageFiles = Array.from(files).filter((f) => f.type.startsWith('image/'));
    if (imageFiles.length === 0) {
      toast.error('Please select image files');
      return;
    }
    setSelectedFiles((prev) => [...prev, ...imageFiles]);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      // Don't submit here, let the form handle it
    }
  };

  const handleSubmit = async (apiType: 'ffmpeg' | 'veo3') => {
    if (!inputText.trim() || selectedFiles.length === 0) {
      return toast.error('Please enter a property description and upload photos');
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputText || 'Uploaded property photos',
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setIsLoading(true);

    try {
      const formData = new FormData();
      formData.append('prompt', inputText);
      for (const file of selectedFiles) {
        formData.append('image', file, file.name);
      }

      const apiBase = import.meta.env.PROD ? '/api' : 'http://localhost:8080/api';
      const endpoint = apiType === 'veo3'
        ? '/generate-video-pro-reels'
        : '/generate-video-reels';
      const url = `${apiBase}${endpoint}`;

      const resp = await useMultiPartFormData(formData, MULTIPART_ACTIONS.SendImageTimeline, url) as TimelineStageResponse | FinalVideoResponse;
      if (!resp.ok) {
        throw new Error(resp.error || 'Failed to submit images');
      }

      if ('timeline' in resp && resp.timeline && Array.isArray(resp.timeline) && resp.timeline.length > 0) {
        setTimelineSegments(resp.timeline);
        setPendingApproval(true);

        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: 'Please review the generated timeline below. Approve to start video rendering.',
          timestamp: new Date(),
        };
        setMessages(prev => [...prev, assistantMessage]);
      } else {
        throw new Error('Unexpected server response. No timeline or SSE URL provided.');
      }
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to generate video');
      console.error(error);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: `Sorry, there was an error generating your property video: ${error instanceof Error ? error.message : 'Unknown error'}`,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
      // Keep inputs so the user can adjust timeline first; clear on completion
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-100 to-white flex flex-col items-center justify-center p-4 md:p-8">
      {/* Title */}
      <h1 className="font-ibm text-3xl md:text-4xl font-bold text-gray-800 mb-8 text-center">
        AI Real Estate Content Creator
      </h1>
      
      <div className="flex flex-col items-center gap-8 w-full max-w-6xl">
        
        {/* Professional Video Display */}
        <div className="relative mb-4">
          <VideoContainer
            timelineSegments={timelineSegments}
            videoUrl={messages.length > 0 ? messages[messages.length - 1].videoUrl : null}
            isMobile={isMobile}
          />
        </div>

        {/* AI Prompt Interface */}
        <div className="w-full max-w-2xl">
          <div className="bg-black/5 dark:bg-white/5 rounded-2xl p-1.5 pt-4">
            <div className="flex items-center gap-2 mb-2.5 mx-2">
              <div className="flex-1 flex items-center gap-2">
                <div className="h-3.5 w-3.5 bg-blue-500 rounded-full"></div>
                <h3 className="text-black dark:text-white/90 text-xs tracking-tighter">
                  Professional Real Estate Videos
                </h3>
              </div>
              <p className="text-black dark:text-white/90 text-xs tracking-tighter">
                Create Now!
              </p>
            </div>
            <div className="relative">
              <div className="relative flex flex-col">
                <div>
                  {selectedFiles.length > 0 && (
                    <div className="mx-4 mb-3">
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-xs text-black/70 dark:text-white/70">
                          {selectedFiles.length} photo{selectedFiles.length > 1 ? 's' : ''} selected
                        </span>
                        <button
                          onClick={clearAllSelected}
                          className="text-xs text-blue-600 hover:underline"
                        >
                          Clear all
                        </button>
                      </div>
                      <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-2">
                        {previewUrls.map((url, idx) => (
                          <div key={url} className="relative rounded-lg overflow-hidden bg-black/5 dark:bg-white/5 aspect-square">
                            <img
                              src={url}
                              alt={selectedFiles[idx]?.name || `Selected image ${idx + 1}`}
                              className="w-full h-full object-cover"
                            />
                            <button
                              onClick={() => removeSelectedFile(idx)}
                              className="absolute top-1 right-1 bg-black/60 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs hover:bg-black/80"
                              aria-label={`Remove ${selectedFiles[idx]?.name || `image ${idx + 1}`}`}
                            >
                              ×
                            </button>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  <Textarea
                    value={inputText}
                    placeholder="Create a virtual tour for this luxury home..."
                    className={cn(
                      "w-full rounded-xl rounded-b-none px-4 py-3 bg-black/5 dark:bg-white/5 border-none dark:text-white placeholder:text-black/70 dark:placeholder:text-white/70 resize-none focus-visible:ring-0 focus-visible:ring-offset-0",
                      "min-h-[72px]"
                    )}
                    ref={textareaRef}
                    onKeyDown={handleKeyDown}
                    onChange={(e) => {
                      setInputText(e.target.value);
                      adjustHeight();
                    }}
                  />
                </div>

                <div className="h-14 bg-black/5 dark:bg-white/5 rounded-b-xl flex items-center">
                  <div className="absolute left-3 right-3 bottom-3 flex items-center justify-between w-[calc(100%-24px)]">
                    <div className="flex items-center gap-2">
                      <input
                        ref={fileInputRef}
                        type="file"
                        accept="image/*"
                        multiple
                        onChange={handleFileSelect}
                        className="hidden"
                      />
                      <input
                        ref={cameraInputRef}
                        type="file"
                        accept="image/*"
                        capture="environment"
                        multiple
                        onChange={handleFileSelect}
                        className="hidden"
                      />
                      
                      <label
                        className={cn(
                          "rounded-lg p-2 bg-black/5 dark:bg-white/5 cursor-pointer",
                          "hover:bg-black/10 dark:hover:bg-white/10 focus-visible:ring-1 focus-visible:ring-offset-0 focus-visible:ring-blue-500",
                          "text-black/40 dark:text-white/40 hover:text-black dark:hover:text-white"
                        )}
                        onClick={() => fileInputRef.current?.click()}
                        aria-label="Upload property photos"
                      >
                        <Upload className="w-4 h-4 transition-colors" />
                      </label>
                      
                      <label
                        className={cn(
                          "rounded-lg p-2 bg-black/5 dark:bg-white/5 cursor-pointer",
                          "hover:bg-black/10 dark:hover:bg-white/10 focus-visible:ring-1 focus-visible:ring-offset-0 focus-visible:ring-blue-500",
                          "text-black/40 dark:text-white/40 hover:text-black dark:hover:text-white"
                        )}
                        onClick={() => cameraInputRef.current?.click()}
                        aria-label="Take property photos"
                      >
                        <Camera className="w-4 h-4 transition-colors" />
                      </label>
                    </div>
                    
                    <div className="flex items-center gap-2">
                      {!isLoading ? (
                          <div className="flex items-center gap-2">
                            <Button
                              onSelect={() => handleSubmit('ffmpeg')}
                              className="flex items-center justify-between gap-2"
                            >
                              <div className="flex items-center gap-2">
                                <div className="w-4 h-4 bg-green-500 rounded-full"></div>
                                <span>FFMPEG (Free)</span>
                              </div>
                            </Button>
                            <Button
                              onSelect={() => handleSubmit('veo3')}
                              className="flex items-center justify-between gap-2"
                            >
                              <div className="flex items-center gap-2">
                                <div className="w-4 h-4 bg-blue-500 rounded-full"></div>
                                <span>Google Veo3 (Pro)</span>
                              </div>
                            </Button>
                          </div>
                      ) : (
                        <div className="rounded-lg p-2 bg-black/5 dark:bg-white/5">
                          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          {isLoading && (
            <div className="mt-3 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <div className="flex items-center space-x-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                <p className="text-sm text-blue-700 dark:text-blue-300 font-ibm">Generating your property video...</p>
              </div>
            </div>
          )}
        </div>
        
        {/* Social Media Links - Show after video generation */}
        {messages.length > 0 && messages[messages.length - 1].videoUrl && (
          <SocialSharePanel
            videoUrl={messages[messages.length - 1].videoUrl!}
            socialLinks={messages[messages.length - 1].socialLinks}
          />
        )}
      </div>
    </div>
  );
}

export default App;