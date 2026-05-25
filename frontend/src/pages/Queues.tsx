import { useState, useEffect } from 'react';
import axios from 'axios';
import { Pause, Play, Trash2, Activity } from 'lucide-react';
import { QueueStats } from '../types';

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default function Queues() {
  const [queues, setQueues] = useState<QueueStats[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = async () => {
    try {
      const { data } = await axios.get(`${API}/api/v1/queues`);
      setQueues(data);
    } catch { /* swallow */ }
    finally { setLoading(false); }
  };

  const pause  = async (name: string) => { await axios.post(`${API}/api/v1/queues/${name}/pause`);  fetch(); };
  const resume = async (name: string) => { await axios.post(`${API}/api/v1/queues/${name}/resume`); fetch(); };

  useEffect(() => { fetch(); const id = setInterval(fetch, 3000); return () => clearInterval(id); }, []);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }} className="fade-in">
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <Activity size={18} color="var(--amber)" />
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>Queues</h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>Manage and monitor queue state</p>
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        {loading
          ? [...Array(3)].map((_, i) => <div key={i} className="skeleton" style={{ height: 80, borderRadius: 10 }} />)
          : queues.map(q => (
            <div key={q.name} style={{
              background: 'var(--bg-surface)', border: `1px solid ${q.paused ? 'rgba(249,115,22,0.3)' : 'var(--border)'}`,
              borderRadius: 12, padding: '18px 22px',
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
                <div style={{
                  width: 42, height: 42, borderRadius: 10,
                  background: q.paused ? 'rgba(249,115,22,0.1)' : 'rgba(245,158,11,0.1)',
                  border: `1px solid ${q.paused ? 'rgba(249,115,22,0.3)' : 'rgba(245,158,11,0.2)'}`,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                  <Activity size={16} color={q.paused ? '#f97316' : 'var(--amber)'} />
                </div>
                <div>
                  <div style={{ fontSize: 15, fontWeight: 600, color: 'var(--text-primary)', display: 'flex', alignItems: 'center', gap: 8 }}>
                    {q.name}
                    {q.paused && <span className="badge badge-failed" style={{ fontSize: 9 }}>PAUSED</span>}
                  </div>
                  <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 3 }}>
                    {q.depth.toLocaleString()} pending jobs
                  </div>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 8 }}>
                {q.paused ? (
                  <button
                    onClick={() => resume(q.name)}
                    style={{
                      background: 'rgba(34,197,94,0.1)', border: '1px solid rgba(34,197,94,0.25)',
                      color: '#4ade80', borderRadius: 8, padding: '7px 14px',
                      cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, fontWeight: 500,
                    }}
                  >
                    <Play size={12} /> Resume
                  </button>
                ) : (
                  <button
                    onClick={() => pause(q.name)}
                    style={{
                      background: 'rgba(249,115,22,0.1)', border: '1px solid rgba(249,115,22,0.25)',
                      color: '#fb923c', borderRadius: 8, padding: '7px 14px',
                      cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontSize: 12, fontWeight: 500,
                    }}
                  >
                    <Pause size={12} /> Pause
                  </button>
                )}
              </div>
            </div>
          ))
        }
      </div>
    </div>
  );
}
