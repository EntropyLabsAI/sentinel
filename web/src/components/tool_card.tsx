import { Tool } from "@/types";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { Button } from "@radix-ui/themes";
import React from "react";

export function ToolCard({ tool }: { tool: Tool }) {
  return (
    <Link to={`/tools/${tool.id}`} key={tool.id}>
      <Card className="flex flex-col ">
        <CardHeader>
          <CardTitle>{tool.name}</CardTitle>
          <CardDescription>
            {tool.description}
            <div>Tool ID: {tool.id}</div>
          </CardDescription>
        </CardHeader>
        <CardContent className="flex-grow">
          {tool.created_at}
          {tool.attributes && <pre className="text-xs mt-1 bg-muted p-2 rounded">{JSON.stringify(tool.attributes, null, 2)}</pre>}
        </CardContent>

        <CardFooter>
          <Button variant="outline">View Tool</Button>
        </CardFooter>
      </Card>
    </Link>
  );
}
