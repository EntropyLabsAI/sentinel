import { ScrollArea } from '@radix-ui/react-scroll-area'
import React, { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'


export default function JsonDisplay({ json }: { json: any }) {
  const [showJson, setShowJson] = useState(true)
  const jsonString = JSON.stringify(json, null, 2)

  console.log('josn is', json)

  return (
    <Card className="overflow-scroll">
      <CardContent className="pt-6">
        {showJson && (
          <ScrollArea className="">
            <pre className="text-xs">{jsonString}</pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
