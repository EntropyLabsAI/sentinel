import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Executions from "./run";
import Page from "./util/page";
import { Button } from "./ui/button";
import { useProject } from "@/contexts/project_context";
import { UUIDDisplay } from "./util/uuid_display";
import { CreatedAgo } from "./util/created_ago";
import { Building2 } from "lucide-react";
import LoadingSpinner from "./util/loading";

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);
  const { setSelectedProject } = useProject();
  const navigate = useNavigate();

  const { data, isLoading, error } = useGetProjects(
    {
      query: {
        refetchInterval: 1000,
        refetchIntervalInBackground: true,
      }
    }
  );

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

  return (
    <Page title="Projects" icon={<Building2 className="w-6 h-6" />} subtitle={projects.length === 0 && <div>No projects found. To register a project, check out the <Link to="/api" className="text-blue-500">docs</Link>.</div>}>
      {isLoading && (
        <LoadingSpinner />
      )}

      {projects.map((project) => (
        <Card key={project.id} className="flex flex-col">
          <CardHeader>
            <CardTitle>{project.name}</CardTitle>
            <CardDescription className="flex flex-col">
              <span>Project <UUIDDisplay uuid={project.id} /></span>
              <span>
                <CreatedAgo datetime={project.created_at} />
              </span>
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
