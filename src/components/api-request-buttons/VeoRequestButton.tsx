import React from "react";
import { Button } from "@/components/ui/button";
import { useSubmission } from "@/context/SubmissionContext";
import { toast } from "sonner";

interface Props {
  prompt: string;
  images: File[];
  disabled?: boolean;
  className?: string;
}

export const VeoRequestButton: React.FC<Props> = ({ prompt, images, disabled, className }) => {
  const { submit } = useSubmission();
  const handleClick = async () => {
    if (prompt.trim() === "" || images.length === 0) {
      toast.error("Please enter a prompt and upload at least one image");
      return;
    }
    await submit("generateVideoProReels", prompt, images);
  };

  return (
    <Button onClick={handleClick} disabled={disabled} className={className}>
      <div className="flex items-center gap-2">
        <div className="w-4 h-4 bg-blue-500 rounded-full"></div>
        <span>Google Veo3 (Pro)</span>
      </div>
    </Button>
  );
};

export default VeoRequestButton;


