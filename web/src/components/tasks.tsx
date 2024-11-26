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
      <div className="flex flex-col gap-4">

        {isLoading && <LoadingSpinner />}
        {error && <div>{error.message}</div>}
        {data?.data?.map((task) => (
          <CompactTask key={task.id} task={task} />
        ))}
      </div>
    </Page>
  )
}

function CompactTask({ task }: { task: Task }) {
  return (
    <Link to={`/tasks/${task.id}`} className="">
      <Card className="w-full hover:bg-accent/50 transition-colors">
        <CardContent className="p-4">
          <div className="flex justify-between items-start gap-4">
            <div className="flex-1">
              <div className="flex items-center gap-2">
                <TaskBadge taskId={task.id} />
                {/* <h3 className="font-medium text-sm truncate">{task.name}</h3> */}
              </div>
              {task.description && (
                <p className="text-xs text-muted-foreground line-clamp-2">
                  {task.description}
                </p>
              )}
            </div>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <CreatedAgo
                    datetime={task.created_at}
                    className="text-xs text-muted-foreground whitespace-nowrap"
                  />
                </TooltipTrigger>
                <TooltipContent>
                  Created: {new Date(task.created_at).toLocaleString()}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </CardContent>
      </Card>
    </Link>
  )
}
