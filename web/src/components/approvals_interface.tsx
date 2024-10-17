import React, { useState, useEffect } from 'react';
import { ReviewRequest, ReviewResult, ToolChoice, Decision, Review } from '@/types';
import ReviewSection from '@/components/review_section'; // Import the new component
import { HubStats as HubStatsType } from '@/types';
import LLMReviews from '@/components/llm_reviews';
import NavBar from './nav';
import SupervisorSelection from './supervisor_selection';
import { HubStatsAccordion } from './hub_stats';

// The API base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
// The websocket base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

const ApprovalsInterface: React.FC = () => {
  const [humanReviewDataList, setHumanReviewDataList] = useState<Review[]>([]);
  const [llmReviewDataList, setLLMReviewDataList] = useState<ReviewResult[]>([]);
  const [selectedRequestId, setSelectedRequestId] = useState<string | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [hubStats, setHubStats] = useState<HubStatsType | null>(null);
  const [selectedSupervisor, setSelectedSupervisor] = useState<string | null>(null);
  const [isSocketConnected, setIsSocketConnected] = useState<boolean>(false);

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
      setLLMReviewDataList([]);
    };
  }, []);

  // Start a timer to fetch the hub stats every second
  // TODO: this is a hack, but it works for now
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/stats`);
        const data: HubStatsType = await response.json();
        setHubStats(data);
      } catch (error) {
        console.error('Error fetching hub stats:', error);
      }
    };

    fetchStats();
    const statsInterval = setInterval(fetchStats, 1000);

    return () => {
      clearInterval(statsInterval);
    };
  }, []);

  // Fetch reviews done by the LLM every 5 seconds
  useEffect(() => {
    const fetchLLMReviews = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/review/llm/list`);
        const data: ReviewResult[] = await response.json();
        setLLMReviewDataList(data);
      } catch (error) {
        console.error('Error fetching LLM reviews:', error);
      }
    };

    fetchLLMReviews();
    const reviewsInterval = setInterval(fetchLLMReviews, 5000);

    return () => {
      clearInterval(reviewsInterval);
    };
  }, []);


  // toolChoiceModified is a helper function to check if the tool choice has been modified
  const toolChoiceModified = (allToolChoices: ToolChoice[], toolChoice: ToolChoice) => {
    const originalToolChoice = allToolChoices.find(t => t.id === toolChoice.id);
    const modified = originalToolChoice && originalToolChoice.arguments !== toolChoice.arguments;
    return modified;
  };

  // Send a response to the Approvals API with the decision and the tool choice
  const sendResponse = (decision: Decision, requestId: string, toolChoice: ToolChoice) => {
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

  // When the user clicks the title, go home
  const handleGoHome = () => {
    setSelectedSupervisor(null);
  };

  return (
    <div className="flex flex-col min-h-screen">
      <NavBar onHome={handleGoHome} isSocketConnected={isSocketConnected} />

      <main className="flex-grow">
        {selectedSupervisor === null ? (
          <SupervisorSelection
            onSelect={setSelectedSupervisor}
            API_BASE_URL={API_BASE_URL}
            WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
          />
        ) : selectedSupervisor === "LLMSupervisor" ? (
          <LLMReviews reviews={llmReviewDataList} />
        ) : (
          <>
            <ReviewSection
              humanReviewDataList={humanReviewDataList}
              llmReviewDataList={llmReviewDataList}
              selectedRequestId={selectedRequestId}
              selectReviewRequest={selectReviewRequest}
              selectedReviewRequest={selectedReviewRequest}
              sendResponse={sendResponse}
            />

            <div className="container mx-auto px-4 py-8 flex flex-col">
              <HubStatsAccordion hubStats={hubStats} />
            </div>
          </>
        )}
      </main>
    </div>
  );
};

export default ApprovalsInterface;
