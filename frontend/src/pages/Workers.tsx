import type { Metrics } from '../types';
import WorkerGrid from '../components/WorkerGrid';
import { Users } from 'lucide-react';

interface Props { metrics: Metrics; loading: boolean; }

export default function Workers({ metrics, loading }: Props) {
  const workers = metrics.workers ?? [];
  const busy = workers.filter(w => w.status === 'busy').length;
  const idle = workers.filter(w => w.status === 'idle').length;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }} className="fade-in">
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <Users size={18} color="var(--amber)" />
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>Workers</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>
            {busy} busy · {idle} idle · {workers.length} total
          </p>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
        {[
          { label: 'Total Workers', value: workers.length, color: 'var(--amber)' },
          { label: 'Busy',          value: busy,           color: '#3b82f6' },
          { label: 'Idle',          value: idle,           color: '#22c55e' },
        ].map(({ label, value, color }) => (
          <div key={label} style={{
            background: 'var(--bg-surface)', border: '1px solid var(--border)',
            borderRadius: 10, padding: '18px 20px',
          }}>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.04em', fontWeight: 500 }}>{label}</div>
            <div style={{ fontSize: 30, fontWeight: 700, color, marginTop: 6 }}>{value}</div>
          </div>
        ))}
      </div>

      <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 12, padding: '20px 22px' }}>
        {loading
          ? <div style={{ color: 'var(--text-muted)', fontSize: 13 }}>Loading workers…</div>
          : <WorkerGrid workers={workers} />
        }
      </div>
    </div>
  );
}
