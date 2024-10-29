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
import { UUIDDisplay } from "./uuid_display";
import { CreatedAgo } from "./created_ago";

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
        <Card className="flex flex-col">
          <CardHeader>
            <CardTitle>{project.name}</CardTitle>
            <CardDescription>
              <span>Project <UUIDDisplay uuid={project.id} /></span>
              <div><CreatedAgo datetime={project.created_at} /></div>
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-grow">
          </CardContent>
          <CardFooter className="flex justify-end">
            <Link
              to={`/projects/${project.id}`}
              key={project.id}
              onClick={(e) => handleProjectSelect(project, e)}
            >
              <Button variant="outline">View Project</Button>
            </Link>
          </CardFooter>
        </Card>
      ))}
    </Page>
  )
}
