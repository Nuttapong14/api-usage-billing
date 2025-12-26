'use client';

import { useMemo, useState } from 'react';
import { RevenueChart } from '../../components/revenue-chart';
import { TopCustomers } from '../../components/top-customers';
import { TopApis } from '../../components/top-apis';
import { useAnalytics, type Money } from '../../hooks/use-analytics';

const compactFormatter = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1
});

function formatDateInput(value: Date) {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, '0');
  const day = String(value.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function formatCurrency(money: Money) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: money.currency,
    maximumFractionDigits: 0
  }).format(money.amount);
}

function formatTimestamp(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return parsed.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

export default function DashboardPage() {
  const today = useMemo(() => new Date(), []);
  const [startDate, setStartDate] = useState(() => {
    const start = new Date(today);
    start.setDate(start.getDate() - 30);
    return formatDateInput(start);
  });
  const [endDate, setEndDate] = useState(() => formatDateInput(today));
  const [granularity, setGranularity] = useState<'day' | 'month'>('day');

  const { overview, revenue, customers, isLoading, isDemo, error, range, refresh } = useAnalytics({
    startDate,
    endDate,
    granularity,
    windowDays: 7,
    limit: 6
  });

  const topEndpoint = overview.topEndpoints[0];

  return (
    <div className="admin-page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Provider intelligence</p>
          <h1 className="page-title">Analytics overview</h1>
          <p className="page-subtitle">
            Track revenue momentum, customer concentration, and API performance in real time.
          </p>
        </div>
        <form
          className="filter-panel"
          onSubmit={(event) => {
            event.preventDefault();
            refresh();
          }}
        >
          <div className="filter-group">
            <label className="filter-label" htmlFor="start-date">
              Start
            </label>
            <input
              id="start-date"
              className="filter-input"
              type="date"
              value={startDate}
              onChange={(event) => setStartDate(event.target.value)}
            />
          </div>
          <div className="filter-group">
            <label className="filter-label" htmlFor="end-date">
              End
            </label>
            <input
              id="end-date"
              className="filter-input"
              type="date"
              value={endDate}
              onChange={(event) => setEndDate(event.target.value)}
            />
          </div>
          <div className="filter-group">
            <label className="filter-label" htmlFor="granularity">
              Grain
            </label>
            <select
              id="granularity"
              className="filter-input"
              value={granularity}
              onChange={(event) => setGranularity(event.target.value as 'day' | 'month')}
            >
              <option value="day">Daily</option>
              <option value="month">Monthly</option>
            </select>
          </div>
          <button className="button" type="submit">
            {isLoading ? 'Refreshing...' : 'Refresh'}
          </button>
          <span className="filter-meta">
            {range.startDate} to {range.endDate}
          </span>
        </form>
      </header>

      {isDemo ? (
        <div className="banner">Demo data shown. Set NEXT_PUBLIC_API_BASE_URL for live analytics.</div>
      ) : null}
      {error ? <div className="banner banner--error">{error}</div> : null}

      <section className="stat-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Revenue MTD</p>
          <h2 className="stat-value">{formatCurrency(overview.revenueMtd.amount)}</h2>
          <p className="stat-meta">{overview.revenueMtd.invoiceCount} invoices</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">API calls today</p>
          <h2 className="stat-value">{compactFormatter.format(overview.requestsToday)}</h2>
          <p className="stat-meta">Rolling 24 hours</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Active customers</p>
          <h2 className="stat-value">{overview.activeCustomers}</h2>
          <p className="stat-meta">Current subscriptions</p>
        </div>
        <div className="card stat-card reveal delay-3">
          <p className="eyebrow">Top API health</p>
          <h2 className="stat-value">
            {topEndpoint ? `${topEndpoint.errorRate.toFixed(1)}%` : '0.0%'}
          </h2>
          <p className="stat-meta">
            {topEndpoint ? `${topEndpoint.avgLatencyMs.toFixed(0)} ms avg latency` : 'No traffic yet'}
          </p>
        </div>
      </section>

      <section className="dashboard-grid">
        <RevenueChart title="Net revenue" report={revenue} isLoading={isLoading} />
        <TopApis endpoints={overview.topEndpoints} />
      </section>

      <TopCustomers
        customers={customers.data}
        subtitle="Customer rankings update with the selected date range."
      />

      <section className="card callout-card reveal">
        <div>
          <p className="eyebrow">Last refreshed</p>
          <h2 className="card-title">Keep finance and product aligned</h2>
          <p className="card-subtitle">
            Last update {formatTimestamp(overview.updatedAt)}. Export reports for leadership reviews.
          </p>
        </div>
        <button className="button button--ghost" type="button">
          Export summary
        </button>
      </section>
    </div>
  );
}
