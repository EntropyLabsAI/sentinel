import { Tool, useGetRunTools, useGetRunMessages, AsteroidMessage, useGetRunChatCount } from "@/types";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import Page from "./util/page";
import { ToolsList } from "@/components/tools_list";
import { UUIDDisplay } from "@/components/util/uuid_display";
import { EyeIcon, ListOrderedIcon, PickaxeIcon, PyramidIcon } from "lucide-react";
import { ToolCard } from "./tool_card";
import { MessagesDisplay } from "./messages";
import { ToolCallState } from "./tool_call_state";
import { Query } from "@tanstack/react-query";

export default function Run() {
  const { runId } = useParams();
  const [tools, setTools] = useState<Tool[]>([]);
  const [selectedToolCallId, setSelectedToolCallId] = useState<string | null>(null);
  const [messages, setMessages] = useState<AsteroidMessage[]>([]);
  const [index, setIndex] = useState<number>(0);

  const { data: toolsData, isLoading: toolsLoading } = useGetRunTools(runId || '');
  const { data: chatCount } = useGetRunChatCount(runId || '');

  const { data: messageData } = useGetRunMessages(
    runId || '',
    index,
    { query: { enabled: !!runId, refetchInterval: 1000 } }
  );

  useEffect(() => {
    if (toolsData?.data) {
      setTools(deduplicateTools(toolsData.data));
    }
  }, [toolsData]);

  // TODO do this serverside 
  function deduplicateTools(tools: Tool[]) {
    return tools.filter((tool, i, self) =>
      i === self.findIndex((t) => t.id === tool.id)
    );
  }

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
            <MessagesDisplay
              chatCount={chatCount?.data || 0}
              index={index}
              setIndex={setIndex}
              expanded={true}
              messages={messages}
              onToolCallClick={setSelectedToolCallId}
              selectedToolCallId={selectedToolCallId || undefined}
            />
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
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {(tools as Tool[]).map((tool) => (
            <div className="col-span-1">
              <ToolCard key={tool.id} tool={tool} runId={runId || ''} />
            </div>
          ))}
        </div>
      </Page >
    </>
  );
}
