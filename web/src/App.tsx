import React, { useState } from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SwaggerUI from '@/components/swagger_ui';
import SupervisorSelection from '@/components/supervisor_selection';
import HumanReviews from '@/components/human_reviews';
import LLMReviews from '@/components/llm_reviews';
import Sidebar from './components/sidebar';
import Home from './components/home';
import ProjectList from './components/project_list';
import Runs from './components/runs';
import Executions from './components/run';
import Tools from './components/tools';
import ToolDetails from './components/tool';

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
      <main className="flex-grow">
        <Router>
          <Sidebar isSocketConnected={isSocketConnected}>
            <Routes>
              <Route path="/" element={
                <Home />
              } />
              <Route path="/projects" element={
                <ProjectList />
              } />
              <Route path="/supervisor" element={
                <SupervisorSelection API_BASE_URL={API_BASE_URL}
                  WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
                />
              } />
              <Route path="/api" element={
                <SwaggerUI />
              } />
              <Route path="/supervisor/human" element={
                <HumanReviews
                  API_BASE_URL={API_BASE_URL}
                  WEBSOCKET_BASE_URL={WEBSOCKET_BASE_URL}
                  setIsSocketConnected={setIsSocketConnected}
                />} />
              <Route path="/supervisor/llm" element={
                <LLMReviews
                  API_BASE_URL={API_BASE_URL}
                />} />
              <Route path="/projects/:projectId" element={<Runs />} />
              <Route path="/projects/:projectId/runs/:runId" element={<Executions />} />
              <Route path="/tools" element={<Tools />} />
              <Route path="/tools/:toolId" element={<ToolDetails />} />
            </Routes>
          </Sidebar>
        </Router>
      </main>
    </div>
  );
};

export default App;
