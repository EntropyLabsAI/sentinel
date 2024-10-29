import React, { createContext, useContext } from 'react';

interface ConfigContextType {
  API_BASE_URL: string;
  WEBSOCKET_BASE_URL: string;
}

const ConfigContext = createContext<ConfigContextType | undefined>(undefined);

export const ConfigProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  // @ts-ignore
  const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;
  // @ts-ignore
  const WEBSOCKET_BASE_URL = import.meta.env.VITE_WEBSOCKET_BASE_URL;

  const value = {
    API_BASE_URL,
    WEBSOCKET_BASE_URL,
  };

  return (
    <ConfigContext.Provider value={value}>
      {children}
    </ConfigContext.Provider>
  );
};

export const useConfig = () => {
  const context = useContext(ConfigContext);
  if (context === undefined) {
    throw new Error('useConfig must be used within a ConfigProvider');
  }
  return context;
};
