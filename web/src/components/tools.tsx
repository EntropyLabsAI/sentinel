import { Tool, useGetTools } from "@/types";
import React, { useEffect, useState } from "react";
import Page from "./page";
import { CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Card } from "@radix-ui/themes";
import { Link } from "react-router-dom";

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
      {tools.map((tool) => (
        <Card className="bg-gray-50" key={tool.id}>
          <CardHeader>
            <CardTitle>{tool.name}</CardTitle>
            <CardDescription>{tool.description}</CardDescription>
            <CardFooter>
              <Link to={`/tools/${tool.id}`}>View</Link>
            </CardFooter>

          </CardHeader>
        </Card>
      ))}
    </Page>
  )
}

