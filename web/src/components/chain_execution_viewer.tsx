import * as React from "react"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { RunExecution } from "@/types"
import ChainStateDisplay from "./chain_state_display"

interface RunExecutionViewerProps {
  runExecution: RunExecution | undefined;
}

export default function RunExecutionViewer({ runExecution }: RunExecutionViewerProps) {
  if (!runExecution) {
    return null;
  }

  return (
    <Card className="">
      <CardHeader>
        <CardTitle>Chain Details</CardTitle>
      </CardHeader>
      <CardContent>
        {runExecution.chains.map((chain) => (
          <ChainStateDisplay chainState={chain} />
        ))}
      </CardContent>
    </Card>
  )
}
