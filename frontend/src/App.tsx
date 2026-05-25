import { useState, useCallback } from 'react';
import Sidebar from './components/Sidebar';
import Overview from './pages/Overview';
import Jobs from './pages/Jobs';
import Queues from './pages/Queues';
import Workers from './pages/Workers';
import DLQ from './pages/DLQ';
import Enqueue from './pages/Enqueue';
import { useMetrics } from './hooks/useMetrics';
import { useWebSocket } from './hooks/useWebSocket';
import { WSEvent } from './types';
import './index.css';

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

export default function App() {
  const [page, setPage] = useState('overview');
  const { metrics, throughput, loading } = useMetrics();

  const handleWSEvent = useCallback((event: WSEvent) => {
    // Real-time events arrive here; pages re-fetch on their own interval
    // but we could update local state here for zero-latency updates
    console.debug('[ws]', event.type);
  }, []);

  useWebSocket(WS_URL, handleWSEvent);

  const deadCount = metrics.total_dead;

  const renderPage = () => {
    switch (page) {
      case 'overview': return <Overview metrics={metrics} throughput={throughput} loading={loading} />;
      case 'jobs':     return <Jobs />;
      case 'queues':   return <Queues />;
      case 'workers':  return <Workers metrics={metrics} loading={loading} />;
      case 'dlq':      return <DLQ />;
      case 'enqueue':  return <Enqueue />;
      default:         return <Overview metrics={metrics} throughput={throughput} loading={loading} />;
    }
  };

  return (
    <div style={{ display: 'flex', minHeight: '100vh', background: 'var(--bg-base)' }}>
      <Sidebar current={page} onChange={setPage} deadCount={deadCount} />
      <main style={{
        flex: 1, padding: '32px 36px',
        maxWidth: 'calc(100vw - 220px)',
        overflowY: 'auto',
      }}>
        {renderPage()}
      </main>
    </div>
  );
}
