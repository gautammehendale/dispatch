import { useEffect, useState } from 'react';
import axios from 'axios';
import { Job } from '../types';
import JobTable from '../components/JobTable';
import { AlertTriangle, RefreshCw } from 'lucide-react';

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default function DLQ() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = async () => {
    setLoading(true);
    try {
      const { data } = await axios.get(`${API}/api/v1/dlq`);
      setJobs(data.jobs ?? []);
    } catch { /* swallow */ }
    finally { setLoading(false); }
  };

  const retry = async (id: string) => {
    await axios.post(`${API}/api/v1/jobs/${id}/retry`);
    fetch();
  };

  useEffect(() => { fetch(); }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }} className="fade-in">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <AlertTriangle size={18} color="#f97316" />
            <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>
              Dead Letter Queue
            </h1>
          </div>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>
            {jobs.length} jobs exhausted all retry attempts
          </p>
        </div>
        <button
          onClick={fetch}
          style={{
            background: 'var(--bg-elevated)', border: '1px solid var(--border)',
            color: 'var(--text-secondary)', borderRadius: 8, padding: '8px 14px',
            cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: 13,
          }}
        >
          <RefreshCw size={13} /> Refresh
        </button>
      </div>

      {!loading && jobs.length > 0 && (
        <div style={{
          background: 'rgba(249,115,22,0.06)', border: '1px solid rgba(249,115,22,0.2)',
          borderRadius: 8, padding: '12px 16px', fontSize: 13, color: '#fb923c',
          display: 'flex', alignItems: 'center', gap: 8,
        }}>
          <AlertTriangle size={14} />
          These jobs failed all retry attempts. Investigate errors before retrying.
        </div>
      )}

      <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 12, padding: '4px 0', overflow: 'hidden' }}>
        <JobTable jobs={jobs} loading={loading} onRetry={retry} showActions />
      </div>
    </div>
  );
}
