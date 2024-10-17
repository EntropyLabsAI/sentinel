import React from 'react';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { Review, ReviewResult } from '@/types';

interface ReviewListProps {
  reviews?: ReviewResult[];
}

export default function ReviewList({ reviews = [] }: ReviewListProps) {
  return (
    <div className="container mx-auto px-4 py-8 space-y-4">
      <h1 className="text-2xl font-bold mb-6">Review List</h1>
      <p className="text-gray-600">
        The LLM reviews interface is still under development, but any reviews performed by the LLM will show up here. We'll make this more user-friendly soon, including allowing you to set the prompt that the LLM uses for reviews. Currently, it's set in <code>callLLMForReview</code> in <code>handlers.go</code>.
      </p>
      {reviews.length === 0 ? (
        <p className="text-gray-600">The LLM has not performed any reviews yet.</p>
      ) : (
        <Accordion type="single" collapsible className="space-y-4">
          {reviews.map((review, index) => (
            <AccordionItem key={review.id} value={`item-${index}`}>
              <AccordionTrigger className="w-full p-4 bg-white shadow-sm rounded-lg hover:bg-gray-50 text-left">
                <div className="grid grid-cols-[auto_1fr_auto] items-center w-full gap-4">
                  <span
                    className={`text-sm font-semibold px-3 py-1 rounded-full ${review.decision === 'approve'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                      }`}
                  >
                    {review.decision.charAt(0).toUpperCase() + review.decision.slice(1)}
                  </span>
                  <span className="text-sm font-mono text-gray-900 truncate">
                    {review.tool_choice.arguments.cmd}
                  </span>
                  <span className="text-sm text-gray-500 justify-self-end">ID: {review.id.slice(0, 8)}</span>
                </div>
              </AccordionTrigger>
              <AccordionContent className="p-4 bg-white shadow-sm rounded-lg mt-2">
                <h2 className="text-lg font-semibold mb-2">Reasoning:</h2>
                <p className="text-gray-700 mb-4 whitespace-pre-wrap">{review.reasoning}</p>
                <div className="bg-gray-100 p-4 rounded-md">
                  <h3 className="text-md font-semibold mb-2">Command:</h3>
                  <code className="text-sm bg-gray-200 px-2 py-1 rounded">
                    {review.tool_choice.arguments.cmd}
                  </code>
                </div>
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      )}
    </div>
  );
}
