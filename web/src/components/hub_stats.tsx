import React, { useEffect, useState } from 'react';

import { HubStats as HubStatsType } from '@/types';
import * as Accordion from '@radix-ui/react-accordion';
import { ChevronDownIcon } from 'lucide-react'; // Using Lucide Icons for the Chevron

const HubStatsAccordion: React.FC<{ API_BASE_URL: string }> = ({ API_BASE_URL }) => {
  const [hubStats, setHubStats] = useState<HubStatsType | null>(null);
  // Fetch hub stats every second
  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/hub/stats`);
        const data: HubStatsType = await response.json();
        setHubStats(data);
      } catch (error) {
        console.error('Error fetching hub stats:', error);
      }
    };

    fetchStats();
    const statsInterval = setInterval(fetchStats, 1000);

    return () => {
      clearInterval(statsInterval);
    };
  }, []);
  return (
    <Accordion.Root type="single" collapsible className="w-full">
      <Accordion.Item value="hub-stats" className="border border-gray-200 rounded-md mb-4">
        <Accordion.Header>
          <Accordion.Trigger className="flex justify-between items-center w-full p-4 rounded-md cursor-pointer focus:outline-none">
            <span className="text-sm font-mono font-semibold text-gray-400">Websocket Hub Statistics</span>
            <ChevronDownIcon className="h-5 w-5 transition-transform duration-200" />
          </Accordion.Trigger>
        </Accordion.Header>
        <Accordion.Content className="p-4 bg-white rounded-md">
          {hubStats ? (
            <HubStats stats={hubStats} />
          ) : (
            <p>Loading hub statistics...</p>
          )}
        </Accordion.Content>
      </Accordion.Item>
    </Accordion.Root>
  );
};

export { HubStatsAccordion };

const HubStats: React.FC<{ stats: HubStatsType }> = ({ stats }) => {
  return (
    <div className="bg-gray-100 p-4 rounded-lg">
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-4">
        <StatItem label="Reviews waiting to be assigned (server-side)" value={stats.queued_reviews} />
        <StatItem label="Reviews in progress (client-side)" value={stats.stored_reviews} />
        <StatItem label="Completed Reviews" value={stats.completed_reviews} />
      </div>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <StatItem label="Connected Clients" value={stats.connected_clients} />
        <StatItem label="Free Clients" value={stats.free_clients} />
        <StatItem label="Busy Clients" value={stats.busy_clients} />
      </div>
      <div className="mt-4">
        <h3 className="font-semibold mb-2">Assigned Reviews</h3>
        <ul className="list-disc list-inside">
          {Object.entries(stats.assigned_reviews).map(([client, count]) => (
            <li key={client}>Client {client.slice(-6)}: {count}</li>
          ))}
        </ul>
      </div>
      <div className="mt-4">
        <h3 className="font-semibold mb-2">Review Distribution</h3>
        <ul className="list-disc list-inside">
          {Object.entries(stats.review_distribution).map(([reviewCount, clientCount]) => (
            <li key={reviewCount}>{clientCount} client(s) with {reviewCount} review(s)</li>
          ))}
        </ul>
      </div>
    </div>
  );
};

const StatItem: React.FC<{ label: string; value: number }> = ({ label, value }) => (
  <div className="bg-white p-3 rounded shadow">
    <div className="text-sm text-gray-600">{label}</div>
    <div className="text-xl font-semibold">{value}</div>
  </div>
);
