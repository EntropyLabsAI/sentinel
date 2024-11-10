import { useGetProjects, Project, useGetProjectRuns, Run, useGetProject, useGetSupervisor, Supervisor, SupervisorType } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import Page from "./util/page";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./util/uuid_display";
import HumanReviews from "./human_reviews";
import { EyeIcon } from "lucide-react";
import JsonDisplay from "./util/json_display";

export default function SupervisorDetails() {
  const [supervisor, setSupervisor] = useState<Supervisor>();
  const { supervisorId } = useParams();

  const { data: supervisorData, isLoading: supervisorLoading, error: supervisorError } = useGetSupervisor(supervisorId || '');


  useEffect(() => {
    if (supervisorData?.data) {
      setSupervisor(supervisorData.data);
    }
  }, [supervisorData]);

  return (
    <Page title={`Supervisor ${supervisor?.name} reviews`} subtitle={<span>Review for supervisor <UUIDDisplay uuid={supervisor?.id} /> will be displayed here</span>} icon={<EyeIcon />}>
      {supervisorLoading && <div>Loading...</div>}
      {supervisorError && <div>Error: {supervisorError.message}</div>}
      {supervisor && <>
        <div className="flex flex-col space-y-4 col-span-3">
          {supervisor?.type === SupervisorType.human_supervisor && <HumanReviews />}
          {supervisor?.type === SupervisorType.client_supervisor && <JsonDisplay json={supervisor} />}
          {supervisor?.type === SupervisorType.no_supervisor && <div>No supervisor</div>}
        </div>
      </>}
    </Page>
  )
}
