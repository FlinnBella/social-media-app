import { Button } from "@/components/ui/button";
import { useSubmission } from "@/context/SubmissionContext";
import { toast } from "sonner";
//types
import type { TimelineCompositionResponse } from "#types/timeline";
import { Spinner } from "@/components/ui/shadcn-io/spinner";

interface Props {
  prompt: string;
  images: File[];
  setHasSubmittedTimeline: (hasSubmittedTimeline: boolean) => void;
}

export const GenerateTimelineButton: React.FC<Props> = ({ prompt, images, setHasSubmittedTimeline }) => {

  const { submitTimeline, isLoading, setIsLoading } = useSubmission();
  const handleClick = async () => {
    if (isLoading) return;
    try {
    if (prompt.trim() === "" || images.length === 0) {
      toast.error("Please enter a prompt and upload at least one image");
      return;
    }
    
    const res = await submitTimeline(prompt, images);
    if (res as TimelineCompositionResponse) {
      setHasSubmittedTimeline(true);
    } else if (res instanceof Error) {
      toast.error(res.message);
    }
  } catch (error) {
    toast.error(error instanceof Error ? error.message : 'Unknown error');
    setHasSubmittedTimeline(false);
  } finally {
    setIsLoading(false);
  }
  };

  return (
    // no need for disabled as should disapear once clicked
    <>
    <Button onClick={handleClick} disabled={isLoading} aria-busy={isLoading} className={isLoading ? "bg-primary/80 hover:bg-primary/80" : undefined}>
      <div className="flex items-center gap-2">
        <div className={isLoading ? "w-4 h-4 bg-purple-400 rounded-full" : "w-4 h-4 bg-purple-500 rounded-full"}></div>
        <span> {isLoading ? "Generating Timeline..." : "Generate Timeline"} </span>
        {isLoading && <Spinner variant="circle" />}
      </div>
    </Button>
    </>
  );
}