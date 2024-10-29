import { useGetProjects, Project, Execution, useGetRunExecutions, useGetProjectRuns, Run, useGetProject, useGetSupervisor, Supervisor, SupervisorType } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./uuid_display";
import HumanReviews from "./human_reviews";

export default function SupervisorDetails() {
  const [supervisor, setSupervisor] = useState<Supervisor>();
  const { supervisorId } = useParams();

  // const { selectedProject, setSelectedProject } = useProject();

  const { data: supervisorData, isLoading: supervisorLoading, error: supervisorError } = useGetSupervisor(supervisorId || '');


  useEffect(() => {
    if (supervisorData?.data) {
      setSupervisor(supervisorData.data);
    }
  }, [supervisorData]);

  if (supervisorLoading) return <Page title="Supervisor">Loading...</Page>;
  if (supervisorError) return <Page title="Supervisor">Error: {supervisorError.message}</Page>;

  return (
    <Page title={`Supervisor ${supervisor?.name} reviews`} subtitle={<span>Review for supervisor <UUIDDisplay uuid={supervisor?.id} /> will be displayed here</span>}>
      <div className="flex flex-col space-y-4 col-span-3">
        {supervisor?.type === SupervisorType.human_supervisor && <HumanReviews />}
        {supervisor?.type === SupervisorType.client_supervisor && <div>Client Supervisor</div>}
      </div>
    </Page>
  )
}
