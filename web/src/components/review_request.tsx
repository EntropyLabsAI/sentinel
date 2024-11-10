import { Check, X, SkullIcon } from "lucide-react"
import { ReviewPayload, ToolRequest, Decision } from "@/types"
import ToolChoiceDisplay from "./tool_call"
import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import ContextDisplay from "@/components/context_display"

interface ReviewRequestProps {
  reviewPayload: ReviewPayload;
  sendResponse: (decision: Decision, toolChoice: ToolRequest) => void;
}

export default function ReviewRequestDisplay({ reviewPayload, sendResponse }: ReviewRequestProps) {
  const { supervision_request, request_group, chain_state } = reviewPayload;
  const [selectedToolIndex, setSelectedToolIndex] = useState(0);

  useEffect(() => {
    setSelectedToolIndex(0); // Initialize the first tool as selected
  }, [reviewPayload]);

  const toolRequests = request_group.tool_requests || [];

  function handleToolChoiceChange(updatedToolChoice: ToolRequest, index: number) {
    // Here you can handle any changes to the tool choice if your UI allows editing
  }

  function handleSendResponse(decision: Decision) {
    const selectedToolChoice = toolRequests[selectedToolIndex];
    sendResponse(decision, selectedToolChoice);
  }

  return (
    <div className="w-full max-w-full mx-auto flex flex-col space-y-4">
      {/* Action Buttons */}
      <div className="w-full flex-shrink-0">
        <h2 className="text-2xl mb-4">
          Agent{' '}
          <code>{supervision_request.chainexecution_id}</code> is requesting approval
        </h2>
        <div className="my-4 flex flex-wrap gap-2">
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-green-500 hover:bg-green-600 text-white"
            onClick={() => handleSendResponse(Decision.approve)}
          >
            <Check className="mr-2 h-4 w-4" /> Approve
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-yellow-500 hover:bg-yellow-600 text-white"
            onClick={() => handleSendResponse(Decision.reject)}
          >
            <X className="mr-2 h-4 w-4" /> Reject
          </Button>
          <Button
            variant="default"
            size="sm"
            className="flex-1 bg-red-500 hover:bg-red-600 text-white"
            onClick={() => handleSendResponse(Decision.terminate)}
          >
            <SkullIcon className="mr-2 h-4 w-4" /> Kill Agent
          </Button>
        </div>

        {/* Tool Choices */}
        <div className="space-y-4">
          {toolRequests.map((toolChoice, index) => (
            <ToolChoiceDisplay
              key={index}
              toolChoice={toolChoice}
              lastMessage={toolChoice.message}
              onToolChoiceChange={(updatedToolChoice) =>
                handleToolChoiceChange(updatedToolChoice, index)
              }
              isSelected={selectedToolIndex === index}
              onSelect={() => setSelectedToolIndex(index)}
              index={index + 1}
            />
          ))}
        </div>
      </div>

      {/* Context Display */}
      <div className="w-full flex-grow overflow-auto">
        <ContextDisplay context={toolRequests[selectedToolIndex].task_state} />
      </div>
    </div>
  )
}
