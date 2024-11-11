import { Check, X, SkullIcon, MessagesSquareIcon, ClockIcon, CodeIcon, Copy } from "lucide-react"
import { ReviewPayload, ToolRequest, Decision } from "@/types"
import ToolChoiceDisplay from "./tool_call"
import React, { useState, useEffect } from "react"
import { Button } from "@/components/ui/button"
import ToolsDisplay from "./tool_display"
import { MessagesDisplay } from "./messages"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "./ui/accordion"
import ChainStateDisplay from "./chain_state_display"
import CopyButton from "./util/copy_button"

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
    // TODO: handle editting tool choice
  }

  function handleSendResponse(decision: Decision) {
    const selectedToolChoice = toolRequests[selectedToolIndex];
    sendResponse(decision, selectedToolChoice);
  }

  return (
    <div className="w-full max-w-full mx-auto flex flex-col space-y-4">
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
      <MessagesDisplay messages={toolRequests[selectedToolIndex].task_state.messages} />

      {/* Chain State Display */}
      <ChainStateDisplay
        chainState={chain_state}
        currentRequestId={supervision_request.id}
      />

      {/* Tools */}
      {toolRequests[selectedToolIndex].task_state.tools && toolRequests[selectedToolIndex].task_state.tools.length > 0 && (
        <ToolsDisplay tools={toolRequests[selectedToolIndex].task_state.tools} />
      )}

      {/* Raw JSON */}
      <Accordion type="single" collapsible className="w-full">
        <AccordionItem value="messages" className="border border-gray-200 rounded-md">
          <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
            <div className="flex flex-row gap-4 items-center justify-between w-full">
              <div className="flex flex-row gap-4">
                <CodeIcon className="w-4 h-4" />
                Raw Task State JSON
              </div>
              <CopyButton className="mr-4 bg-gray-100 hover:bg-gray-200 text-gray-800" text={JSON.stringify(toolRequests[selectedToolIndex].task_state, null, 2)} />
            </div>
          </AccordionTrigger>
          <AccordionContent className="p-4">
            <pre>{JSON.stringify(toolRequests[selectedToolIndex].task_state, null, 2)}</pre>
          </AccordionContent>
        </AccordionItem>
      </Accordion>

    </div>
  )
}
