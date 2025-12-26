'use client';

import { useEffect, useMemo, useState } from 'react';

export interface UsageMetrics {
  totalRequests: number;
  successfulRequests: number;
  failedRequests: number;
  totalBandwidthBytes: number;
  avgLatencyMs: number;
  p95LatencyMs: number;
  p99LatencyMs: number;
}

export interface QuotaItem {
  limit: number | null;
  used: number;
  remaining: number | null;
  percentageUsed: number;
}

export interface UsageQuota {
  requests: QuotaItem;
  bandwidthMb: QuotaItem;
}

export interface UsageSnapshot {
  customerId: string;
  billingPeriod: {
    start: string;
    end: string;
  };
  usage: UsageMetrics;
  quota: UsageQuota;
  lastUpdated: string;
}

export interface UsageSeriesPoint {
  label: string;
  value: number;
  timestamp: string;
}

export interface QuotaProgressItem {
  label: string;
  used: number;
  limit: number | null;
  unit: string;
  percentUsed: number;
}

interface CurrentUsageResponse {
  customer_id: string;
  billing_period: {
    start: string;
    end: string;
  };
  usage: {
    total_requests: number;
    successful_requests: number;
    failed_requests: number;
    total_bandwidth_bytes: number;
    avg_latency_ms: number;
    p95_latency_ms: number;
    p99_latency_ms: number;
  };
  quota: {
    requests?: {
      limit: number | null;
      used: number;
      remaining: number | null;
      percentage_used: number;
    };
    bandwidth_mb?: {
      limit: number | null;
      used: number;
      remaining: number | null;
      percentage_used: number;
    };
  };
  last_updated: string;
}

interface UsageHistoryResponse {
  data: Array<{
    timestamp: string;
    period: string;
    metrics: {
      total_requests: number;
    };
  }>;
}

interface UseUsageOptions {
  baseUrl?: string;
  historyDays?: number;
  autoRefreshMs?: number;
}

interface UsageState {
  overview: UsageSnapshot;
  history: UsageSeriesPoint[];
  isLoading: boolean;
  error?: string;
  isDemo: boolean;
}

const demoUsage: UsageSnapshot = {
  customerId: 'demo-customer',
  billingPeriod: {
    start: '2025-01-01',
    end: '2025-01-31'
  },
  usage: {
    totalRequests: 184230,
    successfulRequests: 180984,
    failedRequests: 3246,
    totalBandwidthBytes: 8940000000,
    avgLatencyMs: 142,
    p95LatencyMs: 320,
    p99LatencyMs: 540
  },
  quota: {
    requests: {
      limit: 250000,
      used: 184230,
      remaining: 65770,
      percentageUsed: 73.7
    },
    bandwidthMb: {
      limit: 12000,
      used: 8640,
      remaining: 3360,
      percentageUsed: 72
    }
  },
  lastUpdated: '2025-01-24T10:12:00Z'
};

const demoHistory: UsageSeriesPoint[] = [
  { label: 'Jan 10', value: 11840, timestamp: '2025-01-10T00:00:00Z' },
  { label: 'Jan 11', value: 12510, timestamp: '2025-01-11T00:00:00Z' },
  { label: 'Jan 12', value: 13320, timestamp: '2025-01-12T00:00:00Z' },
  { label: 'Jan 13', value: 12990, timestamp: '2025-01-13T00:00:00Z' },
  { label: 'Jan 14', value: 14020, timestamp: '2025-01-14T00:00:00Z' },
  { label: 'Jan 15', value: 14980, timestamp: '2025-01-15T00:00:00Z' },
  { label: 'Jan 16', value: 15640, timestamp: '2025-01-16T00:00:00Z' },
  { label: 'Jan 17', value: 16210, timestamp: '2025-01-17T00:00:00Z' },
  { label: 'Jan 18', value: 17040, timestamp: '2025-01-18T00:00:00Z' },
  { label: 'Jan 19', value: 16680, timestamp: '2025-01-19T00:00:00Z' },
  { label: 'Jan 20', value: 17890, timestamp: '2025-01-20T00:00:00Z' },
  { label: 'Jan 21', value: 18520, timestamp: '2025-01-21T00:00:00Z' },
  { label: 'Jan 22', value: 19240, timestamp: '2025-01-22T00:00:00Z' },
  { label: 'Jan 23', value: 19980, timestamp: '2025-01-23T00:00:00Z' }
];

const baseState: UsageState = {
  overview: demoUsage,
  history: demoHistory,
  isLoading: false,
  isDemo: true
};

function mapCurrentUsage(payload: CurrentUsageResponse): UsageSnapshot {
  return {
    customerId: payload.customer_id,
    billingPeriod: payload.billing_period,
    usage: {
      totalRequests: payload.usage.total_requests,
      successfulRequests: payload.usage.successful_requests,
      failedRequests: payload.usage.failed_requests,
      totalBandwidthBytes: payload.usage.total_bandwidth_bytes,
      avgLatencyMs: payload.usage.avg_latency_ms,
      p95LatencyMs: payload.usage.p95_latency_ms,
      p99LatencyMs: payload.usage.p99_latency_ms
    },
    quota: {
      requests: mapQuotaItem(payload.quota.requests),
      bandwidthMb: mapQuotaItem(payload.quota.bandwidth_mb)
    },
    lastUpdated: payload.last_updated
  };
}

function mapQuotaItem(item?: CurrentUsageResponse['quota']['requests']): QuotaItem {
  if (!item) {
    return { limit: null, used: 0, remaining: null, percentageUsed: 0 };
  }

  return {
    limit: item.limit ?? null,
    used: item.used ?? 0,
    remaining: item.remaining ?? null,
    percentageUsed: item.percentage_used ?? 0
  };
}

function mapHistory(payload: UsageHistoryResponse): UsageSeriesPoint[] {
  return payload.data.map((point) => ({
    label: point.period,
    value: point.metrics.total_requests,
    timestamp: point.timestamp
  }));
}

function formatDate(value: Date): string {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, '0');
  const day = String(value.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function useUsage(options: UseUsageOptions = {}) {
  const baseUrl = options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL;
  const historyDays = options.historyDays ?? 14;
  const normalizedBaseUrl = baseUrl ? baseUrl.replace(/\/$/, '') : '';

  const [state, setState] = useState<UsageState>(baseState);
  const [refreshIndex, setRefreshIndex] = useState(0);

  useEffect(() => {
    if (!normalizedBaseUrl) {
      return;
    }

    const controller = new AbortController();

    async function loadUsage() {
      setState((prev) => ({
        ...prev,
        isLoading: true,
        error: undefined
      }));

      try {
        const end = new Date();
        const start = new Date();
        start.setDate(end.getDate() - historyDays);

        const historyParams = new URLSearchParams({
          start_date: formatDate(start),
          end_date: formatDate(end),
          granularity: 'daily'
        });

        const [currentResponse, historyResponse] = await Promise.all([
          fetch(`${normalizedBaseUrl}/v1/usage/current`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
            signal: controller.signal
          }),
          fetch(`${normalizedBaseUrl}/v1/usage/history?${historyParams.toString()}`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
            signal: controller.signal
          })
        ]);

        if (!currentResponse.ok || !historyResponse.ok) {
          throw new Error('Unable to load usage data');
        }

        const currentPayload = (await currentResponse.json()) as CurrentUsageResponse;
        const historyPayload = (await historyResponse.json()) as UsageHistoryResponse;

        setState({
          overview: mapCurrentUsage(currentPayload),
          history: mapHistory(historyPayload),
          isLoading: false,
          isDemo: false
        });
      } catch (error) {
        if (controller.signal.aborted) {
          return;
        }

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to load usage data'
        }));
      }
    }

    loadUsage();

    return () => controller.abort();
  }, [normalizedBaseUrl, historyDays, refreshIndex]);

  useEffect(() => {
    if (!normalizedBaseUrl || !options.autoRefreshMs) {
      return;
    }

    const interval = setInterval(() => {
      setRefreshIndex((value) => value + 1);
    }, options.autoRefreshMs);

    return () => clearInterval(interval);
  }, [normalizedBaseUrl, options.autoRefreshMs]);

  const quotaItems = useMemo<QuotaProgressItem[]>(() => {
    return [
      {
        label: 'API Requests',
        used: state.overview.quota.requests.used,
        limit: state.overview.quota.requests.limit,
        unit: 'requests',
        percentUsed: state.overview.quota.requests.percentageUsed
      },
      {
        label: 'Bandwidth',
        used: state.overview.quota.bandwidthMb.used,
        limit: state.overview.quota.bandwidthMb.limit,
        unit: 'MB',
        percentUsed: state.overview.quota.bandwidthMb.percentageUsed
      }
    ];
  }, [state.overview]);

  const refresh = () => setRefreshIndex((value) => value + 1);

  return {
    ...state,
    quotaItems,
    refresh
  };
}
