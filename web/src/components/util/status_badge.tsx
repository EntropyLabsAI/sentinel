import React from "react";
import { Decision, Status, SupervisionStatus, Supervisor, SupervisorType, useGetSupervisor, useGetTool } from "@/types";
import { Badge } from "@/components/ui/badge";
import { Link } from "react-router-dom";
import { BotIcon, PickaxeIcon } from "lucide-react";
import { UUIDDisplay } from "./uuid_display";
import { useProject } from "@/contexts/project_context";

export function StatusBadge({ status, statuses }: { status?: Status, statuses?: SupervisionStatus[] }) {
  const colors = {
    [Status.pending]: 'bg-gray-400',
    [Status.completed]: 'bg-purple-800',
    [Status.failed]: 'bg-gray-800',
    [Status.assigned]: 'bg-purple-700',
    [Status.timeout]: 'bg-gray-600',
  }

  if (statuses) {
    const mostRecentStatus = statuses.sort((a, b) => a.id - b.id)[statuses.length - 1].status;
    status = mostRecentStatus;
  } else if (!status) {
    return
  }

  return <Badge className={`shadow-none ${colors[status]}`}>{status}</Badge>;
}



export function DecisionBadge({ decision }: { decision: Decision | undefined }) {
  if (!decision) {
    return
  }

  const colors = {
    [Decision.approve]: 'bg-green-600',
    [Decision.modify]: 'bg-green-500',
    [Decision.reject]: 'bg-red-500',
    [Decision.escalate]: 'bg-yellow-500',
    [Decision.terminate]: 'bg-black',
  }

  return <Badge className={`text-center ${colors[decision]} text-white shadow-none whitespace-nowrap`}>{decision}</Badge>;
}

export function SupervisorTypeBadge({ type }: { type: SupervisorType }) {
  const colors = {
    [SupervisorType.client_supervisor]: 'gray',
    [SupervisorType.human_supervisor]: 'gray',
    [SupervisorType.no_supervisor]: 'gray',
  }

  return (
    <Badge className={`text-white bg-${colors[type]}-600`}>
      {type.toString()}
    </Badge>
  );

};

export function RunBadge({ runId }: { runId: string }) {
  const projectId = useProject();
  if (!runId) {
    return
  }
  return <Badge className="text-white bg-gray-400 shadow-none whitespace-nowrap"><Link to={`/projects/${projectId.selectedProject}/runs/${runId}`}>Run {runId.slice(0, 8)}</Link></Badge>;
}

// TODO accept a tool object instead of ID, optionally.
export const ToolBadge: React.FC<{ toolId: string }> = ({ toolId }) => {
  // Load tool name from toolId
  const { data, isLoading, error } = useGetTool(toolId);

  if (isLoading) return <Badge className="text-white bg-gray-400 shadow-none whitespace-nowrap">Loading...</Badge>;
  if (error) return <Badge className="text-white bg-gray-400 shadow-none whitespace-nowrap">Error: {error.message}</Badge>;

  return (
    <Badge key={toolId} className="text-gray-800 shadow-none bg-amber-300 min-w-0 inline-flex">
      <Link to={`/tools/${toolId}`} className="flex flex-row gap-2 items-center overflow-hidden">
        <PickaxeIcon className="w-3 h-3 flex-shrink-0" />
        <span className="truncate overflow-hidden">{data?.data.name}</span>
      </Link>
    </Badge>
  );

};

export const ExecutionStatusBadge: React.FC<{ statuses: SupervisionStatus[] }> = ({ statuses }) => {
  // Sort statuses by ID and return the status of the most recent one
  const mostRecentStatus = statuses.sort((a, b) => a.id - b.id)[statuses.length - 1].status;

  return <StatusBadge status={mostRecentStatus} />
}

type SupervisorBadgeProps = {
  supervisorId: string;
  supervisor?: Supervisor;
}

// TODO stop re-fetching supervisor if already provided
export const SupervisorBadge: React.FC<SupervisorBadgeProps> = ({ supervisorId, supervisor }) => {
  const { data, isLoading, error } = useGetSupervisor(supervisorId);

  if (isLoading) {
    return <Badge key={supervisorId} className="text-white bg-gray-400 shadow-none whitespace-nowrap">Loading...</Badge>;
  }
  if (error) {
    return <Badge key={supervisorId} className="text-white bg-gray-400 shadow-none whitespace-nowrap">Error: {error.message}</Badge>;
  }

  const supervisorData = supervisor || data?.data;
  const baseBadgeClasses = "text-gray-800 shadow-none bg-sky-200 whitespace-nowrap";

  return (
    <Link to={`/supervisors/${supervisorId}`}>
      <Badge key={supervisorId} className={baseBadgeClasses}>
        <div className="flex flex-row gap-2 items-center">
          <BotIcon className="w-3 h-3" />
          {supervisorData?.name}
        </div>
      </Badge>
    </Link>
  );
};
