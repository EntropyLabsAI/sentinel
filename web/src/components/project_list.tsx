import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useGetProjects, Project } from "@/types";
import React, { useEffect, useState } from "react";

export default function ProjectList() {
  const [projects, setProjects] = useState<Project[]>([]);

  const { data, isLoading, error } = useGetProjects();

  useEffect(() => {
    if (data?.data) {
      setProjects(data.data);
    }
  }, [data]);

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    data?.data && (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-6">Projects</h1>
        <p className="text-semibold">Here you will be able to view and manage, configure and export projects, runs and agent information. This feature will be available soon.</p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {projects.map((project) => (
            <Card key={project.id} className="flex flex-col">
              <CardHeader>
                <CardTitle>{project.name}</CardTitle>
                <CardDescription>Project ID: {project.id}</CardDescription>
              </CardHeader>
              <CardContent className="flex-grow">
                <h3 className="font-semibold mb-2">Tools:</h3>
                <ScrollArea className="h-[100px]">
                  <div className="flex flex-wrap gap-2">
                    {project.tools.map((tool, index) => (
                      <Badge key={index} variant="secondary">
                        {tool.name}
                      </Badge>
                    ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    )
  )
}
