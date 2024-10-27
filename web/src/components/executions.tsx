import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project, Execution, useGetRunExecutions } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./page";

export default function Executions() {
  const { runId } = useParams();
  const [executions, setExecutions] = useState<Execution[]>([]);

  if (!runId) {
    return <Page title="Executions">No run selected</Page>;
  }

  const { data, isLoading, error } = useGetRunExecutions(runId);

  useEffect(() => {
    if (data?.data) {
      setExecutions(data.data);
    }
  }, [data]);

  if (isLoading) return <Page title="Executions">Loading...</Page>;
  if (error) return <Page title="Executions">Error: {error.message}</Page>;

  return (
    <Page title={`Executions for run ${runId}`}>
      {executions.length === 0 && <div>No executions found for this run.</div>}
      {executions.map((execution) => (
        <Card key={execution.id}>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              Execution {execution.id.slice(0, 8)}
              <Badge>{execution.status}</Badge>
            </CardTitle>
            <CardDescription>
              {execution.created_at}
              <div>Run ID: {execution.run_id?.slice(0, 8)}</div>
              <div>
                <Link to={`/tools/${execution.tool_id}`}>Tool ID: {execution.tool_id?.slice(0, 8)}
                </Link>
              </div>
            </CardDescription>
            <CardContent>
            </CardContent>
          </CardHeader>
        </Card>
      ))}
    </Page>
  )
}
