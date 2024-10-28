import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { MessageSquare, Info, Hammer, Text, Code, Check, X, SkullIcon } from "lucide-react"
import { LLMMessage, TaskState, SupervisionRequest, Tool, ToolRequest, Decision, Output } from "@/types"
import ToolChoiceDisplay from "./tool_call"
import React, { useState, useEffect, useRef } from "react"
import { Button } from "./ui/button"
import CopyButton from "./copy_button"
import { MessagesDisplay } from "./messages"

interface ReviewRequestProps {
  reviewRequest: SupervisionRequest;
  sendResponse: (decision: Decision, toolChoice: ToolRequest) => void;
}

export default function ReviewRequestDisplay({ reviewRequest, sendResponse }: ReviewRequestProps) {
  const [updatedReviewRequest, setUpdatedReviewRequest] = useState(reviewRequest);
  const [selectedToolIndex, setSelectedToolIndex] = useState(0); // Added state for selected tool

  useEffect(() => {
    setUpdatedReviewRequest(reviewRequest);
    setSelectedToolIndex(0); // Initialize the first tool as selected
  }, [reviewRequest]);

  function handleToolChoiceChange(updatedToolChoice: ToolRequest, index: number) {
    const updatedToolChoices = [...(updatedReviewRequest.tool_requests || [])];
    updatedToolChoices[index] = updatedToolChoice;

    const updatedReview = {
      ...updatedReviewRequest,
      tool_choices: updatedToolChoices,
    };

    setUpdatedReviewRequest(updatedReview);
  }

  function handleSendResponse(decision: Decision) {
    const selectedToolChoice = updatedReviewRequest.tool_requests[selectedToolIndex];
    sendResponse(decision, selectedToolChoice);
  }

  return (
    <div className="w-full max-w-full mx-auto flex flex-col space-y-4">
      {/* Action Buttons */}
      <div className="w-full flex-shrink-0">
        <h2 className="text-2xl mb-4">
          Agent #<code>{updatedReviewRequest.run_id.slice(0, 8)}</code> is requesting approval
        </h2>
        <div className="my-4 flex flex-wrap gap-2">
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-green-500 hover:bg-green-600 text-white"
            onClick={() => handleSendResponse('approve')}
          >
            <Check className="mr-2 h-4 w-4" /> Approve
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-yellow-500 hover:bg-yellow-600 text-white"
            onClick={() => handleSendResponse('reject')}
          >
            <X className="mr-2 h-4 w-4" /> Reject
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-red-500 hover:bg-red-600 text-white"
            onClick={() => handleSendResponse('terminate')}
          >
            <SkullIcon className="mr-2 h-4 w-4" /> Kill Agent
          </Button>
        </div>

        {/* Tool Choices */}
        <div className="space-y-4">
          {updatedReviewRequest.tool_requests &&
            updatedReviewRequest.messages &&
            updatedReviewRequest.tool_requests.map((toolChoice, index) => (
              <ToolChoiceDisplay
                key={index}
                toolChoice={toolChoice}
                lastMessage={updatedReviewRequest.messages[index]}
                onToolChoiceChange={(updatedToolChoice) => handleToolChoiceChange(updatedToolChoice, index)}
                isSelected={selectedToolIndex === index}
                onSelect={() => setSelectedToolIndex(index)}
                index={index + 1}
              />
            ))}
        </div>
      </div>

      {/* Context Display */}
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
    </div>
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
          <p>Input Tokens: {output.usage?.input_tokens}</p>
          <p>Output Tokens: {output.usage?.output_tokens}</p>
          <p>Total Tokens: {output.usage?.total_tokens}</p>
        </div>
      </CardContent>
    </Card>
  )
}

function JsonDisplay({ reviewRequest }: { reviewRequest: SupervisionRequest }) {
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
