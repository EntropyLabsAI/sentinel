import { SupervisionRequest, SupervisionResult } from '@/types'
import { ScrollArea } from '@radix-ui/react-scroll-area'
import { Code } from 'lucide-react'
import React, { useState } from 'react'
import CopyButton from './copy_button'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { Button } from '@/components/ui/button'


export default function JsonDisplay({ json }: { json: any }) {
  const [showJson, setShowJson] = useState(true)
  const jsonString = JSON.stringify(json, null, 2)

  console.log('josn is', json)

  return (
    <Card className="mt-4 overflow-scroll">
      <CardHeader>
      </CardHeader>
      <CardContent>
        {showJson && (
          <ScrollArea className="">
            <pre className="text-xs">{jsonString}</pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
