import React from 'react'
import ReactDOM from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import App from './App'
import './index.css'
import axios from 'axios';
import { ProjectProvider } from '@/contexts/project_context';
import { ConfigProvider } from './contexts/config_context';

axios.defaults.baseURL = 'http://localhost:8099';

const queryClient = new QueryClient();

const rootElement = document.getElementById('root');
if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <React.StrictMode>
      <QueryClientProvider client={queryClient}>
        <ConfigProvider>
          <ProjectProvider>
            <App />
          </ProjectProvider>
        </ConfigProvider>
      </QueryClientProvider>
    </React.StrictMode>,
  );
}
