import { SupervisorChain, Tool, useGetToolSupervisorChains } from "@/types";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import React, { useEffect, useState } from "react";
import { ArrowRightIcon } from "@radix-ui/react-icons";
import { SupervisorBadge, SupervisorTypeBadge, ToolBadge } from "./util/status_badge";
import { ToolAttributes } from "./tool_attributes";

type ToolCardProps = {
  tool: Tool;
  runId?: string;
};

export function ToolCard({ tool, runId }: ToolCardProps) {
  return (
    <Card className="flex flex-col text-muted-foreground">
      <CardHeader className="py-2">
        <CardTitle className="py-4 flex flex-row gap-4">
          <ToolBadge toolId={tool.id || ''} />
        </CardTitle>
        <CardDescription>
          <p className="text-sm font-semibold">Description</p>
          {/* <textarea className="text-xs bg-muted p-2 rounded w-full resize-none" value={JSON.stringify(tool.description, null, 2)} readOnly rows={10} disabled /> */}
          {tool.description && <p className="text-sm">{tool.description}</p>}
        </CardDescription>
      </CardHeader>
      <CardContent className="    flex flex-col gap-2">
        <p className="text-sm font-semibold">Attributes</p>
        {tool.attributes && <ToolAttributes attributes={tool.attributes || ''} ignoredAttributes={tool.ignored_attributes || []} />}
        {runId && tool.id && <RunToolSupervisors runId={runId} toolId={tool.id} />}
      </CardContent>
    </Card>
  );
}

function RunToolSupervisors({ runId, toolId }: { runId: string, toolId: string }) {
  const [supervisorChain, setSupervisorChain] = useState<SupervisorChain[]>([]);
  const { data, isLoading, error } = useGetToolSupervisorChains(toolId);

  useEffect(() => {
    if (data) {
      setSupervisorChain(data.data);
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
      {supervisorChain.map((chain, index) => (
        <div className="flex flex-row overflow-scroll gap-2 bg-muted p-2 rounded-md items-center" key={index}>
          <p className="text-sm font-semibold text-muted-foreground">Chain {index + 1}</p>
          {chain.supervisors.map((supervisor, index) => (
            <>
              {index > 0 && <ArrowRightIcon className="w-4 h-4" />}
              {supervisor.id && <SupervisorBadge supervisorId={supervisor.id} key={index} />}
            </>
          ))}
        </div>
      ))}
    </div>
  );
}
