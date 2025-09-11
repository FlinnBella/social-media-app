import React, { useState } from 'react';
import { useTypeWriter, useSimpleTypeWriter, useStreamingTypeWriter } from '../../hooks/useTypeWriter';

// Basic Example
export function BasicTypewriterExample() {
  const { displayedText, cursor } = useSimpleTypeWriter(
    "Hello! This is a basic typewriter effect.",
    80
  );

  return (
    <div className="p-4 border rounded-lg">
      <h3 className="text-lg font-semibold mb-2">Basic Typewriter</h3>
      <p className="font-mono text-lg">
        {displayedText}
        {cursor}
      </p>
    </div>
  );
}

// Advanced Example with Controls
export function AdvancedTypewriterExample() {
  const [text, setText] = useState("Welcome to our amazing platform! 🚀");
  
  const {
    displayedText,
    cursor,
    isTyping,
    isComplete,
    progress,
    start,
    stop,
    reset
  } = useTypeWriter({
    text,
    speed: 50,
    showCursor: true,
    cursorChar: "█",
    autoStart: false,
    onComplete: () => console.log("Typing complete!"),
    onType: (text, index) => console.log(`Typed: ${text} (${index} chars)`)
  });

  return (
    <div className="p-4 border rounded-lg space-y-4">
      <h3 className="text-lg font-semibold">Advanced Typewriter with Controls</h3>
      
      <div className="font-mono text-lg bg-gray-100 dark:bg-gray-800 p-3 rounded">
        {displayedText}
        {cursor}
      </div>
      
      <div className="flex gap-2">
        <button 
          onClick={start}
          disabled={isTyping}
          className="px-3 py-1 bg-blue-500 text-white rounded disabled:opacity-50"
        >
          Start
        </button>
        <button 
          onClick={stop}
          disabled={!isTyping}
          className="px-3 py-1 bg-red-500 text-white rounded disabled:opacity-50"
        >
          Stop
        </button>
        <button 
          onClick={reset}
          className="px-3 py-1 bg-gray-500 text-white rounded"
        >
          Reset
        </button>
      </div>
      
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <span className="text-sm">Progress:</span>
          <div className="flex-1 bg-gray-200 rounded-full h-2">
            <div 
              className="bg-blue-500 h-2 rounded-full transition-all duration-300"
              style={{ width: `${progress * 100}%` }}
            />
          </div>
          <span className="text-sm">{Math.round(progress * 100)}%</span>
        </div>
        
        <div className="flex gap-4 text-sm">
          <span>Status: {isTyping ? 'Typing...' : isComplete ? 'Complete' : 'Stopped'}</span>
        </div>
      </div>
      
      <div className="space-y-2">
        <label className="block text-sm font-medium">Change Text:</label>
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          className="w-full p-2 border rounded"
          placeholder="Enter text to type..."
        />
      </div>
    </div>
  );
}

// Streaming Example (like ChatGPT/Claude)
export function StreamingTypewriterExample() {
  const { displayedText, cursor, updateText, appendText, fullText } = useStreamingTypeWriter(30);
  const [inputText, setInputText] = useState("");

  const simulateStream = () => {
    const responses = [
      "I'm an AI assistant",
      " that can help you",
      " with various tasks.",
      " This simulates how",
      " ChatGPT and Claude",
      " stream their responses",
      " in real-time."
    ];
    
    updateText(""); // Reset
    
    responses.forEach((chunk, index) => {
      setTimeout(() => {
        appendText(chunk);
      }, index * 800);
    });
  };

  return (
    <div className="p-4 border rounded-lg space-y-4">
      <h3 className="text-lg font-semibold">Streaming Typewriter (AI-Style)</h3>
      
      <div className="bg-gray-50 dark:bg-gray-900 p-4 rounded border min-h-[100px]">
        <div className="font-mono">
          {displayedText}
          {cursor}
        </div>
      </div>
      
      <div className="flex gap-2">
        <button 
          onClick={simulateStream}
          className="px-4 py-2 bg-green-500 text-white rounded hover:bg-green-600"
        >
          Simulate AI Response
        </button>
        
        <button 
          onClick={() => updateText("")}
          className="px-4 py-2 bg-gray-500 text-white rounded hover:bg-gray-600"
        >
          Clear
        </button>
      </div>
      
      <div className="space-y-2">
        <label className="block text-sm font-medium">Add Custom Text:</label>
        <div className="flex gap-2">
          <input
            type="text"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            className="flex-1 p-2 border rounded"
            placeholder="Enter text to append..."
          />
          <button
            onClick={() => {
              appendText(inputText);
              setInputText("");
            }}
            className="px-3 py-2 bg-blue-500 text-white rounded"
          >
            Append
          </button>
        </div>
      </div>
    </div>
  );
}

// Looping Example
export function LoopingTypewriterExample() {
  const messages = [
    "Building amazing apps...",
    "Creating smooth animations...",
    "Delivering great user experiences...",
    "Powered by React and TypeScript..."
  ];
  
  const [currentMessageIndex, setCurrentMessageIndex] = useState(0);
  
  const { displayedText, cursor } = useTypeWriter({
    text: messages[currentMessageIndex],
    speed: 60,
    loop: true,
    loopDelay: 2000,
    onComplete: () => {
      setTimeout(() => {
        setCurrentMessageIndex((prev) => (prev + 1) % messages.length);
      }, 1000);
    }
  });

  return (
    <div className="p-4 border rounded-lg">
      <h3 className="text-lg font-semibold mb-2">Looping Typewriter</h3>
      <div className="text-xl font-mono text-blue-600 dark:text-blue-400 min-h-[2rem]">
        {displayedText}
        {cursor}
      </div>
      <p className="text-sm text-gray-500 mt-2">
        Message {currentMessageIndex + 1} of {messages.length}
      </p>
    </div>
  );
}

// Custom Cursor Styles Example
export function CustomCursorExample() {
  const [cursorStyle, setCursorStyle] = useState<'default' | 'block' | 'underline'>('default');
  
  const { displayedText, cursorVisible } = useTypeWriter({
    text: "This demonstrates different cursor styles!",
    speed: 70,
    showCursor: false // We'll handle cursor manually
  });

  const renderCustomCursor = () => {
    if (!cursorVisible) return null;
    
    switch (cursorStyle) {
      case 'block':
        return <span className="typewriter-cursor-block" />;
      case 'underline':
        return <span className="typewriter-cursor-underline">_</span>;
      default:
        return <span className="typewriter-cursor">|</span>;
    }
  };

  return (
    <div className="p-4 border rounded-lg space-y-4">
      <h3 className="text-lg font-semibold">Custom Cursor Styles</h3>
      
      <div className="font-mono text-lg bg-gray-100 dark:bg-gray-800 p-3 rounded">
        {displayedText}
        {renderCustomCursor()}
      </div>
      
      <div className="space-y-2">
        <label className="block text-sm font-medium">Cursor Style:</label>
        <div className="flex gap-2">
          {(['default', 'block', 'underline'] as const).map((style) => (
            <button
              key={style}
              onClick={() => setCursorStyle(style)}
              className={`px-3 py-1 rounded capitalize ${
                cursorStyle === style 
                  ? 'bg-blue-500 text-white' 
                  : 'bg-gray-200 dark:bg-gray-700'
              }`}
            >
              {style}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

// Complete Demo Component
export function TypewriterDemo() {
  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <div className="text-center space-y-2">
        <h1 className="text-3xl font-bold">useTypeWriter Hook Examples</h1>
        <p className="text-gray-600 dark:text-gray-400">
          Comprehensive examples showcasing the typewriter hook capabilities
        </p>
      </div>
      
      <div className="grid gap-6">
        <BasicTypewriterExample />
        <AdvancedTypewriterExample />
        <StreamingTypewriterExample />
        <LoopingTypewriterExample />
        <CustomCursorExample />
      </div>
    </div>
  );
}