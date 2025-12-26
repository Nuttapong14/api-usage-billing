'use client';

import Link from 'next/link';
import { UsageChart } from '../../components/usage-chart';
import { QuotaProgress } from '../../components/quota-progress';
import { BillingPeriodCard } from '../../components/billing-period-card';
import { useUsage } from '../../hooks/use-usage';

const compactFormatter = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1
});

function formatLatency(value: number) {
  return `${Math.round(value)} ms`;
}

function formatBandwidth(bytes: number) {
  const gb = bytes / (1024 * 1024 * 1024);
  if (gb >= 1) {
    return `${gb.toFixed(1)} GB`;
  }
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
}

export default function DashboardPage() {
  const { overview, history, quotaItems, isLoading, error, isDemo, refresh } = useUsage({
    autoRefreshMs: 60000
  });

  const successRate = overview.usage.totalRequests
    ? (overview.usage.successfulRequests / overview.usage.totalRequests) * 100
    : 0;

  return (
    <div className="page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Welcome back</p>
          <h1 className="page-title">Usage overview</h1>
          <p className="page-subtitle">
            Track real-time performance, quota runway, and billing cycle health in one view.
          </p>
        </div>
        <div className="header-actions">
          <button className="button button--ghost" type="button" onClick={refresh}>
            {isLoading ? 'Refreshing...' : 'Refresh'}
          </button>
          <Link className="button" href="/billing">
            View billing
          </Link>
        </div>
      </header>

      {isDemo ? (
        <div className="banner">Demo data is shown. Set NEXT_PUBLIC_API_BASE_URL to load live usage.</div>
      ) : null}
      {error ? <div className="banner banner--error">{error}</div> : null}

      <section className="stats-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Total requests</p>
          <h2 className="stat-value">{compactFormatter.format(overview.usage.totalRequests)}</h2>
          <p className="stat-meta">This billing period</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">Success rate</p>
          <h2 className="stat-value">{successRate.toFixed(1)}%</h2>
          <p className="stat-meta">{compactFormatter.format(overview.usage.failedRequests)} failed</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Avg latency</p>
          <h2 className="stat-value">{formatLatency(overview.usage.avgLatencyMs)}</h2>
          <p className="stat-meta">P95 {formatLatency(overview.usage.p95LatencyMs)}</p>
        </div>
        <div className="card stat-card reveal delay-3">
          <p className="eyebrow">Bandwidth</p>
          <h2 className="stat-value">{formatBandwidth(overview.usage.totalBandwidthBytes)}</h2>
          <p className="stat-meta">Peak {formatLatency(overview.usage.p99LatencyMs)}</p>
        </div>
      </section>

      <section className="content-grid">
        <UsageChart title="Requests over time" data={history} total={overview.usage.totalRequests} />
        <div className="stack">
          <QuotaProgress items={quotaItems} updatedAt="Just now" />
          <BillingPeriodCard
            start={overview.billingPeriod.start}
            end={overview.billingPeriod.end}
            nextInvoiceDate={overview.billingPeriod.end}
            status="Active"
          />
        </div>
      </section>

      <section className="card callout-card reveal">
        <div>
          <p className="eyebrow">Billing snapshot</p>
          <h2 className="card-title">Get ahead of your next invoice</h2>
          <p className="page-subtitle">
            Review detailed invoice breakdowns and download receipts for your finance team.
          </p>
        </div>
        <Link className="button button--ghost" href="/billing">
          Open billing history
        </Link>
      </section>
    </div>
  );
}
