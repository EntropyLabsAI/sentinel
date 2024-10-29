import React, { createContext, useContext, useState, useEffect } from 'react';

interface ProjectContextType {
  selectedProject: string | null;
  setSelectedProject: (projectId: string | null) => void;
}

const ProjectContext = createContext<ProjectContextType | undefined>(undefined);

export function ProjectProvider({ children }: { children: React.ReactNode }) {
  const [selectedProject, setSelectedProject] = useState<string | null>(() => {
    // Initialize from localStorage
    return localStorage.getItem('selectedProject');
  });

  useEffect(() => {
    // Persist to localStorage whenever it changes
    if (selectedProject) {
      localStorage.setItem('selectedProject', selectedProject);
    } else {
      localStorage.removeItem('selectedProject');
    }
  }, [selectedProject]);

  return (
    <ProjectContext.Provider value={{ selectedProject, setSelectedProject }}>
      {children}
    </ProjectContext.Provider>
  );
}

export function useProject() {
  const context = useContext(ProjectContext);
  if (context === undefined) {
    throw new Error('useProject must be used within a ProjectProvider');
  }
  return context;
}
