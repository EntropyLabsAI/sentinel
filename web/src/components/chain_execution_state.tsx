import React from "react"
import { Table, TableHeader, TableRow, TableHead, TableCell, TableBody } from "./ui/table"
import { SupervisorBadge, StatusBadge, DecisionBadge } from "@/components/util/status_badge"
import { CreatedAgo } from "@/components/util/created_ago"
import { UUIDDisplay } from "@/components/util/uuid_display"
import { ChainExecutionState } from "../types"

export default function ChainExecutionStateDisplay({ chain }: { chain: ChainExecutionState }) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Chain #</TableHead>
          <TableHead className="w-[150px]">Supervisor</TableHead>
          <TableHead className="w-[100px]">Status</TableHead>
          <TableHead className="w-[100px]">Decision</TableHead>
          <TableHead className="w-[150px]">Created At</TableHead>
          <TableHead className="w-[200px]">Request ID</TableHead>
          <TableHead className="w-[200px]">Supervisor Desc</TableHead>
          <TableHead>Reasoning</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {chain.chain.supervisors?.map((supervisor, index) => {
          const supervisionRequest = chain.supervision_requests.find(
            req => req.supervision_request.supervisor_id === supervisor.id
          );

          return (
            <TableRow key={supervisor.id}>
              <TableCell>{index + 1}</TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <SupervisorBadge supervisorId={supervisor.id || ''} />
                  <span className="text-xs text-muted-foreground">({supervisor.type})</span>
                </div>
              </TableCell>
              <TableCell>
                {supervisionRequest ? (
                  <StatusBadge status={supervisionRequest.status.status} />
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                {supervisionRequest?.result ? (
                  <DecisionBadge decision={supervisionRequest.result.decision} />
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                {supervisionRequest ? (
                  <CreatedAgo datetime={supervisionRequest.status.created_at} />
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                {supervisionRequest ? (
                  <UUIDDisplay uuid={supervisionRequest.supervision_request.id || ''} />
                ) : (
                  <span className="text-muted-foreground">-</span>
                )}
              </TableCell>
              <TableCell>
                {/* <span className="text-sm">{supervisor.description}</span> */}
              </TableCell>
              <TableCell>
                {supervisionRequest?.result ? (
                  <span className="text-sm">{supervisionRequest.result.reasoning}</span>
                ) : (
                  <span className="text-muted-foreground italic">No result yet</span>
                )}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
