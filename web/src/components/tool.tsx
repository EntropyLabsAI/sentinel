import { Tool, useGetTool } from "@/types";
import React, { useEffect, useState } from "react";
import Page from "./util/page";
import { useParams } from "react-router-dom";
import { ToolCard } from "./tool_card";
import { PickaxeIcon } from "lucide-react";
import LoadingSpinner from "./util/loading";

export default function ToolDetails() {
  const { toolId } = useParams();
  const [tool, setTool] = useState<Tool | null>(null);

  const { data, isLoading, error } = useGetTool(toolId || '');

  useEffect(() => {
    if (data?.data) {
      setTool(data.data);
    }
  }, [data]);

  return (
    <Page title={`Tool Details`} subtitle={`Details for tool ${tool?.name ?? "Tool"} (${toolId ?? ''})`} icon={<PickaxeIcon className="w-6 h-6" />}>
      <div className="col-span-3">
        {isLoading && <LoadingSpinner />}
        {error && <div>Error loading tool: {error.message}</div>}
        {tool && <ToolCard tool={tool} />}
      </div>
    </Page>
  )
}
