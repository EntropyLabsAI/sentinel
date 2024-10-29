import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useGetProjects, Project, Execution, useGetRunExecutions, useGetProjectRuns, Run, useGetProject } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./uuid_display";
import { Button } from "./ui/button";
import { ArrowRightIcon } from "lucide-react";
import { CreatedAgo } from "./created_ago";

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
      setSelectedProject(projectId);
    } else if (selectedProject && !projectId) {
      // If we have a selected project but no URL parameter,
      // navigate to the correct URL
      navigate(`/projects/${selectedProject}/runs`);
    }
  }, [projectId, selectedProject]);

  useEffect(() => {
    if (runsData?.data) {
      setRuns(runsData.data);
    } else {
      setRuns([]);
    }
  }, [runsData, selectedProject]);

  if (runsLoading) return <Page title="Runs">Loading...</Page>;
  if (runsError) return <Page title="Runs">Error: {runsError.message}</Page>;

  return (
    <Page title={`Runs for project ${projectData?.data?.name}`} subtitle={<span>{runs.length} runs found for project <UUIDDisplay uuid={projectData?.data?.id} /></span>}>
      <div className="flex flex-col space-y-4">
        {runs.length === 0 && <p className="text-sm text-gray-500">No runs found for this project. When you run an agent, it will appear here.</p>}
        {runs.map((run) => (
          <Card key={run.id}>
            <CardHeader>
              <CardTitle>Run <UUIDDisplay uuid={run.id} /></CardTitle>
              <CardDescription className="flex flex-row justify-between gap-2">
                <CreatedAgo datetime={run.created_at} />
                <Link to={`/projects/${projectId}/runs/${run.id}`} key={run.id}>
                  <Button variant="ghost"><ArrowRightIcon className="" /></Button>
                </Link>
              </CardDescription>
            </CardHeader>
          </Card>
        ))}
      </div>
    </Page>
  )
}
