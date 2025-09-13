import { useState, useEffect, useCallback } from "react";
import { Upload, Camera } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useAutoResizeTextarea } from "@/hooks/use-auto-resize-textarea";
// removed unused Button import
// removed unused dropdown menu imports

import SocialSharePanel from "@/components/SocialSharePanel";
// types imported where needed
import VideoContainer from "@/components/VideoScreen/VideoContainer";
// kept for buttons inside the composer zone
// buttons are used inside Timeline, remove local imports to avoid unused warnings
import { useSubmission } from "@/context/SubmissionContext";
import { GenerateTimelineButton } from "@/components/api-request-buttons/GenerateTimelineButton";
import { PROMPT_TYPES } from "@/features/prompt-templates/prompttypes";




//types
import type { Timeline as TimelineType } from "#types/timeline";
import { useChooseTemplatePrompt } from "./features/prompt-templates/usePrompt";
import { useUserState } from "@/context/UserContext";
// MakeTypeFieldsRequired not used in this file

import { PromptTemplateContainer } from "./features/prompt-templates/PromptTemplateContainer";
import { PromptCards, PromptCard } from "./features/prompt-templates/prompttypes";

// Message type handled via SubmissionContext

import { IMAGE_ACCEPT_ATTRIBUTE } from "@/utilites/useImageSelection";

function App() {
  const { isLoading, messages, timeline } = useSubmission();
  const { prompt: inputText, setPrompt: setInputText, selectedFiles, previewUrls, handleFileSelect, removeSelectedFile, clearAllSelected, fileInputRef, cameraInputRef } = useUserState();
  const [isMobile, setIsMobile] = useState(false);
  // removed unused pendingApproval state

  const [hasSubmittedTimeline, setHasSubmittedTimeline] = useState(false);
  
  // Video progress hook for SSE

  const { textareaRef, adjustHeight } = useAutoResizeTextarea({
    minHeight: 72,
    maxHeight: 300,
  });

  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768);
    };

    checkMobile();
    window.addEventListener("resize", checkMobile);
    return () => window.removeEventListener("resize", checkMobile);
  }, []);

  // image selection and validation handled in useImageSelection()

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      // Don't submit here, let the form handle it
    }
  };


 /*
CALLBACKS START
*/
const handlePromptApply = useCallback((promptType: keyof typeof PROMPT_TYPES) => {
  const transformed = useChooseTemplatePrompt(inputText, promptType);
  setInputText(transformed);
}, [inputText, setInputText]);

/*
CALLBACKS END
*/

  // submission handlers moved to SubmissionContext

  return (
    <div className="min-h-screen bg-gradient-to-b from-blue-100 to-white flex flex-col items-center justify-center p-4 md:p-8">
      {/* Title */}
      <h1 className="font-ibm text-3xl md:text-4xl font-bold text-gray-800 mb-8 text-center">
        AI Real Estate Content Creator
      </h1>

      <div className="flex flex-col items-center gap-8 w-full max-w-6xl">
        {/* Professional Video Display */}
        {/* Images passed as file props into container; will allow us to interact with them in the children components */}
        <div className="relative mb-4">
          <VideoContainer
            timeline={timeline as TimelineType}
            videoUrl={
              messages.length > 0
                ? messages[messages.length - 1].videoUrl
                : undefined
            }
            isMobile={isMobile}
            prompt={inputText}
            images={selectedFiles}
            previewUrls={previewUrls}
          />
        </div>


        {/* Prompt Templates */}
        <div>
        <PromptTemplateContainer templates={PromptCards as PromptCard[]} prompt={inputText} onApply={handlePromptApply} />
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
                          {selectedFiles.length} photo
                          {selectedFiles.length > 1 ? "s" : ""} selected
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
                          <div
                            key={url}
                            className="relative rounded-lg overflow-hidden bg-black/5 dark:bg-white/5 aspect-square"
                          >
                            <img
                              src={url}
                              alt={
                                selectedFiles[idx]?.name ||
                                `Selected image ${idx + 1}`
                              }
                              className="w-full h-full object-cover"
                            />
                            <button
                              onClick={() => removeSelectedFile(idx)}
                              className="absolute top-1 right-1 bg-black/60 text-white rounded-full w-6 h-6 flex items-center justify-center text-xs hover:bg-black/80"
                              aria-label={`Remove ${
                                selectedFiles[idx]?.name || `image ${idx + 1}`
                              }`}
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
                        accept={IMAGE_ACCEPT_ATTRIBUTE}
                        multiple
                        onChange={handleFileSelect}
                        className="hidden"
                      />
                      <input
                        ref={cameraInputRef}
                        type="file"
                        accept={IMAGE_ACCEPT_ATTRIBUTE}
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

                    {/* FFMpeg and Veo buttons now rendered in the timeline */}
                    <div className="flex items-center gap-2">
                      {hasSubmittedTimeline ? 
                         (
                          <div className="rounded-lg p-2 bg-black/5 dark:bg-white/5">
                            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                          </div>
                        )
                       : (
                        <div className="flex items-center gap-2">
                          <GenerateTimelineButton
                            prompt={inputText}
                            images={selectedFiles}
                            setHasSubmittedTimeline={setHasSubmittedTimeline}
                          />
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
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
