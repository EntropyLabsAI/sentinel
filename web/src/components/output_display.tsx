import { Output } from '@/types'
import { Text } from 'lucide-react'
import React from 'react'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { Badge } from '@/components/ui/badge'

export default function OutputDisplay({ output }: { output: Output }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center">
          <Text className="mr-2" />
          Output
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm mb-2"><span className="font-semibold">Model:</span> {output.model}</p>
        <div className="space-y-2">
          {output.choices && output.choices.map((choice, index) => (
            <div key={index} className="border-t pt-2 first:border-t-0 first:pt-0">
              <Badge className="mb-1">{choice.message.role}</Badge>
              <p className="text-sm">{choice.message.content}</p>
              <span className="text-xs text-muted-foreground">Stop Reason: {choice.stop_reason}</span>
            </div>
          ))}
        </div>
        <div className="mt-4 text-xs text-muted-foreground">
          <p>Input Tokens: {output.usage?.input_tokens}</p>
          <p>Output Tokens: {output.usage?.output_tokens}</p>
          <p>Total Tokens: {output.usage?.total_tokens}</p>
        </div>
      </CardContent>
    </Card>
  )
}
