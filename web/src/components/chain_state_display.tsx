import React from "react";
import { ChainExecutionState, SupervisionRequestState, Decision } from "@/types";
import { Badge } from "./ui/badge";
import { Card, CardHeader, CardContent } from "./ui/card";
import { CheckCircle2, XCircle, AlertCircle, Clock, ArrowRight, ClockIcon, LinkIcon } from "lucide-react";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import { DecisionBadge, StatusBadge, SupervisorBadge } from "./util/status_badge";
import { UUIDDisplay } from "./util/uuid_display";

interface ChainStateDisplayProps {
  chainState: ChainExecutionState;
  currentRequestId?: string;
  index: number;
}

const getStatusIcon = (state: SupervisionRequestState) => {
  if (!state.result) {
    return <Clock className="w-4 h-4 text-yellow-500" />;
  }
  switch (state.result.decision) {
    case Decision.approve:
      return <CheckCircle2 className="w-4 h-4 text-green-500" />;
    case Decision.reject:
      return <XCircle className="w-4 h-4 text-red-500" />;
    default:
      return <AlertCircle className="w-4 h-4 text-yellow-500" />;
  }
};

const getStatusColor = (state: SupervisionRequestState) => {
  if (!state.result) return "bg-yellow-100 text-yellow-800";
  switch (state.result.decision) {
    case Decision.approve:
      return "bg-green-100 text-green-800";
    case Decision.reject:
      return "bg-red-100 text-red-800";
    default:
      return "bg-yellow-100 text-yellow-800";
  }
};

export default function ChainStateDisplay({ chainState, currentRequestId, index }: ChainStateDisplayProps) {
  return (
    <Accordion type="single" collapsible className="w-full" key={index}>
      <AccordionItem value="messages" className="border border-gray-200 rounded-md">
        <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
          <div className="flex flex-row gap-4 items-center">
            <ClockIcon className="w-4 h-4" />
            <span>Chain {index + 1} Execution Details</span>
            <Badge variant="outline" className="ml-2">
              {chainState.supervision_requests ? chainState.supervision_requests.length : 0} requests
            </Badge>
          </div>
        </AccordionTrigger>
        <AccordionContent className="p-4">
          <Card className="w-full border-none">
            <CardContent>
              <div className="flex flex-col space-y-6">
                {/* Chain Metadata */}
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <h3 className="font-semibold mb-2">Chain Information</h3>
                    <div className="space-y-1">
                      <div>Chain ID: <UUIDDisplay uuid={chainState.chain.chain_id} /></div>
                      <div>Execution ID: <UUIDDisplay uuid={chainState.chain_execution.id} /></div>
                      <div>Created: {new Date(chainState.chain_execution.created_at).toLocaleString()}</div>
                      <div>Toolcall ID: <UUIDDisplay uuid={chainState.chain_execution.toolcall_id} /></div>
                    </div>
                  </div>

                  <div>
                    <h3 className="font-semibold mb-2">Supervisor Chain</h3>
                    <div className="space-y-1">
                      {chainState.chain.supervisors.map((supervisor, index) => (
                        <div key={supervisor.id} className="flex items-center gap-2">
                          <span className="text-gray-500">{index + 1}.</span>
                          <div className="relative group">
                            <Badge variant="outline" className="font-mono">
                              {supervisor.name}
                            </Badge>
                            <span className="absolute left-0 -bottom-6 hidden group-hover:block bg-gray-800 text-white text-xs px-2 py-1 rounded-md z-10">
                              {supervisor.type}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                {/* Supervision Timeline */}
                <div>
                  <h3 className="font-semibold mb-4">Supervision Timeline</h3>
                  <div className="flex flex-col space-y-4">
                    {chainState.supervision_requests ? chainState.supervision_requests.map((request, index) => (
                      <Card key={request.supervision_request.id}
                        className={`border ${request.supervision_request.id === currentRequestId ? 'border-blue-500 bg-blue-50' : 'border-gray-200'}`}>
                        <CardContent className="p-4">
                          <div className="flex flex-col space-y-3">
                            {/* Header */}
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-2">
                                <LinkIcon className="w-4 h-4" />
                                <span className="font-medium">Step {request.supervision_request.position_in_chain}</span>
                                {getStatusIcon(request)}
                              </div>
                              <div className="flex gap-2">
                                <StatusBadge status={request.status.status} />
                                {request.result && <DecisionBadge decision={request.result.decision} />}
                              </div>
                            </div>

                            {/* Supervisor Info */}
                            <div className="flex items-center gap-2">
                              <span className="text-sm text-gray-500">Supervisor:</span>
                              <SupervisorBadge supervisorId={request.supervision_request.supervisor_id} />
                            </div>

                            {/* Result Details (if available) */}
                            {request.result && (
                              <div className="bg-gray-50 p-3 rounded-md space-y-2">
                                <div className="text-sm">
                                  <span className="font-medium">Decision:</span> {request.result.decision}
                                </div>
                                {request.result.toolcall_id && (
                                  <div className="text-sm">
                                    <span className="font-medium">Tool Call:</span>
                                    <UUIDDisplay uuid={request.result.toolcall_id} />
                                  </div>
                                )}
                                <div className="text-sm">
                                  <span className="font-medium">Reasoning:</span>
                                  <p className="mt-1 text-gray-600">{request.result.reasoning}</p>
                                </div>
                              </div>
                            )}

                            {/* Status Details */}
                            <div className="text-xs text-gray-500">
                              Created: {new Date(request.status.created_at).toLocaleString()}
                            </div>
                          </div>
                        </CardContent>
                      </Card>
                    )) : <div>No supervision requests found</div>}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
