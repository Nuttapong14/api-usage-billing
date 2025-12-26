import type { ReactNode } from 'react';
import Link from 'next/link';

export const metadata = {
  title: 'Customer Dashboard',
  description: 'Usage, quota, and billing overview'
};

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <div className="dashboard-shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brand-mark" />
          <div>
            <p className="brand-eyebrow">Customer Portal</p>
            <p className="brand-title">PulseLine</p>
          </div>
        </div>
        <nav className="nav">
          <Link className="nav-link" href="/">
            Overview
          </Link>
          <Link className="nav-link" href="/billing">
            Billing
          </Link>
          <Link className="nav-link" href="/api-keys">
            API Keys
          </Link>
          <Link className="nav-link" href="/webhooks">
            Webhooks
          </Link>
        </nav>
        <div className="sidebar-card">
          <p className="sidebar-card__title">Current plan</p>
          <p className="sidebar-card__value">Growth Tier</p>
          <p className="sidebar-card__meta">Renews on Feb 1</p>
        </div>
        <div className="sidebar-card sidebar-card--outline">
          <p className="sidebar-card__title">Support</p>
          <p className="sidebar-card__meta">support@pulseline.app</p>
        </div>
      </aside>
      <div className="dashboard-main">
        {children}
      </div>
    </div>
  );
}
