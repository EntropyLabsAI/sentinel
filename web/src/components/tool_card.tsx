import { Supervisor, Tool, useGetRunToolSupervisors } from "@/types";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import React, { useEffect, useState } from "react";
import { ArrowRightIcon } from "@radix-ui/react-icons";
import { UUIDDisplay } from "./uuid_display";
import { SupervisorTypeBadge, ToolBadge } from "./status_badge";

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
        {tool.attributes && <pre className="text-xs bg-muted p-2 rounded">{JSON.stringify(tool.attributes, null, 2)}</pre>}
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
        <Card key={supervisor.id}>
          <CardHeader className="flex flex-col gap-2">
            <CardTitle className="flex flex-row justify-between">
              {supervisor.name || <span>Supervisor <UUIDDisplay uuid={supervisor.id} /></span>}
              <SupervisorTypeBadge type={supervisor.type} />
            </CardTitle>
            <CardDescription className="flex flex-row justify-between gap-2">
              {supervisor.description}
              <Link to={`/supervisors/${supervisor.id}`} key={supervisor.id}>
                <Button variant="ghost"><ArrowRightIcon className="" /></Button>
              </Link>
            </CardDescription>


          </CardHeader>
        </Card>
      ))}
    </div>
  );
}
