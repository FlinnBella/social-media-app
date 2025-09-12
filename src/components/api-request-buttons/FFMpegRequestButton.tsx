import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { useSubmission } from "@/context/SubmissionContext";
import { toast } from "sonner";

interface Props {
  prompt: string;
  images: File[];
  disabled: boolean;
  className?: string;
}

export const FFMpegRequestButton: React.FC<Props> = ({ prompt, images, disabled, className }) => {
  
  const { requestVideo, timelineComposition } = useSubmission();
  const handleClick = async () => {
    disabled = true;
    if (prompt.trim() === "" || images.length === 0) {
      toast.error("Please enter a prompt and upload at least one image");
      return;
    }
    if(timelineComposition){
    await requestVideo("generateVideoReels", prompt, images, timelineComposition);
    } else {
      toast.error("You need to order your images in advance to get the best quality!");
      return;
    }
    disabled = false;
  };

  return (    
    <Button onClick={handleClick} disabled={disabled} className={className}>
      <div className={`flex items-center gap-2 ${disabled ? "opacity-50" : ""}`}>
        <div className="w-4 h-4 bg-green-500 rounded-full"></div>
        <span>FFMPEG (Free)</span>
      </div>
    </Button>   
  );
};

export default FFMpegRequestButton;


