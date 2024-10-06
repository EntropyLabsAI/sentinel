import React from 'react';

interface HubStatsProps {
  stats: {
    connected_clients: number;
    queued_reviews: number;
    stored_reviews: number;
    free_clients: number;
    busy_clients: number;
    assigned_reviews: { [key: string]: number };
    review_distribution: { [key: number]: number };
    completed_reviews: number;
  };
}

const HubStats: React.FC<HubStatsProps> = ({ stats }) => {
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

export default HubStats;
