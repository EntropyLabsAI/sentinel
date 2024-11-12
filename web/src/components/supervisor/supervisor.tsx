import { useGetSupervisor, Supervisor, SupervisorType } from "@/types";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import Page from "@/components/util/page";
import HumanReviews from "@/components/supervisor/human_reviews";
import JsonDisplay from "@/components/util/json_display";
import { UUIDDisplay } from "@/components/util/uuid_display";
import { EyeIcon } from "lucide-react";

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
    <div className="">
      {supervisorLoading && <div>Loading...</div>}
      {supervisorError && <div>Error: {supervisorError.message}</div>}
      {supervisor && <>
        {supervisor?.type === SupervisorType.human_supervisor && <HumanReviews supervisor={supervisor} />}
        {supervisor?.type !== SupervisorType.human_supervisor && (
          <Page cols={3} title={`Supervisor "${supervisor?.name}" Details`} subtitle={<span>Details and reviews for supervisor <UUIDDisplay uuid={supervisor?.id} /> will be displayed here</span>} icon={<EyeIcon />}>
            <div className="col-span-3">
              <JsonDisplay json={supervisor} />
            </div>
          </Page>
        )}
      </>}
    </div>
  )
}
