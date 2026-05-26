import type { Job, JobStatus } from '../types';
import { formatDistanceToNow } from 'date-fns';
import { RotateCcw, X } from 'lucide-react';

const PRIORITY_LABELS: Record<number, string> = { 4: 'CRITICAL', 3: 'HIGH', 2: 'NORMAL', 1: 'LOW' };
const PRIORITY_COLORS: Record<number, string> = {
  4: '#ef4444', 3: '#f97316', 2: '#f59e0b', 1: '#6b7280',
};

interface Props {
  jobs: Job[];
  loading: boolean;
  onRetry?: (id: string) => void;
  onCancel?: (id: string) => void;
  showActions?: boolean;
}

function StatusBadge({ status }: { status: JobStatus }) {
  return <span className={`badge badge-${status}`}>{status}</span>;
}

export default function JobTable({ jobs, loading, onRetry, onCancel, showActions = true }: Props) {
  if (loading) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {[...Array(5)].map((_, i) => (
          <div key={i} className="skeleton" style={{ height: 52, borderRadius: 8, opacity: 1 - i * 0.12 }} />
        ))}
      </div>
    );
  }

  if (!jobs?.length) {
    return (
      <div style={{
        padding: '60px 0', textAlign: 'center',
        color: 'var(--text-muted)', fontSize: 13,
        border: '1px dashed var(--border)', borderRadius: 10,
      }}>
        No jobs found
      </div>
    );
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
        <thead>
          <tr style={{ borderBottom: '1px solid var(--border)' }}>
            {['Job ID', 'Type', 'Priority', 'Status', 'Attempts', 'Queue', 'Created', ...(showActions ? [''] : [])].map(h => (
              <th key={h} style={{
                padding: '10px 12px', textAlign: 'left',
                color: 'var(--text-muted)', fontWeight: 500,
                fontSize: 11, letterSpacing: '0.04em', textTransform: 'uppercase',
                whiteSpace: 'nowrap',
              }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {jobs.map((job, i) => (
            <tr
              key={job.id}
              className="fade-in"
              style={{
                borderBottom: '1px solid var(--border)',
                animationDelay: `${i * 0.03}s`,
                transition: 'background 0.15s',
              }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--bg-elevated)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
            >
              <td style={{ padding: '12px 12px' }}>
                <span style={{ fontFamily: 'monospace', fontSize: 12, color: 'var(--amber)' }}>
                  {job.id.slice(0, 8)}…
                </span>
              </td>
              <td style={{ padding: '12px 12px' }}>
                <span style={{
                  background: 'var(--bg-elevated)', border: '1px solid var(--border)',
                  padding: '2px 8px', borderRadius: 5,
                  fontSize: 12, color: 'var(--text-primary)', fontFamily: 'monospace',
                }}>{job.type}</span>
              </td>
              <td style={{ padding: '12px 12px' }}>
                <span style={{ color: PRIORITY_COLORS[job.priority] ?? 'var(--text-secondary)', fontWeight: 600, fontSize: 11 }}>
                  {PRIORITY_LABELS[job.priority] ?? 'NORMAL'}
                </span>
              </td>
              <td style={{ padding: '12px 12px' }}>
                <StatusBadge status={job.status} />
              </td>
              <td style={{ padding: '12px 12px', color: 'var(--text-secondary)' }}>
                {job.attempts}/{job.max_retries}
              </td>
              <td style={{ padding: '12px 12px', color: 'var(--text-secondary)' }}>
                {job.queue}
              </td>
              <td style={{ padding: '12px 12px', color: 'var(--text-muted)', whiteSpace: 'nowrap' }}>
                {formatDistanceToNow(new Date(job.created_at), { addSuffix: true })}
              </td>
              {showActions && (
                <td style={{ padding: '12px 12px' }}>
                  <div style={{ display: 'flex', gap: 6 }}>
                    {(job.status === 'failed' || job.status === 'dead') && onRetry && (
                      <button
                        onClick={() => onRetry(job.id)}
                        style={{
                          background: 'rgba(245,158,11,0.1)', border: '1px solid rgba(245,158,11,0.3)',
                          color: 'var(--amber)', borderRadius: 6, padding: '4px 8px',
                          cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 500,
                        }}
                      >
                        <RotateCcw size={11} /> Retry
                      </button>
                    )}
                    {(job.status === 'pending' || job.status === 'retrying') && onCancel && (
                      <button
                        onClick={() => onCancel(job.id)}
                        style={{
                          background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.25)',
                          color: '#f87171', borderRadius: 6, padding: '4px 8px',
                          cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 500,
                        }}
                      >
                        <X size={11} /> Cancel
                      </button>
                    )}
                  </div>
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
