import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetRunRequestGroups, Tool, ToolRequestGroup, useGetRunTools, RunState, useGetRunState } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./util/page";
import { ToolsList } from "@/components/tools_list";
import { UUIDDisplay } from "@/components/util/uuid_display";
import ExecutionTable from "./execution_table";
import { EyeIcon, ListOrderedIcon, PickaxeIcon, PyramidIcon } from "lucide-react";
import { ToolCard } from "./tool_card";

export default function Run() {
  const { runId } = useParams();
  const [runState, setRunState] = useState<RunState | null>(null);
  const [tools, setTools] = useState<Tool[]>([]);

  const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(runId || '');
  const { data: runStateData, isLoading: runStateLoading } = useGetRunState(runId || '');

  useEffect(() => {
    if (runStateData?.data) {
      setRunState(runStateData.data);
    }
  }, [runStateData]);

  useEffect(() => {
    if (toolsData?.data) {
      setTools(deduplicateTools(toolsData.data));
    }
  }, [toolsData]);

  // TODO do this serverside 
  function deduplicateTools(tools: Tool[]) {
    return tools.filter((tool, index, self) =>
      index === self.findIndex((t) => t.id === tool.id)
    );
  }

  if (!runId) {
    <p>No Run ID found</p>
  }

  return (
    <>
      <Page
        icon={<PyramidIcon className="h-5 w-5" />}
        title={`Executions for run`}
        subtitle={
          <span className="text-sm text-muted-foreground">
            We recorded {runState?.length} execution{runState?.length === 1 ? "" : "s"} for run{' '}
            <UUIDDisplay uuid={runId} />{' '}
            across {tools.length} tool{tools.length === 1 ? "" : "s"}. To see more details for each tool execution, inspect the rows in the table below.
          </span>
        }>
        <div className="col-span-3 space-y-4">
          <ExecutionTable runState={runState || []} />
        </div>
      </Page>
      <Page
        cols={2}
        icon={<PickaxeIcon className="h-5 w-5" />}
        title="Tools used in this run" subtitle={`${tools.length} tool${tools.length === 1 ? "" : "s"} used in this run`}>
        {(tools as Tool[]).map((tool) => (
          <div className="col-span-1">
            <ToolCard key={tool.id} tool={tool} runId={runId} />
          </div>
        ))}
      </Page >
    </>
  );
}
