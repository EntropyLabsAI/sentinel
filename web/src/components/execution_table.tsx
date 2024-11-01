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
import { Execution, Status } from "@/types"
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card"
import { UUIDDisplay } from "@/components/uuid_display"
import { CreatedAgo } from "@/components/created_ago"
import { Link } from "react-router-dom"
import { Badge } from "@/components/ui/badge"
import { SupervisionDetails } from "@/components/supervision_details"
import { StatusBadge, ToolBadge } from "./status_badge"
import { useProject } from "@/contexts/project_context"

export default function ExecutionTable({ executions }: { executions: Execution[] }) {
  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})
  const projectId = useProject()

  const toggleRow = (invoice: string) => {
    setExpandedRows((prev) => ({ ...prev, [invoice]: !prev[invoice] }))
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">Execution ID</TableHead>
          <TableHead>Tool</TableHead>
          <TableHead>Status</TableHead>
          <TableHead className="text-right">Created</TableHead>
          <TableHead className="w-[50px]"></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {executions?.map((execution) => (
          <>
            <TableRow key={execution.id} className="">
              <TableCell className="font-medium"><UUIDDisplay uuid={execution.id} href={`/projects/${projectId}/runs/${execution.run_id}/executions/${execution.id}`} /></TableCell>
              <TableCell>
                <ToolBadge toolId={execution.tool_id || ''} />
              </TableCell>
              <TableCell><StatusBadge status={execution.status || Status.failed} /></TableCell>
              <TableCell className="text-right"><CreatedAgo datetime={execution.created_at || ''} /></TableCell>
              <TableCell className="cursor-pointer w-[100px]" onClick={() => toggleRow(execution.id)}>
                {expandedRows[execution.id] ? (
                  <ChevronUpIcon className="h-4 w-4" />
                ) : (
                  <ChevronDownIcon className="h-4 w-4" />
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell colSpan={5} className="p-0">
                <div
                  className="overflow-hidden transition-[max-height] duration-300 ease-in-out"
                  style={{ maxHeight: expandedRows[execution.id] ? "500px" : "0" }}
                >
                  <div className="p-4 bg-muted/50">
                    <SupervisionDetails executionId={execution.id} />
                  </div>
                </div>
              </TableCell>
            </TableRow>
          </>
        ))
        }
      </TableBody >
      <TableFooter>
        <TableRow>
          <TableCell className="text-xs text-muted-foreground" colSpan={5}>{executions.length} tool executions were found for this run</TableCell>
        </TableRow>
      </TableFooter>
    </Table >
  )
}

export function ExecutionCard({ execution }: { execution: Execution }) {
  const projectId = useProject()
  return (
    <Card key={execution.id} className="w-full ">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          Execution <UUIDDisplay uuid={execution.id} href={`/projects/${projectId}/runs/${execution.run_id}/executions/${execution.id}`} />
          <Badge>{execution.status}</Badge>
        </CardTitle>
        <CardDescription>
          <CreatedAgo datetime={execution.created_at || ''} />
          <div>Run ID: <UUIDDisplay uuid={execution.run_id || ''} /></div>
          <div>
            <Link to={`/tools/${execution.tool_id}`}>
              Tool ID: <UUIDDisplay uuid={execution.tool_id || ''} />
            </Link>
          </div>
        </CardDescription>
      </CardHeader >
      <CardContent>
        <SupervisionDetails executionId={execution.id} />
      </CardContent>
    </Card >
  );
}
