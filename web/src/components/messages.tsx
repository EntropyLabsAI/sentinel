import { Message, StateMessage } from "@/types";
import React, { useRef, useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "./ui/card";
import { MessageSquare } from "lucide-react";

export function MessagesDisplay({ messages }: { messages: Message[] }) {
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    setIsLoaded(true);
  }, []);

  useEffect(() => {
    if (isLoaded && scrollAreaRef.current) {
      setTimeout(() => {
        if (scrollAreaRef.current) {
          scrollAreaRef.current.scrollTop = scrollAreaRef.current.scrollHeight;
        }
      }, 100);
    }
  }, [messages, isLoaded]);

  return (
    <Card className="border-none">
      <CardHeader>
      </CardHeader>
      <CardContent>
        <div className="overflow-auto" ref={scrollAreaRef}>
          {messages.map((message, index) => (
            <MessageDisplay key={index} message={message} index={index} />
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

interface MessageDisplayProps {
  message: StateMessage;
  index: number;
}

export function MessageDisplay({ message, index }: MessageDisplayProps) {
  const getBubbleStyle = (role: string) => {
    const baseStyle = "rounded-2xl p-3 mb-2 break-words";
    switch (role.toLowerCase()) {
      case 'assistant':
        return `${baseStyle} bg-blue-500 text-white`;
      case 'user':
        return `${baseStyle} bg-gray-200 text-gray-800`;
      case 'system':
        return `${baseStyle} bg-gray-300 text-gray-800 italic`;
      default:
        return `${baseStyle} bg-gray-400 text-white`;
    }
  };

  const formatContent = (content: string) => {
    // Split the content by newlines and wrap each line in a <p> tag
    return content.split('\n').map((line, index) => (
      <p key={index} className="whitespace-pre-wrap">{line}</p>
    ));
  };
  return (
    <div key={index} className={`flex flex-col ${message.role.toLowerCase() === 'user' ? 'items-end' : 'items-start'} mb-4 last:mb-0`}>
      <div className={getBubbleStyle(message.role)}>
        <p className="text-sm font-semibold mb-1">{message.role}</p>
        <div className="text-sm">{formatContent(message.content)}</div>
        {message.source && (
          <p className="text-xs opacity-70 mt-1">Source: {message.source}</p>
        )}
        {message.tool_calls && (
          <div className="mt-2">
            <p className="text-xs font-semibold">Tool Calls:</p>
            <code>
              {message.tool_calls.map((toolCall, idx) => (
                <div key={idx} className="ml-2 text-xs">
                  <span className="font-semibold">{toolCall.function}:</span> {JSON.stringify(toolCall.arguments)}
                </div>
              ))}
            </code>
          </div>
        )}
      </div>
    </div>
  )
}

