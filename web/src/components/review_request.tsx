import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageSquare, Info, Hammer, Text, EyeOff, Eye, Code, Check, X, MessageSquare as MessageSquareIcon, SkullIcon } from "lucide-react"
import { Message, Output, TaskState, ReviewRequest, Tool, ToolChoice } from "../review"
import ToolChoiceDisplay from "./tool_call"
import React, { useState, useEffect, useRef } from "react"
import { Button } from "./ui/button"
import CopyButton from "./copy_button"

interface ReviewRequestProps {
  reviewRequest: ReviewRequest;
  sendResponse: (decision: string, updatedReviewRequest: ReviewRequest) => void;
}

export default function ReviewRequestDisplay({ reviewRequest, sendResponse }: ReviewRequestProps) {
  const [updatedReviewRequest, setUpdatedReviewRequest] = useState(reviewRequest);

  // Update the state when the prop changes
  useEffect(() => {
    setUpdatedReviewRequest(reviewRequest);
  }, [reviewRequest]);

  function handleToolChoiceChange(updatedToolChoice: ToolChoice, index: number) {
    const updatedToolChoices = [...(updatedReviewRequest.tool_choices || [])];
    updatedToolChoices[index] = updatedToolChoice;

    const updatedReview = {
      ...updatedReviewRequest,
      tool_choice: updatedToolChoices,
    };

    setUpdatedReviewRequest(updatedReview);
  }

  return (
    <div className="w-full max-w-full mx-auto flex flex-col space-y-4">
      {/* Button/Tool column (always on top) */}
      <div className="w-full flex-shrink-0">
        <h2 className="text-2xl mb-4">
          Agent #<code>{updatedReviewRequest.agent_id.slice(0, 8)}</code> is requesting approval
        </h2>
        <div className="my-4 flex flex-wrap gap-2">
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-green-500 hover:bg-green-600 text-white"
            onClick={() => sendResponse('approve', updatedReviewRequest)}
          >
            <Check className="mr-2 h-4 w-4" /> Approve
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-yellow-500 hover:bg-yellow-600 text-white"
            onClick={() => sendResponse('reject', updatedReviewRequest)}
          >
            <X className="mr-2 h-4 w-4" /> Reject
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-red-500 hover:bg-red-600 text-white"
            onClick={() => sendResponse('terminate', updatedReviewRequest)}
          >
            <SkullIcon className="mr-2 h-4 w-4" /> Kill Agent
          </Button>
        </div>

        {/* Map over the arrays of tool_choice and last_message */}
        {updatedReviewRequest.tool_choices &&
          updatedReviewRequest.last_messages &&
          updatedReviewRequest.tool_choices.map((toolChoice, index) => (
            <ToolChoiceDisplay
              key={index}
              toolChoice={toolChoice}
              lastMessage={updatedReviewRequest.last_messages[index]}
              onToolChoiceChange={(updatedToolChoice) => handleToolChoiceChange(updatedToolChoice, index)}
            />
          ))}
      </div>

      {/* Context column (always below) */}
      <div className="w-full flex-grow overflow-auto">
        <ContextDisplay context={updatedReviewRequest.task_state} />
        <JsonDisplay reviewRequest={updatedReviewRequest} />
      </div>
    </div>
  )
}

function ContextDisplay({ context }: { context: TaskState }) {
  return (
    <div className="space-y-4">
      <MessagesDisplay messages={context.messages} />
      {context.tools && context.tools.length > 0 && <ToolsDisplay tools={context.tools} />}
      {context.output && <OutputDisplay output={context.output} />}
      {/* <div className="flex items-center justify-between">
        <Badge variant={context.completed ? "default" : "secondary"}>
          {context.completed ? "Completed" : "In Progress"}
        </Badge>
        {context.metadata && Object.keys(context.metadata).length > 0 && (
          <span className="text-sm text-muted-foreground">
            Metadata: {JSON.stringify(context.metadata)}
          </span>
        )}
      </div> */}
    </div>
  )
}

function MessagesDisplay({ messages }: { messages: Message[] }) {
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
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <MessageSquare className="mr-2" />
          Messages
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="h-[500px] overflow-auto" ref={scrollAreaRef}>
          {messages.map((message, index) => (
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
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

function ToolsDisplay({ tools }: { tools: Tool[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Hammer className="mr-2" />
          Tools
        </CardTitle>
      </CardHeader>
      <CardContent>
        {tools.map((tool, index) => (
          <div key={index} className="mb-2 last:mb-0">
            <Badge variant="outline" className="mb-1">{tool.name}</Badge>
            {tool.description && <p className="text-sm">{tool.description}</p>}
            {tool.attributes && (
              <pre className="text-xs mt-1 bg-muted p-2 rounded">
                {JSON.stringify(tool.attributes, null, 2)}
              </pre>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}

function OutputDisplay({ output }: { output: Output }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Text className="mr-2" />
          Output
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm mb-2"><span className="font-semibold">Model:</span> {output.model}</p>
        <div className="space-y-2">
          {output.choices && output.choices.map((choice, index) => (
            <div key={index} className="border-t pt-2 first:border-t-0 first:pt-0">
              <Badge className="mb-1">{choice.message.role}</Badge>
              <p className="text-sm">{choice.message.content}</p>
              <span className="text-xs text-muted-foreground">Stop Reason: {choice.stop_reason}</span>
            </div>
          ))}
        </div>
        <div className="mt-4 text-xs text-muted-foreground">
          <p>Input Tokens: {output.usage.input_tokens}</p>
          <p>Output Tokens: {output.usage.output_tokens}</p>
          <p>Total Tokens: {output.usage.total_tokens}</p>
        </div>
      </CardContent>
    </Card>
  )
}

function JsonDisplay({ reviewRequest }: { reviewRequest: ReviewRequest }) {
  const [showJson, setShowJson] = useState(true)
  const jsonString = JSON.stringify(reviewRequest, null, 2)

  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span className="flex items-center">
            <Code className="mr-2" />
            Task State JSON
          </span>
          <div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setShowJson(!showJson)}
              className="mr-2"
            >
              {showJson ? "Hide" : "Show"} JSON
            </Button>
            {showJson && <CopyButton text={jsonString} />}
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {showJson && (
          <ScrollArea className="h-[300px]">
            <pre className="text-xs">{jsonString}</pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
