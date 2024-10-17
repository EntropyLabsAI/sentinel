import React from 'react';
import { Review, ReviewResult, Decision, ToolChoice } from '@/types';
import ReviewRequestDisplay from '@/components/review_request';

interface ReviewSectionProps {
  humanReviewDataList: Review[];
  llmReviewDataList: ReviewResult[];
  selectedRequestId: string | null;
  selectReviewRequest: (requestId: string) => void;
  selectedReviewRequest: Review | undefined;
  sendResponse: (decision: Decision, requestId: string, toolChoice: ToolChoice) => void;
}

const ReviewSection: React.FC<ReviewSectionProps> = ({
  humanReviewDataList,
  selectedRequestId,
  selectReviewRequest,
  selectedReviewRequest,
  sendResponse,
}) => {
  return (
    <div className="container mx-auto px-4 py-8 flex">
      {/* Sidebar */}
      <div className="w-full md:w-1/4 pr-4 border-r">
        <h2 className="text-xl font-semibold mb-4">Review Requests</h2>
        {humanReviewDataList.length === 0 ? (
          <p>No review requests at the moment.</p>
        ) : (
          <ul className="space-y-2">
            {humanReviewDataList.map((req) => (
              <li
                key={req.id}
                className={`cursor-pointer p-2 rounded-md ${req.id === selectedRequestId
                    ? 'bg-blue-500 text-white'
                    : 'bg-gray-100 text-gray-800'
                  }`}
                onClick={() => selectReviewRequest(req.id)}
              >
                <div className="font-semibold">Agent #{req.request.agent_id}</div>
                <div className="text-sm">Request ID: {req.id.slice(0, 8)}</div>
                {req.request.tool_choices && (
                  <div className="text-xs italic mt-1">Tool: {req.request.tool_choices[0].function}</div>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Main Content */}
      <div className="w-full md:w-3/4 pl-4">
        {!selectedReviewRequest ? (
          <div id="loading" className="text-left">
            <p className="text-lg">Select a review request from the sidebar.</p>
          </div>
        ) : (
          <div id="content" className="space-y-6">
            <ReviewRequestDisplay
              reviewRequest={selectedReviewRequest}
              sendResponse={(decision: Decision, toolChoice: ToolChoice) =>
                sendResponse(decision, selectedReviewRequest.id, toolChoice)
              }
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default ReviewSection;
