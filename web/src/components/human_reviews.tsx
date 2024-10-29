import React, { useState, useEffect } from 'react';
import { SupervisionResult, ToolRequest, Decision, SupervisionRequest } from '@/types';
import ReviewRequestDisplay from '@/components/review_request';
import { HubStatsAccordion } from './hub_stats';
import { useConfig } from '@/contexts/config_context';

interface ReviewSectionProps {
}

const HumanReviews: React.FC<ReviewSectionProps> = ({
}) => {
  const [humanReviewDataList, setHumanReviewDataList] = useState<SupervisionRequest[]>([]);
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isSocketConnected, setIsSocketConnected] = useState(false);

  const { API_BASE_URL, WEBSOCKET_BASE_URL } = useConfig();

  // Initialize WebSocket connection
  useEffect(() => {
    const ws = new WebSocket(WEBSOCKET_BASE_URL);
    setSocket(ws);

    ws.onopen = () => {
      console.log('WebSocket connection opened');
    };

    ws.onmessage = (event) => {
      const data: SupervisionRequest = JSON.parse(event.data);

      if (!data.id) {
        console.error('Received a message with no ID');
        return;
      }

      // Use functional update to ensure we have the latest state
      setHumanReviewDataList((prevList) => {
        const newList = [...prevList, data];
        return newList;
      });

      // If no review is selected, automatically select the first one in the list
      setSelectedRequestId(humanReviewDataList[0]?.id || null);
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
  const toolChoiceModified = (allToolChoices: ToolRequest[], toolChoice: ToolRequest) => {
    const originalToolChoice = allToolChoices.find(t => t.id === toolChoice.id);
    const modified = originalToolChoice && originalToolChoice.arguments !== toolChoice.arguments;
    return modified;
  };

  // Send a response to the API with the decision and the tool choice
  const sendResponse = (decision: Decision, requestId: string, toolChoice: ToolRequest) => {
    const selectedReviewRequest = humanReviewDataList.find(
      (req) => req.id === requestId
    );

    // Check if the tool args of the tool the user chose is not the same was it was originally
    if (selectedReviewRequest && toolChoiceModified(selectedReviewRequest.tool_requests, toolChoice)) {
      decision = Decision.modify;
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
      const response: SupervisionResult = {
        id: requestId,
        decision: decision,
        reasoning: "Human decided via interface",
        toolrequest: toolChoice,
        created_at: new Date().toISOString(),
        supervision_request_id: requestId,
      };
      socket.send(JSON.stringify(response));

      // Remove the handled review request from the list
      setHumanReviewDataList((prevList) => prevList.filter((req) => req.id !== requestId));
      // If the selected review request was the one that was handled, select the first one in the list
      setSelectedRequestId(humanReviewDataList[0]?.id || null);
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
                  onClick={() => selectReviewRequest(req.id || '')}
                >
                  <div className="font-semibold">Agent #{req.run_id}</div>
                  <div className="text-sm">Request ID: {req.id?.slice(0, 8)}</div>
                  {req.tool_requests && (
                    <div className="text-xs italic mt-1">Tool: {req.tool_requests[0].id}</div>
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
                  sendResponse={(decision: Decision, toolChoice: ToolRequest) =>
                    sendResponse(decision, selectedReviewRequest.id || '', toolChoice)
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
