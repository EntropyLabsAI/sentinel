import { Tool, useGetTools } from "@/types";
import React, { useEffect, useState } from "react";
import Page from "./page";
import { CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Card } from "@radix-ui/themes";
import { Link } from "react-router-dom";
import { ToolsList } from "@/components/tools_list";

export default function Tools() {
  const [tools, setTools] = useState<Tool[]>([]);
  const { data, isLoading, error } = useGetTools();

  useEffect(() => {
    if (data?.data) {
      setTools(data.data);
    }
  }, [data]);

  if (isLoading) return <Page title="Tools">Loading...</Page>;
  if (error) return <Page title="Tools">Error: {error.message}</Page>;

  return (
    <Page title="Tools">
      <ToolsList tools={tools} variant="card" />
    </Page>
  );
}
