import { Button } from "@/components/ui/button";
import { useSubmission } from "@/context/SubmissionContext";
import { toast } from "sonner";
import type { TimelineCompositionResponse } from "#types/timeline";

interface Props {
  prompt: string;
  images: File[];
  setHasSubmittedTimeline: (hasSubmittedTimeline: boolean) => void;
}

export const GenerateTimelineButton: React.FC<Props> = ({ prompt, images, setHasSubmittedTimeline }) => {

  const { submit } = useSubmission();
  const handleClick = async () => {
    if (prompt.trim() === "" || images.length === 0) {
      toast.error("Please enter a prompt and upload at least one image");
      return;
    }
    
    const res = await submit("generateVideoTimeline", prompt, images);
    if (res) {
      setHasSubmittedTimeline(true);
    }
  };

  return (
    // no need for disabled as should disapear once clicked
    <>
    <Button onClick={handleClick}>
      <div className="flex items-center gap-2">
        <div className="w-4 h-4 bg-purple-500 rounded-full"></div>
        <span>Generate Timeline</span>
      </div>
    </Button>
    </>
  );
}