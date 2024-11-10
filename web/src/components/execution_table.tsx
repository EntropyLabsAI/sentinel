import { useState } from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react"
import { RunState, Status } from "@/types"
import { UUIDDisplay } from "@/components/uuid_display"
import { CreatedAgo } from "@/components/created_ago"
import { StatusBadge, ToolBadge } from "./status_badge"
import { useProject } from "@/contexts/project_context"
import React from "react"

export default function ExecutionTable({ executions }: { executions: RunState }) {
  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})
  const { selectedProject } = useProject()

  const toggleRow = (groupId: string) => {
    setExpandedRows((prev) => ({ ...prev, [groupId]: !prev[groupId] }))
  }

  // Helper to determine the overall status of a request group
  const getGroupStatus = (execution: RunState[0]): Status => {
    // If any chain has a failed status, the group is failed
    const hasFailedChain = execution.chains.some(chain =>
      chain.supervision_requests.some(req => req.status.status === Status.failed)
    )
    if (hasFailedChain) return Status.failed

    // If all chains are completed, the group is completed
    const allChainsCompleted = execution.chains.every(chain =>
      chain.supervision_requests.every(req => req.status.status === Status.completed)
    )
    if (allChainsCompleted) return Status.completed

    // If any chain is pending, the group is pending
    const hasPendingChain = execution.chains.some(chain =>
      chain.supervision_requests.some(req => req.status.status === Status.pending)
    )
    if (hasPendingChain) return Status.pending

    // Default to pending if we can't determine status
    return Status.pending
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Request Group ID</TableHead>
          <TableHead className="w-[20px]">Tool</TableHead>
          <TableHead className="w-[120px]">Status</TableHead>
          <TableHead className="w-[120px] text-right">Created</TableHead>
          <TableHead className="w-[150px]"></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {executions?.map((execution) => (
          <>
            <TableRow key={execution.request_group.id} className="">
              <TableCell className="font-medium">
                <UUIDDisplay
                  uuid={execution.request_group.id || ''}
                  href={`/projects/${selectedProject}/runs/${execution.request_group.id}`}
                />
              </TableCell>
              <TableCell>
                <ToolBadge toolId={execution.request_group.tool_requests[0]?.tool_id || ''} />
              </TableCell>
              <TableCell>
                <StatusBadge status={getGroupStatus(execution)} />
              </TableCell>
              <TableCell className="text-right">
                <CreatedAgo datetime={execution.request_group.created_at || ''} />
              </TableCell>
              <TableCell
                className="cursor-pointer w-[200px] text-right"
                onClick={() => toggleRow(execution.request_group.id || '')}
              >
                {expandedRows[execution.request_group.id || ''] ? (
                  <span className="flex flex-row gap-4 text-xs text-muted-foreground">
                    Supervision chains
                    <ChevronUpIcon className="h-4 w-4" />
                  </span>
                ) : (
                  <span className="flex flex-row gap-4 text-xs text-muted-foreground">
                    Supervision chains
                    <ChevronDownIcon className="h-4 w-4" />
                  </span>
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={5} className="p-0 bg-muted/50">
                <div
                  className="overflow-hidden transition-[max-height] duration-300 ease-in-out"
                  style={{
                    maxHeight: expandedRows[execution.request_group.id || ''] ? "none" : "0",
                  }}
                >
                  <div className="p-4">
                    <p className="text-sm text-gray-500 mb-4">
                      In this execution, the agent requested to execute the{" "}
                      <ToolBadge toolId={execution.request_group.tool_requests[0]?.tool_id || ''} /> tool.
                      The request was supervised by {execution.chains.length} chain(s):
                    </p>

                    {execution.chains.map((chain, chainIndex) => (
                      <div key={chain.chain.chain_id} className="mb-4 last:mb-0">
                        <h4 className="text-sm font-medium mb-2">Chain {chainIndex + 1}</h4>
                        <div className="space-y-2">
                          {chain.supervision_requests.map((request) => (
                            <div
                              key={request.supervision_request.id}
                              className="flex items-center justify-between bg-background p-2 rounded-md"
                            >
                              <div className="flex items-center gap-2">
                                <StatusBadge status={request.status.status} />
                                <span className="text-sm">
                                  {request.result ? (
                                    <span className="text-muted-foreground">
                                      Decision: {request.result.decision} - {request.result.reasoning}
                                    </span>
                                  ) : (
                                    <span className="text-muted-foreground">Awaiting decision</span>
                                  )}
                                </span>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </TableCell>
            </TableRow>
          </>
        ))}
      </TableBody>
      <TableFooter>
        <TableRow>
          <TableCell className="text-xs text-muted-foreground" colSpan={5}>
            {executions?.length || 0} request groups were found for this run
          </TableCell>
        </TableRow>
      </TableFooter>
    </Table>
  )
}
