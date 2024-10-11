import React, { useState } from 'react';
import { HelpCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ExplainButtonProps {
  text: string;
  onExplanation: (explanation: string) => void;
  onScore: (score: string) => void;
}

// @ts-ignore
const API_BASE_URL = import.meta.env.VITE_APPROVAL_API_BASE_URL || `http://localhost:8080`;

// Ask the language model for an explanation of the agents actions
export default function ExplainButton({ text, onExplanation, onScore }: ExplainButtonProps) {
  const [isLoading, setIsLoading] = useState(false);

  const handleExplain = async () => {
    setIsLoading(true);
    try {
      const completion = await fetch(`${API_BASE_URL}/api/explain`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ text }),
      });

      const explanation = await completion.json();

      onExplanation(explanation.explanation || 'No explanation provided.');
      onScore(explanation.score || 'No score provided.');
    } catch (error) {
      console.error('Error getting explanation:', error);
      onExplanation('Failed to get explanation. Please try again.');
      onScore('Failed to get score. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Button
      size="icon"
      onClick={handleExplain}
      className="ml-2"
    >
      {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <HelpCircle className="h-4 w-4" />}
    </Button>
  );
}
