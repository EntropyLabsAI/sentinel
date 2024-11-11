import React from "react";
import { ChainExecutionState, SupervisionRequestState, Decision } from "@/types";
import { Badge } from "./ui/badge";
import { Card, CardHeader, CardContent } from "./ui/card";
import { CheckCircle2, XCircle, AlertCircle, Clock, ArrowRight, ClockIcon } from "lucide-react";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";

interface ChainStateDisplayProps {
  chainState: ChainExecutionState;
  currentRequestId?: string;
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

export default function ChainStateDisplay({ chainState, currentRequestId }: ChainStateDisplayProps) {
  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="messages" className="border border-gray-200 rounded-md">
        <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
          <div className="flex flex-row gap-4">
            <ClockIcon className="w-4 h-4" />
            Supervision Chain State
          </div>
        </AccordionTrigger>
        <AccordionContent className="p-4">
          <Card className="w-full border-none">
            <CardContent>
              <div className="flex flex-col space-y-4">
                {/* Chain Info */}
                <div className="text-sm text-gray-500">
                  Chain ID: {chainState.chain.chain_id}
                </div>

                {/* Supervision Timeline */}
                <div className="flex flex-col space-y-2">
                  {chainState.supervision_requests.map((request, index) => (
                    <div
                      key={request.supervision_request.id}
                      className={`flex items-center space-x-2 p-2 rounded-lg ${request.supervision_request.id === currentRequestId ? 'bg-blue-50' : ''
                        }`}
                    >
                      {/* Position indicator */}
                      <span className="text-sm font-medium w-6">
                        {request.supervision_request.position_in_chain}
                      </span>

                      {/* Status icon */}
                      {getStatusIcon(request)}

                      {/* Supervisor info */}
                      <div className="flex-grow">
                        <span className="text-sm font-medium">
                          Supervisor {request.supervision_request.supervisor_id}
                        </span>
                      </div>

                      {/* Status badge */}
                      <Badge className={getStatusColor(request)}>
                        {request.result ? request.result.decision : request.status.status}
                      </Badge>

                      {/* Show connector line if not last item */}
                      {index < chainState.supervision_requests.length - 1 && (
                        <ArrowRight className="w-4 h-4 text-gray-400" />
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </CardContent>
          </Card>
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  );
}
