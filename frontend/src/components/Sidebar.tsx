import { type LucideIcon, LayoutDashboard, Layers, Users, AlertTriangle, PlusCircle, Activity } from 'lucide-react';

interface NavItem {
  icon: LucideIcon;
  label: string;
  page: string;
  badge?: number;
}

interface Props {
  current: string;
  onChange: (page: string) => void;
  deadCount?: number;
}

const NAV: NavItem[] = [
  { icon: LayoutDashboard, label: 'Overview',  page: 'overview' },
  { icon: Layers,          label: 'Jobs',       page: 'jobs' },
  { icon: Activity,        label: 'Queues',     page: 'queues' },
  { icon: Users,           label: 'Workers',    page: 'workers' },
  { icon: AlertTriangle,   label: 'Dead Letter', page: 'dlq' },
  { icon: PlusCircle,      label: 'Enqueue',    page: 'enqueue' },
];

export default function Sidebar({ current, onChange, deadCount }: Props) {
  return (
    <aside style={{
      width: 220,
      minHeight: '100vh',
      background: 'var(--bg-surface)',
      borderRight: '1px solid var(--border)',
      display: 'flex',
      flexDirection: 'column',
      padding: '0',
      flexShrink: 0,
    }}>
      {/* Logo */}
      <div style={{ padding: '24px 20px 20px', borderBottom: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{
            width: 32, height: 32,
            background: 'linear-gradient(135deg, #f59e0b, #f97316)',
            borderRadius: 8,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 14, fontWeight: 700, color: '#000',
          }}>D</div>
          <div>
            <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--text-primary)', letterSpacing: '-0.3px' }}>
              Dispatch
            </div>
            <div style={{ fontSize: 11, color: 'var(--text-muted)', letterSpacing: '0.02em' }}>
              Task Queue Engine
            </div>
          </div>
        </div>
      </div>

      {/* Nav */}
      <nav style={{ flex: 1, padding: '12px 10px' }}>
        {NAV.map(({ icon: Icon, label, page, badge }) => {
          const isActive = current === page;
          const count = page === 'dlq' ? deadCount : badge;
          return (
            <button
              key={page}
              onClick={() => onChange(page)}
              style={{
                width: '100%',
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                gap: 10, padding: '9px 12px',
                borderRadius: 8, border: 'none', cursor: 'pointer',
                background: isActive ? 'rgba(245,158,11,0.1)' : 'transparent',
                color: isActive ? 'var(--amber)' : 'var(--text-secondary)',
                fontSize: 13, fontWeight: isActive ? 600 : 400,
                transition: 'all 0.15s',
                marginBottom: 2,
              }}
              onMouseEnter={e => { if (!isActive) (e.currentTarget as HTMLButtonElement).style.background = 'var(--bg-elevated)'; }}
              onMouseLeave={e => { if (!isActive) (e.currentTarget as HTMLButtonElement).style.background = 'transparent'; }}
            >
              <span style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <Icon size={15} strokeWidth={isActive ? 2.2 : 1.8} />
                {label}
              </span>
              {count != null && count > 0 && (
                <span style={{
                  background: page === 'dlq' ? 'rgba(239,68,68,0.2)' : 'rgba(245,158,11,0.15)',
                  color: page === 'dlq' ? '#f87171' : 'var(--amber)',
                  fontSize: 10, fontWeight: 700,
                  padding: '1px 6px', borderRadius: 999,
                  minWidth: 18, textAlign: 'center',
                }}>{count}</span>
              )}
            </button>
          );
        })}
      </nav>

      {/* Footer */}
      <div style={{ padding: '16px 20px', borderTop: '1px solid var(--border)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
          <div className="live-dot" />
          <span style={{ fontSize: 11, color: 'var(--text-muted)' }}>Live · v1.0.0</span>
        </div>
      </div>
    </aside>
  );
}
