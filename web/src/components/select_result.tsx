import React from 'react'
import { useState, useRef, useEffect } from 'react'
import { Button } from "@/components/ui/button"
import { ChevronDown, ChevronUp } from 'lucide-react'
import { SupervisionResultBadge } from "@/components/util/status_badge"
import { useUpdateRunResult } from '@/types'
import { useToast } from '@/hooks/use-toast'

interface SelectResultProps {
  result?: string
  possibleResults: string[]
  onResultChange?: (result: string) => void
  runId: string
}

export default function SelectResult({ result, possibleResults, onResultChange, runId }: SelectResultProps) {
  const [isOpen, setIsOpen] = useState(false)
  const [selectedResult, setSelectedResult] = useState(result)
  const dropdownRef = useRef<HTMLDivElement>(null)
  const { toast } = useToast()
  const updateRunResult = useUpdateRunResult()

  const handleToggle = () => setIsOpen(!isOpen)

  const handleResultClick = async (newResult: string) => {
    setSelectedResult(newResult)
    setIsOpen(false)

    const runResult = { result: newResult }

    try {
      await updateRunResult.mutateAsync({
        runId,
        data: runResult
      })

      if (onResultChange) {
        onResultChange(newResult)
      }
    } catch (error) {
      console.error('Failed to update run result:', error)
      toast({
        title: 'Failed to update run result',
        description: 'Error: ' + error,
        variant: 'destructive'
      })
    }
  }

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  return (
    <div className="relative" ref={dropdownRef}>
      <Button
        variant="ghost"
        onClick={handleToggle}
        className="h-6 px-1 rounded-md text-xs font-medium hover:bg-transparent"
        aria-haspopup="listbox"
        aria-expanded={isOpen}
      >
        {selectedResult ? (
          <SupervisionResultBadge result={selectedResult} />
        ) : (
          <span>Select result</span>
        )}
        {isOpen ? <ChevronUp className="h-3 w-3 ml-1" /> : <ChevronDown className="h-3 w-3 ml-1" />}
      </Button>
      {isOpen && (
        <div className="fixed mt-1 w-[150px] rounded-md bg-background shadow-lg ring-1 ring-black ring-opacity-5 z-50">
          <div className="py-1" role="listbox">
            {possibleResults.map((option) => (
              <Button
                key={option}
                variant="ghost"
                className="w-full justify-start rounded-none text-left px-2"
                onClick={() => handleResultClick(option)}
              >
                <SupervisionResultBadge result={option} />
              </Button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
