import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project, Execution, useGetRunExecutions, useGetRunTools, Tool } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./page";
import { ToolsList } from "@/components/tools_list";
import ExecutionCard from "./execution_card";

export default function Run() {
  const { runId } = useParams();
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [tools, setTools] = useState<Tool[]>([]);

  if (!runId) {
    return <Page title="Executions">No run selected</Page>;
  }

  const { data, isLoading, error } = useGetRunExecutions(runId);
  const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(runId);

  useEffect(() => {
    if (data?.data) {
      setExecutions(data.data);
    }
  }, [data]);

  useEffect(() => {
    if (toolsData?.data) {
      setTools(toolsData.data);
    }
  }, [toolsData]);

  if (isLoading || toolsLoading) return <Page title="Executions">Loading...</Page>;
  if (error) return <Page title="Executions">Error: {error.message}</Page>;

  return (
    <>
      <Page title={`Run ${runId}`}>
        <div className="mb-4">
          There {executions.length === 1 ? "is" : "are"} {executions.length} execution{executions.length === 1 ? "" : "s"} for this run.
        </div>
      </Page>

      <Page title={`Executions for run`}>
        <div className="mb-4">
          {executions.length === 0 && <div>No executions found for this run.</div>}
          {executions.map((execution) => (
            <>
              <ExecutionCard key={execution.id} execution={execution} />
            </>
          ))}
        </div>
      </Page>
      <Page title="Tools used in this run">
        <ToolsList tools={tools} runId={runId} variant="card" />
      </Page>
    </>
  );
}
