import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import type { Job, JobStatus } from '../types';

const API = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function useJobs(status?: JobStatus, queue?: string) {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = { limit: '50', offset: '0' };
      if (status) params.status = status;
      if (queue) params.queue = queue;
      const { data } = await axios.get(`${API}/api/v1/jobs`, { params });
      setJobs(data.jobs ?? []);
      setTotal(data.total ?? 0);
    } catch { /* backend may be starting */ }
    finally { setLoading(false); }
  }, [status, queue]);

  useEffect(() => {
    fetch();
  }, [fetch]);

  const retry = useCallback(async (id: string) => {
    await axios.post(`${API}/api/v1/jobs/${id}/retry`);
    fetch();
  }, [fetch]);

  const cancel = useCallback(async (id: string) => {
    await axios.post(`${API}/api/v1/jobs/${id}/cancel`);
    fetch();
  }, [fetch]);

  return { jobs, total, loading, refresh: fetch, retry, cancel };
}

export function useEnqueue() {
  const enqueue = useCallback(async (payload: {
    type: string;
    priority: string;
    queue: string;
    max_retries: number;
    payload: Record<string, unknown>;
  }) => {
    const { data } = await axios.post(`${API}/api/v1/jobs`, payload);
    return data as Job;
  }, []);
  return { enqueue };
}
