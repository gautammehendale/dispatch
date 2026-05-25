import { LucideIcon } from 'lucide-react';

interface Props {
  label: string;
  value: string | number;
  sub?: string;
  icon: LucideIcon;
  color?: string;
  delta?: number;
}

export default function MetricCard({ label, value, sub, icon: Icon, color = 'var(--amber)', delta }: Props) {
  return (
    <div style={{
      background: 'var(--bg-surface)',
      border: '1px solid var(--border)',
      borderRadius: 12,
      padding: '20px 22px',
      display: 'flex', flexDirection: 'column', gap: 12,
      position: 'relative', overflow: 'hidden',
      transition: 'border-color 0.2s',
      cursor: 'default',
    }}
      onMouseEnter={e => (e.currentTarget.style.borderColor = 'var(--border-strong)')}
      onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}
    >
      {/* top stripe on hover via pseudo won't work inline — use a div overlay */}
      <div style={{
        position: 'absolute', top: 0, left: 0, right: 0, height: 2,
        background: `linear-gradient(90deg, transparent, ${color}, transparent)`,
        opacity: 0.6,
      }} />

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <span style={{ fontSize: 12, color: 'var(--text-muted)', fontWeight: 500, letterSpacing: '0.04em', textTransform: 'uppercase' }}>
          {label}
        </span>
        <div style={{
          width: 32, height: 32, borderRadius: 8,
          background: `${color}18`,
          border: `1px solid ${color}30`,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <Icon size={15} color={color} strokeWidth={2} />
        </div>
      </div>

      <div>
        <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.5px', lineHeight: 1 }}>
          {typeof value === 'number' ? value.toLocaleString() : value}
        </div>
        {sub && (
          <div style={{ fontSize: 12, color: 'var(--text-muted)', marginTop: 5, display: 'flex', alignItems: 'center', gap: 5 }}>
            {delta !== undefined && (
              <span style={{ color: delta >= 0 ? 'var(--green)' : 'var(--red)', fontWeight: 600 }}>
                {delta >= 0 ? '↑' : '↓'} {Math.abs(delta)}
              </span>
            )}
            {sub}
          </div>
        )}
      </div>
    </div>
  );
}
