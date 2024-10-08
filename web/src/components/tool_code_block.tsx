import React, { useState, useRef, useEffect } from "react";
import { Code, X, AlertTriangle, AlertCircle, CheckCircle } from "lucide-react";
import { Textarea } from "./ui/textarea";
import { Button } from "./ui/button";
import CopyButton from "./copy_button";
import ExplainButton from "./ask_lm";

interface ToolCodeBlockProps {
  isBashCommand: boolean;
  code: string;
  handleCodeChange: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
  explanation: string | null;
  setExplanation: React.Dispatch<React.SetStateAction<string | null>>;
  score: string | null;
  setScore: React.Dispatch<React.SetStateAction<string | null>>;
  resetExplanation: () => void;
  resetScore: () => void;
}

export default function ToolCodeBlock({
  isBashCommand,
  code,
  handleCodeChange,
  explanation,
  setExplanation,
  score,
  setScore,
  resetExplanation,
  resetScore,
}: ToolCodeBlockProps) {
  const [editting, setEditting] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setEditting(false);
      }
    }

    document.addEventListener("mousedown", handleClickOutside);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [ref]);

  // Auto-resize effect
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = textareaRef.current.scrollHeight + "px";
    }
  }, [code]);

  const getScoreStyle = (score: string) => {
    switch (score.toLowerCase()) {
      case 'harmless':
        return {
          bgColor: 'bg-green-800',
          textColor: 'text-green-200',
          Icon: CheckCircle
        };
      case 'risky':
        return {
          bgColor: 'bg-yellow-800',
          textColor: 'text-yellow-200',
          Icon: AlertTriangle
        };
      case 'dangerous':
        return {
          bgColor: 'bg-red-800',
          textColor: 'text-red-200',
          Icon: AlertCircle
        };
      default:
        return {
          bgColor: 'bg-gray-800',
          textColor: 'text-gray-200',
          Icon: AlertCircle
        };
    }
  };

  return (
    <div
      ref={ref}
      className="mt-2 bg-black p-4 rounded-md font-mono"
    >
      <div
        className="flex items-center"
      >
        {/* Left side: $ or Python icon */}
        <div className="flex mr-2">
          {isBashCommand ? (
            <span className="text-green-400">$</span>
          ) : (
            <>
              <Code className="text-green-400 mr-1" size={18} />
              <span className="text-green-400">python</span>
            </>
          )}
        </div>
        {/* Middle: Textarea or code display */}
        <div className="flex-grow"
          onClick={() => setEditting(true)}

        >
          {editting ? (
            <Textarea
              ref={textareaRef}
              value={code}
              onChange={(e) => {
                handleCodeChange(e);
                if (textareaRef.current) {
                  textareaRef.current.style.height = "auto";
                  textareaRef.current.style.height =
                    textareaRef.current.scrollHeight + "px";
                }
              }}
              className="w-full bg-gray-800 text-white text-md border-none resize-none overflow-hidden"
              style={{
                lineHeight: "1.5",
                paddingTop: "0.375rem",
                paddingBottom: "0.375rem",
              }}
            />
          ) : (
            <Textarea
              ref={textareaRef}
              value={code}
              className="w-full bg-transparent text-white text-md border-none resize-none overflow-hidden"
              style={{
                lineHeight: "1.5",
                paddingTop: "0.375rem",
                paddingBottom: "0.375rem",
              }}
            />
          )}
        </div>
        {/* Right side: Buttons */}
        <div className="flex items-center ml-2">
          <CopyButton text={code} />
          <ExplainButton text={code} onExplanation={setExplanation} onScore={setScore} />
        </div>
      </div>
      {explanation && (
        <div className="mt-2 text-sm text-gray-300 bg-gray-800 p-2 rounded flex flex-row justify-between">
          <p>{explanation}</p>
          <Button
            size="icon"
            onClick={() => {
              resetExplanation();
              resetScore();
            }}
            className="ml-2 p-2 bg-gray-700 hover:bg-gray-600 outline-none"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      )}
      {score && (
        <div className={`mt-2 text-sm p-2 rounded flex flex-row items-center ${getScoreStyle(score).bgColor} ${getScoreStyle(score).textColor}`}>
          {React.createElement(getScoreStyle(score).Icon, { className: "h-4 w-4 mr-2" })}
          <p className="flex-grow">{score}</p>
        </div>
      )}
    </div>
  );
}
