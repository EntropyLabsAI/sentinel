import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageSquare, Info, Hammer, Text, EyeOff, Eye, Code, Check, X, MessageSquare as MessageSquareIcon, SkullIcon } from "lucide-react"
import { Message, Output, TaskState, ReviewRequest, Tool } from "../review"
import ToolChoiceDisplay from "./tool_call"
import React, { useState } from "react"
import { Button } from "./ui/button"

interface ReviewRequestProps {
  reviewRequest: ReviewRequest;
  sendResponse: (decision: string) => void;
}

export default function ReviewRequestDisplay({ reviewRequest, sendResponse }: ReviewRequestProps) {
  return (
    <div className="w-full h-screen max-w-full mx-auto overflow-hidden flex flex-col xl:flex-row">
      {/* Left column (top on large and smaller screens) */}
      <div className="w-full xl:w-1/3 flex-shrink-0 py-4 border-b xl:border-b-0">
        <h2 className="text-2xl font-bold mb-4">Review Request: {reviewRequest.id}</h2>
        {reviewRequest.tool_choice && <ToolChoiceDisplay toolChoice={reviewRequest.tool_choice} />}
        <div className="mt-4 flex flex-wrap gap-2">
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-green-500 hover:bg-green-600"
            onClick={() => sendResponse('approve')}
          >
            <Check className="mr-2 h-4 w-4" /> Approve
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-red-500 hover:bg-red-600"
            onClick={() => sendResponse('reject')}
          >
            <X className="mr-2 h-4 w-4" /> Reject
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-yellow-500 hover:bg-yellow-600"
            onClick={() => sendResponse('terminate')}
          >
            <SkullIcon className="mr-2 h-4 w-4" /> Kill Agent
          </Button>
        </div>
      </div>

      {/* Right column (bottom on large and smaller screens) */}
      <div className="w-full xl:w-2/3 flex-grow overflow-auto p-4">
        <ContextDisplay context={reviewRequest.task_state} />
        <JsonDisplay reviewRequest={reviewRequest} />
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
      <div className="flex items-center justify-between">
        <Badge variant={context.completed ? "default" : "secondary"}>
          {context.completed ? "Completed" : "In Progress"}
        </Badge>
        {context.metadata && Object.keys(context.metadata).length > 0 && (
          <span className="text-sm text-muted-foreground">
            Metadata: {JSON.stringify(context.metadata)}
          </span>
        )}
      </div>
    </div>
  )
}

function MessagesDisplay({ messages }: { messages: Message[] }) {
  const getBadgeColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'assistant':
        return 'bg-blue-500 hover:bg-blue-600';
      case 'user':
        return 'bg-green-500 hover:bg-green-600';
      case 'system':
        return 'bg-purple-500 hover:bg-purple-600';
      default:
        return 'bg-gray-500 hover:bg-gray-600';
    }
  };

  return (
    <Card className="">
      <CardHeader>
        <CardTitle className="flex items-center">
          <MessageSquare className="mr-2" />
          Messages
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[500px]">
          {messages.map((message, index) => (
            <div key={index} className="mb-4 last:mb-0">
              <Badge className={`mb-1 ${getBadgeColor(message.role)}`}>
                {message.role}
              </Badge>
              <p className="text-sm">{message.content}</p>
              {message.source && (
                <span className="text-xs text-muted-foreground">Source: {message.source}</span>
              )}
              {message.tool_calls && (
                <div className="mt-2">
                  <span className="text-xs font-semibold">Tool Calls:</span>
                  {message.tool_calls.map((toolCall, idx) => (
                    <div key={idx} className="ml-2 text-xs">
                      <span className="font-semibold">{toolCall.function}:</span> {JSON.stringify(toolCall.arguments)}
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </ScrollArea>
      </CardContent>
    </Card>
  )
}

function ToolsDisplay({ tools }: { tools: Tool[] }) {
  console.log(tools)
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Hammer className="mr-2" />
          Tools
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[150px]">
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
        </ScrollArea>
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
          {output.choices.map((choice, index) => (
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
  const [showJson, setShowJson] = useState(false)
  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span className="flex items-center">
            <Code className="mr-2" />
            JSON Data
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setShowJson(!showJson)}
          >
            {showJson ? "Hide" : "Show"} JSON
          </Button>
        </CardTitle>
      </CardHeader>
      <CardContent>
        {showJson && (
          <ScrollArea className="h-[300px]">
            <pre className="text-xs">{JSON.stringify(reviewRequest, null, 2)}</pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
