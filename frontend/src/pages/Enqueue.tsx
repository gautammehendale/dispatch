import { useState } from 'react';
import { useEnqueue } from '../hooks/useJobs';
import { CheckCircle2, Send } from 'lucide-react';
import { Job } from '../types';

const JOB_TYPES = ['send_email', 'process_payment', 'resize_image', 'generate_report', 'sync_data'];
const PRIORITIES = ['CRITICAL', 'HIGH', 'NORMAL', 'LOW'];
const QUEUES = ['default', 'email', 'notifications'];

const DEMO_PAYLOADS: Record<string, Record<string, unknown>> = {
  send_email:       { to: 'user@example.com', subject: 'Welcome to Dispatch!', template: 'welcome' },
  process_payment:  { amount: 99.99, currency: 'USD', customer_id: 'cust_123' },
  resize_image:     { url: 'https://example.com/image.jpg', width: 800, height: 600 },
  generate_report:  { report_type: 'monthly', format: 'pdf', period: '2025-01' },
  sync_data:        { source: 'postgres', target: 's3', table: 'users' },
};

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <label style={{ fontSize: 12, fontWeight: 500, color: 'var(--text-secondary)', letterSpacing: '0.03em' }}>
        {label}
      </label>
      {children}
    </div>
  );
}

const inputStyle: React.CSSProperties = {
  background: 'var(--bg-elevated)', border: '1px solid var(--border)',
  borderRadius: 8, padding: '9px 12px', color: 'var(--text-primary)',
  fontSize: 13, outline: 'none', width: '100%',
  transition: 'border-color 0.15s',
};

export default function Enqueue() {
  const { enqueue } = useEnqueue();
  const [form, setForm] = useState({
    type: 'send_email', priority: 'NORMAL', queue: 'default',
    max_retries: 3,
    payload: JSON.stringify(DEMO_PAYLOADS.send_email, null, 2),
  });
  const [result, setResult] = useState<Job | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setResult(null);
    setLoading(true);
    try {
      let parsed: Record<string, unknown> = {};
      try { parsed = JSON.parse(form.payload); } catch { setError('Payload must be valid JSON'); setLoading(false); return; }
      const job = await enqueue({ ...form, payload: parsed });
      setResult(job);
    } catch (err: any) {
      setError(err?.response?.data?.error ?? 'Failed to enqueue job');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24, maxWidth: 680 }} className="fade-in">
      <div>
        <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>Enqueue Job</h1>
        <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>Submit a new job to the queue</p>
      </div>

      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16,
        background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 12, padding: 24 }}>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
          <Field label="Job Type">
            <select
              value={form.type}
              onChange={e => setForm(f => ({ ...f, type: e.target.value, payload: JSON.stringify(DEMO_PAYLOADS[e.target.value] ?? {}, null, 2) }))}
              style={{ ...inputStyle, cursor: 'pointer' }}
            >
              {JOB_TYPES.map(t => <option key={t} value={t}>{t}</option>)}
            </select>
          </Field>

          <Field label="Priority">
            <select value={form.priority} onChange={e => setForm(f => ({ ...f, priority: e.target.value }))} style={{ ...inputStyle, cursor: 'pointer' }}>
              {PRIORITIES.map(p => <option key={p} value={p}>{p}</option>)}
            </select>
          </Field>

          <Field label="Queue">
            <select value={form.queue} onChange={e => setForm(f => ({ ...f, queue: e.target.value }))} style={{ ...inputStyle, cursor: 'pointer' }}>
              {QUEUES.map(q => <option key={q} value={q}>{q}</option>)}
            </select>
          </Field>

          <Field label="Max Retries">
            <input
              type="number" min={0} max={10} value={form.max_retries}
              onChange={e => setForm(f => ({ ...f, max_retries: parseInt(e.target.value) || 0 }))}
              style={inputStyle}
            />
          </Field>
        </div>

        <Field label="Payload (JSON)">
          <textarea
            value={form.payload}
            onChange={e => setForm(f => ({ ...f, payload: e.target.value }))}
            rows={8}
            style={{ ...inputStyle, fontFamily: 'monospace', fontSize: 12, resize: 'vertical', lineHeight: 1.6 }}
          />
        </Field>

        {error && (
          <div style={{ background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.25)', borderRadius: 8, padding: '10px 14px', fontSize: 13, color: '#f87171' }}>
            {error}
          </div>
        )}

        <button
          type="submit" disabled={loading}
          style={{
            background: loading ? 'var(--bg-elevated)' : 'linear-gradient(135deg, #f59e0b, #f97316)',
            color: loading ? 'var(--text-muted)' : '#000',
            border: 'none', borderRadius: 8, padding: '11px 20px',
            fontWeight: 700, fontSize: 14, cursor: loading ? 'not-allowed' : 'pointer',
            display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8,
            transition: 'opacity 0.15s',
          }}
        >
          <Send size={15} />
          {loading ? 'Enqueueing…' : 'Enqueue Job'}
        </button>
      </form>

      {result && (
        <div style={{
          background: 'rgba(34,197,94,0.06)', border: '1px solid rgba(34,197,94,0.2)',
          borderRadius: 12, padding: 20,
        }} className="fade-in">
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12 }}>
            <CheckCircle2 size={16} color="#4ade80" />
            <span style={{ fontSize: 14, fontWeight: 600, color: '#4ade80' }}>Job enqueued successfully</span>
          </div>
          <pre style={{
            background: 'var(--bg-elevated)', border: '1px solid var(--border)',
            borderRadius: 8, padding: '14px 16px', fontSize: 12, color: 'var(--text-secondary)',
            overflow: 'auto', fontFamily: 'monospace', lineHeight: 1.6,
          }}>{JSON.stringify(result, null, 2)}</pre>
        </div>
      )}
    </div>
  );
}
