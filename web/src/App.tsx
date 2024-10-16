import React, { useState, useEffect } from 'react';
import { ReviewRequest, ReviewResult, ToolChoice, Decision, Review } from '@/types';
import ReviewRequestDisplay from '@/components/review_request';
import HubStats from '@/components/hub_stats';
import { HubStats as HubStatsType } from '@/types';
import { UserIcon, BrainCircuitIcon } from 'lucide-react';
import LLMReviews from '@/components/llm_reviews';

// SupervisorNames is a list of names of the supervisors
const SupervisorNames = [
  "HumanSupervisor",
  "LLMSupervisor",
];

const SupervisorIcons = {
  HumanSupervisor: UserIcon,
  LLMSupervisor: BrainCircuitIcon,
};

// The API base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
// The websocket base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

const SupervisorSelection: React.FC<{ onSelect: (supervisor: string) => void }> = ({ onSelect }) => {
  if (!API_BASE_URL || !WEBSOCKET_BASE_URL) {
    return <div>No API or WebSocket base URL set: API is: {API_BASE_URL} and WebSocket is: {WEBSOCKET_BASE_URL}</div>;
  }
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {SupervisorNames.map((supervisor) => {
          const Icon = SupervisorIcons[supervisor as keyof typeof SupervisorIcons];
          return (
            <div
              key={supervisor}
              className="border p-4 rounded-lg cursor-pointer hover:bg-gray-100 flex flex-col items-center"
              onClick={() => onSelect(supervisor)}
            >
              <Icon size={24} className="mb-2" />
              <h2 className="text-xl text-center">{supervisor}</h2>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const NavBar: React.FC<{ onHome: () => void; isSocketConnected: boolean }> = ({ onHome, isSocketConnected }) => {
  return (
    <nav className="bg-gray-800 text-white p-4">
      <div className="container mx-auto flex justify-between items-center">
        <h1
          className="text-xl font-bold cursor-pointer hover:text-gray-300"
          onClick={onHome}
        >
          Sentinel
        </h1>
        <div className="text-sm flex items-center space-x-4">
          <div>
            <p>API: {API_BASE_URL}</p>
          </div>
          <div className="flex items-center">
            <p>WebSocket: {WEBSOCKET_BASE_URL}</p>
            <span
              className={`ml-2 h-3 w-3 rounded-full ${isSocketConnected ? 'bg-green-500' : 'bg-red-500'
                }`}
            ></span>
          </div>
        </div>
      </div>
    </nav>
  );
};

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
          <SupervisorSelection onSelect={setSelectedSupervisor} />
        ) : selectedSupervisor === "LLMsupervisor" ? (
          <LLMReviews reviews={llmReviewDataList} />
        ) : (
          <>
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

            <div className="container mx-auto px-4 py-8 flex flex-col">
              {/* Hub Stats */}
              <div className="mb-8">
                <h2 className="text-2xl font-semibold mb-4">Hub Statistics</h2>
                {hubStats ? (
                  <HubStats stats={hubStats} />
                ) : (
                  <p>Loading hub statistics...</p>
                )}
              </div>
            </div>
          </>
        )
        }
      </main >
    </div >
  );
};

export default ApprovalsInterface;
