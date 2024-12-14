import { Check, X, SkullIcon, MessagesSquareIcon, ClockIcon, CodeIcon, Copy } from "lucide-react"
import { ReviewPayload, Decision, SentinelToolCall } from "@/types"
// import ToolChoiceDisplay from "../tool_call"
import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import ToolsDisplay from "../tool_display"
import { MessagesDisplay } from "../messages"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "../ui/accordion"
import ChainStateDisplay from "../chain_state_display"
import CopyButton from "../util/copy_button"
import { ToolCallState } from "../tool_call_state"

interface ReviewRequestProps {
  reviewPayload: ReviewPayload;
  sendResponse: (decision: Decision, toolcall: SentinelToolCall, feedback?: string) => void;
}

export default function ReviewRequestDisplay({ reviewPayload, sendResponse }: ReviewRequestProps) {
  const { supervision_request, toolcall, chain_state } = reviewPayload;
  const [selectedToolIndex, setSelectedToolIndex] = useState(0);

  useEffect(() => {
    setSelectedToolIndex(0); // Initialize the first tool as selected
  }, [reviewPayload]);

  function handleSendResponse(decision: Decision, feedback?: string) {
    sendResponse(decision, toolcall, feedback);
  }

  return (
    <div className="flex flex-col space-y-4">
      {/* Action Buttons */}
      <div className="w-full flex-shrink-0">
        <div className="mb-4 flex flex-wrap gap-2">
          <Button
            variant="outline"
            size="lg"
            className="flex-1 hover:bg-green-500 hover:text-white text-green-500"
            onClick={() => handleSendResponse(Decision.approve)}
          >
            <Check className="mr-2 h-4 w-4" /> Approve
          </Button>
          <Button
            variant="outline"
            size="lg"
            className="flex-1 hover:bg-yellow-500 hover:text-white text-yellow-500"
            onClick={() => handleSendResponse(Decision.reject)}
          >
            <X className="mr-2 h-4 w-4" />
            <p className="font-bold">Reject</p>
          </Button>
          <Button
            variant="outline"
            size="lg"
            className="flex-1 hover:bg-red-500 hover:text-white text-red-500"
            onClick={() => handleSendResponse(Decision.terminate)}
          >
            <SkullIcon className="mr-2 h-4 w-4" /> Kill Agent
          </Button>
        </div>

        {/* Tool Choices */}
        {/* <div className="space-y-4">
          <ToolChoiceDisplay
            toolChoice={toolcall}
            lastMessage={toolcall.message}
            onToolChoiceChange={(updatedToolChoice) =>
              handleToolChoiceChange(updatedToolChoice, index)
            }
              isSelected={selectedToolIndex === index}
              onSelect={() => setSelectedToolIndex(index)}
              index={index + 1}
              runId={reviewPayload.run_id}
            />
        </div> */}
      </div>
      <ToolCallState toolCallId={toolcall.call_id} />

      {/* Context Display */}
      <MessagesDisplay messages={reviewPayload.messages} onToolCallClick={() => { }} expanded={true} />

      {/* Chain State Display */}
      <ChainStateDisplay
        chainState={chain_state}
        currentRequestId={supervision_request.id}
      />

      {/* Tools
      {toolRequests[selectedToolIndex].task_state.tools && toolRequests[selectedToolIndex].task_state.tools.length > 0 && (
        <ToolsDisplay tools={toolRequests[selectedToolIndex].task_state.tools} />
      )} */}

    </div>
  )
}
