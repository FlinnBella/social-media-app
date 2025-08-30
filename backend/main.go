import React, { useState, useRef } from 'react';
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
      const formData = new FormData();
      formData.append('prompt', inputText);
      if (selectedFile) {
        formData.append('file', selectedFile);
      }

      const response = await fetch('http://localhost:8080/api/generate-video', {
        method: 'POST',
        body: formData,
      });

      const result = await response.json();

      if (response.ok) {
        const assistantMessage: Message = {
          id: (Date.now() + 1).toString(),
          type: 'assistant',
          content: 'Your video has been generated! You can now share it on your favorite social media platforms.',
          timestamp: new Date(),
          videoUrl: result.videoUrl,
          socialLinks: {
            instagram: `https://www.instagram.com/create/story/`,
            tiktok: `https://www.tiktok.com/upload/`,
            twitter: `https://twitter.com/compose/tweet`,
            facebook: `https://www.facebook.com/`,
          },
        };
        setMessages(prev => [...prev, assistantMessage]);
      } else {
        throw new Error(result.error || 'Failed to generate video');
      }
    } catch (error) {
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
            </div>
            
            {/* Social Media Links - Show after video generation */}
            {messages.length > 0 && messages[messages.length - 1].videoUrl && (
              <div className="border-t border-gray-200 p-4">
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
      </div>
    </div>
  );
}

export default App;