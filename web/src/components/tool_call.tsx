import { Message, ToolChoice } from "@/review";
import { Code, Code2, Link, X, MessageSquare } from "lucide-react"
import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import CopyButton from "./copy_button"
import ExplainButton from "./ask_lm"
import { Button } from "./ui/button";
import { Textarea } from "./ui/textarea"; // Make sure to import Textarea
import ToolCodeBlock from "./tool_code_block";
import { MessageDisplay, MessagesDisplay } from "./messages";

interface ToolChoiceDisplayProps {
  toolChoice: ToolChoice;
  lastMessage: Message;
  onToolChoiceChange: (updatedToolChoice: ToolChoice) => void;
  isSelected: boolean; // Added isSelected prop
  onSelect: () => void; // Added onSelect prop
}

const ToolChoiceDisplay: React.FC<ToolChoiceDisplayProps> = ({
  toolChoice,
  lastMessage,
  onToolChoiceChange,
  isSelected,
  onSelect,
}) => {
  const isBashCommand = toolChoice.function === "bash";
  const [code, setCode] = useState(
    isBashCommand ? toolChoice.arguments.cmd : toolChoice.arguments.code
  );
  const [explanation, setExplanation] = useState<string | null>(null);
  const [score, setScore] = useState<string | null>(null);
  const [showMessage, setShowMessage] = useState(false);

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
    <Card className={isSelected ? "border-2 border-blue-500" : ""}>
      <CardHeader className="py-2">
        <CardTitle className="flex justify-between items-center">
          <div className="flex items-center">
            <Code2 className="mr-2" />
            Tool Call
          </div>
          <div className="flex items-center">
            <Button
              size="icon"
              className={`outline-none bg-transparent shadow-none text-gray-600 hover:text-gray-400 hover:bg-transparent`}
              onClick={() => setShowMessage(!showMessage)}
            >
              <MessageSquare size={16} />
            </Button>
            <span className="font-semibold"></span>
            <code className="">{toolChoice.id}</code>
            <CopyButton className="bg-transparent shadow-none text-gray-600 hover:text-gray-400 hover:bg-transparent outline-none" text={toolChoice.id} />
          </div>
          <div className="flex items-center">
            <span className="font-semibold mr-2"></span>
            <code>{toolChoice.function}</code>
            <Button
              size="sm"
              variant={isSelected ? "default" : "outline"}
              onClick={onSelect}
              disabled={isSelected}
              className="ml-4"
            >
              {isSelected ? "Selected" : "Select"}
            </Button>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
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
          {showMessage && (
            <MessageDisplay message={lastMessage} index={0} />
          )}
        </div>
      </CardContent>
    </Card>
  );
};

export default ToolChoiceDisplay;
