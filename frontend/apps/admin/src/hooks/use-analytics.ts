'use client';

import { useEffect, useMemo, useState } from 'react';

export interface Money {
  amount: number;
  currency: string;
}

export interface RevenueSnapshot {
  amount: Money;
  invoiceCount: number;
  periodStart: string;
  periodEnd: string;
}

export interface RevenuePoint {
  period: string;
  date: string;
  amount: Money;
  invoiceCount: number;
}

export interface RevenueReport {
  total: Money;
  periodStart: string;
  periodEnd: string;
  granularity: string;
  data: RevenuePoint[];
}

export interface EndpointStat {
  endpoint: string;
  method: string;
  requests: number;
  errorRate: number;
  avgLatencyMs: number;
}

export interface CustomerMetric {
  customerId: string;
  name: string;
  email: string;
  revenue: Money;
  requestCount: number;
  invoiceCount: number;
  lastInvoiceAt?: string;
}

export interface CustomerRanking {
  periodStart: string;
  periodEnd: string;
  data: CustomerMetric[];
}

export interface Overview {
  revenueMtd: RevenueSnapshot;
  requestsToday: number;
  activeCustomers: number;
  topEndpoints: EndpointStat[];
  updatedAt: string;
}

interface AnalyticsState {
  overview: Overview;
  revenue: RevenueReport;
  customers: CustomerRanking;
  isLoading: boolean;
  isDemo: boolean;
  error?: string;
}

interface UseAnalyticsOptions {
  baseUrl?: string;
  startDate?: string;
  endDate?: string;
  granularity?: 'day' | 'month';
  windowDays?: number;
  limit?: number;
}

interface OverviewResponse {
  revenue_mtd: {
    amount: Money;
    invoice_count: number;
    period_start: string;
    period_end: string;
  };
  requests_today: number;
  active_customers: number;
  top_endpoints: Array<{
    endpoint: string;
    method: string;
    requests: number;
    error_rate: number;
    avg_latency_ms: number;
  }>;
  updated_at: string;
}

interface RevenueReportResponse {
  total: Money;
  period_start: string;
  period_end: string;
  granularity: string;
  data: Array<{
    period: string;
    date: string;
    amount: Money;
    invoice_count: number;
  }>;
}

interface CustomerRankingResponse {
  period_start: string;
  period_end: string;
  data: Array<{
    customer_id: string;
    name: string;
    email: string;
    revenue: Money;
    request_count: number;
    invoice_count: number;
    last_invoice_at?: string;
  }>;
}

const defaultRangeDays = 30;

const demoCustomers: CustomerMetric[] = [
  {
    customerId: 'cust-1032',
    name: 'Northwind Health',
    email: 'ops@northwind.health',
    revenue: { amount: 84200, currency: 'USD' },
    requestCount: 324000,
    invoiceCount: 5,
    lastInvoiceAt: '2025-01-22T08:00:00Z'
  },
  {
    customerId: 'cust-5521',
    name: 'Voltline Systems',
    email: 'finance@voltline.io',
    revenue: { amount: 69350, currency: 'USD' },
    requestCount: 298400,
    invoiceCount: 4,
    lastInvoiceAt: '2025-01-21T08:00:00Z'
  },
  {
    customerId: 'cust-2445',
    name: 'Nebula Retail',
    email: 'billing@nebula.store',
    revenue: { amount: 58400, currency: 'USD' },
    requestCount: 252100,
    invoiceCount: 4,
    lastInvoiceAt: '2025-01-20T08:00:00Z'
  },
  {
    customerId: 'cust-8870',
    name: 'Atlas Mobility',
    email: 'accounts@atlasmobility.com',
    revenue: { amount: 43120, currency: 'USD' },
    requestCount: 198900,
    invoiceCount: 3,
    lastInvoiceAt: '2025-01-18T08:00:00Z'
  },
  {
    customerId: 'cust-4401',
    name: 'Copperhead Labs',
    email: 'team@copperhead.ai',
    revenue: { amount: 32500, currency: 'USD' },
    requestCount: 156400,
    invoiceCount: 2,
    lastInvoiceAt: '2025-01-17T08:00:00Z'
  },
  {
    customerId: 'cust-9933',
    name: 'Juniper Grove',
    email: 'ops@junipergrove.co',
    revenue: { amount: 21240, currency: 'USD' },
    requestCount: 98100,
    invoiceCount: 2,
    lastInvoiceAt: '2025-01-15T08:00:00Z'
  }
];

const demoEndpoints: EndpointStat[] = [
  {
    endpoint: '/v1/usage/current',
    method: 'GET',
    requests: 72000,
    errorRate: 0.7,
    avgLatencyMs: 138
  },
  {
    endpoint: '/v1/usage/history',
    method: 'GET',
    requests: 64200,
    errorRate: 1.1,
    avgLatencyMs: 182
  },
  {
    endpoint: '/v1/subscription',
    method: 'GET',
    requests: 58800,
    errorRate: 0.4,
    avgLatencyMs: 120
  },
  {
    endpoint: '/v1/invoices',
    method: 'GET',
    requests: 50400,
    errorRate: 0.9,
    avgLatencyMs: 176
  },
  {
    endpoint: '/v1/api-keys',
    method: 'POST',
    requests: 33200,
    errorRate: 2.4,
    avgLatencyMs: 204
  }
];

function parseDate(value?: string): Date | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(parsed.getTime())) {
    return undefined;
  }
  return parsed;
}

function formatDate(value: Date): string {
  const year = value.getUTCFullYear();
  const month = String(value.getUTCMonth() + 1).padStart(2, '0');
  const day = String(value.getUTCDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function resolveRange(start?: string, end?: string, fallbackDays = defaultRangeDays) {
  const now = new Date();
  const endDate = parseDate(end) ?? now;
  let startDate = parseDate(start);
  if (!startDate) {
    startDate = new Date(endDate);
    startDate.setUTCDate(startDate.getUTCDate() - fallbackDays);
  }
  const msPerDay = 24 * 60 * 60 * 1000;
  if (startDate > endDate) {
    const swap = startDate;
    startDate = endDate;
    const days = Math.max(1, Math.round((swap.getTime() - startDate.getTime()) / msPerDay) + 1);
    return { startDate, endDate: swap, days };
  }
  const days = Math.max(1, Math.round((endDate.getTime() - startDate.getTime()) / msPerDay) + 1);
  return { startDate, endDate, days };
}

function formatPeriodLabel(value: Date, granularity: 'day' | 'month'): string {
  const options =
    granularity === 'month'
      ? { month: 'short', year: 'numeric' }
      : { month: 'short', day: 'numeric' };
  return new Intl.DateTimeFormat('en-US', options).format(value);
}

function generateRevenueSeries(start: Date, end: Date, granularity: 'day' | 'month'): RevenuePoint[] {
  const points: RevenuePoint[] = [];
  let cursor = new Date(Date.UTC(start.getUTCFullYear(), start.getUTCMonth(), start.getUTCDate()));
  const endDate = new Date(Date.UTC(end.getUTCFullYear(), end.getUTCMonth(), end.getUTCDate()));
  let index = 0;

  while (cursor <= endDate) {
    const wave = Math.sin(index / 1.8) * 2200;
    const trend = 14000 + index * 320;
    const amountValue = Math.max(4800, Math.round(trend + wave));
    const invoiceCount = Math.max(6, Math.round(amountValue / 1800));

    points.push({
      period: formatPeriodLabel(cursor, granularity),
      date: cursor.toISOString(),
      amount: { amount: amountValue, currency: 'USD' },
      invoiceCount
    });

    if (granularity === 'month') {
      cursor = new Date(Date.UTC(cursor.getUTCFullYear(), cursor.getUTCMonth() + 1, 1));
    } else {
      cursor.setUTCDate(cursor.getUTCDate() + 1);
    }
    index += 1;
  }

  return points;
}

function buildDemoState(options: UseAnalyticsOptions): AnalyticsState {
  const range = resolveRange(options.startDate, options.endDate, defaultRangeDays);
  const granularity =
    options.granularity ??
    (range.days > 90 ? 'month' : 'day');
  const revenueSeries = generateRevenueSeries(range.startDate, range.endDate, granularity);
  const totalAmount = revenueSeries.reduce((sum, point) => sum + point.amount.amount, 0);

  const now = new Date();
  const monthStart = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1));
  const monthSeries = generateRevenueSeries(monthStart, now, 'day');
  const monthTotal = monthSeries.reduce((sum, point) => sum + point.amount.amount, 0);
  const monthInvoices = monthSeries.reduce((sum, point) => sum + point.invoiceCount, 0);

  const rangeFactor = Math.max(0.35, Math.min(1.75, range.days / defaultRangeDays));

  const customers = demoCustomers.map((customer) => ({
    ...customer,
    revenue: {
      amount: Math.round(customer.revenue.amount * rangeFactor),
      currency: customer.revenue.currency
    },
    requestCount: Math.round(customer.requestCount * rangeFactor),
    invoiceCount: Math.max(1, Math.round(customer.invoiceCount * rangeFactor))
  }));

  const endpoints = demoEndpoints.map((endpoint) => ({
    ...endpoint,
    requests: Math.round(endpoint.requests * rangeFactor)
  }));

  return {
    overview: {
      revenueMtd: {
        amount: { amount: Math.round(monthTotal), currency: 'USD' },
        invoiceCount: monthInvoices,
        periodStart: monthStart.toISOString(),
        periodEnd: now.toISOString()
      },
      requestsToday: Math.round(18250 * rangeFactor),
      activeCustomers: 126,
      topEndpoints: endpoints,
      updatedAt: now.toISOString()
    },
    revenue: {
      total: { amount: Math.round(totalAmount), currency: 'USD' },
      periodStart: range.startDate.toISOString(),
      periodEnd: range.endDate.toISOString(),
      granularity,
      data: revenueSeries
    },
    customers: {
      periodStart: range.startDate.toISOString(),
      periodEnd: range.endDate.toISOString(),
      data: customers
    },
    isLoading: false,
    isDemo: true
  };
}

function mapOverview(payload: OverviewResponse): Overview {
  return {
    revenueMtd: {
      amount: payload.revenue_mtd.amount,
      invoiceCount: payload.revenue_mtd.invoice_count,
      periodStart: payload.revenue_mtd.period_start,
      periodEnd: payload.revenue_mtd.period_end
    },
    requestsToday: payload.requests_today,
    activeCustomers: payload.active_customers,
    topEndpoints: payload.top_endpoints.map((endpoint) => ({
      endpoint: endpoint.endpoint,
      method: endpoint.method,
      requests: endpoint.requests,
      errorRate: endpoint.error_rate,
      avgLatencyMs: endpoint.avg_latency_ms
    })),
    updatedAt: payload.updated_at
  };
}

function mapRevenue(payload: RevenueReportResponse): RevenueReport {
  return {
    total: payload.total,
    periodStart: payload.period_start,
    periodEnd: payload.period_end,
    granularity: payload.granularity,
    data: payload.data.map((point) => ({
      period: point.period,
      date: point.date,
      amount: point.amount,
      invoiceCount: point.invoice_count
    }))
  };
}

function mapCustomers(payload: CustomerRankingResponse): CustomerRanking {
  return {
    periodStart: payload.period_start,
    periodEnd: payload.period_end,
    data: payload.data.map((customer) => ({
      customerId: customer.customer_id,
      name: customer.name,
      email: customer.email,
      revenue: customer.revenue,
      requestCount: customer.request_count,
      invoiceCount: customer.invoice_count,
      lastInvoiceAt: customer.last_invoice_at
    }))
  };
}

export function useAnalytics(options: UseAnalyticsOptions = {}) {
  const baseUrl = options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL;
  const normalizedBaseUrl = baseUrl ? baseUrl.replace(/\/$/, '') : '';
  const startDate = options.startDate;
  const endDate = options.endDate;
  const windowDays = options.windowDays ?? 7;
  const limit = options.limit ?? 8;
  const range = useMemo(
    () => resolveRange(startDate, endDate, defaultRangeDays),
    [startDate, endDate]
  );
  const granularity =
    options.granularity ?? (range.days > 90 ? 'month' : 'day');
  const demoState = useMemo(
    () =>
      buildDemoState({
        startDate: formatDate(range.startDate),
        endDate: formatDate(range.endDate),
        granularity
      }),
    [range.startDate, range.endDate, granularity]
  );

  const [state, setState] = useState<AnalyticsState>(demoState);
  const [refreshIndex, setRefreshIndex] = useState(0);

  useEffect(() => {
    if (!normalizedBaseUrl) {
      setState(demoState);
      return;
    }

    const controller = new AbortController();

    async function loadAnalytics() {
      setState((prev) => ({
        ...prev,
        isLoading: true,
        isDemo: false,
        error: undefined
      }));

      try {
        const start = formatDate(range.startDate);
        const end = formatDate(range.endDate);
        const overviewParams = new URLSearchParams({
          window_days: String(windowDays)
        });
        const revenueParams = new URLSearchParams({
          start_date: start,
          end_date: end,
          granularity
        });
        const customerParams = new URLSearchParams({
          start_date: start,
          end_date: end,
          limit: String(limit)
        });

        const [overviewResponse, revenueResponse, customerResponse] = await Promise.all([
          fetch(`${normalizedBaseUrl}/v1/analytics/overview?${overviewParams.toString()}`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
            signal: controller.signal
          }),
          fetch(`${normalizedBaseUrl}/v1/analytics/revenue?${revenueParams.toString()}`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
            signal: controller.signal
          }),
          fetch(`${normalizedBaseUrl}/v1/analytics/customers?${customerParams.toString()}`, {
            headers: { Accept: 'application/json' },
            credentials: 'include',
            signal: controller.signal
          })
        ]);

        if (!overviewResponse.ok || !revenueResponse.ok || !customerResponse.ok) {
          throw new Error('Unable to load analytics data');
        }

        const [overviewPayload, revenuePayload, customerPayload] = await Promise.all([
          overviewResponse.json(),
          revenueResponse.json(),
          customerResponse.json()
        ]);

        setState({
          overview: mapOverview(overviewPayload as OverviewResponse),
          revenue: mapRevenue(revenuePayload as RevenueReportResponse),
          customers: mapCustomers(customerPayload as CustomerRankingResponse),
          isLoading: false,
          isDemo: false
        });
      } catch (error) {
        if (error instanceof DOMException && error.name === 'AbortError') {
          return;
        }
        setState((prev) => ({
          ...prev,
          isLoading: false,
          isDemo: false,
          error: 'Unable to load analytics right now.'
        }));
      }
    }

    loadAnalytics();

    return () => controller.abort();
  }, [
    normalizedBaseUrl,
    windowDays,
    limit,
    range.startDate,
    range.endDate,
    granularity,
    refreshIndex
  ]);

  const refresh = () => setRefreshIndex((prev) => prev + 1);

  return {
    overview: state.overview,
    revenue: state.revenue,
    customers: state.customers,
    isLoading: state.isLoading,
    isDemo: state.isDemo,
    error: state.error,
    range: {
      startDate: formatDate(range.startDate),
      endDate: formatDate(range.endDate),
      days: range.days
    },
    refresh
  };
}
