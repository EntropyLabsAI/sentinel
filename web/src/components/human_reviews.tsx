import React, { useState, useEffect } from 'react';
import { Review, ReviewResult, Decision, ToolChoice } from '@/types';
import ReviewRequestDisplay from '@/components/review_request';
import { HubStatsAccordion } from './hub_stats';

interface ReviewSectionProps {
  API_BASE_URL: string;
  WEBSOCKET_BASE_URL: string;
  setIsSocketConnected: (isConnected: boolean) => void;
}

const HumanReviews: React.FC<ReviewSectionProps> = ({
  API_BASE_URL,
  WEBSOCKET_BASE_URL,
  setIsSocketConnected,
}) => {
  const [humanReviewDataList, setHumanReviewDataList] = useState<Review[]>([]);
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);

  // Initialize WebSocket connection
  useEffect(() => {
    const ws = new WebSocket(WEBSOCKET_BASE_URL);
    setSocket(ws);

    ws.onopen = () => {
      console.log('WebSocket connection opened');
      setIsSocketConnected(true);
    };

    ws.onmessage = (event) => {
      const data: Review = JSON.parse(event.data);

      // Use functional update to ensure we have the latest state
      setHumanReviewDataList((prevList) => {
        const newList = [...prevList, data];
        if (newList.length > 10) {
          newList.shift(); // Remove the oldest item
        }
        return newList;
      });

      // If no review is selected, automatically select the first one
      setSelectedRequestId((prevSelectedId) => prevSelectedId || data.id);
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
      setIsSocketConnected(false);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      setIsSocketConnected(false);
    };

    return () => {
      ws.close();
      // Wipe the review data list. The reviews will be reloaded from the server when the connection is re-established
      setHumanReviewDataList([]);
    };
  }, [WEBSOCKET_BASE_URL, setIsSocketConnected]);

  // toolChoiceModified is a helper function to check if the tool choice has been modified
  const toolChoiceModified = (allToolChoices: ToolChoice[], toolChoice: ToolChoice) => {
    const originalToolChoice = allToolChoices.find(t => t.id === toolChoice.id);
    const modified = originalToolChoice && originalToolChoice.arguments !== toolChoice.arguments;
    return modified;
  };

  // Send a response to the Approvals API with the decision and the tool choice
  const sendResponse = (decision: Decision, requestId: string, toolChoice: ToolChoice) => {
    const selectedReviewRequest = humanReviewDataList.find(
      (req) => req.id === requestId
    );

    // Check if the tool args of the tool the user chose is not the same was it was originally
    if (selectedReviewRequest && toolChoiceModified(selectedReviewRequest.request.tool_choices, toolChoice)) {
      decision = Decision.modify;
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
      const response: ReviewResult = {
        id: requestId,
        decision: decision,
        reasoning: "Human decided via interface",
        tool_choice: toolChoice
      };
      socket.send(JSON.stringify(response));

      // Remove the handled review request from the list
      setHumanReviewDataList((prevList) => {
        const newList = prevList.filter((req) => req.id !== requestId);
        setSelectedRequestId((prevSelectedId) => {
          if (prevSelectedId === requestId) {
            return newList.length > 0 ? newList[0].id : null;
          } else {
            return prevSelectedId;
          }
        });
        return newList;
      });
    }
  };

  // When the user selects a review request, set the selected request ID
  const selectReviewRequest = (requestId: string) => {
    setSelectedRequestId(requestId);
  };

  // Find the selected review request
  const selectedReviewRequest = humanReviewDataList.find(
    (req) => req.id === selectedRequestId
  );

  return (
    <div>
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
            <>
              <div id="content" className="space-y-6">
                <ReviewRequestDisplay
                  reviewRequest={selectedReviewRequest}
                  sendResponse={(decision: Decision, toolChoice: ToolChoice) =>
                    sendResponse(decision, selectedReviewRequest.id, toolChoice)
                  }
                />
              </div>
            </>
          )}
        </div>
      </div>
      <div className="container mx-auto px-4 py-8 flex flex-col">
        <HubStatsAccordion API_BASE_URL={API_BASE_URL} />
      </div>
    </div>
  );
};

export default HumanReviews;
