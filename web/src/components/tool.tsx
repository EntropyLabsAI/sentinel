import { Tool, useGetTool } from "@/types";
import React, { useEffect, useState } from "react";
import Page from "./page";
import { useParams } from "react-router-dom";
import { ToolCard } from "./tool_card";

export default function ToolDetails() {
  const { toolId } = useParams();
  const [tool, setTool] = useState<Tool | null>(null);

  if (!toolId) return <Page title="Tool">Tool ID is required</Page>

  const { data, isLoading, error } = useGetTool(toolId);

  useEffect(() => {
    if (data?.data) {
      setTool(data.data);
    }
  }, [data]);

  if (isLoading) return <Page title="Tool">Loading...</Page>;
  if (error) return <Page title="Tool">Error: {error.message}</Page>;

  if (!tool) return <Page title="Tool">Tool not found</Page>;

  return (
    <Page title={`Tool details`} subtitle={`Details for tool ${tool?.name || "Tool"}`}>
      <ToolCard tool={tool} />
    </Page>
  )
}
