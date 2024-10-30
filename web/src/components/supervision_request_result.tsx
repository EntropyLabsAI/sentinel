import { Execution, ExecutionSupervisions, SupervisionRequest, SupervisionResult, SupervisionStatus, useGetExecutionSupervisions } from "@/types";
import { Badge } from "./ui/badge";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card";
import { UUIDDisplay } from './uuid_display';
import { CreatedAgo } from "./created_ago";

export function SupervisionPairCard({ request, result }: { request: SupervisionRequest, result?: SupervisionResult }) {
  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center justify-between text-base">
          Supervision Request <UUIDDisplay uuid={request.id} />
          <Badge>{result?.decision || 'Pending'}</Badge>
        </CardTitle>
        <CardDescription>
          Supervisor: <UUIDDisplay uuid={request.supervisor_id || ''} />
          {request.status && (
            <div>Status: {request.status.status}</div>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <div>
          <div className="font-semibold">Messages:</div>
          {(request.messages ?? []).map((msg, idx) => (
            <div key={idx} className="ml-2">
              {msg.role}: {msg.content}
            </div>
          ))}
        </div>
        {result && (
          <div>
            <div className="font-semibold">Result:</div>
            <div className="ml-2">
              <div>Decision: {result.decision}</div>
              <div>Reasoning: {result.reasoning}</div>
              {result.toolrequest && (
                <div>Tool Request ID: <UUIDDisplay uuid={result.toolrequest.id} /></div>
              )}
            </div>
          </div>
        )}
        {(request.tool_requests ?? []).length > 0 && (
          <div>
            <div className="font-semibold">Tool Requests:</div>
            {request.tool_requests?.map((toolRequest, idx) => (
              <div key={idx} className="ml-2">
                Tool <UUIDDisplay uuid={toolRequest.tool_id} />: {JSON.stringify(toolRequest.arguments)}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}


