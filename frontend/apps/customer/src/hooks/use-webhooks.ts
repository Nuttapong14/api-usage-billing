'use client';

import { useCallback, useEffect, useState } from 'react';

export type WebhookEventType =
  | 'usage.quota_warning'
  | 'usage.quota_exceeded'
  | 'invoice.created'
  | 'invoice.paid'
  | 'payment.received'
  | 'subscription.upgraded'
  | 'subscription.downgraded'
  | 'subscription.cancelled'
  | 'api_key.created'
  | 'api_key.revoked';

export interface WebhookEndpoint {
  id: string;
  url: string;
  description?: string | null;
  events: WebhookEventType[];
  isActive: boolean;
  lastTriggeredAt?: string | null;
  consecutiveFailures: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWebhookInput {
  url: string;
  description?: string;
  events: WebhookEventType[];
}

export interface CreateWebhookResult {
  webhook: WebhookEndpoint;
  secret: string;
}

export interface TestWebhookResult {
  deliveryId: string;
  status: string;
  responseStatus?: number | null;
  responseTimeMs?: number | null;
  message?: string;
}

interface WebhookListResponse {
  data: Array<{
    id: string;
    url: string;
    description?: string | null;
    events: WebhookEventType[];
    is_active: boolean;
    last_triggered_at?: string | null;
    consecutive_failures: number;
    created_at: string;
    updated_at: string;
  }>;
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

interface CreateWebhookResponse {
  id: string;
  url: string;
  description?: string | null;
  events: WebhookEventType[];
  is_active: boolean;
  last_triggered_at?: string | null;
  consecutive_failures: number;
  created_at: string;
  updated_at: string;
  secret: string;
}

interface TestWebhookResponse {
  delivery_id: string;
  status: string;
  response_status?: number | null;
  response_time_ms?: number | null;
  message?: string;
}

interface UseWebhooksOptions {
  baseUrl?: string;
  limit?: number;
}

interface WebhookState {
  webhooks: WebhookEndpoint[];
  isLoading: boolean;
  error?: string;
  isDemo: boolean;
}

const demoWebhooks: WebhookEndpoint[] = [
  {
    id: 'wh-001',
    url: 'https://hooks.pulseline.app/usage',
    description: 'Send quota alerts to Opsgenie.',
    events: ['usage.quota_warning', 'usage.quota_exceeded'],
    isActive: true,
    lastTriggeredAt: '2025-01-24T09:30:00Z',
    consecutiveFailures: 0,
    createdAt: '2024-11-20T12:12:00Z',
    updatedAt: '2025-01-18T15:45:00Z'
  },
  {
    id: 'wh-002',
    url: 'https://finance.northwind.io/billing/webhooks',
    description: 'Billing and payment confirmations.',
    events: ['invoice.created', 'invoice.paid', 'payment.received'],
    isActive: true,
    lastTriggeredAt: '2025-01-23T17:10:00Z',
    consecutiveFailures: 1,
    createdAt: '2024-10-01T09:00:00Z',
    updatedAt: '2025-01-22T08:00:00Z'
  },
  {
    id: 'wh-003',
    url: 'https://teams.aurora.io/hooks/subscriptions',
    description: 'Subscription lifecycle updates.',
    events: ['subscription.upgraded', 'subscription.downgraded', 'subscription.cancelled'],
    isActive: false,
    lastTriggeredAt: '2024-12-08T11:12:00Z',
    consecutiveFailures: 3,
    createdAt: '2024-07-18T10:40:00Z',
    updatedAt: '2025-01-05T10:00:00Z'
  }
];

const baseState: WebhookState = {
  webhooks: demoWebhooks,
  isLoading: false,
  isDemo: true
};

function mapWebhook(payload: WebhookListResponse['data'][number]): WebhookEndpoint {
  return {
    id: payload.id,
    url: payload.url,
    description: payload.description ?? null,
    events: payload.events,
    isActive: payload.is_active,
    lastTriggeredAt: payload.last_triggered_at ?? null,
    consecutiveFailures: payload.consecutive_failures,
    createdAt: payload.created_at,
    updatedAt: payload.updated_at
  };
}

function buildDemoWebhook(input: CreateWebhookInput): CreateWebhookResult {
  const now = new Date().toISOString();
  const secret = `whsec_${Math.random().toString(36).slice(2, 10)}${Math.random().toString(36).slice(2, 10)}`;
  const webhook: WebhookEndpoint = {
    id: `wh-${Date.now()}`,
    url: input.url,
    description: input.description ?? null,
    events: input.events,
    isActive: true,
    lastTriggeredAt: null,
    consecutiveFailures: 0,
    createdAt: now,
    updatedAt: now
  };
  return { webhook, secret };
}

export function useWebhooks(options: UseWebhooksOptions = {}) {
  const baseUrl = options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL;
  const limit = options.limit ?? 20;
  const normalizedBaseUrl = baseUrl ? baseUrl.replace(/\/$/, '') : '';

  const [state, setState] = useState<WebhookState>(baseState);
  const [refreshIndex, setRefreshIndex] = useState(0);

  const fetchWebhooks = useCallback(
    async (signal?: AbortSignal) => {
      if (!normalizedBaseUrl) {
        return;
      }

      setState((prev) => ({
        ...prev,
        isLoading: true,
        error: undefined
      }));

      try {
        const params = new URLSearchParams({
          limit: String(limit),
          offset: '0'
        });
        const response = await fetch(`${normalizedBaseUrl}/v1/webhooks?${params.toString()}`, {
          headers: { Accept: 'application/json' },
          credentials: 'include',
          signal
        });

        if (!response.ok) {
          throw new Error('Unable to load webhooks');
        }

        const payload = (await response.json()) as WebhookListResponse;
        const webhooks = payload.data.map(mapWebhook);

        setState({
          webhooks,
          isLoading: false,
          isDemo: false
        });
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          return;
        }
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isDemo: false,
          error: err instanceof Error ? err.message : 'Unable to load webhooks'
        }));
      }
    },
    [limit, normalizedBaseUrl]
  );

  useEffect(() => {
    const controller = new AbortController();
    fetchWebhooks(controller.signal);
    return () => controller.abort();
  }, [fetchWebhooks, refreshIndex]);

  const refresh = useCallback(() => setRefreshIndex((prev) => prev + 1), []);

  const createWebhook = useCallback(
    async (input: CreateWebhookInput): Promise<CreateWebhookResult> => {
      if (!normalizedBaseUrl) {
        const demo = buildDemoWebhook(input);
        setState((prev) => ({
          ...prev,
          webhooks: [demo.webhook, ...prev.webhooks]
        }));
        return demo;
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/webhooks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          url: input.url,
          description: input.description,
          events: input.events
        })
      });

      if (!response.ok) {
        throw new Error('Unable to create webhook');
      }

      const payload = (await response.json()) as CreateWebhookResponse;
      const webhook = mapWebhook({
        id: payload.id,
        url: payload.url,
        description: payload.description,
        events: payload.events,
        is_active: payload.is_active,
        last_triggered_at: payload.last_triggered_at,
        consecutive_failures: payload.consecutive_failures,
        created_at: payload.created_at,
        updated_at: payload.updated_at
      });

      setState((prev) => ({
        ...prev,
        webhooks: [webhook, ...prev.webhooks]
      }));

      return { webhook, secret: payload.secret };
    },
    [normalizedBaseUrl]
  );

  const testWebhook = useCallback(
    async (webhookId: string): Promise<TestWebhookResult> => {
      if (!normalizedBaseUrl) {
        return {
          deliveryId: `whd-${Date.now()}`,
          status: 'delivered',
          responseStatus: 200,
          responseTimeMs: 180,
          message: 'Test event delivered'
        };
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/webhooks/${webhookId}/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ event_type: 'test.webhook' })
      });

      if (!response.ok) {
        throw new Error('Unable to send test webhook');
      }

      const payload = (await response.json()) as TestWebhookResponse;
      return {
        deliveryId: payload.delivery_id,
        status: payload.status,
        responseStatus: payload.response_status ?? null,
        responseTimeMs: payload.response_time_ms ?? null,
        message: payload.message
      };
    },
    [normalizedBaseUrl]
  );

  return {
    webhooks: state.webhooks,
    isLoading: state.isLoading,
    error: state.error,
    isDemo: state.isDemo,
    refresh,
    createWebhook,
    testWebhook
  };
}
