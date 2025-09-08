import React from 'react';
import { Button } from '@/components/ui/button';
import { Download } from 'lucide-react';
import { toast } from 'sonner';

type SocialLinks = {
  instagram?: string;
  tiktok?: string;
  twitter?: string;
  facebook?: string;
};

interface SocialSharePanelProps {
  videoUrl: string;
  socialLinks?: SocialLinks;
}

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

export default function SocialSharePanel({ videoUrl, socialLinks }: SocialSharePanelProps) {
  const handleDownload = () => {
    if (!videoUrl) return;
    const link = document.createElement('a');
    link.href = videoUrl;
    link.download = `property-video-${Date.now()}.mp4`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    toast.success('Video downloaded successfully!');
  };

  const links: Required<SocialLinks> = {
    instagram: socialLinks?.instagram || 'https://www.instagram.com/create/story/',
    tiktok: socialLinks?.tiktok || 'https://www.tiktok.com/upload/',
    twitter: socialLinks?.twitter || 'https://twitter.com/compose/tweet',
    facebook: socialLinks?.facebook || 'https://www.facebook.com/',
  };

  return (
    <div className="w-full max-w-2xl bg-white/95 backdrop-blur-sm rounded-2xl shadow-2xl p-4">
      <div className="text-center mb-4">
        <h3 className="text-lg font-semibold text-gray-800 mb-2">Share Your Property Video</h3>
        <p className="text-sm text-gray-600">Download locally or share directly to social media</p>
      </div>

      <div className="flex justify-center mb-4">
        <Button
          onClick={handleDownload}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium transition-colors"
        >
          <Download className="w-4 h-4" />
          Download Video
        </Button>
      </div>

      <div className="flex justify-center gap-4">
        <a href={links.instagram} target="_blank" rel="noopener noreferrer"
           className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg"
           title="Share to Instagram Stories">
          <InstagramSVG />
        </a>
        <a href={links.tiktok} target="_blank" rel="noopener noreferrer"
           className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg"
           title="Upload to TikTok">
          <TikTokSVG />
        </a>
        <a href={links.twitter} target="_blank" rel="noopener noreferrer"
           className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg"
           title="Share on X (Twitter)">
          <TwitterSVG />
        </a>
        <a href={links.facebook} target="_blank" rel="noopener noreferrer"
           className="p-3 bg-white rounded-xl hover:scale-105 transition-transform shadow-lg"
           title="Share on Facebook">
          <FacebookSVG />
        </a>
      </div>

      <p className="text-xs text-gray-500 text-center mt-3">
        Click download first, then use the social media links to upload your saved video
      </p>
    </div>
  );
}


