import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import ApprovalsInterface from '@/components/approvals_interface';
import SwaggerUI from 'swagger-ui-react';
// import SwaggerUI from '@/components/swagger_ui';

const App: React.FC = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<ApprovalsInterface />} />
        <Route path="/api/docs" element={<SwaggerUI />} />
      </Routes>
    </Router>
  );
};

export default App;
