import React, { useState, useEffect } from 'react';
import {
  SupervisionResult,
  ToolRequest,
  Decision,
  ReviewPayload,
  SupervisionRequest,
  useGetSupervisionReviewPayload
} from '@/types';
import ReviewRequestDisplay from '@/components/review_request';
import { HubStatsAccordion } from './util/hub_stats';
import { useConfig } from '@/contexts/config_context';
import axios from 'axios';
import { EyeIcon } from 'lucide-react';
import { UUIDDisplay } from './util/uuid_display';
import { Supervisor } from '@/types';
import JSONDisplay from './util/json_display';

interface ReviewSectionProps {
  supervisor: Supervisor;
}

// Map the review request to the payload
type ReviewPayloadMap = {
  [key: string]: ReviewPayload;
};

const HumanReviews: React.FC<ReviewSectionProps> = ({ supervisor }) => {
  const { API_BASE_URL, WEBSOCKET_BASE_URL } = useConfig();
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isSocketConnected, setIsSocketConnected] = useState(false);

  // Queue for incoming requests that need their payloads fetched
  const [requestQueue, setRequestQueue] = useState<string[]>([]);
  // Map of request ID to review payload
  const [reviews, setReviews] = useState<ReviewPayloadMap>({});
  // Currently selected request ID
  const [selectedRequestId, setSelectedRequestId] = useState<string>();

  // Hook to fetch payload for the next request in queue
  const nextRequestId = requestQueue[0];
  const { data: nextReviewPayload } = useGetSupervisionReviewPayload(nextRequestId || '');

  // Process the next item in the queue when payload is received
  useEffect(() => {
    if (nextReviewPayload && nextRequestId) {
      // Add the new review to the map
      setReviews(prev => ({
        ...prev,
        [nextRequestId]: nextReviewPayload.data
      }));
      // Remove the processed request from queue
      setRequestQueue(prev => prev.slice(1));
    }
  }, [nextReviewPayload, nextRequestId]);

  // WebSocket initialization
  useEffect(() => {
    const ws = new WebSocket(WEBSOCKET_BASE_URL);
    setSocket(ws);

    ws.onopen = () => {
      setIsSocketConnected(true);
    };

    ws.onmessage = (event) => {
      const data: SupervisionRequest = JSON.parse(event.data);
      if (!data.id) {
        console.error('Received a message with no ID');
        return;
      }
      // Add new request to queue
      if (data.id) {
        setRequestQueue(prev => [...prev, data.id || '']);
      }
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
      setReviews({});
    };
  }, [WEBSOCKET_BASE_URL, API_BASE_URL]);

  // Send response and remove the review
  const sendResponse = (decision: Decision, requestId: string, toolChoice: ToolRequest) => {
    if (socket?.readyState === WebSocket.OPEN) {
      const response: SupervisionResult = {
        decision: decision,
        reasoning: 'Human decided via interface',
        created_at: new Date().toISOString(),
        supervision_request_id: requestId,
        chosen_toolrequest_id: toolChoice.id,
      };
      socket.send(JSON.stringify(response));

      // Remove the handled review and update selection
      setReviews(prev => {
        const newReviews = { ...prev };
        delete newReviews[requestId];

        // Update selection to next available review if exists
        const remainingIds = Object.keys(newReviews);
        setSelectedRequestId(remainingIds.length > 0 ? remainingIds[0] : undefined);

        return newReviews;
      });
    }
  };

  // When the user selects a review request, set the selected request ID
  const selectReviewRequest = (requestId: string) => {
    setSelectedRequestId(requestId);
  };

  // Find the selected review payload only when needed for rendering
  const selectedReviewPayload = selectedRequestId
    ? reviews[selectedRequestId]
    : null;

  return (
    <div className="p-16 flex flex-col gap-16">

      {/* Main Content */}
      <div className="flex">
        {/* Sidebar */}
        <div className="w-full md:w-1/4 pr-4 border-r">
          <h2 className="text-xl font-semibold mb-4">

            Review requests for human supervisor <UUIDDisplay uuid={supervisor.id} /> will be displayed here.
          </h2>
          {Object.keys(reviews).length === 0 ? (
            <p>No review requests at the moment.</p>
          ) : (
            <ul className="space-y-2">
              {Object.keys(reviews).map((id) => {
                const payload = reviews[id];
                return (
                  <li
                    key={id}
                    className={`cursor-pointer p-2 rounded-md ${payload.supervision_request.id === selectedRequestId
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-800'
                      }`}
                    onClick={() =>
                      selectReviewRequest(payload.supervision_request.id || '')
                    }
                  >
                    <div className="">
                      Agent{' '}
                      <code>{payload.supervision_request.chainexecution_id}</code>
                    </div>
                    <div className="text-sm">
                      Request{' '}
                      <code>{payload.supervision_request.id}</code>
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        {/* Main Content */}
        <div className="w-full md:w-3/4 pl-4">
          {!selectedReviewPayload ? (<></>) : (
            <>
              <div id="content" className="space-y-6">
                <ReviewRequestDisplay
                  reviewPayload={selectedReviewPayload}
                  sendResponse={(decision: Decision, toolChoice: ToolRequest) =>
                    sendResponse(
                      decision,
                      selectedReviewPayload.supervision_request.id || '',
                      toolChoice
                    )
                  }
                />
              </div>
            </>
          )}
        </div>

      </div>

      <div className="flex flex-col">
        <h2 className="text-xl font-semibold mb-4">Supervisor Config</h2>
        <JSONDisplay json={supervisor} />
      </div>

      <div className="flex flex-col">
        <h2 className="text-xl font-semibold mb-4">Hub Stats</h2>
        <HubStatsAccordion API_BASE_URL={API_BASE_URL} />
      </div>
    </div>
  );
};

export default HumanReviews;
