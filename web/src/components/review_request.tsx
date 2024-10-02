import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { MessageSquare, Info, Hammer, Text, EyeOff, Eye } from "lucide-react"
import { Message, Output, ReviewContext, ReviewRequest, Tool, ToolChoice } from "../review"
import React, { useState } from "react"
import { Button } from "./ui/button"

interface ReviewRequestProps {
  reviewRequest: ReviewRequest
}

export default function ReviewRequestDisplay({ reviewRequest }: ReviewRequestProps) {
  const [showJson, setShowJson] = useState(false)
  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>Review Request: {reviewRequest.id}</span>
          <Badge variant="outline">{reviewRequest.proposed_action}</Badge>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <Accordion type="single" collapsible className="w-full">
          <AccordionItem value="context">
            <AccordionTrigger>
              <div className="flex items-center">
                <Info className="mr-2" />
                Context
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <ContextDisplay context={reviewRequest.context} />
              <div className="mt-4">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowJson(!showJson)}
                  className="flex items-center"
                >
                  {showJson ? <EyeOff className="mr-2 h-4 w-4" /> : <Eye className="mr-2 h-4 w-4" />}
                  {showJson ? "Hide" : "Show"} JSON
                </Button>
                {showJson && (
                  <ScrollArea className="h-[300px] mt-2 p-4 border rounded-md">
                    <pre className="text-xs">{JSON.stringify(reviewRequest, null, 2)}</pre>
                  </ScrollArea>
                )}
              </div>
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </CardContent>
    </Card>
  )
}

function ContextDisplay({ context }: { context: ReviewContext }) {
  return (
    <div className="space-y-4">
      <MessagesDisplay messages={context.messages} />
      {context.tools && context.tools.length > 0 && <ToolsDisplay tools={context.tools} />}
      {context.tool_choice && <ToolChoiceDisplay toolChoice={context.tool_choice} />}
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
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <MessageSquare className="mr-2" />
          Messages
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[200px]">
          {messages.map((message, index) => (
            <div key={index} className="mb-4 last:mb-0">
              <Badge className="mb-1">{message.role}</Badge>
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

function ToolChoiceDisplay({ toolChoice }: { toolChoice: ToolChoice }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm">Tool Choice</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-sm">
          <p><span className="font-semibold">ID:</span> {toolChoice.id}</p>
          <p><span className="font-semibold">Function:</span> {toolChoice.function}</p>
          <p><span className="font-semibold">Type:</span> {toolChoice.type}</p>
          <p><span className="font-semibold">Arguments:</span></p>
          <pre className="text-xs mt-1 bg-muted p-2 rounded">
            {JSON.stringify(toolChoice.arguments, null, 2)}
          </pre>
        </div>
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
