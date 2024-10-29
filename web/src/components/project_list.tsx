import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Executions from "./run";
import Page from "./page";
import { Button } from "./ui/button";
import { useProject } from "@/contexts/project_context";

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);
  const { setSelectedProject } = useProject();
  const navigate = useNavigate();

  const { data, isLoading, error } = useGetProjects();

  useEffect(() => {
    if (data?.data) {
      setProjects(data.data);
    }
  }, [data]);

  const handleProjectSelect = (project: Project, e: React.MouseEvent) => {
    e.preventDefault();
    setSelectedProject(project.id);
    navigate(`/projects/${project.id}`);
  };

  if (isLoading) return <Page title="Projects">Loading...</Page>;
  if (error) return <Page title="Projects">Error: {error.message}</Page>;

  return (
    <Page title="Projects">
      {projects.length === 0 && <div>No projects found. To register a project, check out the <Link to="/api" className="text-blue-500">docs</Link>.</div>}

      {projects.map((project) => (
        <Link
          to={`/projects/${project.id}`}
          key={project.id}
          onClick={(e) => handleProjectSelect(project, e)}
        >
          <Card className="flex flex-col">
            <CardHeader>
              <CardTitle>{project.name}</CardTitle>
              <CardDescription>Project ID: {project.id}</CardDescription>
            </CardHeader>
            <CardContent className="flex-grow">
              {project.created_at}
            </CardContent>
            <CardFooter>
              <Button variant="outline">View Project</Button>
            </CardFooter>
          </Card>
        </Link>
      ))}
    </Page>
  )
}
