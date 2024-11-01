import { Decision, Execution, ExecutionSupervisions, SupervisionResult, SupervisionStatus, Tool, useGetExecutionSupervisions, useGetRunExecutions, useGetRunTools } from '@/types';
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Page from './page';
import ContextDisplay from '@/components/context_display'
import { UUIDDisplay } from './uuid_display';
import JsonDisplay from './json_display';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { DecisionBadge, ExecutionStatusBadge, StatusBadge, SupervisorBadge, ToolBadge } from './status_badge';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from './ui/collapsible';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from './ui/accordion';
import { ChevronDownIcon, FileJsonIcon, MessagesSquareIcon } from 'lucide-react';


// TODO allow execution supervision to be passed in
export default function ExecutionComponent() {
  const { executionId } = useParams()

  // const [executions, setExecutions] = useState<Execution[]>([]);
  const [supervision, setSupervisions] = useState<ExecutionSupervisions>();

  const { data, isLoading, error } = useGetExecutionSupervisions(executionId || '');
  // const { data, isLoading, error } = useGetExecution(executionID || '');

  // const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(executionID || '');

  useEffect(() => {
    if (data?.data) {
      setSupervisions(data.data);
    }
  }, [data]);

  if (error) {
    return (<Page title="Supervision request & result">
      {error.message}
    </Page>)
  }

  if (!supervision) {
    return (
      <Page title="Supervision requests & results">
        No supervision found.
      </Page>
    )
  }

  function findResultForRequest(results: SupervisionResult[], request_id: string) {
    var result = results.find(result => result.supervision_request_id === request_id)

    return result
  }

  return (
    <Page title="Supervision requests & results" subtitle={<span>{supervision.requests.length} supervision requests have been made so far for execution <UUIDDisplay uuid={executionId} /> which is currently in status {` `} <ExecutionStatusBadge statuses={supervision.statuses} />
    </span>} >
      <div className="col-span-3 flex flex-col space-y-4">
        <div>
          {supervision.requests && supervision.requests[0].tool_requests && (
            <ToolBadge toolId={supervision.requests[0].tool_requests[0].tool_id} />
          )}
        </div>
        <div>
          {supervision.requests?.map((request, index) => (
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
                      <DecisionBadge decision={findResultForRequest(supervision.results, request.id || '')?.decision} />
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
                        <div className="flex flex-row gap-4">
                          <FileJsonIcon className="w-4 h-4" />
                          Full Task State JSON
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <JsonDisplay reviewRequest={request} />
                      </AccordionContent>
                    </AccordionItem>
                  </Accordion>

                  <SupervisionResultAccordion result={findResultForRequest(supervision.results, request.id || '')} />
                </AccordionContent>
              </AccordionItem>
            </Accordion>

          ))}
        </div>
      </div ></Page >
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
