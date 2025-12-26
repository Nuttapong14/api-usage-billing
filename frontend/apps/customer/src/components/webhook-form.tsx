'use client';

import { type FormEvent, useState } from 'react';
import type { CreateWebhookInput, CreateWebhookResult, WebhookEventType } from '../hooks/use-webhooks';

interface WebhookFormProps {
  onCreate: (input: CreateWebhookInput) => Promise<CreateWebhookResult>;
}

const eventOptions: Array<{ value: WebhookEventType; label: string; description: string }> = [
  {
    value: 'usage.quota_warning',
    label: 'Quota warning',
    description: 'Triggered when usage crosses 80% or 95% of quota.'
  },
  {
    value: 'usage.quota_exceeded',
    label: 'Quota exceeded',
    description: 'Triggered when usage reaches or exceeds 100%.'
  },
  {
    value: 'invoice.created',
    label: 'Invoice created',
    description: 'Triggered whenever a new invoice is issued.'
  },
  {
    value: 'invoice.paid',
    label: 'Invoice paid',
    description: 'Triggered when an invoice is marked as paid.'
  },
  {
    value: 'payment.received',
    label: 'Payment received',
    description: 'Triggered when a payment is confirmed.'
  },
  {
    value: 'subscription.upgraded',
    label: 'Subscription upgraded',
    description: 'Triggered when a tier upgrade is completed.'
  },
  {
    value: 'subscription.downgraded',
    label: 'Subscription downgraded',
    description: 'Triggered when a tier downgrade is completed.'
  },
  {
    value: 'subscription.cancelled',
    label: 'Subscription cancelled',
    description: 'Triggered when a subscription is cancelled.'
  },
  {
    value: 'api_key.created',
    label: 'API key created',
    description: 'Triggered when a new API key is issued.'
  },
  {
    value: 'api_key.revoked',
    label: 'API key revoked',
    description: 'Triggered when an API key is revoked.'
  }
];

export function WebhookForm({ onCreate }: WebhookFormProps) {
  const [url, setUrl] = useState('');
  const [description, setDescription] = useState('');
  const [selectedEvents, setSelectedEvents] = useState<WebhookEventType[]>([
    'usage.quota_warning',
    'usage.quota_exceeded'
  ]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<CreateWebhookResult | null>(null);

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    if (selectedEvents.length === 0) {
      setError('Select at least one event type.');
      setIsSubmitting(false);
      return;
    }

    try {
      const result = await onCreate({
        url: url.trim(),
        description: description.trim() || undefined,
        events: selectedEvents
      });
      setCreated(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unable to create webhook');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleCopy() {
    if (!created?.secret) {
      return;
    }
    try {
      await navigator.clipboard.writeText(created.secret);
    } catch {
      setError('Unable to copy secret to clipboard');
    }
  }

  function toggleEvent(value: WebhookEventType) {
    setSelectedEvents((prev) => {
      if (prev.includes(value)) {
        return prev.filter((item) => item !== value);
      }
      return [...prev, value];
    });
  }

  function resetForm() {
    setUrl('');
    setDescription('');
    setSelectedEvents(['usage.quota_warning', 'usage.quota_exceeded']);
    setCreated(null);
    setError(null);
  }

  return (
    <div className="card webhook-form-card reveal">
      <div className="card-header">
        <div>
          <p className="eyebrow">New endpoint</p>
          <h2 className="card-title">Create webhook</h2>
          <p className="page-subtitle">
            Choose event triggers and share the signing secret with your security team.
          </p>
        </div>
        {created ? (
          <button className="button button--ghost button--sm" type="button" onClick={resetForm}>
            Create another
          </button>
        ) : null}
      </div>

      {created ? (
        <div className="key-display">
          <p className="eyebrow">Signing secret</p>
          <p className="key-value">{created.secret}</p>
          <p className="helper-text">
            Store this secret securely. It will not be shown again after leaving this page.
          </p>
          {error ? <div className="banner banner--error">{error}</div> : null}
          <div className="modal-actions">
            <button className="button" type="button" onClick={handleCopy}>
              Copy secret
            </button>
            <button className="button button--ghost" type="button" onClick={resetForm}>
              Done
            </button>
          </div>
        </div>
      ) : (
        <form className="form-grid" onSubmit={handleSubmit}>
          <div className="field">
            <label className="label" htmlFor="webhook-url">
              Endpoint URL
            </label>
            <input
              id="webhook-url"
              className="input"
              value={url}
              onChange={(event) => setUrl(event.target.value)}
              placeholder="https://example.com/webhooks"
              required
            />
          </div>

          <div className="field">
            <label className="label" htmlFor="webhook-description">
              Description
            </label>
            <textarea
              id="webhook-description"
              className="textarea"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder="Route to billing or operations workflows"
              rows={3}
            />
          </div>

          <div className="field">
            <span className="label">Event triggers</span>
            <div className="event-grid">
              {eventOptions.map((option) => (
                <label key={option.value} className="event-pill">
                  <input
                    type="checkbox"
                    checked={selectedEvents.includes(option.value)}
                    onChange={() => toggleEvent(option.value)}
                  />
                  <span>{option.label}</span>
                  <span className="muted">{option.description}</span>
                </label>
              ))}
            </div>
          </div>

          {error ? <div className="banner banner--error">{error}</div> : null}

          <div className="modal-actions">
            <button className="button" type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Create webhook'}
            </button>
            <button className="button button--ghost" type="button" onClick={resetForm}>
              Reset
            </button>
          </div>
        </form>
      )}
    </div>
  );
}
