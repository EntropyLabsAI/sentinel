import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useGetProjects, Project, Execution, useGetRunExecutions, useGetProjectRuns, Run, useGetProject } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./page";
import { useProject } from "@/contexts/project_context";

export default function Runs() {
  const [runs, setRuns] = useState<Run[]>([]);
  const { projectId } = useParams();
  const navigate = useNavigate();
  // Sync global state with URL parameter
  const { selectedProject, setSelectedProject } = useProject();

  const { data: runsData, isLoading: runsLoading, error: runsError } = useGetProjectRuns(projectId || '');
  const { data: projectData, isLoading: projectLoading, error: projectError } = useGetProject(projectId || '');

  useEffect(() => {
    if (projectId && projectId !== selectedProject) {
      setSelectedProject(selectedProject);
    }
  }, [selectedProject]);

  useEffect(() => {
    if (runsData?.data) {
      setRuns(runsData.data);
    }
  }, [runsData]);

  if (runsLoading) return <Page title="Runs">Loading...</Page>;
  if (runsError) return <Page title="Runs">Error: {runsError.message}</Page>;

  return (
    <Page title={`Runs for project ${projectData?.data?.name}`} subtitle={`${runs.length} runs found for project ${projectData?.data?.id}`}>
      <div className="flex flex-col space-y-4">
        {runs.length === 0 && <p className="text-sm text-gray-500">No runs found for this project. When you run an agent, it will appear here.</p>}
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
      </div>
    </Page>
  )
}
