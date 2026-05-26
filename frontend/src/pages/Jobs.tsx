import { useState } from 'react';
import { RefreshCw } from 'lucide-react';
import JobTable from '../components/JobTable';
import { useJobs } from '../hooks/useJobs';
import type { JobStatus } from '../types';

const STATUSES: { label: string; value: JobStatus | '' }[] = [
  { label: 'All',       value: '' },
  { label: 'Pending',   value: 'pending' },
  { label: 'Running',   value: 'running' },
  { label: 'Completed', value: 'completed' },
  { label: 'Failed',    value: 'failed' },
  { label: 'Retrying',  value: 'retrying' },
  { label: 'Dead',      value: 'dead' },
];

export default function Jobs() {
  const [status, setStatus] = useState<JobStatus | ''>('');
  const { jobs, total, loading, refresh, retry, cancel } = useJobs(status || undefined);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }} className="fade-in">
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>Jobs</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>
            {total.toLocaleString()} total jobs
          </p>
        </div>
        <button
          onClick={refresh}
          style={{
            background: 'var(--bg-elevated)', border: '1px solid var(--border)',
            color: 'var(--text-secondary)', borderRadius: 8, padding: '8px 14px',
            cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: 13,
          }}
        >
          <RefreshCw size={13} />
          Refresh
        </button>
      </div>

      {/* Filter tabs */}
      <div style={{ display: 'flex', gap: 4, borderBottom: '1px solid var(--border)', paddingBottom: 0 }}>
        {STATUSES.map(({ label, value }) => (
          <button
            key={value}
            onClick={() => setStatus(value as JobStatus | '')}
            style={{
              background: 'none', border: 'none', cursor: 'pointer',
              padding: '8px 14px', fontSize: 13, fontWeight: status === value ? 600 : 400,
              color: status === value ? 'var(--amber)' : 'var(--text-secondary)',
              borderBottom: status === value ? '2px solid var(--amber)' : '2px solid transparent',
              transition: 'all 0.15s', marginBottom: -1,
            }}
          >
            {label}
          </button>
        ))}
      </div>

      <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 12, padding: '4px 0', overflow: 'hidden' }}>
        <JobTable jobs={jobs} loading={loading} onRetry={retry} onCancel={cancel} />
      </div>
    </div>
  );
}
