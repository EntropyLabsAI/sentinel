import React, { useState } from 'react'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from '@radix-ui/react-accordion'
import { FileJsonIcon, Copy, Check } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function JsonDisplay({ json }: { json: any }) {
  const [showJson, setShowJson] = useState(true)
  const [copied, setCopied] = useState(false)
  const jsonString = JSON.stringify(json, null, 2)

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(jsonString)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Accordion type="single" collapsible className="w-full">
      <AccordionItem value="messages" className="border border-gray-200 rounded-md">
        <AccordionTrigger className="w-full p-4 rounded-md cursor-pointer focus:outline-none">
          <div className="flex flex-row gap-4 items-center">
            <FileJsonIcon className="w-4 h-4" />
            Supervision JSON
          </div>
        </AccordionTrigger>
        <AccordionContent className="p-4">
          {showJson && (
            <ScrollArea className="max-h-[1000px] w-full rounded-md border overflow-auto">
              <Button
                variant="ghost"
                size="icon"
                className="absolute right-4 top-4 z-10"
                onClick={copyToClipboard}
              >
                {copied ? (
                  <Check className="h-4 w-4" />
                ) : (
                  <Copy className="h-4 w-4" />
                )}
              </Button>
              <pre className="text-xs p-4">{jsonString}</pre>
            </ScrollArea>
          )}
        </AccordionContent>
      </AccordionItem>
    </Accordion>
  )
}
