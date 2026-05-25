import { CheckCircle2, XCircle, AlertOctagon, Zap, Clock, Users } from 'lucide-react';
import MetricCard from '../components/MetricCard';
import ThroughputChart from '../components/ThroughputChart';
import QueueDepthChart from '../components/QueueDepthChart';
import WorkerGrid from '../components/WorkerGrid';
import { Metrics, ThroughputPoint } from '../types';

interface Props { metrics: Metrics; throughput: ThroughputPoint[]; loading: boolean; }

export default function Overview({ metrics, throughput, loading }: Props) {
  const successRate = metrics.total_enqueued > 0
    ? ((metrics.total_completed / metrics.total_enqueued) * 100).toFixed(1)
    : '—';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }} className="fade-in">
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h1 style={{ fontSize: 22, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.4px' }}>
            Overview
          </h1>
          <p style={{ fontSize: 13, color: 'var(--text-muted)', marginTop: 3 }}>
            Real-time job pipeline metrics
          </p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 8, padding: '7px 14px' }}>
          <div className="live-dot" />
          <span style={{ fontSize: 12, color: 'var(--text-secondary)', fontWeight: 500 }}>Live</span>
        </div>
      </div>

      {/* Metric Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 14 }}>
        <MetricCard
          label="Total Enqueued"
          value={loading ? '—' : metrics.total_enqueued}
          sub="all time"
          icon={Zap}
          color="#f59e0b"
        />
        <MetricCard
          label="Completed"
          value={loading ? '—' : metrics.total_completed}
          sub={`${successRate}% success rate`}
          icon={CheckCircle2}
          color="#22c55e"
        />
        <MetricCard
          label="Failed"
          value={loading ? '—' : metrics.total_failed}
          sub="retried automatically"
          icon={XCircle}
          color="#ef4444"
        />
        <MetricCard
          label="Dead Letter"
          value={loading ? '—' : metrics.total_dead}
          sub="exhausted retries"
          icon={AlertOctagon}
          color="#f97316"
        />
        <MetricCard
          label="Active Workers"
          value={loading ? '—' : `${metrics.active_workers} / ${metrics.workers?.length ?? 0}`}
          sub="processing now"
          icon={Users}
          color="#3b82f6"
        />
        <MetricCard
          label="Avg Latency"
          value={loading ? '—' : `${(metrics.avg_latency_ms || 0).toFixed(1)}ms`}
          sub="job execution time"
          icon={Clock}
          color="#a855f7"
        />
      </div>

      {/* Charts */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 380px', gap: 16 }}>
        <ThroughputChart data={throughput} />
        <QueueDepthChart queues={metrics.queues ?? []} />
      </div>

      {/* Workers */}
      <div style={{ background: 'var(--bg-surface)', border: '1px solid var(--border)', borderRadius: 12, padding: '20px 22px' }}>
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-primary)' }}>Worker Pool</div>
          <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 2 }}>
            {metrics.workers?.filter(w => w.status === 'busy').length ?? 0} busy · {metrics.workers?.filter(w => w.status === 'idle').length ?? 0} idle
          </div>
        </div>
        <WorkerGrid workers={metrics.workers ?? []} />
      </div>
    </div>
  );
}
