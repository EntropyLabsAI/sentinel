import { Supervisor, Tool } from "@/types";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import React from "react";
import { ToolCard } from "./tool_card";

interface ToolsListProps {
  tools: Tool[] | string[];
  variant?: "card" | "badge";
  // Optionally, a run ID which we will use to get supervisors for this tool for this run
  runId?: string;
}

export function ToolsList({ tools, variant = "card", runId }: ToolsListProps) {
  if (variant === "badge") {
    return (
      <div className="flex gap-2">
        {(tools as string[]).map((toolId) => (
          <Link key={toolId} to={`/tools/${toolId}`}>
            <Badge variant="secondary">{toolId.slice(0, 8)}</Badge>
          </Link>
        ))}
      </div>
    );
  }

  return (
    <>
      {(tools as Tool[]).map((tool) => (
        <ToolCard key={tool.id} tool={tool} runId={runId} />
      ))}
    </>
  );
}
