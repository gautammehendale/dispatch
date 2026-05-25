import { WorkerStatus } from '../types';
import { formatDistanceToNow } from 'date-fns';
import { Cpu } from 'lucide-react';

interface Props { workers: WorkerStatus[]; }

function WorkerCard({ w }: { w: WorkerStatus }) {
  const isBusy = w.status === 'busy';
  const isIdle = w.status === 'idle';
  const accentColor = isBusy ? '#f59e0b' : isIdle ? '#22c55e' : '#525252';

  return (
    <div style={{
      background: 'var(--bg-elevated)', border: `1px solid ${isBusy ? '#3d2e00' : 'var(--border)'}`,
      borderRadius: 10, padding: '14px 16px',
      transition: 'border-color 0.2s, background 0.2s',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <div style={{
            width: 28, height: 28, borderRadius: 7,
            background: `${accentColor}18`, border: `1px solid ${accentColor}30`,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <Cpu size={13} color={accentColor} />
          </div>
          <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-secondary)', fontFamily: 'monospace' }}>
            {w.id.split('-').pop()}
          </span>
        </div>
        <span className={`badge badge-${w.status === 'busy' ? 'running' : w.status === 'idle' ? 'completed' : 'dead'}`}>
          {w.status}
        </span>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 5 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>Jobs run</span>
          <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--text-primary)' }}>{w.jobs_run.toLocaleString()}</span>
        </div>
        {w.current_job && (
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>Current</span>
            <span style={{ fontSize: 10, color: 'var(--amber)', fontFamily: 'monospace', maxWidth: 100, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {w.current_job.slice(0, 8)}…
            </span>
          </div>
        )}
        <div style={{ display: 'flex', justifyContent: 'space-between' }}>
          <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>Started</span>
          <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>
            {formatDistanceToNow(new Date(w.started_at), { addSuffix: true })}
          </span>
        </div>
      </div>
    </div>
  );
}

export default function WorkerGrid({ workers }: Props) {
  if (!workers?.length) {
    return (
      <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-muted)', fontSize: 13 }}>
        No workers registered
      </div>
    );
  }
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 10 }}>
      {workers.map(w => <WorkerCard key={w.id} w={w} />)}
    </div>
  );
}
