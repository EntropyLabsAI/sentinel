import React from 'react';
import { Check, Copy } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';

interface UUIDDisplayProps {
  uuid: string | undefined;
  className?: string;
  href?: string;
  label?: string;
}

export const UUIDDisplay: React.FC<UUIDDisplayProps> = ({ uuid, className = '', href, label = '' }) => {
  if (!uuid) return null;

  const [copied, setCopied] = useState(false);

  const copyToClipboard = async (e: React.MouseEvent) => {
    e.preventDefault();
    await navigator.clipboard.writeText(uuid);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const content = (
    <>
      {label ? <span className="text-muted-foreground">{label}</span> : uuid.slice(0, 8)}
      {copied ? (
        <Check className="h-4 w-4 text-green-500" />
      ) : (<></>)}
    </>
  );

  return href ? (
    <Link to={href}>
      <code
        className={`inline-flex items-center gap-2 cursor-pointer hover:bg-gray-100 rounded px-1 ${className}`}
        title="Click to copy UUID, click link to navigate"
      >
        {content}
      </code>
    </Link>
  ) : (
    <code
      className={`inline-flex items-center gap-2 cursor-pointer hover:bg-gray-100 rounded px-1 ${className}`}
      onClick={copyToClipboard}
      title="Click to copy"
    >
      {content}
    </code>
  );
};
