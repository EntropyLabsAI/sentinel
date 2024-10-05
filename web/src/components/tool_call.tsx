import { ToolChoice } from "@/review";
import { Code, Code2 } from "lucide-react"
import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import CopyButton from "./copy_button"
import ExplainButton from "./ask_lm"

interface ToolChoiceProps {
  toolChoice: ToolChoice
}

export default function ToolChoiceDisplay({ toolChoice }: ToolChoiceProps) {
  const isBashCommand = toolChoice.function === "bash"
  const code = isBashCommand ? toolChoice.arguments.cmd : toolChoice.arguments.code
  const [explanation, setExplanation] = useState<string | null>(null)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Code2 className="mr-2" />
          Tool Call
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div>
            <span className="font-semibold">ID: </span><code>{toolChoice.id}</code>
          </div>
          <div>
            <span className="font-semibold">Function:</span> <code>{toolChoice.function}</code>
          </div>
          <div>
            {isBashCommand ? (
              <div className="mt-2 bg-black p-4 rounded-md font-mono">
                <div className="flex items-center mb-2 justify-between">
                  <div className="pl-6 text-white">
                    <span className="text-green-400">$ </span>
                    {toolChoice.arguments.cmd}
                  </div>
                  <div className="flex items-center">
                    <CopyButton text={toolChoice.arguments.cmd} />
                    <ExplainButton text={toolChoice.arguments.cmd} onExplanation={setExplanation} />
                  </div>
                </div>
                {explanation && (
                  <div className="mt-2 text-sm text-gray-300 bg-gray-800 p-2 rounded">
                    <strong>Explanation:</strong> {explanation}
                  </div>
                )}
              </div>
            ) : (
              <div className="mt-2 bg-black text-white p-4 rounded-md font-mono">
                <div className="flex items-center mb-2 justify-between">
                  <div className="flex items-center">
                    <Code className="mr-2" size={18} />
                    <span className="text-green-400">python</span>
                  </div>
                  <div className="flex items-center">
                    <CopyButton text={code} />
                    <ExplainButton text={code} onExplanation={setExplanation} />
                  </div>
                </div>
                <div className="pl-6">
                  {code}
                </div>
                {explanation && (
                  <div className="mt-2 text-sm text-gray-300 bg-gray-800 p-2 rounded">
                    <strong>Explanation:</strong> {explanation}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
