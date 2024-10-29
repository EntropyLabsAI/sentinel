import React from 'react';
import { Check, Copy } from 'lucide-react';
import { useState } from 'react';

interface UUIDDisplayProps {
  uuid: string | undefined;
  className?: string;
}

export const UUIDDisplay: React.FC<UUIDDisplayProps> = ({ uuid, className = '' }) => {
  if (!uuid) return null;

  const [copied, setCopied] = useState(false);

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(uuid);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000); // Reset after 2 seconds
  };

  return (
    <code
      className={`font-mono inline-flex items-center gap-2 cursor-pointer hover:bg-gray-100 rounded px-1 ${className}`}
      onClick={copyToClipboard}
      title="Click to copy"
    >
      {uuid.slice(0, 8)}
      {copied ? (
        <Check className="h-4 w-4 text-green-500" />
      ) : (<></>)}
    </code>
  );
};
