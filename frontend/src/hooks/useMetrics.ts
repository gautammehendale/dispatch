import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import type { Metrics, ThroughputPoint } from '../types';

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080';

const EMPTY_METRICS: Metrics = {
  total_enqueued: 0,
  total_completed: 0,
  total_failed: 0,
  total_dead: 0,
  active_workers: 0,
  throughput_per_sec: 0,
  avg_latency_ms: 0,
  p99_latency_ms: 0,
  queues: [],
  workers: [],
};

export function useMetrics() {
  const [metrics, setMetrics] = useState<Metrics>(EMPTY_METRICS);
  const [throughput, setThroughput] = useState<ThroughputPoint[]>([]);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    try {
      const { data } = await axios.get<Metrics>(`${API}/api/v1/metrics`);
      setMetrics(data);
      setThroughput((prev) => {
        const now = new Date().toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        const next = [...prev, { time: now, completed: data.total_completed, failed: data.total_failed }];
        return next.slice(-30);
      });
    } catch { /* swallow — backend may be starting */ }
    finally { setLoading(false); }
  }, []);

  useEffect(() => {
    fetch();
    const id = setInterval(fetch, 2000);
    return () => clearInterval(id);
  }, [fetch]);

  return { metrics, throughput, loading };
}
