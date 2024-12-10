import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tool, useGetRunTools, RunState, useGetRunState, useGetRunMessages, SentinelMessage } from "@/types";
import React, { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import Page from "./util/page";
import { ToolsList } from "@/components/tools_list";
import { UUIDDisplay } from "@/components/util/uuid_display";
import { EyeIcon, ListOrderedIcon, PickaxeIcon, PyramidIcon } from "lucide-react";
import { ToolCard } from "./tool_card";
import { MessagesDisplay } from "./messages";
import { ToolCallState } from "./tool_call_state";

export default function Run() {
  const { runId } = useParams();
  // const [runState, setRunState] = useState<RunState | null>(null);
  const [tools, setTools] = useState<Tool[]>([]);
  const [selectedToolCallId, setSelectedToolCallId] = useState<string | null>(null);

  const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(runId || '');
  // const { data: runStateData, isLoading: runStateLoading } = useGetRunState(runId || '', {
  //   query: {
  //     refetchInterval: 1000, // Refetch every 1 second
  //     refetchIntervalInBackground: true, // Optional: continue polling when window is in background
  //   }
  // });

  // useEffect(() => {
  //   if (runStateData?.data) {
  //     setRunState(runStateData.data);
  //   }
  // }, [runStateData]);

  useEffect(() => {
    if (toolsData?.data) {
      setTools(deduplicateTools(toolsData.data));
    }
  }, [toolsData]);

  // TODO do this serverside 
  function deduplicateTools(tools: Tool[]) {
    return tools.filter((tool, index, self) =>
      index === self.findIndex((t) => t.id === tool.id)
    );
  }

  const { data: messageData } = useGetRunMessages(runId || '');
  const [messages, setMessages] = useState<SentinelMessage[]>([]);

  useEffect(() => {
    if (messageData?.data) {
      setMessages(messageData.data);
    }
  }, [messageData]);

  if (!runId) {
    <p>No Run ID found</p>
  }

  return (
    <>
      <Page
        icon={<PyramidIcon className="h-5 w-5" />}
        title={`Run Details`}
        subtitle={
          <span className="text-sm text-muted-foreground"></span>
        }>
        <div className="grid grid-cols-1 xl:grid-cols-4 lg:grid-cols-4 gap-4">
          <div className="xl:col-span-2 lg:col-span-2">
            <MessagesDisplay expanded={true} messages={messages} onToolCallClick={setSelectedToolCallId} />
          </div>
          <div className="xl:col-span-2 lg:col-span-2">
            <ToolCallState toolCallId={selectedToolCallId || undefined} />
          </div>
        </div>
      </Page>
      <Page
        cols={2}
        icon={<PickaxeIcon className="h-5 w-5" />}
        title="Tools used in this run" subtitle={`${tools.length} tool${tools.length === 1 ? "" : "s"} used in this run`}>
        {(tools as Tool[]).map((tool) => (
          <div className="col-span-1">
            <ToolCard key={tool.id} tool={tool} runId={runId} />
          </div>
        ))}
      </Page >
    </>
  );
}

// <span className="text-sm text-muted-foreground">
//   The agent has made {runState?.length} tool execution{runState?.length === 1 ? "" : "s"} for run{' '}
//   <UUIDDisplay uuid={runId} />{' '}
//   across {tools.length} tool{tools.length === 1 ? "" : "s"}. To see more details for each tool execution, inspect the rows in the table below.
// </span>
