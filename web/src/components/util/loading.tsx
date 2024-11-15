import { Loader2 } from 'lucide-react'
import React from 'react'

export default function LoadingSpinner() {
  return (
    <div className="flex w-full h-[500px] justify-center items-center h-full">
      <Loader2 className="animate-spin h-4 w-4" />
    </div>
  )
}
