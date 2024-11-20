import React from "react";
import { Link, useParams } from "react-router-dom";
import { useGetProjectTasks, Task } from "@/types";
import { ListIcon } from "lucide-react";
import Page from "./util/page";
import LoadingSpinner from "./util/loading";
import { ProjectBadge, TaskBadge } from "./util/status_badge";
import { Info } from 'lucide-react'
import { Card, CardContent } from "@/components/ui/card"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { CreatedAgo } from "./util/created_ago";

export default function Tasks() {
  const { projectId } = useParams();
  const { data, isLoading, error } = useGetProjectTasks(projectId || '');

  return (
    <Page title="Tasks" icon={<ListIcon className="w-6 h-6" />} subtitle={<span>{data?.data?.length && data?.data?.length > 0 ? `${data?.data?.length} task${data?.data?.length === 1 ? "" : "s"}` : 'No tasks'} found for this project</span>}>
      {isLoading && <LoadingSpinner />}
      {error && <div>{error.message}</div>}
      {data?.data?.map((task) => (
        <CompactTask key={task.id} task={task} />
      ))}
    </Page>
  )
}

function CompactTask({ task }: { task: Task }) {
  return (
    <Link to={`/tasks/${task.id}`}>
      <Card className="w-64 shadow-none ">
        <CardContent className="p-4">
          <div className="flex flex-col items-left gap-2">
            <div>
              <TaskBadge taskId={task.id} />
            </div>
            {task.description && (
              <p className="text-xs text-muted-foreground mt-1">{task.description}</p>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            <CreatedAgo datetime={task.created_at} />
          </p>
        </CardContent>
      </Card>
    </Link>
  )
}
