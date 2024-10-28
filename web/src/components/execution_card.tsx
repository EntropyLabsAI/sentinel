import { Execution, ExecutionSupervisions, SupervisionRequest, SupervisionResult, SupervisionStatus, useGetExecutionSupervisions } from "@/types";
import { Badge } from "./ui/badge";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "./ui/card";

function SupervisionPairCard({ request, result }: { request: SupervisionRequest, result?: SupervisionResult }) {
  return (
    <Card className="mt-4">
      <CardHeader>
        <CardTitle className="flex items-center justify-between text-base">
          Supervision Request {request.id}
          <Badge>{result?.decision || 'Pending'}</Badge>
        </CardTitle>
        <CardDescription>
          Supervisor: {request.supervisor_id || 'Unassigned'}
          {request.status && (
            <div>Status: {request.status.status}</div>
          )}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-2">
        <div>
          <div className="font-semibold">Messages:</div>
          {request.messages?.map((msg, idx) => (  // Add optional chaining
            <div key={idx} className="ml-2">
              {msg.role}: {msg.content}
            </div>
          )) || 'No messages'}
        </div>
        {result && (
          <div>
            <div className="font-semibold">Result:</div>
            <div className="ml-2">
              <div>Decision: {result.decision}</div>
              <div>Reasoning: {result.reasoning}</div>
              {result.toolrequest && (
                <div>Tool Request ID: {result.toolrequest.id}</div>
              )}
            </div>
          </div>
        )}
        {(request.tool_requests ?? []).length > 0 && (
          <div>
            <div className="font-semibold">Tool Requests:</div>
            {request.tool_requests?.map((toolRequest, idx) => (
              <div key={idx} className="ml-2">
                Tool {toolRequest.tool_id}: {JSON.stringify(toolRequest.arguments)}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function SupervisionDetails({ executionId }: { executionId: string }) {
  const [supervisions, setSupervisions] = useState<ExecutionSupervisions>();
  const { data, isLoading, error } = useGetExecutionSupervisions(executionId);


  if (isLoading) {
    return <p>Loading supervisions...</p>;
  }

  if (!data) {
    return <p>No supervisions found for this execution.</p>;
  }

  useEffect(() => {
    if (data) {
      setSupervisions(data.data);
    }
  }, [data]);

  if (!supervisions) {
    return <p>Agent hasn't made a requests yet.</p>;
  }

  console.log(supervisions)

  const results: SupervisionResult[] = supervisions.results || [];  // Add default empty array
  const statuses: SupervisionStatus[] = supervisions.statuses;
  const requests: SupervisionRequest[] = supervisions.requests;

  type row = { request: SupervisionRequest, result?: SupervisionResult }
  // Create a data structure with one entry for each request, then add the result if it exists


  console.log('results', results)
  const rows: row[] = requests.map(request => ({
    request,
    result: results?.find(result => result.supervision_request_id === request.id)
  }));

  console.log('rows', rows)

  return (
    <div>
      <div className="flex gap-2 mb-4">
        <Badge variant="outline">
          Requests: {requests.length}
        </Badge>
        <Badge variant="outline">
          Results: {results.length}
        </Badge>
        <Badge variant="outline">
          Statuses: {statuses.length}
        </Badge>
      </div>
      <div className="space-y-4">
        {rows.map(({ request, result }) => (
          <SupervisionPairCard
            key={request.id}
            request={request}
            result={result}
          />
        ))}
      </div>
    </div>
  );
}

export default function ExecutionCard({ execution }: { execution: Execution }) {
  return (
    <Card key={execution.id}>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          Execution {execution.id}
          <Badge>{execution.status}</Badge>
        </CardTitle>
        <CardDescription>
          {execution.created_at}
          <div>Run ID: {execution.run_id?.slice(0, 8)}</div>
          <div>
            <Link to={`/tools/${execution.tool_id}`}>Tool ID: {execution.tool_id?.slice(0, 8)}</Link>
          </div>
        </CardDescription>
      </CardHeader>
      <CardContent>
        <SupervisionDetails executionId={execution.id} />
      </CardContent>
    </Card>
  );
}
