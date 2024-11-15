import { Tool } from '@/types';
import { Hammer } from 'lucide-react';
import React from 'react';
import { Card, CardHeader, CardTitle, CardContent } from './ui/card';
import { Badge } from '@/components/ui/badge';


export default function ToolsDisplay({ tools }: { tools: Tool[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Hammer className="mr-2" />
          Tools
        </CardTitle>
      </CardHeader>
      <CardContent>
        {tools.map((tool, index) => (
          <div key={index} className="mb-2 last:mb-0">
            <Badge variant="outline" className="mb-1">{tool.name}</Badge>
            {tool.description && <p className="text-sm">{tool.description}</p>}
            {tool.attributes && (
              <pre className="text-xs mt-1 bg-muted p-2 rounded">
                {JSON.stringify(tool.attributes, null, 2)}
              </pre>
            )}
          </div>
        ))}
      </CardContent>
    </Card>
  )
}
