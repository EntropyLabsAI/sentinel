import { Message, Tool, ToolChoice, ToolRequest, ToolRequestArguments, useGetTool } from "@/types";
import { Code, Code2, Link, X, MessageSquare } from "lucide-react"
import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import CopyButton from "./copy_button"
import { Button } from "./ui/button";
import ToolCodeBlock from "./tool_code_block";
import { MessageDisplay } from "./messages";
import { UUIDDisplay } from "./uuid_display";
import { ToolBadge } from "./status_badge";

interface ToolChoiceDisplayProps {
  toolChoice: ToolRequest;
  lastMessage: Message;
  onToolChoiceChange: (updatedToolChoice: ToolRequest) => void;
  isSelected: boolean;
  onSelect: () => void;
  index: number;
}

const ToolChoiceDisplay: React.FC<ToolChoiceDisplayProps> = ({
  toolChoice,
  lastMessage,
  onToolChoiceChange,
  isSelected,
  onSelect,
  index,
}) => {
  const [explanation, setExplanation] = useState<string | null>(null);
  const [score, setScore] = useState<string | null>(null);
  const [showMessage, setShowMessage] = useState(false);
  const [tool, setTool] = useState<Tool>();

  const [args, setArgs] = useState<ToolRequestArguments>(toolChoice.arguments);

  const [hiddenArgs, setHiddenArgs] = useState<Partial<ToolRequestArguments>>({});

  const getVisibleArgs = (fullArgs: ToolRequestArguments): ToolRequestArguments => {
    if (!tool?.ignored_attributes) return fullArgs;

    const visibleArgs = { ...fullArgs };
    tool.ignored_attributes.forEach(key => {
      delete visibleArgs[key];
    });
    return visibleArgs;
  };

  const toolQuery = useGetTool(toolChoice.tool_id);

  function resetExplanation() {
    setExplanation(null);
  }

  function resetScore() {
    setScore(null);
  }

  useEffect(() => {
    if (tool?.ignored_attributes) {
      const hidden = tool.ignored_attributes.reduce((acc, key) => {
        if (key in toolChoice.arguments) {
          acc[key] = toolChoice.arguments[key];
        }
        return acc;
      }, {} as Partial<ToolRequestArguments>);

      setHiddenArgs(hidden);
      setArgs(getVisibleArgs(toolChoice.arguments));
    }
  }, [tool, toolChoice.arguments]);

  function handleCodeChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const newArgs = e.target.value;
    const newVisibleArgs: ToolRequestArguments = JSON.parse(newArgs);

    setArgs(newVisibleArgs);

    const updatedToolChoice = {
      ...toolChoice,
      arguments: {
        ...newVisibleArgs,
        ...hiddenArgs, // Merge back hidden values
      },
    };

    onToolChoiceChange(updatedToolChoice);
  }

  useEffect(() => {
    if (toolQuery.data) {
      setTool(toolQuery.data.data);
    }
  }, [toolQuery.data]);

  return (
    <Card className={isSelected ? "border-2 border-blue-500" : ""}>
      <CardHeader className="py-2">
        <CardTitle className="flex justify-between items-center">
          <div className="flex items-center space-x-2">
            <Code2 className="mr-2" />
            <p>Tool Call</p>
            <p className="text-xs text-gray-500">option {index}</p>
          </div>
          <div className="flex items-center">
            <Button
              size="icon"
              className={`outline-none bg-transparent shadow-none text-gray-600 hover:text-gray-400 hover:bg-transparent`}
              onClick={() => setShowMessage(!showMessage)}
            >
              <MessageSquare size={16} />
            </Button>
            <span className="text-sm text-gray-500">
              Tool Choice ID <UUIDDisplay className="" uuid={toolChoice.id} />
            </span>
          </div>
          <div className="flex items-center">
            <span className="font-semibold mr-2"></span>
            <ToolBadge toolId={toolChoice.tool_id} />
            <Button
              size="sm"
              variant={isSelected ? "outline" : "outline"}
              onClick={onSelect}
              disabled={isSelected}
              className="ml-4 bg-blue-500 hover:bg-blue-600 hover:text-white text-white"
            >
              {isSelected ? "Selected" : "Select"}
            </Button>
          </div>
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <ToolCodeBlock
            code={JSON.stringify(getVisibleArgs(args), null, 2)}
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
