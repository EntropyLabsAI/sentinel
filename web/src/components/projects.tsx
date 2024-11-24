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
import { Building2, Code2, CogIcon, Database, FileAudio, FileCode, FileIcon, FileImage, FileMusic, FileText, FileVideo, Gauge, GitBranch, Globe, Layout, ListIcon, Server, Terminal } from "lucide-react";
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
      // Sort the projects by most recently created
      setProjects(data.data.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()));
    }
  }, [data]);

  const handleProjectSelect = (project: Project, e: React.MouseEvent) => {
    e.preventDefault();
    setSelectedProject(project.id);
    navigate(`/projects/${project.id}`);
  };

  return (
    <Page
      title="Projects"
      icon={<Building2 className="w-6 h-6" />}
      subtitle={projects.length === 0 && <div>No projects found. To register a project, check out the <Link to="/api" className="text-blue-500">docs</Link>.</div>}
      cols={1}
    >
      {isLoading && (
        <LoadingSpinner />
      )}

      <div className="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">

        {projects.map((project) => (
          <div className="">
            <Link
              to={`/project/${project.id}`}
              key={project.id}
              onClick={(e) => handleProjectSelect(project, e)}
            >
              <Card key={project.id} className="flex">
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <FileIcon className="w-4 h-4 min-w-4" />
                    {project.name}
                  </CardTitle>
                  <CardDescription className="flex flex-col">
                    <span>Project <UUIDDisplay uuid={project.id} /></span>
                    <span>
                      <CreatedAgo datetime={project.created_at} />
                    </span>
                  </CardDescription>
                </CardHeader>
                <CardContent className="flex-grow">
                </CardContent>
              </Card>
            </Link>
          </div>
        ))}
      </div>
    </Page>
  )
}
