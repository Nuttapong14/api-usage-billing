'use client';

import Link from 'next/link';
import { useState } from 'react';
import { WebhookForm } from '../../../components/webhook-form';
import { useWebhooks, type WebhookEndpoint } from '../../../hooks/use-webhooks';

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit'
});

function formatDate(value?: string | null) {
  if (!value) {
    return 'Never';
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return dateFormatter.format(parsed);
}

function statusClass(webhook: WebhookEndpoint) {
  if (!webhook.isActive) {
    return 'status-badge status-badge--revoked';
  }
  if (webhook.consecutiveFailures > 0) {
    return 'status-badge status-badge--pending';
  }
  return 'status-badge status-badge--active';
}

function statusLabel(webhook: WebhookEndpoint) {
  if (!webhook.isActive) {
    return 'inactive';
  }
  if (webhook.consecutiveFailures > 0) {
    return 'retrying';
  }
  return 'healthy';
}

export default function WebhooksPage() {
  const { webhooks, isLoading, error, isDemo, refresh, createWebhook, testWebhook } = useWebhooks({
    limit: 20
  });
  const [actionMessage, setActionMessage] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const activeCount = webhooks.filter((webhook) => webhook.isActive).length;
  const failingCount = webhooks.filter((webhook) => webhook.consecutiveFailures > 0).length;

  async function handleTest(webhook: WebhookEndpoint) {
    setActionMessage(null);
    setActionError(null);
    try {
      const result = await testWebhook(webhook.id);
      setActionMessage(`${webhook.url} responded with ${result.status.toUpperCase()}.`);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Unable to send test webhook');
    }
  }

  return (
    <div className="page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Alerting center</p>
          <h1 className="page-title">Webhooks</h1>
          <p className="page-subtitle">
            Route quota and billing alerts to your incident response tools in seconds.
          </p>
        </div>
        <div className="header-actions">
          <button className="button button--ghost" type="button" onClick={refresh}>
            Refresh
          </button>
          <Link className="button button--ghost" href="/">
            Back to overview
          </Link>
        </div>
      </header>

      {isDemo ? (
        <div className="banner">Demo data is shown. Connect the API to manage live webhook traffic.</div>
      ) : null}
      {error ? <div className="banner banner--error">{error}</div> : null}
      {actionMessage ? <div className="banner">{actionMessage}</div> : null}
      {actionError ? <div className="banner banner--error">{actionError}</div> : null}

      <section className="stats-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Total endpoints</p>
          <h2 className="stat-value">{webhooks.length}</h2>
          <p className="stat-meta">Across all alert types</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">Active</p>
          <h2 className="stat-value">{activeCount}</h2>
          <p className="stat-meta">Currently receiving events</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Needs attention</p>
          <h2 className="stat-value">{failingCount}</h2>
          <p className="stat-meta">Recent delivery failures</p>
        </div>
        <div className="card stat-card reveal delay-3">
          <p className="eyebrow">Last activity</p>
          <h2 className="stat-value">
            {webhooks[0] ? formatDate(webhooks[0].lastTriggeredAt) : 'No activity'}
          </h2>
          <p className="stat-meta">Latest webhook delivery</p>
        </div>
      </section>

      <WebhookForm onCreate={createWebhook} />

      <section className="card webhook-card reveal">
        <div className="card-header">
          <div>
            <p className="eyebrow">Endpoints</p>
            <h2 className="card-title">Active webhooks</h2>
          </div>
          <span className="pill">{webhooks.length} total</span>
        </div>

        {isLoading ? (
          <div className="empty-state">Loading webhooks...</div>
        ) : webhooks.length === 0 ? (
          <div className="empty-state">No webhooks configured yet.</div>
        ) : (
          <div className="webhook-table">
            <div className="webhook-row webhook-row--header">
              <span>Endpoint</span>
              <span>Events</span>
              <span>Status</span>
              <span>Last triggered</span>
              <span>Actions</span>
            </div>
            {webhooks.map((webhook) => (
              <div key={webhook.id} className="webhook-row">
                <div>
                  <p className="webhook-url">{webhook.url}</p>
                  <p className="muted">{webhook.description ?? 'No description'}</p>
                </div>
                <div className="event-list">
                  {webhook.events.map((event) => (
                    <span key={event} className="event-pill">
                      {event}
                    </span>
                  ))}
                </div>
                <div>
                  <span className={statusClass(webhook)}>{statusLabel(webhook)}</span>
                  {webhook.consecutiveFailures > 0 ? (
                    <p className="muted">{webhook.consecutiveFailures} failures</p>
                  ) : null}
                </div>
                <div>
                  <p className="webhook-meta">{formatDate(webhook.lastTriggeredAt)}</p>
                </div>
                <div className="webhook-actions">
                  <button
                    className="button button--ghost button--sm"
                    type="button"
                    onClick={() => handleTest(webhook)}
                  >
                    Send test
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
