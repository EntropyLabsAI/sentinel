import React from "react";
import { Decision, Status, SupervisionStatus, SupervisorType, useGetTool } from "@/types";
import { Badge } from "@/components/ui/badge";
import { Link } from "react-router-dom";
import { PickaxeIcon } from "lucide-react";

export function StatusBadge({ status }: { status: Status }) {
  const colors = {
    [Status.pending]: 'bg-gray-400',
    [Status.completed]: 'bg-purple-800',
    [Status.failed]: 'bg-gray-800',
    [Status.assigned]: 'bg-purple-700',
    [Status.timeout]: 'bg-gray-600',
  }
  return <Badge className={`shadow-none ${colors[status]}`}>{status}</Badge>;
}

export function DecisionBadge({ decision }: { decision: Decision }) {
  const colors = {
    [Decision.approve]: 'bg-green-600',
    [Decision.modify]: 'bg-green-500',
    [Decision.reject]: 'bg-red-500',
    [Decision.escalate]: 'bg-yellow-500',
    [Decision.terminate]: 'bg-black',
  }

  return <Badge className={`text-center ${colors[decision]} text-white shadow-none whitespace-nowrap`}>{decision}</Badge>;
}

export const SupervisorTypeBadge: React.FC<{ type: SupervisorType }> = ({ type }) => {
  const label = type === SupervisorType.client_supervisor ? "client-side supervision" : "human supervision";
  const color = type === SupervisorType.client_supervisor ? "blue" : "blue";
  return <Badge className={`text-white bg-${color}-900 shadow-none whitespace-nowrap hover:bg-${color}-700`}>{label}</Badge>;

};

// TODO accept a tool object instead of ID, optionally.
export const ToolBadge: React.FC<{ toolId: string }> = ({ toolId }) => {
  // Load tool name from toolId
  const { data, isLoading, error } = useGetTool(toolId);

  if (isLoading) return <Badge className="text-white bg-gray-400 shadow-none whitespace-nowrap">Loading...</Badge>;
  if (error) return <Badge className="text-white bg-gray-400 shadow-none whitespace-nowrap">Error: {error.message}</Badge>;

  return <Badge className="text-gray-800 shadow-none bg-gray-100 hover:bg-gray-200 whitespace-nowrap"><Link to={`/tools/${toolId}`} className="flex flex-row gap-2 items-center"><PickaxeIcon className="w-3 h-3" />{data?.data.name}</Link></Badge>;

};
