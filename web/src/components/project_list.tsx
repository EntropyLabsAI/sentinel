import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project } from "@/types";
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import Executions from "./executions";
import Page from "./page";

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);

  const { data, isLoading, error } = useGetProjects();

  useEffect(() => {
    if (data?.data) {
      setProjects(data.data);
    }
  }, [data]);

  if (isLoading) return <Page title="Projects">Loading...</Page>;
  if (error) return <Page title="Projects">Error: {error.message}</Page>;

  return (
    <Page title="Projects">
      {projects.length === 0 && <div>No projects found. To register a project, check out the <Link to="/api" className="text-blue-500">docs</Link>.</div>}
      {projects.map((project) => (
        <Link to={`/projects/${project.id}`} key={project.id}>
          <Card key={project.id} className="flex flex-col">
            <CardHeader>
              <CardTitle>{project.name}</CardTitle>
              <CardDescription>Project ID: {project.id}</CardDescription>
            </CardHeader>
            <CardContent className="flex-grow">
              {project.created_at}
            </CardContent>
          </Card>
        </Link>
      ))}
    </Page>
  )
}
