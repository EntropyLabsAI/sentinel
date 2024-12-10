import * as React from 'react';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Clock, CheckCircle2, XCircle } from 'lucide-react'
import { SentinelToolCall } from '@/types';
import { CreatedAgo } from './util/created_ago';
import { StatusBadge } from './util/status_badge';
import { UUIDDisplay } from './util/uuid_display';

interface ToolCallCardProps {
  status: "completed" | "failed" | "pending";
  toolCall: SentinelToolCall | undefined;
}

export default function ToolCallCard({ status, toolCall }: ToolCallCardProps) {
  if (!toolCall) {
    return null;
  }

  const parsedArguments = JSON.parse(toolCall.arguments || '{}');
  const functionArguments = JSON.parse(parsedArguments.arguments || '{}');

  const getStatusIcon = () => {
    switch (status) {
      case "completed":
        return <CheckCircle2 className="h-5 w-5 text-green-500" />;
      case "failed":
        return <XCircle className="h-5 w-5 text-red-500" />;
      case "pending":
        return <Clock className="h-5 w-5 text-yellow-500" />;
    }
  };

  const getStatusColor = () => {
    switch (status) {
      case "completed":
        return "bg-green-100 text-green-800";
      case "failed":
        return "bg-red-100 text-red-800";
      case "pending":
        return "bg-yellow-100 text-yellow-800";
    }
  };

  return (
    <Card className="w-full">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-semibold">Tool Call</CardTitle>
        {/* <Badge className={`${getStatusColor()} capitalize`}>
          {getStatusIcon()}
          <span className="ml-1">{status}</span>
        </Badge> */}
        <StatusBadge status={status} />
      </CardHeader>
      <CardContent>
        <div className="mt-2 space-y-2">
          <code className="rounded bg-muted p-2 font-mono text-sm">
            {parsedArguments.name}({Object.entries(functionArguments).map(([key, value]) => `${key}: "${value}"`).join(", ")})
          </code>
        </div>
      </CardContent>
      <CardFooter className="text-xs text-muted-foreground flex flex-row justify-between">
        <CreatedAgo datetime={toolCall.created_at || ''} />
        <UUIDDisplay uuid={toolCall.id} />
      </CardFooter>
    </Card>
  )
}

