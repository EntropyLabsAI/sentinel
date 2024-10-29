import React, { useEffect, useState } from 'react';
import { UserIcon, BrainCircuitIcon } from 'lucide-react';
import { Badge } from './ui/badge';
import { Link } from 'react-router-dom';
import { Supervisor, useGetProject } from '@/types';
import { useGetSupervisors } from '@/types';
import { useProject } from '@/contexts/project_context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import Page from './page';

interface SupervisorSelectionProps {
  API_BASE_URL: string;
  WEBSOCKET_BASE_URL: string;
}

const SupervisorSelection: React.FC<SupervisorSelectionProps> = ({ API_BASE_URL, WEBSOCKET_BASE_URL }) => {
  const [supervisors, setSupervisors] = useState<Supervisor[]>([]);
  const { selectedProject } = useProject();
  const { data, isLoading, error } = useGetSupervisors(
    { projectId: selectedProject! }
  );

  const { data: projectData } = useGetProject(selectedProject!);

  useEffect(() => {
    if (data?.data) {
      setSupervisors(data.data);
    }
  }, [data]);

  if (!selectedProject) {
    return <div>Please select a project first</div>;
  }

  if (isLoading) return <Page title="Supervisors">Loading supervisors...</Page>;
  if (error) return <Page title="Supervisors">Error loading supervisors: {error.message}</Page>;

  if (!API_BASE_URL || !WEBSOCKET_BASE_URL) {
    return (
      <div className="text-center text-red-500">
        No API or WebSocket base URL set: API is: {API_BASE_URL || 'Not Set'} and WebSocket is: {WEBSOCKET_BASE_URL || 'Not Set'}
      </div>
    );
  }

  return (
    <Page title="Supervisors" subtitle={`${supervisors.length} supervisors registered against runs in ${projectData?.data.name}`}>
      {/* Introductory Section */}
      {/* Supervisor Selection Grid */}
      {supervisors.map((supervisor) => {
        return (
          <div>
            <Link to={`/supervisor/${supervisor.id}`}>
              <Card key={supervisor.id}>
                <CardHeader>
                  <CardTitle>{supervisor.name}</CardTitle>
                  <CardDescription>{supervisor.description}</CardDescription>
                </CardHeader>
                <CardContent>
                  <p>{supervisor.description}</p>
                </CardContent>
              </Card>
            </Link>
          </div>
        )
      })}
      <div className="mb-12 space-y-12 col-span-3">
        <p className="text-lg text-gray-700">
          Supervisors are used to review agent actions. To get started, ensure that your agent is running and making requests to the Sentinel API when it wants to take an action. Requests will be paused until a supervisor approves the action.
        </p>
        <p>Supervisors will then return one of the following responses to your agent:
        </p>
        <div className="grid grid-cols-[auto,1fr] gap-x-4 gap-y-2 mt-2">
          <Badge variant="outline" className="flex flex-col text-center bg-green-500 text-white whitespace-nowrap">APPROVE</Badge>
          <div>The agent can proceed</div>

          <Badge variant="outline" className="flex flex-col text-center bg-blue-500 text-white whitespace-nowrap">MODIFY</Badge>
          <div>The action has been modified and should be approved in its new form</div>

          <Badge variant="outline" className="flex flex-col text-center bg-red-500 text-white whitespace-nowrap">REJECT</Badge>
          <div>The agent action is blocked and the agent should try again</div>

          <Badge variant="outline" className="flex flex-col text-center bg-yellow-500 text-white whitespace-nowrap">ESCALATE</Badge>
          <div>The action should be escalated to the next supervisor if one is configured</div>

          <Badge variant="outline" className="flex flex-col text-center bg-black text-white whitespace-nowrap">TERMINATE</Badge>
          <div>The agent process should be killed</div>
        </div>
      </div>

    </Page>
  );
};

export default SupervisorSelection;
