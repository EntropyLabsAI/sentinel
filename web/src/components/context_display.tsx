import { TaskState } from '@/types';
import React from 'react';
import { MessagesDisplay } from './messages';
import ToolsDisplay from './tool_display';
import OutputDisplay from './output_display';

export default function ContextDisplay({ context }: { context: TaskState }) {
  return (
    <div className="space-y-4">
      <MessagesDisplay messages={context.messages} />
      {context.tools && context.tools.length > 0 && <ToolsDisplay tools={context.tools} />}
      {context.output && <OutputDisplay output={context.output} />}
    </div>
  )
}
