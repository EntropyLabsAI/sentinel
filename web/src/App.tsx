import React, { useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SwaggerUI from '@/components/swagger_ui';
import SupervisorSelection from '@/components/supervisor_selection';
import NavBar from '@/components/nav';
import HumanReviews from '@/components/human_reviews';
import LLMReviews from '@/components/llm_reviews';

// The API base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
// The websocket base URL is set via an environment variable in the docker-compose.yml file
// @ts-ignore
const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

const App: React.FC = () => {
  const [isSocketConnected, setIsSocketConnected] = useState<boolean>(false);

  return (
    <div className="flex flex-col min-h-screen">
      <NavBar isSocketConnected={isSocketConnected} />

      <main className="flex-grow">
        <Router>
          <Routes>
            <Route path="/" element={
              <SupervisorSelection
                API_BASE_URL={API_BASE_URL}
                WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
              />
            } />
            <Route path="/api/docs" element={<SwaggerUI />} />
            <Route path="/supervisor/human" element={<HumanReviews
              API_BASE_URL={API_BASE_URL}
              WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
              setIsSocketConnected={setIsSocketConnected}
            />} />
            <Route path="/supervisor/llm" element={<LLMReviews
              API_BASE_URL={API_BASE_URL}
            />} />

          </Routes>
        </Router>
      </main>
    </div>
  );
};

export default App;
