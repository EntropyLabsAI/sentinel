import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Decision, Supervisor, useGetProject } from '@/types';
import { useGetSupervisors } from '@/types';
import { useProject } from '@/contexts/project_context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import Page from '@/components/util/page';
import { UUIDDisplay } from '@/components/util/uuid_display';
import { CreatedAgo } from '@/components/util/created_ago';
import { Button } from '@/components/ui/button';
import { ArrowRightIcon, EyeIcon, ScanEyeIcon } from 'lucide-react';
import { SupervisorTypeBadge } from '@/components/util/status_badge';
import { DecisionBadge } from '@/components/util/status_badge';
import LoadingSpinner from '@/components/util/loading';

const SupervisorSelection: React.FC = () => {
  const [supervisors, setSupervisors] = useState<Supervisor[]>([]);
  const { selectedProject } = useProject();

  const { data, isLoading, error } = useGetSupervisors(selectedProject!);

  const { data: projectData } = useGetProject(selectedProject!);

  useEffect(() => {
    if (data?.data) {
      setSupervisors(dedupeSupervisors(data.data));
    }
  }, [data]);

  function dedupeSupervisors(supervisors: Supervisor[]) {
    return supervisors.filter((supervisor, index, self) =>
      index === self.findIndex((t) => t.id === supervisor.id)
    );
  }

  if (!selectedProject) {
    return <div>Please select a project first</div>;
  }
  console.log(supervisors);

  return (
    <Page title="Supervisors" subtitle={`${supervisors.length} supervisors registered against runs in ${projectData?.data.name}`} icon={<ScanEyeIcon className="w-6 h-6" />}>
      {isLoading && (
        <LoadingSpinner />
      )}
      {error && (
        <div>Error loading supervisors: {error.message}</div>
      )}
      {supervisors.length > 0 && <div className="py-12 col-span-3 space-y-4">
        {supervisors.map((supervisor) => {
          return (
            <Card key={supervisor.id}>
              <CardHeader className="">
                <CardTitle className="flex flex-row justify-between gap-2">
                  <div className="flex flex-row gap-2 items-center">
                    <EyeIcon className="w-4 h-4" />
                    <p>
                      {supervisor.name || <span>Supervisor <UUIDDisplay uuid={supervisor.id} /></span>}
                    </p>
                  </div>
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
          <DecisionBadge decision={Decision.approve} />
          <div>The agent can proceed</div>

          <DecisionBadge decision={Decision.modify} />
          <div>The action has been modified and should be approved in its new form</div>

          <DecisionBadge decision={Decision.reject} />
          <div>The agent action is blocked and the agent should try again</div>

          <DecisionBadge decision={Decision.escalate} />
          <div>The action should be escalated to the next supervisor if one is configured</div>

          <DecisionBadge decision={Decision.terminate} />
          <div>The agent process should be killed</div>
        </div>
      </div>

    </Page>
  );
};

export default SupervisorSelection;
