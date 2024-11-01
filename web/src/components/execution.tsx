import { Execution, ExecutionSupervisions, Tool, useGetExecutionSupervisions, useGetRunExecutions, useGetRunTools } from '@/types';
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Page from './page';
import ContextDisplay from '@/components/context_display'
import { UUIDDisplay } from './uuid_display';
import JsonDisplay from './json_display';


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

  return (
    <Page title="Supervision request & result" subtitle={<span>For execution <UUIDDisplay uuid={executionId} /></span>} >
      <div className="col-span-3">
        {supervision.requests?.map((request) => (
          <div>
            <ContextDisplay context={request.task_state} />
            <JsonDisplay reviewRequest={request} />
          </div>
        ))}
      </div></Page>
  )
}
