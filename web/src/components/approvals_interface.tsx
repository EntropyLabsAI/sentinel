import React, { useState, useEffect } from 'react';
import { ReviewResult } from '@/types';
import ReviewSection from '@/components/human_reviews';
import { HubStats as HubStatsType } from '@/types';
import LLMReviews from '@/components/llm_reviews';
import NavBar from './nav';
import SupervisorSelection from './supervisor_selection';
import { HubStatsAccordion } from './hub_stats';
import HumanReviews from '@/components/human_reviews';

// The API base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
// The websocket base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

const ApprovalsInterface: React.FC = () => {
  const [llmReviewDataList, setLLMReviewDataList] = useState<ReviewResult[]>([]);
  const [hubStats, setHubStats] = useState<HubStatsType | null>(null);
  const [selectedSupervisor, setSelectedSupervisor] = useState<string | null>(null);
  const [isSocketConnected, setIsSocketConnected] = useState<boolean>(false);

  // Fetch LLM reviews every 5 seconds
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
            <HumanReviews
              API_BASE_URL={API_BASE_URL}
              WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
              setIsSocketConnected={setIsSocketConnected}
            />
          </>
        )}
      </main>
    </div>
  );
};

export default ApprovalsInterface;
