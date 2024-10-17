import React from 'react';
import { UserIcon, BrainCircuitIcon } from 'lucide-react';

// Define a Supervisor type to include additional information
interface Supervisor {
  name: string;
  Icon: React.ComponentType<{ size: number }>;
  apiEndpoint: string;
  description: string;
}

// Array of supervisors with detailed information
const Supervisors: Supervisor[] = [
  {
    name: "HumanSupervisor",
    Icon: UserIcon as React.ComponentType<{ size: number }>,
    apiEndpoint: "/api/review/human",
    description: "Review agent actions using human judgment and expertise.",
  },
  {
    name: "LLMSupervisor",
    Icon: BrainCircuitIcon as React.ComponentType<{ size: number }>,
    apiEndpoint: "/api/review/llm",
    description: "Automatically review agent actions using LLMs.",
  },
  // Add more supervisors here if needed
];

interface SupervisorSelectionProps {
  onSelect: (supervisor: string) => void;
  API_BASE_URL: string;
  WEBSOCKET_BASE_URL: string;
}

const SupervisorSelection: React.FC<SupervisorSelectionProps> = ({ onSelect, API_BASE_URL, WEBSOCKET_BASE_URL }) => {

  if (!API_BASE_URL || !WEBSOCKET_BASE_URL) {
    return (
      <div className="text-center text-red-500">
        No API or WebSocket base URL set: API is: {API_BASE_URL || 'Not Set'} and WebSocket is: {WEBSOCKET_BASE_URL || 'Not Set'}
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 ">
      <h2 className="text-3xl font-semibold mb-6 text-center">Select a Supervisor</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {Supervisors.map((supervisor) => {
          const { name, Icon, apiEndpoint, description } = supervisor;
          return (
            <div
              key={name}
              className="border p-6 rounded-lg cursor-pointer hover:bg-gray-100 transition duration-300 flex flex-col items-center space-y-4"
              onClick={() => onSelect(name)}
            >
              <Icon size={32} />
              <h3 className="text-2xl font-semibold mb-2">{name}</h3>
              <p className="text-gray-600 mb-4 text-center">{description}</p>
              <p className="text-sm text-gray-500"><span className="font-mono">{API_BASE_URL}{apiEndpoint}</span></p>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default SupervisorSelection;
