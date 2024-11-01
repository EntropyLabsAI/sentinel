import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project, Execution, useGetRunExecutions, useGetRunTools, Tool } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./page";
import { ToolsList } from "@/components/tools_list";
import { UUIDDisplay } from "@/components/uuid_display";
import ExecutionTable from "./execution_table";

export default function Run() {
  const { runId } = useParams();
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [tools, setTools] = useState<Tool[]>([]);

  const { data, isLoading, error } = useGetRunExecutions(runId || '');
  const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(runId || '');

  useEffect(() => {
    if (data?.data) {
      // Sort executions by created_at
      data.data.sort((a, b) => new Date(a.created_at || '').getTime() - new Date(b.created_at || '').getTime());
      setExecutions(data.data);
    }
  }, [data]);

  useEffect(() => {
    if (toolsData?.data) {
      setTools(toolsData.data);
    }
  }, [toolsData]);

  if (!runId) {
    return <Page title="Executions">No run selected</Page>;
  }

  if (isLoading || toolsLoading) return <Page title="Executions">Loading...</Page>;
  if (error) return <Page title="Executions">Error: {error.message}</Page>;

  return (
    <>
      <Page
        title="Run details"
        subtitle={
          <span>
            We recorded {executions.length} execution{executions.length === 1 ? "" : "s"} for run{' '}
            <UUIDDisplay uuid={runId} />{' '}
            across {tools.length} tool{tools.length === 1 ? "" : "s"}. To see more details for each tool execution, inspect the rows in the table below.
          </span>
        }
      >
        <div className="mb-4"></div>
      </Page>

      <Page title={`Executions for run`} subtitle={`During this agent run, we recorded ${executions.length} execution${executions.length === 1 ? "" : "s"} of ${tools.length} tool${tools.length === 1 ? "" : "s"}. `}>
        <div className="col-span-3">
          <ExecutionTable executions={executions} />
        </div>
      </Page>
      <Page title="Tools used in this run" subtitle={`${tools.length} tool${tools.length === 1 ? "" : "s"} used in this run`}>
        <ToolsList tools={tools} runId={runId} variant="card" />
      </Page>
    </>
  );
}

{/* <div className="mb-4 flex flex-col col-span-3 gap-4">
          {executions.map((execution) => (
            <div className="flex flex-row">
              <ExecutionCard key={execution.id} execution={execution} />
            </div>
          ))}
        </div> */}
