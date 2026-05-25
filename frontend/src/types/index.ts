export type Priority = 'CRITICAL' | 'HIGH' | 'NORMAL' | 'LOW';
export type JobStatus = 'pending' | 'running' | 'completed' | 'failed' | 'retrying' | 'dead' | 'cancelled';

export interface Job {
  id: string;
  type: string;
  payload: Record<string, unknown>;
  priority: number;
  status: JobStatus;
  queue: string;
  max_retries: number;
  attempts: number;
  run_at: string;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
  worker_id?: string;
}

export interface JobExecution {
  id: string;
  job_id: string;
  worker_id: string;
  attempt: number;
  started_at: string;
  ended_at: string;
  duration_ms: number;
  status: JobStatus;
  error?: string;
}

export interface WorkerStatus {
  id: string;
  status: 'idle' | 'busy' | 'stopped';
  current_job?: string;
  jobs_run: number;
  started_at: string;
  last_beat_at: string;
}

export interface QueueStats {
  name: string;
  depth: number;
  paused: boolean;
}

export interface Metrics {
  total_enqueued: number;
  total_completed: number;
  total_failed: number;
  total_dead: number;
  active_workers: number;
  throughput_per_sec: number;
  avg_latency_ms: number;
  p99_latency_ms: number;
  queues: QueueStats[];
  workers: WorkerStatus[];
}

export interface WSEvent {
  type: string;
  payload: Job | WorkerStatus;
}

export interface ThroughputPoint {
  time: string;
  completed: number;
  failed: number;
}
