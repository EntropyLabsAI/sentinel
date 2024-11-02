import { Supervisor, Tool, useGetRunToolSupervisors } from "@/types";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import React, { useEffect, useState } from "react";
import { ArrowRightIcon } from "@radix-ui/react-icons";
import { UUIDDisplay } from "./uuid_display";
import { SupervisorBadge, SupervisorTypeBadge, ToolBadge } from "./status_badge";
import { ToolAttributes } from "./tool_attributes";

type ToolCardProps = {
  tool: Tool;
  runId?: string;
};

export function ToolCard({ tool, runId }: ToolCardProps) {
  return (
    <Card className="flex flex-col ">
      <CardHeader className="py-2">
        <CardTitle className="py-4 flex flex-row gap-4">
          <ToolBadge toolId={tool.id || ''} />
          <p>
            {tool.description}
          </p>
        </CardTitle>
        <CardDescription>
        </CardDescription>
      </CardHeader>
      <CardContent className="    flex flex-col gap-2">
        {tool.created_at}
        {tool.attributes && <ToolAttributes attributes={tool.attributes || ''} ignoredAttributes={tool.ignored_attributes || []} />}
        {runId && tool.id && <RunToolSupervisors runId={runId} toolId={tool.id} />}
      </CardContent>
    </Card>
  );
}

function RunToolSupervisors({ runId, toolId }: { runId: string, toolId: string }) {
  const [supervisors, setSupervisors] = useState<Supervisor[]>([]);
  const { data, isLoading, error } = useGetRunToolSupervisors(runId, toolId);

  useEffect(() => {
    if (data) {
      setSupervisors(data.data);
    }
  }, [data]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm font-semibold">Supervisors configured for this tool</p>
      {supervisors.map((supervisor) => (
        supervisor.id && <SupervisorBadge supervisorId={supervisor.id} />
      ))}
    </div>
  );
}
