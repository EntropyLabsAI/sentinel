import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SwaggerUI from '@/components/swagger_ui';
import SupervisorSelection from '@/components/supervisors';
import Sidebar from './components/sidebar';
import Home from './components/home';
import ProjectList from './components/project_list';
import Runs from './components/runs';
import Executions from './components/run';
import Tools from './components/tools';
import ToolDetails from './components/tool';
import SupervisorDetails from './components/supervisor_details';
import Execution from './components/execution';

const App: React.FC = () => {
  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-grow">
        <Router>
          <Sidebar >
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/api" element={<SwaggerUI />} />
              <Route path="/projects" element={<ProjectList />} />
              <Route path="/projects/:projectId" element={<Runs />} />
              <Route path="/projects/:projectId/runs/:runId" element={<Executions />} />
              <Route path="/projects/:projectId/runs/:runId/executions/:executionId" element={<Execution />} />
              <Route path="/tools" element={<Tools />} />
              <Route path="/tools/:toolId" element={<ToolDetails />} />
              <Route path="/supervisors" element={<SupervisorSelection />} />
              <Route path="/supervisors/:supervisorId" element={<SupervisorDetails />} />
            </Routes>
          </Sidebar>
        </Router>
      </main>
    </div>
  );
};

export default App;
