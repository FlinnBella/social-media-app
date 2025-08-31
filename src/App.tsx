import React, { useState, useRef } from 'react';
import { toast } from 'sonner';
import { Send, Upload, Camera, Video, Image, Instagram, Twitter, Facebook } from 'lucide-react';

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
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const cameraInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file && (file.type.startsWith('image/') || file.type.startsWith('video/'))) {
      setSelectedFile(file);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!inputText.trim() && !selectedFile) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputText || 'Uploaded media',
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setIsLoading(true);

    try {
      let response: Response;
      if (selectedFile) {
        const formData = new FormData();
        formData.append('prompt', inputText);
        formData.append('file', selectedFile);
        response = await fetch('http://localhost:8080/api/generate-video', {
          method: 'POST',
          body: formData,
        });
      } else {
        response = await fetch('http://localhost:8080/api/generate-video', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ prompt: inputText }),
        });
      }

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
        // Backend returned an MP4 (or other video) directly
        const blob = await response.blob();
        videoObjectUrl = URL.createObjectURL(blob);
      } else if (contentType.includes('application/json')) {
        // Backend returned JSON (fallback/compatibility)
        const result = await response.json();
        if (result && result.videoUrl) {
          videoObjectUrl = result.videoUrl;
        } else if (result && result.dataUrl) {
          videoObjectUrl = result.dataUrl;
        } else {
          throw new Error('Unexpected response: no video found');
        }
      } else {
        // Fallback: try to treat as binary
        const blob = await response.blob();
        videoObjectUrl = URL.createObjectURL(blob);
      }

      const previousVideoUrl = messages.length > 0 ? messages[messages.length - 1].videoUrl : undefined;

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: 'Your video has been generated! You can now share it on your favorite social media platforms.',
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
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: `Sorry, there was an error generating your video: ${error instanceof Error ? error.message : 'Unknown error'}`,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
      setInputText('');
      setSelectedFile(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
      if (cameraInputRef.current) cameraInputRef.current.value = '';
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-pink-200 to-white flex flex-col items-center justify-center p-8">
      {/* Title */}
      <h1 className="font-ibm text-4xl font-bold text-gray-800 mb-8 text-center">
        Social Media Content Maker
      </h1>
      
      <div className="flex flex-col items-center gap-8 max-w-md w-full">
        
        {/* iPhone 15 Frame */}
        <div className="relative mb-4">
          <div className="w-80 h-[640px] bg-black rounded-[3rem] p-2 shadow-2xl">
            <div className="w-full h-full bg-white rounded-[2.5rem] relative overflow-hidden">
              {/* Dynamic Island */}
              <div className="absolute top-6 left-1/2 transform -translate-x-1/2 w-32 h-8 bg-black rounded-full z-10"></div>
              
              {/* Video Display Area */}
              <div className="w-full h-full flex items-center justify-center bg-gray-50">
                {messages.length > 0 && messages[messages.length - 1].videoUrl ? (
                  <video 
                    src={messages[messages.length - 1].videoUrl} 
                    controls 
                    className="w-full h-full object-cover"
                    autoPlay
                    muted
                    loop
                  />
                ) : (
                  <div className="text-center text-gray-400 p-8">
                    <Video className="w-16 h-16 mx-auto mb-4 opacity-50" />
                    <p className="text-sm">Your AI-generated video will appear here</p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Chat Interface */}
        <div className="w-full max-w-md">
          <div className="bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl overflow-hidden">
            <div className="p-4">
              {selectedFile && (
                <div className="mb-3 p-2 bg-gray-50 rounded-lg flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    {selectedFile.type.startsWith('image/') ? (
                      <Image className="w-4 h-4 text-gray-500" />
                    ) : (
                      <Video className="w-4 h-4 text-gray-500" />
                    )}
                    <span className="text-sm text-gray-600 truncate">{selectedFile.name}</span>
                  </div>
                  <button 
                    onClick={() => setSelectedFile(null)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    ×
                  </button>
                </div>
              )}

              {!isLoading ? (
              <form onSubmit={handleSubmit} className="space-y-3">
                <div className="flex items-end space-x-2">
                  <div className="flex-1">
                    <textarea
                      value={inputText}
                      onChange={(e) => setInputText(e.target.value)}
                      placeholder="Describe the video you want to create..."
                      className="w-full p-3 border border-gray-300 rounded-xl resize-none focus:ring-2 focus:ring-pink-500 focus:border-transparent font-ibm"
                      rows={2}
                    />
                  </div>
                  <button
                    type="submit"
                    disabled={isLoading || (!inputText.trim() && !selectedFile)}
                    className="p-3 bg-gradient-to-r from-pink-400 to-pink-500 text-white rounded-xl hover:from-pink-500 hover:to-pink-600 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 hover:scale-105"
                  >
                    <Send className="w-5 h-5" />
                  </button>
                </div>
                
                <div className="flex space-x-2">
                  <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*,video/*"
                    onChange={handleFileSelect}
                    className="hidden"
                  />
                  <input
                    ref={cameraInputRef}
                    type="file"
                    accept="image/*,video/*"
                    capture="environment"
                    onChange={handleFileSelect}
                    className="hidden"
                  />
                  
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="flex items-center space-x-1 px-3 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors font-ibm"
                  >
                    <Upload className="w-4 h-4" />
                    <span>Upload</span>
                  </button>
                  
                  <button
                    type="button"
                    onClick={() => cameraInputRef.current?.click()}
                    className="flex items-center space-x-1 px-3 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors font-ibm"
                  >
                    <Camera className="w-4 h-4" />
                    <span>Camera</span>
                  </button>
                </div>
              </form>
              ):    
              (
                <div className="mt-3 p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center space-x-2">
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-pink-500"></div>
                    <p className="text-sm text-gray-600 font-ibm">Generating your video...</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
        
        {/* Social Media Links - Show after video generation */}
        {messages.length > 0 && messages[messages.length - 1].videoUrl && (
          <div className="w-full max-w-md bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl p-4">
            <p className="text-sm text-gray-600 mb-3 font-ibm text-center">Share your video:</p>
            <div className="flex justify-center gap-3">
              <a href="https://www.instagram.com/create/story/" target="_blank" rel="noopener noreferrer" 
                 className="p-3 bg-gradient-to-r from-pink-400 to-pink-500 rounded-xl hover:scale-105 transition-transform">
                <Instagram className="w-5 h-5 text-white" />
              </a>
              <a href="https://www.tiktok.com/upload/" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-black rounded-xl hover:scale-105 transition-transform">
                <Video className="w-5 h-5 text-white" />
              </a>
              <a href="https://twitter.com/compose/tweet" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-blue-500 rounded-xl hover:scale-105 transition-transform">
                <Twitter className="w-5 h-5 text-white" />
              </a>
              <a href="https://www.facebook.com/" target="_blank" rel="noopener noreferrer"
                 className="p-3 bg-blue-600 rounded-xl hover:scale-105 transition-transform">
                <Facebook className="w-5 h-5 text-white" />
              </a>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

export default App;