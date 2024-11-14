import React from 'react'
import { useState, useRef, useEffect } from 'react'
import { Button } from "@/components/ui/button"

interface SlideoverProps {
  isOpen: boolean;
  setIsOpen: (isOpen: boolean) => void;
}

export default function Slideover({ isOpen, setIsOpen }: SlideoverProps) {
  const divRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (divRef.current && !divRef.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  return (
    <div>
      {/* Overlay */}
      <div className={`fixed inset-0 bg-black bg-opacity-50 transition-opacity duration-300 z-40 ${isOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'}`} />

      {/* Sliding Div */}
      <div
        ref={divRef}
        className={`fixed top-0 right-0 w-3/4 h-full bg-white shadow-lg transform transition-all duration-300 ease-in-out z-50 ${isOpen ? 'translate-x-0' : 'translate-x-full'
          }`}
      >
        <div className="p-6">
          <h2 className="text-2xl font-bold mb-4">Sliding Div Content</h2>
          <p>This div slides in from the right and takes up about 75% of the screen.</p>
          <p className="mt-4">Click outside this div to close it.</p>
        </div>
      </div>
    </div>
  )
}
