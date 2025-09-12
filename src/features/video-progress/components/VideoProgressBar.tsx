import React from 'react';
import { cn } from '@/lib/utils';

export interface VideoProgressData {
  stage: string;
  message: string;
  progress: number;
}

export interface VideoErrorData {
  error: string;
  stage: string;
}

interface VideoProgressBarProps {
  progress?: VideoProgressData | null;
  error?: VideoErrorData | null;
  isVisible: boolean;
  className?: string;
}

const getStageColor = (stage: string) => {
  switch (stage) {
    case 'compiling':
      return 'bg-blue-500';
    case 'streaming':
      return 'bg-green-500';
    case 'completed':
      return 'bg-green-600';
    case 'compilation_failed':
    case 'streaming_failed':
      return 'bg-red-500';
    default:
      return 'bg-blue-500';
  }
};

const getStageMessage = (stage: string) => {
  switch (stage) {
    case 'compiling':
      return 'Compiling your video...';
    case 'streaming':
      return 'Streaming video to you...';
    case 'completed':
      return 'Video generation completed!';
    case 'compilation_failed':
      return 'Video compilation failed';
    case 'streaming_failed':
      return 'Video streaming failed';
    default:
      return 'Processing...';
  }
};

export const VideoProgressBar: React.FC<VideoProgressBarProps> = ({
  progress,
  error,
  isVisible,
  className
}) => {
  if (!isVisible) return null;

  const displayProgress = progress || { stage: 'processing', message: 'Starting...', progress: 0 };
  const displayError = error;
  
  const progressValue = Math.max(0, Math.min(100, displayProgress.progress));
  const stageColor = getStageColor(displayProgress.stage);
  const stageMessage = displayError ? displayError.error : (displayProgress.message || getStageMessage(displayProgress.stage));

  return (
    <div className={cn(
      "w-full max-w-2xl mx-auto mt-4 p-4 bg-white dark:bg-gray-800 rounded-lg shadow-lg border",
      className
    )}>
      <div className="flex items-center justify-between mb-2">
        <h3 className="text-sm font-medium text-gray-900 dark:text-gray-100">
          Video Generation Progress
        </h3>
        <span className="text-xs text-gray-500 dark:text-gray-400">
          {progressValue}%
        </span>
      </div>
      
      <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mb-2">
        <div
          className={cn(
            "h-2 rounded-full transition-all duration-300 ease-out",
            stageColor,
            displayError && "bg-red-500"
          )}
          style={{ width: `${progressValue}%` }}
        />
      </div>
      
      <div className="flex items-center justify-between">
        <p className={cn(
          "text-sm",
          displayError 
            ? "text-red-600 dark:text-red-400" 
            : "text-gray-700 dark:text-gray-300"
        )}>
          {stageMessage}
        </p>
        
        {displayProgress.stage === 'completed' && !displayError && (
          <div className="flex items-center text-green-600 dark:text-green-400">
            <svg className="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
            <span className="text-xs font-medium">Complete</span>
          </div>
        )}
      </div>
    </div>
  );
};
