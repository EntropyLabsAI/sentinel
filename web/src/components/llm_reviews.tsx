import React, { useEffect, useState } from 'react';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { SupervisionResult, SupervisionRequest } from '@/types';

interface ReviewListProps {
  API_BASE_URL: string;
}

export default function ReviewList({ API_BASE_URL }: ReviewListProps) {
  const [llmPrompt, setLLMPrompt] = useState('');
  const [llmReviewDataList, setLLMReviewDataList] = useState<SupervisionResult[]>([]);

  // Fetch LLM reviews and current prompt on component mount
  useEffect(() => {
    const fetchLLMReviews = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/review/llm/list`);
        const data: SupervisionResult[] = await response.json();
        setLLMReviewDataList(data);
      } catch (error) {
        console.error('Error fetching LLM reviews:', error);
      }
    };

    const fetchCurrentPrompt = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/review/llm/prompt`);
        const data = await response.json();
        setLLMPrompt(data.prompt);
      } catch (error) {
        console.error('Error fetching current LLM prompt:', error);
      }
    };

    fetchLLMReviews();
    fetchCurrentPrompt();
    const reviewsInterval = setInterval(fetchLLMReviews, 5000);

    return () => {
      clearInterval(reviewsInterval);
    };
  }, [API_BASE_URL]);

  const handleSubmitPrompt = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/review/llm/prompt`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ prompt: llmPrompt }),
      });

      if (!response.ok) {
        throw new Error('Failed to update LLM prompt');
      }

      const data = await response.json();
      alert(data.message);
    } catch (error) {
      console.error('Error updating LLM prompt:', error);
    }
  };

  return (
    <> </>
    // <div className="container mx-auto px-4 py-8 space-y-4">
    //   <h1 className="text-2xl font-bold mb-6">Review List</h1>
    //   <p className="text-gray-600">
    //     The LLM reviews interface is still under active development.
    //   </p>
    //   <p>
    //     Any reviews performed by the LLM will show up here. You can set the prompt that the LLM uses for reviews below. Make sure it includes <code>{`{function}`}</code> and <code>{`{arguments}`}</code> somewhere in the prompt.
    //   </p>

    //   {/* New prompt submission form */}
    //   <div className="mb-6">
    //     <h2 className="text-xl font-bold mb-2">Set LLM Review Prompt</h2>
    //     <textarea
    //       className="w-full p-2 border border-gray-300 rounded mb-2"
    //       rows={5}
    //       value={llmPrompt}
    //       onChange={(e) => setLLMPrompt(e.target.value)}
    //       placeholder="Enter the prompt for LLM reviews here..."
    //     ></textarea>
    //     <button
    //       onClick={handleSubmitPrompt}
    //       className="px-4 py-2 bg-blue-500 text-white rounded"
    //     >
    //       Update Prompt
    //     </button>
    //   </div>

    //   {llmReviewDataList.length === 0 ? (
    //     <p className="text-gray-600">The LLM has not performed any reviews yet.</p>
    //   ) : (
    //     <Accordion type="single" collapsible className="space-y-4">
    //       {llmReviewDataList.map((review, index) => (
    //         <AccordionItem key={review.id} value={`item-${index}`}>
    //           <AccordionTrigger className="w-full p-4 bg-white shadow-sm rounded-lg hover:bg-gray-50 text-left">
    //             <div className="grid grid-cols-[auto_1fr_auto] items-center w-full gap-4">
    //               <span
    //                 className={`text-sm font-semibold px-3 py-1 rounded-full ${review.decision === 'approve'
    //                   ? 'bg-green-100 text-green-800'
    //                   : 'bg-red-100 text-red-800'
    //                   }`}
    //               >
    //                 {review.decision.charAt(0).toUpperCase() + review.decision.slice(1)}
    //               </span>
    //               <span className="text-sm font-mono text-gray-900 truncate">
    //                 {review.toolrequest?.id?.slice(0, 8)}
    //               </span>
    //               <span className="text-sm text-gray-500 justify-self-end">ID: {review.id.slice(0, 8)}</span>
    //             </div>
    //           </AccordionTrigger>
    //           <AccordionContent className="p-4 bg-white shadow-sm rounded-lg mt-2">
    //             <h2 className="text-lg font-semibold mb-2">Reasoning:</h2>
    //             <p className="text-gray-700 mb-4 whitespace-pre-wrap">{review.reasoning}</p>
    //             <div className="bg-gray-100 p-4 rounded-md">
    //               <h3 className="text-md font-semibold mb-2">Command:</h3>
    //               <code className="text-sm bg-gray-200 px-2 py-1 rounded">
    //                 {JSON.stringify(review.toolrequest?.arguments, null, 2)}
    //               </code>
    //             </div>
    //           </AccordionContent>
    //         </AccordionItem>
    //       ))}
    //     </Accordion>
    //   )}
    // </div>
  );
}
