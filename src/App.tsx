import React, { useState, useRef, useEffect } from 'react';
import { toast } from 'sonner';
import { ArrowRight, Upload, Camera, Monitor, Smartphone } from 'lucide-react';
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
import { motion, AnimatePresence } from 'motion/react';

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

// Social Media Brand SVGs
const InstagramSVG = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
    <defs>
      <linearGradient id="instagram-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
        <stop offset="0%" stopColor="#f09433" />
        <stop offset="25%" stopColor="#e6683c" />
        <stop offset="50%" stopColor="#dc2743" />
        <stop offset="75%" stopColor="#cc2366" />
        <stop offset="100%" stopColor="#bc1888" />
      </linearGradient>
    </defs>
    <rect x="2" y="2" width="20" height="20" rx="5" ry="5" fill="url(#instagram-gradient)" />
    <path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z" fill="white" />
    <line x1="17.5" y1="6.5" x2="17.51" y2="6.5" stroke="white" strokeWidth="2" strokeLinecap="round" />
  </svg>
);

const TikTokSVG = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
    <path d="M19.59 6.69a4.83 4.83 0 0 1-3.77-4.25V2h-3.45v13.67a2.89 2.89 0 0 1-5.2 1.74 2.89 2.89 0 0 1 2.31-4.64 2.93 2.93 0 0 1 .88.13V9.4a6.84 6.84 0 0 0-1-.05A6.33 6.33 0 0 0 5 20.1a6.34 6.34 0 0 0 10.86-4.43v-7a8.16 8.16 0 0 0 4.77 1.52v-3.4a4.85 4.85 0 0 1-1-.1z" fill="#000"/>
  </svg>
);

const TwitterSVG = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" fill="#000"/>
  </svg>
);

const FacebookSVG = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
    <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" fill="#1877F2"/>
  </svg>
);

function App() {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputText, setInputText] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);
  
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

      const url = apiType === 'veo3'
        ? 'http://localhost:8080/api/generate-video-pro-reels'
        : 'http://localhost:8080/api/generate-video-reels';
      
      const response = await fetch(url, { method: 'POST', body: formData });
      const contentType = response.headers.get('content-type') || '';

      if (!response.ok) {
        try {
          if (contentType.includes('application/json')) {
            const errorJson = await response.json();
            throw new Error(errorJson.error || 'Failed to generate video');
          }
          const errorText = await response.text();
          throw new Error(errorText || 'Failed to generate video');
        } catch (e) {
          throw new Error(e instanceof Error ? e.message : 'Failed to generate video');
        }
      }

      let videoObjectUrl: string | undefined;
      if (contentType.includes('video/')) {
        const blob = await response.blob();
        videoObjectUrl = URL.createObjectURL(blob);
      } else if (contentType.includes('application/json')) {
        const result = await response.json();
        if (result && result.videoUrl) {
          videoObjectUrl = result.videoUrl;
        } else if (result && result.dataUrl) {
          videoObjectUrl = result.dataUrl;
        } else {
          throw new Error('Unexpected response: no video found');
        }
      } else {
        const blob = await response.blob();
        videoObjectUrl = URL.createObjectURL(blob);
      }

      const previousVideoUrl = messages.length > 0 ? messages[messages.length - 1].videoUrl : undefined;

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: 'Your property video has been generated! Share it across your marketing channels to attract potential buyers.',
        timestamp: new Date(),
        videoUrl: videoObjectUrl,
        socialLinks: {
          instagram: `https://www.instagram.com/create/story/`,
          tiktok: `https://www.tiktok.com/upload/`,
          twitter: `https://twitter.com/compose/tweet`,
          facebook: `https://www.facebook.com/`,
        },
      };
      setMessages(prev => [...prev, assistantMessage]);

      if (previousVideoUrl && previousVideoUrl.startsWith('blob:')) {
        try { URL.revokeObjectURL(previousVideoUrl); } catch {}
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
      setInputText('');
      setSelectedFiles([]);
      adjustHeight(true);
      if (fileInputRef.current) fileInputRef.current.value = '';
      if (cameraInputRef.current) cameraInputRef.current.value = '';
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
          <div className={cn(
            "bg-gray-900 rounded-2xl p-4 shadow-2xl",
            isMobile ? "w-full max-w-sm" : "w-full max-w-4xl"
          )}>
            <div className={cn(
              "bg-black rounded-xl relative overflow-hidden",
              isMobile ? "aspect-[9/16] max-h-[600px]" : "aspect-video h-[400px]"
            )}>
              {/* Video Display Area */}
              <div className="w-full h-full flex items-center justify-center bg-gray-900">
                {messages.length > 0 && messages[messages.length - 1].videoUrl ? (
                  <video 
                    src={messages[messages.length - 1].videoUrl} 
                    controls 
                    className="w-full h-full object-cover rounded-xl"
                    autoPlay
                    muted
                    loop
                  />
                ) : (
                  <div className="text-center text-gray-400 p-8">
                    {isMobile ? <Smartphone className="w-16 h-16 mx-auto mb-4 opacity-50" /> : <Monitor className="w-16 h-16 mx-auto mb-4 opacity-50" />}
                    <p className="text-sm">Your AI-generated property video will appear here</p>
                    <p className="text-xs mt-2 opacity-75">
                      Optimized for {isMobile ? "mobile viewing" : "desktop presentation"}
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
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
                <div className="overflow-y-auto" style={{ maxHeight: "400px" }}>
                  {selectedFiles.length > 0 && (
                    <div className="mx-4 mb-3 p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <div className="w-4 h-4 text-blue-500">📷</div>
                        <span className="text-sm text-blue-700 dark:text-blue-300 truncate">
                          {selectedFiles.length === 1 ? selectedFiles[0].name : `${selectedFiles.length} property photos selected`}
                        </span>
                      </div>
                      <button 
                        onClick={() => setSelectedFiles([])}
                        className="text-blue-400 hover:text-blue-600"
                      >
                        ×
                      </button>
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
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button
                              variant="ghost"
                              className={cn(
                                "rounded-lg p-2 bg-black/5 dark:bg-white/5",
                                "hover:bg-black/10 dark:hover:bg-white/10 focus-visible:ring-1 focus-visible:ring-offset-0 focus-visible:ring-blue-500"
                              )}
                              disabled={!inputText.trim() || selectedFiles.length === 0}
                            >
                              <ArrowRight
                                className={cn(
                                  "w-4 h-4 dark:text-white transition-opacity duration-200",
                                  (inputText.trim() && selectedFiles.length > 0)
                                    ? "opacity-100"
                                    : "opacity-30"
                                )}
                              />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent
                            className={cn(
                              "min-w-[12rem]",
                              "border-black/10 dark:border-white/10",
                              "bg-gradient-to-b from-white via-white to-neutral-100 dark:from-neutral-950 dark:via-neutral-900 dark:to-neutral-800"
                            )}
                          >
                            <DropdownMenuItem
                              onSelect={() => handleSubmit('ffmpeg')}
                              className="flex items-center justify-between gap-2"
                            >
                              <div className="flex items-center gap-2">
                                <div className="w-4 h-4 bg-green-500 rounded-full"></div>
                                <span>FFMPEG (Free)</span>
                              </div>
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onSelect={() => handleSubmit('veo3')}
                              className="flex items-center justify-between gap-2"
                            >
                              <div className="flex items-center gap-2">
                                <div className="w-4 h-4 bg-blue-500 rounded-full"></div>
                                <span>Google Veo3 (Pro)</span>
                              </div>
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
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
          <div className="w-full max-w-2xl bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl p-4">
            <p className="text-sm text-gray-600 mb-3 font-ibm text-center">Share your property video:</p>
            <div className="flex justify-center gap-4">
              <a href="https://www.instagram.com/create/story/" target="_blank" rel="noopener noreferrer" 
                 className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg">
                <InstagramSVG />
              </a>
              <a href="https://www.tiktok.com/upload/" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg">
                <TikTokSVG />
              </a>
              <a href="https://twitter.com/compose/tweet" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg">
                <TwitterSVG />
              </a>
              <a href="https://www.facebook.com/" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg">
                <FacebookSVG />
              </a>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;