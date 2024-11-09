import { useState } from "react"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableFooter,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { ChevronDownIcon, ChevronUpIcon } from "lucide-react"
import React from "react"
import { ToolRequestGroup, Status } from "@/types"
import { UUIDDisplay } from "@/components/uuid_display"
import { CreatedAgo } from "@/components/created_ago"
import { StatusBadge, ToolBadge } from "./status_badge"
import { useProject } from "@/contexts/project_context"
// import { SupervisionResultsForExecution, SupervisionsForExecution } from "./execution_old"

export default function ExecutionTable({ executions }: { executions: ToolRequestGroup[] }) {
  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})
  const { selectedProject } = useProject()

  const toggleRow = (invoice: string) => {
    setExpandedRows((prev) => ({ ...prev, [invoice]: !prev[invoice] }))
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Execution ID</TableHead>
          <TableHead className="w-[20px]">Tool</TableHead>
          <TableHead className="w-[120px]">Status</TableHead>
          <TableHead className="w-[120px] text-right">Created</TableHead>
          <TableHead className="w-[150px]"></TableHead>
        </TableRow>
      </TableHeader>
      {/* <TableBody> */}
      {/* {executions?.map((execution, index) => (
          <>
            <TableRow key={execution.id} className="">
              <TableCell className="font-medium"><UUIDDisplay uuid={execution.id} href={`/projects/${selectedProject}/runs/${execution.requestgroup_id}/executions/${execution.id}`} /></TableCell>
              <TableCell>
                <ToolBadge toolId={execution.tool_id || ''} />
              </TableCell>
              <TableCell><StatusBadge status={execution.status || Status.failed} /></TableCell>
              <TableCell className="text-right"><CreatedAgo datetime={execution.created_at || ''} /></TableCell>
              <TableCell className="cursor-pointer w-[200px] text-right" onClick={() => toggleRow(execution.id)}>
                {expandedRows[execution.id] ? (
                  <span className="flex flex-row gap-4 text-xs text-muted-foreground">
                    Execution summary
                    <ChevronUpIcon className="h-4 w-4" />
                  </span>
                ) : (
                  <span className="flex flex-row gap-4 text-xs text-muted-foreground">
                    Execution summary
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
                    maxHeight: expandedRows[execution.id] ? "" : "0",
                  }}
                >
                  <p className="text-sm text-gray-500 p-4">In this execution, the agent requested to execute the <ToolBadge toolId={execution.tool_id || ''} /> tool. The agent was supervised by the configured supervisors, resulting in these supervision results:</p>
                  <div className="p-4 bg-muted/50">
                    <SupervisionsForExecution executionId={execution.id} />
                  </div>
                </div>
              </TableCell>
            </TableRow>
          </>
        ))
        }
      </TableBody > */}
      <TableFooter>
        <TableRow>
          <TableCell className="text-xs text-muted-foreground" colSpan={5}>{executions.length} tool executions were found for this run</TableCell>
        </TableRow>
      </TableFooter>
    </Table >
  )
}
