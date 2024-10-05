import React, { useState } from 'react';
import { Copy, HelpCircle, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ExplainButtonProps {
  text: string;
  onExplanation: (explanation: string) => void;
}

export default function ExplainButton({ text, onExplanation }: ExplainButtonProps) {
  const [isLoading, setIsLoading] = useState(false);


  const handleExplain = async () => {
    setIsLoading(true);
    try {
      const completion = await fetch("http://localhost:8080/api/explain", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ text }),
      });

      const explanation = await completion.json();
      console.log(explanation);

      onExplanation(explanation.explanation || 'No explanation provided.');
    } catch (error) {
      console.error('Error getting explanation:', error);
      onExplanation('Failed to get explanation. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    // <Button onClick={handleExplain} variant="ghost" size="sm" className="ml-2" disabled={isLoading}>
    //   <HelpCircle className="mr-2" size={18} />
    // </Button>
    <Button
      size="icon"
      onClick={handleExplain}
      className="ml-2 bg-gray-700 hover:bg-gray-600 outline-none"
    >
      {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <HelpCircle className="h-4 w-4" />}
    </Button>
  );
}
