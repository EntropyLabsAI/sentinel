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
import ContextDisplay from "./context_display"
import JsonDisplay from "./json_display"

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
    const selectedToolChoice = updatedReviewRequest.tool_requests?.[selectedToolIndex];
    sendResponse(decision, selectedToolChoice);
  }

  return (
    <div className="w-full max-w-full mx-auto flex flex-col space-y-4">
      {/* Action Buttons */}
      <div className="w-full flex-shrink-0">
        <h2 className="text-2xl mb-4">
          Agent #<code>{updatedReviewRequest?.run_id}</code> is requesting approval
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
