import { useState, useEffect, useCallback, useRef, useMemo } from "react";

export interface UseTypeWriterOptions {
  /**
   * Text to type out character by character
   */
  text: string;
  
  /**
   * Typing speed in milliseconds per character
   * @default 50
   */
  speed?: number;
  
  /**
   * Whether to show a blinking cursor
   * @default true
   */
  showCursor?: boolean;
  
  /**
   * Custom cursor character
   * @default "|"
   */
  cursorChar?: string;
  
  /**
   * Cursor blink speed in milliseconds
   * @default 1000
   */
  cursorBlinkSpeed?: number;
  
  /**
   * Whether to start typing immediately
   * @default true
   */
  autoStart?: boolean;
  
  /**
   * Delay before starting to type (in milliseconds)
   * @default 0
   */
  startDelay?: number;
  
  /**
   * Whether to loop the animation
   * @default false
   */
  loop?: boolean;
  
  /**
   * Delay between loops (in milliseconds)
   * @default 1000
   */
  loopDelay?: number;
  
  /**
   * Callback when typing is complete
   */
  onComplete?: () => void;
  
  /**
   * Callback on each character typed
   */
  onType?: (displayedText: string, currentIndex: number) => void;
}

export interface UseTypeWriterReturn {
  /**
   * The currently displayed text
   */
  displayedText: string;
  
  /**
   * Whether the typing animation is currently running
   */
  isTyping: boolean;
  
  /**
   * Whether the typing is complete
   */
  isComplete: boolean;
  
  /**
   * Start the typing animation
   */
  start: () => void;
  
  /**
   * Stop the typing animation
   */
  stop: () => void;
  
  /**
   * Reset the typing animation
   */
  reset: () => void;
  
  /**
   * Current cursor visibility state (for custom cursor implementations)
   */
  cursorVisible: boolean;
  
  /**
   * Cursor element with blinking animation (ready to render)
   */
  cursor: JSX.Element | null;
  
  /**
   * Current typing progress (0-1)
   */
  progress: number;
}

export function useTypeWriter(options: UseTypeWriterOptions): UseTypeWriterReturn {
  const {
    text,
    speed = 50,
    showCursor = true,
    cursorChar = "|",
    cursorBlinkSpeed = 1000,
    autoStart = true,
    startDelay = 0,
    loop = false,
    loopDelay = 1000,
    onComplete,
    onType
  } = options;

  // State
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isTyping, setIsTyping] = useState(false);
  const [cursorVisible, setCursorVisible] = useState(true);
  const [hasStarted, setHasStarted] = useState(false);

  // Refs for cleanup
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const cursorTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const startDelayTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const loopTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Derived state
  const displayedText = useMemo(() => text.slice(0, currentIndex), [text, currentIndex]);
  const isComplete = currentIndex >= text.length;
  const progress = text.length > 0 ? currentIndex / text.length : 0;

  // Cleanup function
  const cleanup = useCallback(() => {
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
      typingTimeoutRef.current = null;
    }
    if (cursorTimeoutRef.current) {
      clearInterval(cursorTimeoutRef.current);
      cursorTimeoutRef.current = null;
    }
    if (startDelayTimeoutRef.current) {
      clearTimeout(startDelayTimeoutRef.current);
      startDelayTimeoutRef.current = null;
    }
    if (loopTimeoutRef.current) {
      clearTimeout(loopTimeoutRef.current);
      loopTimeoutRef.current = null;
    }
  }, []);

  // Cursor blinking effect
  useEffect(() => {
    if (showCursor) {
      cursorTimeoutRef.current = setInterval(() => {
        setCursorVisible(prev => !prev);
      }, cursorBlinkSpeed);

      return () => {
        if (cursorTimeoutRef.current) {
          clearInterval(cursorTimeoutRef.current);
        }
      };
    } else {
      setCursorVisible(false);
    }
  }, [showCursor, cursorBlinkSpeed]);

  // Typing animation
  useEffect(() => {
    if (!isTyping || currentIndex >= text.length) return;

    typingTimeoutRef.current = setTimeout(() => {
      setCurrentIndex(prev => {
        const newIndex = prev + 1;
        const newDisplayedText = text.slice(0, newIndex);
        
        // Call onType callback
        if (onType) {
          onType(newDisplayedText, newIndex);
        }
        
        return newIndex;
      });
    }, speed);

    return () => {
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
      }
    };
  }, [isTyping, currentIndex, text, speed, onType]);

  // Handle completion and looping
  useEffect(() => {
    if (isComplete && isTyping) {
      setIsTyping(false);
      
      // Call onComplete callback
      if (onComplete) {
        onComplete();
      }
      
      // Handle looping
      if (loop) {
        loopTimeoutRef.current = setTimeout(() => {
          setCurrentIndex(0);
          setIsTyping(true);
        }, loopDelay);
      }
    }
  }, [isComplete, isTyping, loop, loopDelay, onComplete]);

  // Auto start with delay
  useEffect(() => {
    if (autoStart && !hasStarted && text.length > 0) {
      setHasStarted(true);
      
      if (startDelay > 0) {
        startDelayTimeoutRef.current = setTimeout(() => {
          setIsTyping(true);
        }, startDelay);
      } else {
        setIsTyping(true);
      }
    }
  }, [autoStart, hasStarted, text.length, startDelay]);

  // Reset when text changes
  useEffect(() => {
    cleanup();
    setCurrentIndex(0);
    setIsTyping(false);
    setHasStarted(false);
    
    if (autoStart && text.length > 0) {
      setHasStarted(true);
      
      if (startDelay > 0) {
        startDelayTimeoutRef.current = setTimeout(() => {
          setIsTyping(true);
        }, startDelay);
      } else {
        setIsTyping(true);
      }
    }
  }, [text, autoStart, startDelay, cleanup]);

  // Control functions
  const start = useCallback(() => {
    if (!isComplete) {
      setIsTyping(true);
    }
  }, [isComplete]);

  const stop = useCallback(() => {
    setIsTyping(false);
    cleanup();
  }, [cleanup]);

  const reset = useCallback(() => {
    cleanup();
    setCurrentIndex(0);
    setIsTyping(false);
    setHasStarted(false);
  }, [cleanup]);

  // Cursor component
  const cursor = useMemo(() => {
    if (!showCursor) return null;
    
    return (
      <span 
        className="typewriter-cursor"
        style={{
          opacity: cursorVisible ? 1 : 0,
          transition: 'opacity 0.1s ease-in-out'
        }}
      >
        {cursorChar}
      </span>
    );
  }, [showCursor, cursorVisible, cursorChar]);

  // Cleanup on unmount
  useEffect(() => {
    return cleanup;
  }, [cleanup]);

  return {
    displayedText,
    isTyping,
    isComplete,
    start,
    stop,
    reset,
    cursorVisible,
    cursor,
    progress
  };
}

// Convenience hook for simple use cases
export function useSimpleTypeWriter(text: string, speed = 50) {
  return useTypeWriter({ text, speed });
}

// Hook for streaming text (like ChatGPT/Claude)
export function useStreamingTypeWriter(speed = 30) {
  const [fullText, setFullText] = useState("");
  
  const typewriter = useTypeWriter({
    text: fullText,
    speed,
    showCursor: true,
    autoStart: true
  });

  const updateText = useCallback((newText: string) => {
    setFullText(newText);
  }, []);

  const appendText = useCallback((textToAppend: string) => {
    setFullText(prev => prev + textToAppend);
  }, []);

  return {
    ...typewriter,
    updateText,
    appendText,
    fullText
  };
}