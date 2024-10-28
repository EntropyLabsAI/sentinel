import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useGetProjects, Project, Execution, useGetRunExecutions, useGetProjectRuns, Run } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./page";

export default function Runs() {
  const [runs, setRuns] = useState<Run[]>([]);
  const { projectId } = useParams();

  // Add early return if projectId is undefined
  if (!projectId) {
    return (<Page title="Runs"><p>No project selected</p></Page>);
  }

  const { data, isLoading, error } = useGetProjectRuns(projectId);

  useEffect(() => {
    if (data?.data) {
      setRuns(data.data);
    }
  }, [data]);

  if (isLoading) return <Page title="Runs">Loading...</Page>;
  if (error) return <Page title="Runs">Error: {error.message}</Page>;

  return (
    <Page title={`Runs for project ${projectId}`}>
      {runs.length === 0 && <p>No runs found for this project.</p>}
      {runs.map((run) => (
        <Link to={`/projects/${projectId}/runs/${run.id}`} key={run.id}>
          <Card key={run.id}>
            <CardHeader>
              <CardTitle>Run {run.id}</CardTitle>
              <CardDescription>{run.created_at}</CardDescription>
            </CardHeader>
          </Card>
        </Link>
      ))}
    </Page>
  )
}
