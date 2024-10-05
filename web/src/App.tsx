import React, { useState, useEffect } from 'react';
import { ReviewRequest } from './review';
import ReviewRequestDisplay from './components/review_request';

const ApprovalsInterface: React.FC = () => {
  const [reviewDataList, setReviewDataList] = useState<ReviewRequest[]>([]);
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket(`ws://localhost:8080/ws`);
    setSocket(ws);

    ws.onmessage = (event) => {
      const data: ReviewRequest = JSON.parse(event.data);

      console.log(`Received data: ${JSON.stringify(data.request_id)}`);

      // Use functional update to ensure we have the latest state
      setReviewDataList((prevList) => {
        const newList = [...prevList, data];
        if (newList.length > 10) {
          newList.shift(); // Remove the oldest item
        }
        return newList;
      });

      // If no review is selected, automatically select the first one
      setSelectedRequestId((prevSelectedId) => prevSelectedId || data.request_id);
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    return () => {
      ws.close();
    };
  }, []);

  const sendResponse = (decision: string, requestId: string) => {
    console.log(`Sending response for request ${requestId}: ${decision}`);
    if (socket && socket.readyState === WebSocket.OPEN) {
      const response = {
        id: requestId,
        decision: decision,
      };
      socket.send(JSON.stringify(response));

      // Remove the handled review request from the list using functional update
      setReviewDataList((prevList) => {
        const newList = prevList.filter((req) => req.request_id !== requestId);

        // If the handled request was selected, select the next one
        setSelectedRequestId((prevSelectedId) => {
          if (prevSelectedId === requestId) {
            return newList.length > 0 ? newList[0].request_id : null;
          } else {
            return prevSelectedId;
          }
        });

        return newList;
      });
    }
  };

  const selectReviewRequest = (requestId: string) => {
    setSelectedRequestId(requestId);
  };

  const selectedReviewRequest = reviewDataList.find(
    (req) => req.request_id === selectedRequestId
  );

  return (
    <div className="container mx-auto px-4 py-8 flex">
      {/* Sidebar */}
      <div className="w-full md:w-1/4 pr-4 border-r">
        <h2 className="text-xl font-semibold mb-4">Review Requests</h2>
        {reviewDataList.length === 0 ? (
          <p>No review requests at the moment.</p>
        ) : (
          <ul className="space-y-2">
            {reviewDataList.map((req) => (
              <li
                key={req.request_id}
                className={`cursor-pointer p-2 rounded-md ${req.request_id === selectedRequestId
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-800'
                  }`}
                onClick={() => selectReviewRequest(req.request_id)}
              >
                <div className="font-semibold">Agent #{req.agent_id.slice(0, 8)}</div>
                <div className="text-sm">Request ID: {req.request_id.slice(0, 8)}</div>
                {req.tool_choice && (
                  <div className="text-xs italic mt-1">Tool: {req.tool_choice.function}</div>
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
              sendResponse={(decision) =>
                sendResponse(decision, selectedReviewRequest.request_id)
              }
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default ApprovalsInterface;
