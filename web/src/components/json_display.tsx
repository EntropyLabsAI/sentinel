import { SupervisionRequest, SupervisionResult } from '@/types'
import { ScrollArea } from '@radix-ui/react-scroll-area'
import { Code } from 'lucide-react'
import React, { useState } from 'react'
import CopyButton from './copy_button'
import { Card, CardHeader, CardTitle, CardContent } from './ui/card'
import { Button } from '@/components/ui/button'


export default function JsonDisplay({ reviewRequest }: { reviewRequest: SupervisionRequest | SupervisionResult }) {
  const [showJson, setShowJson] = useState(true)
  const jsonString = JSON.stringify(reviewRequest, null, 2)

  return (
    <Card className="mt-4 overflow-scroll">
      <CardHeader>
      </CardHeader>
      <CardContent>
        {showJson && (
          <ScrollArea className="h-[300px]">
            <pre className="text-xs">{jsonString}</pre>
          </ScrollArea>
        )}
      </CardContent>
    </Card>
  )
}
