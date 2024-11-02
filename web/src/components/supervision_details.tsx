import { ExecutionSupervisions, useGetExecutionSupervisions, SupervisionResult, SupervisionStatus, SupervisionRequest } from "@/types";
import { Badge } from "@/components/ui/badge";
import React, { useState, useEffect } from "react";
import { SupervisionPairCard } from "@/components/supervision_request_result";
import { SupervisionResultsForExecution } from "./execution";

export function SupervisionDetails({ executionId }: { executionId: string }) {
  const [supervisions, setSupervisions] = useState<ExecutionSupervisions>();
  const { data, isLoading, error } = useGetExecutionSupervisions(executionId);

  useEffect(() => {
    if (data) {
      setSupervisions(data.data);
    }
  }, [data]);

  if (isLoading) {
    return <p>Loading supervisions...</p>;
  }

  if (!data) {
    return <p>No supervisions found for this execution.</p>;
  }

  if (!supervisions) {
    return <p>Agent hasn't made a requests yet.</p>;
  }

  const results: SupervisionResult[] = supervisions?.results ?? [];
  const statuses: SupervisionStatus[] = supervisions?.statuses ?? [];
  const requests: SupervisionRequest[] = supervisions?.requests ?? [];

  type row = { request: SupervisionRequest, result?: SupervisionResult }

  const rows: row[] = requests.map(request => ({
    request,
    result: results.find(result => result.supervision_request_id === request.id)
  }));

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
        {/* {rows.map(({ request, result }) => (
          <SupervisionPairCard
            key={request.id}
            request={request}
            result={result}
          />
        ))} */}
        <SupervisionResultsForExecution results={supervisions.results} requests={supervisions.requests} />
      </div>
    </div>
  );
}
