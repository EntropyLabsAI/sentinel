import { Decision, Execution, ExecutionSupervisions, SupervisionRequest, SupervisionResult, SupervisionStatus, Supervisor, useGetSupervisor } from "@/types";
import { Badge } from "./ui/badge";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card";
import { UUIDDisplay } from './util/uuid_display';
import { CreatedAgo } from "./util/created_ago";
import { DecisionBadge } from "./util/status_badge";

export function SupervisionPairCard({ request, result }: { request: SupervisionRequest, result?: SupervisionResult }) {
  const [supervisor, setSupervisor] = useState<Supervisor | null>(null);
  const { data, isLoading, error } = useGetSupervisor(request.supervisor_id ?? '');

  useEffect(() => {
    if (data) {
      setSupervisor(data.data);
    }
  }, [data]);

  if (isLoading) return <div>Loading supervisor...</div>;
  if (error) return <div>Error loading supervisor: {error.message}</div>;

  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center justify-between text-base">
          <span>Supervision Request <UUIDDisplay uuid={request.id} /></span>
          {result && (
            <div><DecisionBadge decision={result.decision} /></div>
          )}
        </CardTitle>
        <CardDescription>
          Supervisor: {supervisor?.name || <UUIDDisplay uuid={request.supervisor_id || ''} />}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <div>
          <div className="font-semibold">Messages</div>
          {(request.messages ?? []).map((msg, idx) => (
            <div key={idx} className="">
              {msg.role}: {msg.content}
            </div>
          ))}
        </div>
        {result && (
          <div>
            <div className="font-semibold">Result</div>
            <div className="">
              <div>Reasoning: {result.reasoning}</div>
              {result.toolrequest && (
                <div>Tool Request ID: <UUIDDisplay uuid={result.toolrequest.id} /></div>
              )}
            </div>
          </div>
        )}
        {(request.tool_requests ?? []).length > 0 && (
          <div>
            <div className="font-semibold">Tool Requests</div>
            {request.tool_requests?.map((toolRequest, idx) => (
              <div key={idx} className="">
                Tool <UUIDDisplay uuid={toolRequest.tool_id} />: {JSON.stringify(toolRequest.arguments)}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}


