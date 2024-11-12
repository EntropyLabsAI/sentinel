import React, { useState, useEffect } from 'react';
import {
  SupervisionResult,
  ToolRequest,
  Decision,
  ReviewPayload,
  SupervisionRequest,
  useGetSupervisionReviewPayload
} from '@/types';
import ReviewRequestDisplay from '@/components/supervisor/review_request';
import { HubStatsAccordion } from '../util/hub_stats';
import { useConfig } from '@/contexts/config_context';
import axios from 'axios';
import { EyeIcon } from 'lucide-react';
import { UUIDDisplay } from '../util/uuid_display';
import { Supervisor } from '@/types';
import JSONDisplay from '../util/json_display';
import { SupervisorBadge, ToolBadge } from '../util/status_badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";

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

  const [showFeedbackDialog, setShowFeedbackDialog] = useState(false);
  const [feedbackText, setFeedbackText] = useState("");
  const [pendingToolChoice, setPendingToolChoice] = useState<ToolRequest | null>(null);

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

      // If no review is selected, select the first one
      if (!selectedRequestId && (Object.keys(reviews).length > 0 || nextRequestId)) {
        setSelectedRequestId(nextRequestId);
      }

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
  const sendResponse = (decision: Decision, requestId: string, toolChoice: ToolRequest, feedback?: string) => {
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
    <div className="p-16 flex flex-col gap-16 container">
      {/* Main Content */}
      <div className="flex gap-6">
        {/* Sidebar */}
        <div className="w-full max-w-[200px] pr-4 border-r flex-shrink-0">
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
                    <div className="flex flex-col gap-1 min-w-0">
                      <div className="flex flex-wrap items-center gap-1">
                        <span>Agent wants to use</span>
                        <ToolBadge toolId={payload.request_group.tool_requests[0].tool_id || ''} />
                      </div>
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        {/* Main Content */}
        <div className="w-full pl-4 min-w-0">
          {!selectedReviewPayload ? (<></>) : (
            <>
              <div id="content" className="space-y-6 break-words">
                <ReviewRequestDisplay
                  reviewPayload={selectedReviewPayload}
                  sendResponse={(decision: Decision, toolChoice: ToolRequest, feedback?: string) => {
                    if (decision === Decision.reject) {
                      setShowFeedbackDialog(true);
                      setPendingToolChoice(toolChoice);
                    } else {
                      sendResponse(decision, selectedReviewPayload.supervision_request.id || '', toolChoice);
                    }
                  }}
                />
              </div>

              <Dialog open={showFeedbackDialog} onOpenChange={setShowFeedbackDialog}>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Provide Rejection Feedback</DialogTitle>
                  </DialogHeader>
                  <textarea
                    className="w-full p-2 border rounded"
                    value={feedbackText}
                    onChange={(e) => setFeedbackText(e.target.value)}
                    placeholder="Enter feedback for rejection..."
                    rows={4}
                  />
                  <DialogFooter className="flex gap-2">
                    <button
                      className="px-4 py-2 bg-red-500 text-white rounded"
                      onClick={() => {
                        if (pendingToolChoice) {
                          sendResponse(
                            Decision.reject,
                            selectedReviewPayload.supervision_request.id || '',
                            pendingToolChoice,
                            feedbackText
                          );
                        }
                        setShowFeedbackDialog(false);
                        setFeedbackText("");
                        setPendingToolChoice(null);
                      }}
                    >
                      Submit Rejection
                    </button>
                    <button
                      className="px-4 py-2 bg-gray-300 rounded"
                      onClick={() => {
                        setShowFeedbackDialog(false);
                        setFeedbackText("");
                        setPendingToolChoice(null);
                      }}
                    >
                      Cancel
                    </button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </>
          )}
        </div>

      </div>

      <div className="">
        <h2 className="text-xl font-semibold mb-4">Supervisor Config</h2>
        <p className="text-sm text-muted-foreground mb-4">
          Configuration for human supervisor <UUIDDisplay uuid={supervisor.id} />.
        </p>
        <JSONDisplay json={supervisor} />
      </div>

      <div className="">
        <h2 className="text-xl font-semibold mb-4">Hub Stats</h2>
        <HubStatsAccordion API_BASE_URL={API_BASE_URL} />
      </div>
    </div>
  );
};

export default HumanReviews;
