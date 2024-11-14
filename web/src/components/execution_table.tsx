import { useEffect, useState } from "react"
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
import { RunExecution, RunState, Status } from "@/types"
import { UUIDDisplay } from "@/components/util/uuid_display"
import { CreatedAgo } from "@/components/util/created_ago"
import { DecisionBadge, StatusBadge, SupervisorBadge, ToolBadge } from "./util/status_badge"
import { useProject } from "@/contexts/project_context"
import React from "react"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from './ui/accordion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card'
import { FileJsonIcon, MessagesSquareIcon, PickaxeIcon, LinkIcon } from 'lucide-react'
import JsonDisplay from './util/json_display'
import { MessagesDisplay } from "./messages"
import { Button } from "./ui/button"
import Slideover from "./supervisor/slideover"
import ChainExecutionState from "./chain_execution_state"
import LoadingSpinner from "./util/loading"

export default function ExecutionTable({ runState }: { runState: RunState }) {
  const [expandedRows, setExpandedRows] = useState<Record<string, boolean>>({})
  const { selectedProject } = useProject()
  const [rows, setRows] = useState<RunExecution[]>(runState)

  useEffect(() => {
    // Sort the rows by the created_at field
    const sortedRows = runState.sort(
      (a, b) => new Date(a.request_group.created_at || '').getTime() - new Date(b.request_group.created_at || '').getTime()
    )
    setRows(sortedRows)
  }, [runState])

  const toggleRow = (groupId: string) => {
    setExpandedRows((prev) => ({ ...prev, [groupId]: !prev[groupId] }))
  }

  const [isSlideoverOpen, setIsSlideoverOpen] = useState(false)

  return (
    <>
      <Slideover isOpen={isSlideoverOpen} setIsOpen={setIsSlideoverOpen} />
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[100px]">Request</TableHead>
            <TableHead className="w-[20px]">Tool</TableHead>
            <TableHead className="w-[120px]">Status</TableHead>
            <TableHead className="w-[120px] text-right">Created</TableHead>
            <TableHead className="w-[150px]"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 && (
            <TableRow>
              <TableCell colSpan={5} className="text-center text-muted-foreground h-[100px]">
                <LoadingSpinner />
              </TableCell>
            </TableRow>
          )}
          {rows?.map((execution, idx) => (
            <>
              <TableRow key={execution.request_group.id} className="">
                <TableCell className="font-medium">
                  <UUIDDisplay
                    uuid={execution.request_group.id || ''}
                    label={`Request ${idx + 1}`}
                  />
                </TableCell>
                <TableCell>
                  <ToolBadge toolId={execution.request_group.tool_requests[0]?.tool_id || ''} />
                </TableCell>
                <TableCell>
                  <StatusBadge status={execution.status} />
                </TableCell>
                <TableCell className="text-right">
                  <CreatedAgo datetime={execution.request_group.created_at || ''} label='' />
                </TableCell>
                <TableCell
                  className="cursor-pointer w-[200px] text-right"
                  onClick={() => toggleRow(execution.request_group.id || '')}
                >
                  {expandedRows[execution.request_group.id || ''] ? (
                    <span className="flex flex-row gap-4 text-xs text-muted-foreground items-center justify-center">
                      <ChevronUpIcon className="h-4 w-4" />
                    </span>
                  ) : (
                    <span className="flex flex-row gap-4 text-xs text-muted-foreground items-center justify-center">
                      <ChevronDownIcon className="h-4 w-4" />
                    </span>
                  )}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell colSpan={5} className="p-0 bg-muted/50">
                  <div
                    className="overflow-hidden transition-[max-height] duration-500 ease-in-out"
                    style={{
                      maxHeight: expandedRows[execution.request_group.id || ''] ? "5000px" : "0",
                    }}
                  >
                    <div className="p-4">
                      <p className="text-sm text-gray-500 mb-4">
                        In this execution, the agent requested to execute the{" "}
                        <ToolBadge toolId={execution.request_group.tool_requests[0]?.tool_id || ''} /> tool.
                        The request was supervised by {execution.chains.length} chain(s):
                      </p>

                      {execution.chains.map((chain, chainIndex) => (
                        <>
                          {/* <ChainExecutionState chain={chain} /> */}
                          <div key={chain.chain.chain_id} className="w-full space-y-4 bg-muted/50 rounded-md px-4 mb-4">
                            <div className="flex flex-row items-center gap-2 py-2">
                              <LinkIcon className="w-4 h-4" />
                              <p className="text-xs text-gray-500">
                                Chain {chainIndex + 1} - Supervisors:{' '}
                                {chain.chain.supervisors?.map((supervisor, idx) => (
                                  <span key={supervisor.id} className="inline-flex items-center gap-1">
                                    {idx > 0 && " → "}
                                    <SupervisorBadge supervisorId={supervisor.id || ''} />
                                  </span>
                                ))}
                              </p>
                            </div>

                            {/* Show all supervisors in the chain, with their requests if they exist */}
                            {chain.chain.supervisors?.map((supervisor, supervisorIndex) => {
                              const supervisionRequest = chain.supervision_requests.find(
                                req => req.supervision_request.supervisor_id === supervisor.id
                              );

                              return (
                                <Accordion
                                  type="single"
                                  collapsible
                                  className="w-full"
                                  key={supervisor.id}
                                >
                                  <AccordionItem
                                    value="supervision-details"
                                    className={`border border-gray-200 rounded-md mb-4 ${!supervisionRequest ? 'opacity-50' : ''}`}
                                  >
                                    <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
                                      <div className="flex flex-row w-full justify-between">
                                        <div className="flex flex-row gap-2">
                                          <span className="text-sm">Supervisor #{supervisorIndex + 1}</span>
                                          <SupervisorBadge supervisorId={supervisor.id || ''} />
                                          {supervisionRequest ? (
                                            <>
                                              is in status
                                              <StatusBadge status={supervisionRequest.status.status} />
                                              {supervisionRequest.result && (
                                                <>
                                                  because supervisor decided to
                                                  <DecisionBadge decision={supervisionRequest.result.decision} />
                                                </>
                                              )}
                                            </>
                                          ) : (
                                            <span className="text-muted-foreground">
                                              (No supervision request yet)
                                            </span>
                                          )}
                                        </div>
                                      </div>
                                    </AccordionTrigger>

                                    <AccordionContent className="p-4 bg-white rounded-md space-y-4">
                                      {supervisionRequest && supervisionRequest.status.status === Status.pending && (
                                        <Button variant="outline" size="sm" onClick={() => setIsSlideoverOpen(true)}>
                                          Review
                                        </Button>
                                      )}
                                      {supervisionRequest ? (
                                        <>
                                          <p className="text-xs text-gray-500">
                                            Supervision info for request{" "}
                                            <UUIDDisplay uuid={supervisionRequest.supervision_request.id || ''} />
                                          </p>

                                          {/* Tool Requests Section */}
                                          <Accordion type="single" collapsible className="w-full">
                                            <AccordionItem value="tool-requests" className="border border-gray-200 rounded-md">
                                              <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
                                                <div className="flex flex-row gap-4 text-center">
                                                  <PickaxeIcon className="w-4 h-4" />
                                                  Tool Requests
                                                </div>
                                              </AccordionTrigger>
                                              <AccordionContent className="p-4">
                                                {execution.request_group.tool_requests.map((tool_request, idx) => (
                                                  <div key={idx} className="px-4">
                                                    <JsonDisplay json={tool_request} />
                                                  </div>
                                                ))}
                                              </AccordionContent>
                                            </AccordionItem>
                                          </Accordion>

                                          <MessagesDisplay messages={execution.request_group.tool_requests[0]?.task_state?.messages || []} />

                                          {/* Supervision Result Card */}
                                          {supervisionRequest.result && (
                                            <Card>
                                              <CardHeader>
                                                <CardTitle>
                                                  Supervision Result:{" "}
                                                  <SupervisorBadge supervisorId={supervisionRequest.supervision_request.supervisor_id} />{" "}
                                                  returned <DecisionBadge decision={supervisionRequest.result.decision} />
                                                </CardTitle>
                                                <CardDescription>
                                                  <CreatedAgo datetime={supervisionRequest.result.created_at} label="Supervision result occurred" />.
                                                  ID is <UUIDDisplay uuid={supervisionRequest.result.id || ''} />
                                                </CardDescription>
                                              </CardHeader>
                                              <CardContent>
                                                <p>Reasoning: {supervisionRequest.result.reasoning || "No reasoning given"}</p>
                                              </CardContent>
                                            </Card>
                                          )}
                                        </>
                                      ) : (
                                        <p className="text-sm text-muted-foreground">
                                          This supervisor hasn't been called for supervision yet.
                                        </p>
                                      )}
                                    </AccordionContent>
                                  </AccordionItem>
                                </Accordion>
                              );
                            })}
                          </div>
                        </>
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
              {runState?.length || 0} request groups were found for this run
            </TableCell>
          </TableRow>
        </TableFooter>
      </Table>
    </>
  )
}
