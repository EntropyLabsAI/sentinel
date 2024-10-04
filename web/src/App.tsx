import React, { useState, useEffect } from 'react';
import { ReviewRequest } from './review';
import ReviewRequestDisplay from './components/review_request';

const ApprovalsInterface: React.FC = () => {
  const [reviewData, setReviewData] = useState<ReviewRequest | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const ws = new WebSocket(`ws://localhost:8080/ws`);
    setSocket(ws);

    ws.onmessage = (event) => {
      const data: ReviewRequest = JSON.parse(event.data);

      // Update state with received data
      setReviewData(data);
      setIsLoading(false);
    };

    ws.onclose = () => {
      console.log('WebSocket connection closed');
    };

    return () => {
      ws.close();
    };
  }, []);

  const sendResponse = (decision: string) => {
    if (socket && socket.readyState === WebSocket.OPEN && reviewData) {
      const response = {
        id: reviewData.request_id,
        decision: decision
      };
      socket.send(JSON.stringify(response));

      // Reset state
      setReviewData(null);
      setIsLoading(true);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-6">Approval Interface</h1>

      {isLoading ? (
        <div id="loading" className="text-left">
          <p className="text-lg">Waiting for review requests...</p>
        </div>
      ) : (
        <div id="content" className="space-y-6">
          {reviewData && <ReviewRequestDisplay reviewRequest={reviewData} sendResponse={sendResponse} />}
        </div>
      )}
    </div>
  );
};

export default ApprovalsInterface;
