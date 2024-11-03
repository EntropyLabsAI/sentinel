import React from "react"
import { ToolRequest } from "@/types"

export default function ToolRequestDisplay({ tool_request }: { tool_request: ToolRequest }) {
  return <p>{JSON.stringify(tool_request)}</p>

}
