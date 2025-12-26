import type { ReactNode } from 'react';
import Link from 'next/link';

export const metadata = {
  title: 'Provider Analytics',
  description: 'Revenue, customers, and API performance insights'
};

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="admin-brand">
          <div className="brand-badge">
            <span className="brand-dot" />
          </div>
          <div>
            <p className="brand-kicker">Provider Console</p>
            <p className="brand-title">AtlasPulse</p>
          </div>
        </div>

        <nav className="admin-nav">
          <Link className="admin-link" href="/">
            Overview
          </Link>
          <Link className="admin-link" href="/customers">
            Customers
          </Link>
          <Link className="admin-link" href="/tiers">
            Tiers
          </Link>
        </nav>

        <div className="sidebar-card">
          <p className="sidebar-label">Revenue MTD</p>
          <p className="sidebar-value">$482k</p>
          <p className="sidebar-meta">12% above last month</p>
        </div>
        <div className="sidebar-card sidebar-card--accent">
          <p className="sidebar-label">System health</p>
          <p className="sidebar-value">Stable</p>
          <p className="sidebar-meta">0.7% error rate</p>
        </div>
        <div className="sidebar-footer">
          <p className="sidebar-meta">Auto-refresh every 60 seconds</p>
        </div>
      </aside>

      <main className="admin-main">{children}</main>
    </div>
  );
}
