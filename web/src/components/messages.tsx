import { Message, MessageType } from "@/types";
import React, { useRef, useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { MessagesSquareIcon } from "lucide-react";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Dialog, DialogContent, DialogTrigger } from "@/components/ui/dialog";

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
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="messages" className="border border-gray-200 rounded-md">
        <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
          <div className="flex flex-row gap-4">
            <MessagesSquareIcon className="w-4 h-4" />
            Messages
          </div>
        </AccordionTrigger>
        <AccordionContent className="">
          <Card className="border-none">
            <CardContent>
              <div className="max-h-[1000px] overflow-y-auto" ref={scrollAreaRef}>
                {messages.map((message, index) => (
                  <MessageDisplay key={index} message={message} index={index} />
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
  message: Message;
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
        return `${baseStyle} bg-amber-400 text-white`;

    }
  };

  return (
    <div key={index} className={`flex flex-col ${message.role.toLowerCase() === 'user' ? 'items-end' : 'items-start'} mb-4 last:mb-0`}>
      <div className={getBubbleStyle(message.role)}>
        <p className="text-sm font-semibold mb-1">{message.role}</p>
        <MessageTypeDisplay message={message} />
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

const MessageTypeDisplay = ({ message }: { message: Message }) => {
  if (!message.type) {
    return null;
  }

  const formatContent = (content: string) => {
    // Split the content by newlines and wrap each line in a <p> tag
    return content.split('\n').map((line, index) => (
      <p key={index} className="whitespace-pre-wrap">{line}</p>
    ));
  };

  switch (message.type) {
    case MessageType.image_url:
    case MessageType.image:
      return (
        <Dialog>
          <DialogTrigger asChild>
            <img
              className="max-w-[500px] cursor-pointer hover:opacity-90 transition-opacity"
              src={message.content}
              alt="Image"
            />
          </DialogTrigger>
          <DialogContent className="max-w-[90vw] max-h-[90vh] p-0">
            <div className="w-full h-full max-h-[85vh] overflow-auto">
              <img
                className="w-full h-auto"
                src={message.content}
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
