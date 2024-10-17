import React from 'react';
import { UserIcon, BrainCircuitIcon } from 'lucide-react';
import { Badge } from './ui/badge';

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
    description: "Review agent actions with a human operator",
  },
  {
    name: "LLMSupervisor",
    Icon: BrainCircuitIcon as React.ComponentType<{ size: number }>,
    apiEndpoint: "/api/review/llm",
    description: "Review agent actions using an LLM",
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
    <div className="container mx-auto px-4 py-24">
      {/* Introductory Section */}
      <div className="mb-12 space-y-12">
        <h1 className="text-2xl font-bold mb-4">Select a Supervisor</h1>
        <p className="text-lg text-gray-700">
          Supervisors are used to review agent actions. To get started, ensure that your agent is running and making requests to the Sentinel API when it wants to take an action. Requests will be paused until a supervisor approves the action.
        </p>
        <p>Supervisors will then return one of the following responses to your agent:
          <div className="grid grid-cols-[auto,1fr] gap-x-4 gap-y-2 mt-2">
            <Badge variant="outline" className="whitespace-nowrap">APPROVE</Badge>
            <div>The agent can proceed</div>

            <Badge variant="outline" className="whitespace-nowrap">REJECT</Badge>
            <div>The agent action is blocked and the agent should try again</div>

            <Badge variant="outline" className="whitespace-nowrap">ESCALATE</Badge>
            <div>The action should be escalated to the next supervisor if one is configured</div>

            <Badge variant="outline" className="whitespace-nowrap">TERMINATE</Badge>
            <div>The agent process should be killed</div>
          </div>
        </p>
        <p className="mt-4 text-sm text-gray-500">
          Select a supervisor below to get started.
        </p>
      </div>

      {/* Supervisor Selection Grid */}
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
