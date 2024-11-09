import { Tool, useGetProject, useGetProjectTools } from "@/types";
import React, { useEffect, useState } from "react";
import Page from "./page";
import { ToolsList } from "@/components/tools_list";
import { useProject } from '@/contexts/project_context';
import { UUIDDisplay } from "./uuid_display";
import LoadingSpinner from "./loading";
import { PickaxeIcon } from "lucide-react";

export default function Tools() {
  const [tools, setTools] = useState<Tool[]>([]);
  const { selectedProject } = useProject();
  const { data: projectData, isLoading: projectLoading, error: projectError } = useGetProject(selectedProject || '');

  const { data, isLoading, error } = useGetProjectTools(
    selectedProject || '',
  );

  useEffect(() => {
    if (data?.data) {
      setTools(data.data);
    } else {
      setTools([]);
    }
  }, [data, selectedProject]);


  if (!selectedProject) {
    return <div>Please select a project first</div>;
  }

  return (
    <Page title={`Tools`} subtitle={<span>{tools.length} tool{tools.length === 1 ? '' : 's'} found for project {projectData?.data?.name} (<UUIDDisplay uuid={projectData?.data?.id} />)</span>} icon={<PickaxeIcon className="w-6 h-6" />}>
      <div className="col-span-3">

        {isLoading && (
          <LoadingSpinner />
        )}
        {error && (
          <div>Error loading tools: {error.message}</div>
        )}
        {tools.length > 0 && <ToolsList tools={tools} variant="card" />}
        {tools.length === 0 && <p className="text-sm text-gray-500">No tools found for this project. When your agent registers a tool, it will appear here.</p>}
      </div>

    </Page>
  );
}
