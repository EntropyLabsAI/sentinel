import { ToolChoice } from "@/review";
import { Code, Code2, X } from "lucide-react"
import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import CopyButton from "./copy_button"
import ExplainButton from "./ask_lm"
import { Button } from "./ui/button";
import { Textarea } from "./ui/textarea"; // Make sure to import Textarea
import ToolCodeBlock from "./tool_code_block";

interface ToolChoiceProps {
  toolChoice: ToolChoice;
  onToolChoiceChange: (updatedToolChoice: ToolChoice) => void; // New prop for handling changes
}

export default function ToolChoiceDisplay({ toolChoice, onToolChoiceChange }: ToolChoiceProps) {
  const isBashCommand = toolChoice.function === "bash";
  const [code, setCode] = useState(
    isBashCommand ? toolChoice.arguments.cmd : toolChoice.arguments.code
  );
  const [explanation, setExplanation] = useState<string | null>(null);
  const [score, setScore] = useState<string | null>(null);

  useEffect(() => {
    setCode(isBashCommand ? toolChoice.arguments.cmd : toolChoice.arguments.code);
  }, [toolChoice, isBashCommand]);

  function resetExplanation() {
    setExplanation(null);
  }

  function resetScore() {
    setScore(null);
  }

  function handleCodeChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const newCode = e.target.value;
    setCode(newCode);
    const updatedToolChoice = {
      ...toolChoice,
      arguments: isBashCommand
        ? { ...toolChoice.arguments, cmd: newCode }
        : { ...toolChoice.arguments, code: newCode },
    };
    onToolChoiceChange(updatedToolChoice);
  }

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
            <span className="font-semibold">ID: </span>
            <code>{toolChoice.id}</code>
          </div>
          <div>
            <span className="font-semibold">Function:</span> <code>{toolChoice.function}</code>
          </div>
          <ToolCodeBlock
            isBashCommand={isBashCommand}
            code={code}
            handleCodeChange={handleCodeChange}
            explanation={explanation}
            setExplanation={setExplanation}
            score={score}
            setScore={setScore}
            resetExplanation={resetExplanation}
            resetScore={resetScore}
          />
        </div>
      </CardContent>
    </Card>
  );
}
