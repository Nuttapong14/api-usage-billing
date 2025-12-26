'use client';

import { useMemo } from 'react';
import { TopCustomers } from '../../../components/top-customers';
import { useAnalytics } from '../../../hooks/use-analytics';

const compactFormatter = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1
});

function formatCurrency(amount: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0
  }).format(amount);
}

export default function CustomersPage() {
  const today = useMemo(() => new Date(), []);
  const startDate = useMemo(() => {
    const start = new Date(today);
    start.setDate(start.getDate() - 30);
    const year = start.getFullYear();
    const month = String(start.getMonth() + 1).padStart(2, '0');
    const day = String(start.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }, [today]);
  const endDate = useMemo(() => {
    const year = today.getFullYear();
    const month = String(today.getMonth() + 1).padStart(2, '0');
    const day = String(today.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }, [today]);

  const { customers } = useAnalytics({ startDate, endDate, granularity: 'day', limit: 10 });

  const totalRevenue = customers.data.reduce((sum, customer) => sum + customer.revenue.amount, 0);
  const avgRevenue =
    customers.data.length > 0 ? totalRevenue / customers.data.length : 0;
  const totalRequests = customers.data.reduce((sum, customer) => sum + customer.requestCount, 0);

  return (
    <div className="admin-page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Customer intelligence</p>
          <h1 className="page-title">Customer management</h1>
          <p className="page-subtitle">
            Monitor revenue concentration, engagement intensity, and account health signals.
          </p>
        </div>
        <div className="header-actions">
          <button className="button button--ghost" type="button">
            Export list
          </button>
          <button className="button" type="button">
            Add customer
          </button>
        </div>
      </header>

      <section className="stat-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Top customer revenue</p>
          <h2 className="stat-value">
            {customers.data[0] ? formatCurrency(customers.data[0].revenue.amount, customers.data[0].revenue.currency) : '$0'}
          </h2>
          <p className="stat-meta">Largest account this period</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">Total revenue</p>
          <h2 className="stat-value">{formatCurrency(totalRevenue, 'USD')}</h2>
          <p className="stat-meta">{customers.data.length} active accounts</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Avg account value</p>
          <h2 className="stat-value">{formatCurrency(avgRevenue, 'USD')}</h2>
          <p className="stat-meta">Per account</p>
        </div>
        <div className="card stat-card reveal delay-3">
          <p className="eyebrow">Total requests</p>
          <h2 className="stat-value">{compactFormatter.format(totalRequests)}</h2>
          <p className="stat-meta">All tracked usage</p>
        </div>
      </section>

      <TopCustomers
        customers={customers.data}
        title="Customer leaderboard"
        subtitle="Ranked by revenue and request volume for the last 30 days."
      />

      <section className="insight-grid">
        <div className="card insight-card reveal">
          <p className="eyebrow">Expansion ready</p>
          <h2 className="card-title">14 accounts</h2>
          <p className="card-subtitle">Usage up more than 18% this period.</p>
        </div>
        <div className="card insight-card reveal delay-1">
          <p className="eyebrow">At risk</p>
          <h2 className="card-title">3 accounts</h2>
          <p className="card-subtitle">Requests down by more than 10% week over week.</p>
        </div>
        <div className="card insight-card reveal delay-2">
          <p className="eyebrow">High growth</p>
          <h2 className="card-title">8 accounts</h2>
          <p className="card-subtitle">Exceeding quota commitments before mid-month.</p>
        </div>
      </section>
    </div>
  );
}
