import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import SwaggerUI from '@/components/util/swagger_ui';
import SupervisorSelection from '@/components/supervisor/supervisors';
import Sidebar from '@/components/sidebar';
import Home from '@/components/home';
import ProjectList from '@/components/projects';
import Runs from '@/components/runs';
import Executions from '@/components/run';
import Tools from '@/components/tools';
import ToolDetails from '@/components/tool';
import SupervisorDetails from '@/components/supervisor/supervisor';
import Tasks from '@/components/tasks';
import { Toaster } from '@/components/ui/toaster';

const App: React.FC = () => {
  return (
    <main className="relative flex min-h-svh flex-1 flex-col">
      <div className="flex min-w-0 flex-1 flex-col overflow-x-hidden">
        <Router>
          <Sidebar>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/api" element={<SwaggerUI />} />

              <Route path="/projects" element={<ProjectList />} />
              <Route path="/projects/:projectId" element={<Tasks />} />
              <Route path="/tasks/:taskId" element={<Runs />} />
              <Route path="/tasks/:taskId/runs/:runId" element={<Executions />} />
              <Route path="/runs/:runId" element={<Executions />} />

              <Route path="/tools" element={<Tools />} />
              <Route path="/tools/:toolId" element={<ToolDetails />} />
              <Route path="/supervisors" element={<SupervisorSelection />} />
              <Route path="/supervisors/:supervisorId" element={<SupervisorDetails />} />
            </Routes>
          </Sidebar>
        </Router>
      </div>
      <Toaster />
    </main>
  );
};

export default App;
