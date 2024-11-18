import { useGetProjects, Project, useGetTaskRuns, Run, useGetProject, useGetRunTools, Tool, useGetRunRequestGroups } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./util/page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./util/uuid_display";
import { Button } from "./ui/button";
import { ArrowRightIcon, RailSymbol } from "lucide-react";
import { CreatedAgo } from "./util/created_ago";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableFooter,
} from "@/components/ui/table"
import { ProjectBadge, StatusBadge, TaskBadge, ToolBadge, ToolBadges } from "./util/status_badge";

export default function Runs() {
  const [runs, setRuns] = useState<Run[]>([]);
  const { taskId } = useParams();
  const { selectedProject } = useProject();

  const { data: runsData, isLoading: runsLoading, error: runsError } = useGetTaskRuns(taskId || '');

  useEffect(() => {
    if (runsData?.data) {
      setRuns(runsData.data);
    } else {
      setRuns([]);
    }
  }, [runsData]);

  return (
    <Page title={`Runs`}
      subtitle={<span>{runs.length > 0 ? `${runs.length} runs` : 'No runs'} found for task <TaskBadge taskId={taskId ?? ''} /></span>}
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
                <TableHead className="">Tools Assigned</TableHead>
                <TableHead className="w-[100px] text-right">Tool Executions</TableHead>
                <TableHead className="w-[100px] text-right">Created</TableHead>
                <TableHead className="w-[100px] text-right">Status</TableHead>
                <TableHead className="w-[50px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {runs.map((run) => (
                <TableRow key={run.id}>
                  <TableCell className="font-medium">
                    <UUIDDisplay uuid={run.id} href={`/tasks/${taskId}/runs/${run.id}`} />
                  </TableCell>
                  <TableCell>
                    <ToolsBadgeList runId={run.id} />
                  </TableCell>
                  <TableCell className="text-right">
                    <ExecutionsCount runId={run.id} />
                  </TableCell>
                  <TableCell className="text-right">
                    <CreatedAgo datetime={run.created_at} label='' />
                  </TableCell>
                  <TableCell className="text-right">
                    <StatusBadge status={run.status} />
                  </TableCell>
                  <TableCell>
                    <Link to={`/tasks/${taskId}/runs/${run.id}`}>
                      <Button variant="ghost"><ArrowRightIcon className="h-4 w-4" /></Button>
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
            <TableFooter>
              <TableRow>
                <TableCell className="text-xs text-muted-foreground" colSpan={6}>
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
    <div className="flex flex-row gap-2 min-w-0">
      {toolsLoading && <p>Loading...</p>}
      {toolsError && <p>Error: {toolsError.message}</p>}
      <div className="flex-shrink-0">
        <ToolBadges tools={tools} maxTools={4} />
      </div>
    </div>
  )
}

function ExecutionsCount({ runId }: { runId: string }) {
  const { data: executionsData, isLoading: executionsLoading, error: executionsError } = useGetRunRequestGroups(runId || '');
  return (
    <>
      {executionsLoading && <p>Loading...</p>}
      {executionsError && <p>Error: {executionsError.message}</p>}
      <p>{executionsData?.data?.length || 0}</p>
    </>
  )
}
