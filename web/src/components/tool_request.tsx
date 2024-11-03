import React from "react"
import { ToolRequest } from "@/types"

export default function ToolRequestDisplay({ tool_request }: { tool_request: ToolRequest }) {
  return <p>Tool request: {tool_request.tool_id}</p>
}
