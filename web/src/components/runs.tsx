import { useGetProjects, Project, Execution, useGetRunExecutions, useGetProjectRuns, Run, useGetProject, useGetRunTools, Tool } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./uuid_display";
import { Button } from "./ui/button";
import { ArrowRightIcon, RailSymbol } from "lucide-react";
import { CreatedAgo } from "./created_ago";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableFooter,
} from "@/components/ui/table"
import { ToolBadge } from "./status_badge";

export default function Runs() {
  const [runs, setRuns] = useState<Run[]>([]);
  const { projectId } = useParams();
  const navigate = useNavigate();
  // Sync global state with URL parameter
  const { selectedProject, setSelectedProject } = useProject();
  const [executionsCount, setExecutionsCount] = useState(0);

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

  return (
    <Page title={`Agent Runs`}
      subtitle={<span>{runs.length} runs found for project {projectData?.data?.name} (<UUIDDisplay uuid={projectData?.data?.id} />)</span>}
      icon={<RailSymbol className="w-6 h-6" />}
    >
      {runs.length === 0 &&
        <p className="text-sm text-gray-500">No runs found for this project. When you run an agent, it will appear here.</p>
      }
      {runs.length > 0 && (
        <div className="col-span-3">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[100px]">Run ID</TableHead>
                <TableHead className="">Tools Used</TableHead>
                <TableHead className="w-[100px] text-right">Tool Executions</TableHead>
                <TableHead className="w-[100px] text-right">Created</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {runs.map((run) => (
                <TableRow key={run.id}>
                  <TableCell className="font-medium">
                    <UUIDDisplay uuid={run.id} href={`/projects/${projectId}/runs/${run.id}`} />
                  </TableCell>
                  <TableCell>
                    <ToolsBadgeList runId={run.id} />
                  </TableCell>
                  <TableCell className="text-right">
                    <ExecutionsCount runId={run.id} />
                  </TableCell>
                  <TableCell className="text-right">
                    <CreatedAgo datetime={run.created_at} />
                  </TableCell>
                  <TableCell>
                    <Link to={`/projects/${projectId}/runs/${run.id}`}>
                      <Button variant="ghost"><ArrowRightIcon className="h-4 w-4" /></Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell className="text-xs text-muted-foreground" colSpan={5}>
                  {runs.length} runs found for this project
                </TableCell>
              </TableRow>
            </TableFooter>
          </Table>
        </div>
      )
      }
    </Page >
  )
}

function ToolsBadgeList({ runId }: { runId: string }) {
  const { data: toolsData, isLoading: toolsLoading, error: toolsError } = useGetRunTools(runId || '');
  const [tools, setTools] = useState<Tool[]>([]);

  function deduplicateTools(tools: Tool[]) {
    return tools.filter((tool, index, self) =>
      index === self.findIndex((t) => t.id === tool.id)
    );
  }

  useEffect(() => {
    if (toolsData?.data) {
      setTools(deduplicateTools(toolsData.data));
    } else {
      setTools([]);
    }
  }, [toolsData]);

  return (
    <div className="flex flex-row flex-wrap gap-2">

      {toolsLoading && <p>Loading...</p>}
      {toolsError && <p>Error: {toolsError.message}</p>}
      {tools.map((tool) => (
        <ToolBadge toolId={tool.id || ''} />
      ))}
    </div>
  )
}

function ExecutionsCount({ runId }: { runId: string }) {
  const { data: executionsData, isLoading: executionsLoading, error: executionsError } = useGetRunExecutions(runId || '');
  return (
    <>
      {executionsLoading && <p>Loading...</p>}
      {executionsError && <p>Error: {executionsError.message}</p>}
      <p>{executionsData?.data?.length || 0}</p>
    </>
  )
}
