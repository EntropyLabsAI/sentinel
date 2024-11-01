import { Execution, ExecutionSupervisions, Tool, useGetExecutionSupervisions, useGetRunExecutions, useGetRunTools } from '@/types';
import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import Page from './page';


// TODO allow execution supervision to be passed in
export default function ExecutionComponent() {
  const { executionID } = useParams()

  // const [executions, setExecutions] = useState<Execution[]>([]);
  const [supervision, setSupervisions] = useState<ExecutionSupervisions>();

  const { data, isLoading, error } = useGetExecutionSupervisions(executionID || '');
  // const { data, isLoading, error } = useGetExecution(executionID || '');

  // const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(executionID || '');

  useEffect(() => {
    if (data?.data) {
      setSupervisions(data.data);
    }
  }, [data]);

  return (
    <p>{JSON.stringify(supervision)}</p>
  )
}
