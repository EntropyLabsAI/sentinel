import React from 'react';

// Access environment variables
// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
// @ts-ignore
const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

interface NavBarProps {
  onHome: () => void;
  isSocketConnected: boolean;
}

const NavBar: React.FC<NavBarProps> = ({ onHome, isSocketConnected }) => {
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

export default NavBar;
