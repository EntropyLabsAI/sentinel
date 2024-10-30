import React, { useEffect, useState } from 'react';
import { Badge } from './ui/badge';
import { Link } from 'react-router-dom';
import { Supervisor, useGetProject } from '@/types';
import { useGetSupervisors } from '@/types';
import { useProject } from '@/contexts/project_context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import Page from './page';
import { UUIDDisplay } from './uuid_display';
import { CreatedAgo } from './created_ago';
import { Button } from './ui/button';
import { ArrowRightIcon } from 'lucide-react';
import { SupervisorTypeBadge } from './supervisor_type_badge';

const SupervisorSelection: React.FC = () => {
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

  return (
    <Page title="Supervisors" subtitle={`${supervisors.length} supervisors registered against runs in ${projectData?.data.name}`}>
      {supervisors.length === 0 && <div>No supervisors registered. Register a supervisor to get started.</div>}
      {supervisors.length > 0 && <div className="py-12 col-span-3 space-y-4">
        {supervisors.map((supervisor) => {
          return (
            <Card key={supervisor.id}>
              <CardHeader className="">
                <CardTitle className="flex flex-row justify-between gap-2">
                  <p>
                    {supervisor.name || <span>Supervisor <UUIDDisplay uuid={supervisor.id} /></span>}
                  </p>
                  <SupervisorTypeBadge type={supervisor.type} />
                </CardTitle>

                <CardDescription className="">
                  <CreatedAgo datetime={supervisor.created_at} />
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-row justify-between gap-2">
                <p>{supervisor.description}</p>
                <Link to={`/supervisors/${supervisor.id}`} key={supervisor.id}>
                  <Button variant="ghost"><ArrowRightIcon className="" /></Button>
                </Link>
              </CardContent>
            </Card>
          )
        })}
      </div>
      }
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
