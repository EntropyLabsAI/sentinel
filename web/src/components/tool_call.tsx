import { ToolChoice } from "@/review";
import { Code, Code2, Terminal } from "lucide-react"
import React from "react";

interface ToolChoiceProps {
  toolChoice: ToolChoice
}

export default function ToolChoiceDisplay({ toolChoice }: ToolChoiceProps) {
  const isBashCommand = toolChoice.function === "bash"
  const code = isBashCommand ? toolChoice.arguments.cmd : toolChoice.arguments.code

  console.log(toolChoice)
  return (

    <div className="bg-white shadow-md rounded-lg p-6 max-w-2xl">
      <h2 className="text-2xl font-bold mb-4 flex items-center">
        <Code2 className="mr-2" />
        Tool Call
      </h2>
      <div className="space-y-4">
        <div>
          <span className="font-semibold">ID: </span><code>{toolChoice.id}</code>
        </div>
        <div>
          <span className="font-semibold">Function:</span> <code>{toolChoice.function}</code>
        </div>
        {/* <div>
          <span className="font-semibold">Type:</span> {toolChoice.type}
        </div> */}
        <div>
          {isBashCommand ? (
            <div className="mt-2 bg-black text-white p-4 rounded-md font-mono">
              <div className="flex items-center mb-2">
                <Terminal className="mr-2" size={18} />
                <span className="text-green-400">bash</span>
              </div>
              <div className="pl-6">
                $ {toolChoice.arguments.cmd}
              </div>
            </div>
          ) : (
            <div className="mt-2 bg-black text-white p-4 rounded-md font-mono">
              <div className="flex items-center mb-2">
                <Code className="mr-2" size={18} />
                <span className="text-green-400">python</span>
              </div>
              <div className="pl-6">
                {code}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
