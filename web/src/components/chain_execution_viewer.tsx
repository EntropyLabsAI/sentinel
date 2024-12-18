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
        <div className="flex flex-col gap-2">
          {runExecution.chains.map((chain, index) => (
            <ChainStateDisplay chainState={chain} index={index} />
          ))}
        </div>
      </CardContent>
    </Card>
  )
}
