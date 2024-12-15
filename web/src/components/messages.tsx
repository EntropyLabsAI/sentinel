import { AsteroidMessage, MessageType } from "@/types";
import React, { useRef, useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { ChevronLeftIcon, ChevronRightIcon, Key, MessagesSquareIcon } from "lucide-react";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { Button } from "./ui/button";

interface MessagesDisplayProps {
  expanded: boolean;
  messages: AsteroidMessage[];
  onToolCallClick: (toolCallId: string) => void;
  selectedToolCallId?: string;
  index?: number;
  setIndex?: (index: number) => void;
  chatCount?: number;
}

export function MessagesDisplay({
  expanded,
  messages,
  onToolCallClick,
  selectedToolCallId,
  index,
  setIndex,
  chatCount
}: MessagesDisplayProps) {
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    setIsLoaded(true);
  }, []);

  function handlePrevious() {
    if (!setIndex || !chatCount || index === undefined) return;

    if (index < chatCount - 1) {
      setIndex(index + 1);
    } else {
      setIndex(0);
    }
    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  }

  function handleNext() {
    if (!setIndex || !chatCount || index === undefined) return;

    if (index > 0) {
      setIndex(index - 1);
    } else {
      setIndex(chatCount - 1);
    }

    if (scrollAreaRef.current) {
      scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
    }
  }

  const highlightedToolCallId = selectedToolCallId ? selectedToolCallId : undefined;

  return (
    <Accordion type="single" collapsible className="w-full" defaultValue={expanded ? "messages" : undefined}>
      <AccordionItem value="messages" className="border border-gray-200 rounded-md">
        <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
          <div className="flex flex-row gap-4 items-center">
            <MessagesSquareIcon className="w-4 h-4" />
            Messages
          </div>
        </AccordionTrigger>
        <AccordionContent>
          {index !== undefined && setIndex && chatCount && (
            <div className="flex flex-row gap-2 justify-end p-4">
              <Button
                onClick={handlePrevious}
                disabled={index >= chatCount - 1}
                className={cn("hover:bg-gray-500 bg-gray-400 text-black", index >= chatCount - 1 && "opacity-50 cursor-not-allowed")}
              >
                <ChevronLeftIcon className="w-4 h-4" />
                Previous
              </Button>
              <Button
                onClick={handleNext}
                disabled={index === 0}
                className={cn("hover:bg-gray-500 bg-gray-400 text-black", index === 0 && "opacity-50 cursor-not-allowed")}
              >
                Next
                <ChevronRightIcon className="w-4 h-4" />
              </Button>
            </div>
          )}
          <Card className="border-none">
            <CardContent>
              <div className="max-h-[1000px] overflow-y-auto" ref={scrollAreaRef}>
                {messages.map((message, index) => (
                  <MessageDisplay
                    key={`${index}-${highlightedToolCallId}`}
                    message={message}
                    index={index}
                    highlightedToolCallId={highlightedToolCallId}
                    onToolCallClick={onToolCallClick}
                  />
                ))}
              </div>
            </CardContent>
          </Card>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  )
}

interface MessageDisplayProps {
  message: AsteroidMessage;
  index: number;
  highlightedToolCallId?: string;
  onToolCallClick: (toolCallId: string) => void;
}

export function MessageDisplay({ message, index, highlightedToolCallId, onToolCallClick }: MessageDisplayProps) {
  const getBubbleStyle = (role: string) => {
    const baseStyle = "rounded-2xl p-3 mb-2 break-words";
    switch (role.toLowerCase()) {
      case 'assistant':
        return `${baseStyle} bg-blue-500 text-white`;
      case 'user':
        return `${baseStyle} bg-gray-200 text-gray-800`;
      case 'system':
        return `${baseStyle} bg-gray-300 text-gray-800 italic`;
      case 'asteroid':
        return `${baseStyle} bg-teal-800 text-white`;
      default:
        return `${baseStyle} bg-amber-400 text-white`;
    }
  };

  return (
    <div key={index} className={`flex flex-col ${message.role.toLowerCase() === 'user' ? 'items-end' : 'items-start'} mb-4 pr-4 last:mb-0`}>
      <div className={getBubbleStyle(message.role)}>
        <p className="text-sm font-semibold mb-1">{message.role}</p>
        <MessageTypeDisplay message={message} />
        {message.tool_calls && message.tool_calls.length > 0 && (
          <div className="mt-2">
            <p className="text-xs font-semibold">{message.tool_calls.length} tool call{message.tool_calls.length === 1 ? "" : "s"} in this message</p>
            <div className="flex flex-wrap mt-2">
              {message.tool_calls.map((toolCall, idx) => (
                <Badge
                  key={`${index}-${idx}`}
                  className={cn(`mr-2 mb-2 cursor-pointer`, highlightedToolCallId === toolCall.call_id ? "bg-teal-100 text-teal-800" : "bg-gray-100 text-gray-800")}
                  onClick={() => onToolCallClick(toolCall.call_id || '')}
                >
                  {toolCall.name || 'No name provided'}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

const MessageTypeDisplay = ({ message }: { message: AsteroidMessage }) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const MAX_CHARS = 1000;

  if (!message.type) {
    return null;
  }

  const formatContent = (content: string) => {
    if (message.type === MessageType.text && content.length > MAX_CHARS && !isExpanded) {
      // Show truncated content with "Show More" button
      return (
        <>
          <p className="whitespace-pre-wrap">
            {content.slice(0, MAX_CHARS)}...
          </p>
          <button
            onClick={() => setIsExpanded(true)}
            className="text-sm underline mt-1 opacity-80 hover:opacity-100"
          >
            Show More
          </button>
        </>
      );
    }

    // Show full content with "Show Less" button if expanded
    return (
      <>
        <p className="whitespace-pre-wrap">{content}</p>
        {message.type === MessageType.text && content.length > MAX_CHARS && (
          <button
            onClick={() => setIsExpanded(false)}
            className="text-sm underline mt-2 opacity-80 hover:opacity-100"
          >
            Show Less
          </button>
        )}
      </>
    );
  };

  switch (message.type) {
    case MessageType.image_url:
    case MessageType.image:
      console.log("Image message: ", message);
      // Check if the content already has data:image/jpeg;base64, in it
      const imgSrc = message.content.startsWith("data:image/jpeg;base64,")
        ? message.content
        : `data:image/jpeg;base64,${message.content}`;
      return (
        <Dialog>
          <DialogTrigger asChild>
            <img
              className="max-w-[500px] cursor-pointer hover:opacity-90 transition-opacity"
              src={imgSrc}
              alt="Image"
            />
          </DialogTrigger>
          <DialogContent className="max-w-[90vw] max-h-[90vh] p-0">
            <div className="w-full h-full max-h-[85vh] overflow-auto">
              <img
                className="w-full h-auto"
                src={imgSrc}
                alt="Image"
              />
            </div>
          </DialogContent>
        </Dialog>
      );
    case MessageType.text:
      return <div>{formatContent(message.content)}</div>
    case MessageType.audio:
      return <audio src={message.content} controls />
    default:
      return <div>Unknown message type: {message.type}</div>
  }
}
