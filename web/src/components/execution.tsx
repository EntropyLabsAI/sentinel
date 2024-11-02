import { Decision, Execution, ExecutionSupervisions, SupervisionRequest, SupervisionResult, SupervisionStatus, Tool, useGetExecutionSupervisions, useGetRunExecutions, useGetRunTools } from '@/types';
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Page from './page';
import ContextDisplay from '@/components/context_display'
import { UUIDDisplay } from './uuid_display';
import JsonDisplay from './json_display';
import { DecisionBadge, ExecutionStatusBadge, StatusBadge, SupervisorBadge, ToolBadge } from './status_badge';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from './ui/accordion';
import { FileJsonIcon, GitPullRequestIcon, MessagesSquareIcon, PinIcon } from 'lucide-react';


// TODO allow execution supervision to be passed in
export default function ExecutionComponent() {
  const { executionId } = useParams()

  // const [executions, setExecutions] = useState<Execution[]>([]);
  const [supervision, setSupervisions] = useState<ExecutionSupervisions>();

  const { data, isLoading, isError } = useGetExecutionSupervisions(executionId || '');
  // const { data, isLoading, error } = useGetExecution(executionID || '');

  // const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(executionID || '');

  useEffect(() => {
    if (data?.data) {
      setSupervisions(data.data);
    }
  }, [data]);

  if (!supervision) {
    return (
      <p>No supervision found</p>
    )
  }

  return (
    <Page
      title="Supervision requests & results"
      subtitle={
        <span>
          {supervision.requests.length} supervision requests have been made so far for execution <UUIDDisplay uuid={executionId} /> which is currently in status {` `}
          <ExecutionStatusBadge statuses={supervision.statuses} />
        </span>
      }
      icon={<GitPullRequestIcon className="w-6 h-6" />}
    >
      {isLoading && (
        <p>Loading</p>
      )}
      {isError && (
        <p>Error</p>
      )}
      <div className="col-span-3 flex flex-col space-y-4">
        <div>
          {supervision.requests && supervision.requests[0].tool_requests && (
            <ToolBadge toolId={supervision.requests[0].tool_requests[0].tool_id} />
          )}
        </div>
        <div>
          <SupervisionResultsForExecution requests={supervision.requests} results={supervision.results} />

        </div>
      </div >
    </Page >
  )
}

export function SupervisionResultsForExecution({ requests, results }: { requests: SupervisionRequest[], results: SupervisionResult[] }) {
  // It's dumb that we don't have a Supervision parent type with a SupervisionRequest and SupervisionResult attribute,
  // as then we wouldn't have to search through arrays looking for the matching result for the request
  function findResultForRequest(results: SupervisionResult[], request_id: string) {
    if (!requests || !results) {
      return undefined
    }
    var result = results.find(result => result.supervision_request_id === request_id)

    if (!result) {
      return undefined
    }

    return result
  }

  return (
    <div>
      {
        requests?.map((request, index) => (
          <Accordion type="single" collapsible className="w-full">
            <AccordionItem value="hub-stats" className="border border-gray-200 rounded-md mb-4">
              <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
                <div className="flex flex-row w-full justify-between">
                  <div className="flex flex-row gap-2">

                    <span className="text-sm text-gray-500">Supervision Request #{index + 1} to supervisor</span>
                    <SupervisorBadge supervisorId={request.supervisor_id || ''} />
                    is in status
                    {request.status?.status && (
                      <StatusBadge status={request.status?.status} />
                    )}
                    because supervisor decided to
                    <DecisionBadge decision={findResultForRequest(results, request.id || '')?.decision} />
                  </div>


                  <span>
                  </span>

                </div>
              </AccordionTrigger>
              <AccordionContent className="p-4 bg-white rounded-md space-y-4">
                <p className="text-xs text-gray-500">

                  Supervision info for request <UUIDDisplay uuid={request.id} /> as part of execution <UUIDDisplay uuid={request.execution_id} />
                </p>
                <Accordion type="single" collapsible className="w-full">
                  <AccordionItem value="hub-stats" className="border border-gray-200 rounded-md mb-4">
                    <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
                      <div className="flex flex-row gap-4">
                        <MessagesSquareIcon className="w-4 h-4" />Messages
                      </div>
                    </AccordionTrigger>
                    <AccordionContent>
                      <ContextDisplay context={request.task_state} />
                    </AccordionContent>
                  </AccordionItem>
                </Accordion>

                <Accordion type="single" collapsible className="w-full">
                  <AccordionItem value="hub-stats" className="border border-gray-200 rounded-md mb-4">
                    <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
                      <div className="flex flex-row gap-4 text-center">
                        <FileJsonIcon className="w-4 h-4" />
                        Full Task State JSON
                      </div>
                    </AccordionTrigger>
                    <AccordionContent>
                      <JsonDisplay reviewRequest={request} />
                    </AccordionContent>
                  </AccordionItem>
                </Accordion>

                <SupervisionResultAccordion result={findResultForRequest(results, request.id || '')} />
              </AccordionContent>
            </AccordionItem>
          </Accordion>

        ))
      }
    </div>
  )
}

function SupervisionResultAccordion({ result }: { result: SupervisionResult | undefined }) {
  if (!result) {
    return <p>No result has yet been recorded for this request</p>
  }

  return (
    <div>
      {result.reasoning != "" ? result.reasoning : "No reasoning given"}

    </div>

  )
}
