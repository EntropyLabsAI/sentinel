import { Check, Copy } from "lucide-react"
import { useState } from "react"
import { Button } from "./ui/button"
import React from "react"

export default function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      size="icon"
      onClick={handleCopy}
      className="ml-2 bg-gray-700 hover:bg-gray-600 outline-none"
    >
      {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
    </Button>
  )
}
