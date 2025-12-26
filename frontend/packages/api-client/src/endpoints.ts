import { APIClient, getDefaultClient } from './client';
import {
  APIKey,
  APIResponse,
  Customer,
  Invoice,
  Organization,
  PaginatedResponse,
  PaginationParams,
  Plan,
  SearchParams,
  Subscription,
  UsageRecord,
  UsageSummary,
  Webhook,
  WebhookDelivery,
} from './types';

// Helper to build query params
function buildParams(params?: PaginationParams | SearchParams): Record<string, string | number | boolean | undefined> {
  if (!params) return {};
  
  const result: Record<string, string | number | boolean | undefined> = {};
  
  if (params.page) result.page = params.page;
  if (params.pageSize) result.pageSize = params.pageSize;
  if (params.sortBy) result.sortBy = params.sortBy;
  if (params.sortOrder) result.sortOrder = params.sortOrder;
  
  if ('q' in params && params.q) {
    result.q = params.q;
  }
  
  if ('filters' in params && params.filters) {
    Object.entries(params.filters).forEach(([key, value]) => {
      result[`filter[${key}]`] = Array.isArray(value) ? value.join(',') : value;
    });
  }
  
  return result;
}

// =====================================
// Customers API
// =====================================

export const customersAPI = {
  list(params?: SearchParams, client?: APIClient): Promise<PaginatedResponse<Customer>> {
    const c = client || getDefaultClient();
    return c.get<Customer[]>('v1/customers', { params: buildParams(params) }) as Promise<PaginatedResponse<Customer>>;
  },

  get(id: string, client?: APIClient): Promise<APIResponse<Customer>> {
    const c = client || getDefaultClient();
    return c.get<Customer>(`v1/customers/${id}`);
  },

  create(data: Partial<Customer>, client?: APIClient): Promise<APIResponse<Customer>> {
    const c = client || getDefaultClient();
    return c.post<Customer>('v1/customers', data);
  },

  update(id: string, data: Partial<Customer>, client?: APIClient): Promise<APIResponse<Customer>> {
    const c = client || getDefaultClient();
    return c.patch<Customer>(`v1/customers/${id}`, data);
  },

  delete(id: string, client?: APIClient): Promise<APIResponse<void>> {
    const c = client || getDefaultClient();
    return c.delete<void>(`v1/customers/${id}`);
  },

  getByExternalId(externalId: string, client?: APIClient): Promise<APIResponse<Customer>> {
    const c = client || getDefaultClient();
    return c.get<Customer>(`v1/customers/external/${externalId}`);
  },
};

// =====================================
// API Keys API
// =====================================

export const apiKeysAPI = {
  list(customerId: string, params?: PaginationParams, client?: APIClient): Promise<PaginatedResponse<APIKey>> {
    const c = client || getDefaultClient();
    return c.get<APIKey[]>(`v1/customers/${customerId}/api-keys`, { params: buildParams(params) }) as Promise<PaginatedResponse<APIKey>>;
  },

  create(customerId: string, data: { name: string; scopes?: string[]; expiresAt?: string }, client?: APIClient): Promise<APIResponse<APIKey & { key: string }>> {
    const c = client || getDefaultClient();
    return c.post<APIKey & { key: string }>(`v1/customers/${customerId}/api-keys`, data);
  },

  revoke(customerId: string, keyId: string, client?: APIClient): Promise<APIResponse<void>> {
    const c = client || getDefaultClient();
    return c.delete<void>(`v1/customers/${customerId}/api-keys/${keyId}`);
  },

  rotate(customerId: string, keyId: string, client?: APIClient): Promise<APIResponse<APIKey & { key: string }>> {
    const c = client || getDefaultClient();
    return c.post<APIKey & { key: string }>(`v1/customers/${customerId}/api-keys/${keyId}/rotate`);
  },
};

// =====================================
// Subscriptions API
// =====================================

export const subscriptionsAPI = {
  list(customerId: string, params?: PaginationParams, client?: APIClient): Promise<PaginatedResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.get<Subscription[]>(`v1/customers/${customerId}/subscriptions`, { params: buildParams(params) }) as Promise<PaginatedResponse<Subscription>>;
  },

  get(customerId: string, subscriptionId: string, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.get<Subscription>(`v1/customers/${customerId}/subscriptions/${subscriptionId}`);
  },

  create(customerId: string, data: { planId: string; metadata?: Record<string, unknown> }, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.post<Subscription>(`v1/customers/${customerId}/subscriptions`, data);
  },

  update(customerId: string, subscriptionId: string, data: Partial<Subscription>, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.patch<Subscription>(`v1/customers/${customerId}/subscriptions/${subscriptionId}`, data);
  },

  cancel(customerId: string, subscriptionId: string, atPeriodEnd = true, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.post<Subscription>(`v1/customers/${customerId}/subscriptions/${subscriptionId}/cancel`, { atPeriodEnd });
  },

  reactivate(customerId: string, subscriptionId: string, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.post<Subscription>(`v1/customers/${customerId}/subscriptions/${subscriptionId}/reactivate`);
  },

  changePlan(customerId: string, subscriptionId: string, newPlanId: string, client?: APIClient): Promise<APIResponse<Subscription>> {
    const c = client || getDefaultClient();
    return c.post<Subscription>(`v1/customers/${customerId}/subscriptions/${subscriptionId}/change-plan`, { planId: newPlanId });
  },
};

// =====================================
// Plans API
// =====================================

export const plansAPI = {
  list(params?: SearchParams, client?: APIClient): Promise<PaginatedResponse<Plan>> {
    const c = client || getDefaultClient();
    return c.get<Plan[]>('v1/plans', { params: buildParams(params) }) as Promise<PaginatedResponse<Plan>>;
  },

  get(id: string, client?: APIClient): Promise<APIResponse<Plan>> {
    const c = client || getDefaultClient();
    return c.get<Plan>(`v1/plans/${id}`);
  },

  create(data: Partial<Plan>, client?: APIClient): Promise<APIResponse<Plan>> {
    const c = client || getDefaultClient();
    return c.post<Plan>('v1/plans', data);
  },

  update(id: string, data: Partial<Plan>, client?: APIClient): Promise<APIResponse<Plan>> {
    const c = client || getDefaultClient();
    return c.patch<Plan>(`v1/plans/${id}`, data);
  },

  archive(id: string, client?: APIClient): Promise<APIResponse<void>> {
    const c = client || getDefaultClient();
    return c.post<void>(`v1/plans/${id}/archive`);
  },
};

// =====================================
// Usage API
// =====================================

export const usageAPI = {
  record(customerId: string, data: { metricName: string; quantity: number; timestamp?: string; properties?: Record<string, unknown> }, client?: APIClient): Promise<APIResponse<UsageRecord>> {
    const c = client || getDefaultClient();
    return c.post<UsageRecord>(`v1/customers/${customerId}/usage`, data);
  },

  recordBatch(customerId: string, records: Array<{ metricName: string; quantity: number; timestamp?: string; properties?: Record<string, unknown> }>, client?: APIClient): Promise<APIResponse<{ recorded: number }>> {
    const c = client || getDefaultClient();
    return c.post<{ recorded: number }>(`v1/customers/${customerId}/usage/batch`, { records });
  },

  getSummary(customerId: string, params: { metricName: string; periodStart: string; periodEnd: string }, client?: APIClient): Promise<APIResponse<UsageSummary>> {
    const c = client || getDefaultClient();
    return c.get<UsageSummary>(`v1/customers/${customerId}/usage/summary`, { params });
  },

  getBreakdown(customerId: string, params: { metricName: string; periodStart: string; periodEnd: string; granularity?: 'hour' | 'day' | 'week' | 'month' }, client?: APIClient): Promise<APIResponse<UsageSummary>> {
    const c = client || getDefaultClient();
    return c.get<UsageSummary>(`v1/customers/${customerId}/usage/breakdown`, { params });
  },

  getCurrentUsage(customerId: string, client?: APIClient): Promise<APIResponse<UsageSummary[]>> {
    const c = client || getDefaultClient();
    return c.get<UsageSummary[]>(`v1/customers/${customerId}/usage/current`);
  },
};

// =====================================
// Invoices API
// =====================================

export const invoicesAPI = {
  list(customerId: string, params?: PaginationParams, client?: APIClient): Promise<PaginatedResponse<Invoice>> {
    const c = client || getDefaultClient();
    return c.get<Invoice[]>(`v1/customers/${customerId}/invoices`, { params: buildParams(params) }) as Promise<PaginatedResponse<Invoice>>;
  },

  get(customerId: string, invoiceId: string, client?: APIClient): Promise<APIResponse<Invoice>> {
    const c = client || getDefaultClient();
    return c.get<Invoice>(`v1/customers/${customerId}/invoices/${invoiceId}`);
  },

  getUpcoming(customerId: string, client?: APIClient): Promise<APIResponse<Invoice>> {
    const c = client || getDefaultClient();
    return c.get<Invoice>(`v1/customers/${customerId}/invoices/upcoming`);
  },

  pay(customerId: string, invoiceId: string, client?: APIClient): Promise<APIResponse<Invoice>> {
    const c = client || getDefaultClient();
    return c.post<Invoice>(`v1/customers/${customerId}/invoices/${invoiceId}/pay`);
  },

  void(customerId: string, invoiceId: string, client?: APIClient): Promise<APIResponse<Invoice>> {
    const c = client || getDefaultClient();
    return c.post<Invoice>(`v1/customers/${customerId}/invoices/${invoiceId}/void`);
  },

  downloadPdf(customerId: string, invoiceId: string, client?: APIClient): Promise<APIResponse<{ url: string }>> {
    const c = client || getDefaultClient();
    return c.get<{ url: string }>(`v1/customers/${customerId}/invoices/${invoiceId}/pdf`);
  },
};

// =====================================
// Organizations API
// =====================================

export const organizationsAPI = {
  getCurrent(client?: APIClient): Promise<APIResponse<Organization>> {
    const c = client || getDefaultClient();
    return c.get<Organization>('v1/organization');
  },

  update(data: Partial<Organization>, client?: APIClient): Promise<APIResponse<Organization>> {
    const c = client || getDefaultClient();
    return c.patch<Organization>('v1/organization', data);
  },

  getSettings(client?: APIClient): Promise<APIResponse<Record<string, unknown>>> {
    const c = client || getDefaultClient();
    return c.get<Record<string, unknown>>('v1/organization/settings');
  },

  updateSettings(settings: Record<string, unknown>, client?: APIClient): Promise<APIResponse<Record<string, unknown>>> {
    const c = client || getDefaultClient();
    return c.patch<Record<string, unknown>>('v1/organization/settings', settings);
  },
};

// =====================================
// Webhooks API
// =====================================

export const webhooksAPI = {
  list(params?: PaginationParams, client?: APIClient): Promise<PaginatedResponse<Webhook>> {
    const c = client || getDefaultClient();
    return c.get<Webhook[]>('v1/webhooks', { params: buildParams(params) }) as Promise<PaginatedResponse<Webhook>>;
  },

  get(id: string, client?: APIClient): Promise<APIResponse<Webhook>> {
    const c = client || getDefaultClient();
    return c.get<Webhook>(`v1/webhooks/${id}`);
  },

  create(data: { url: string; events: string[] }, client?: APIClient): Promise<APIResponse<Webhook>> {
    const c = client || getDefaultClient();
    return c.post<Webhook>('v1/webhooks', data);
  },

  update(id: string, data: Partial<Webhook>, client?: APIClient): Promise<APIResponse<Webhook>> {
    const c = client || getDefaultClient();
    return c.patch<Webhook>(`v1/webhooks/${id}`, data);
  },

  delete(id: string, client?: APIClient): Promise<APIResponse<void>> {
    const c = client || getDefaultClient();
    return c.delete<void>(`v1/webhooks/${id}`);
  },

  rotateSecret(id: string, client?: APIClient): Promise<APIResponse<{ secret: string }>> {
    const c = client || getDefaultClient();
    return c.post<{ secret: string }>(`v1/webhooks/${id}/rotate-secret`);
  },

  test(id: string, client?: APIClient): Promise<APIResponse<WebhookDelivery>> {
    const c = client || getDefaultClient();
    return c.post<WebhookDelivery>(`v1/webhooks/${id}/test`);
  },

  getDeliveries(id: string, params?: PaginationParams, client?: APIClient): Promise<PaginatedResponse<WebhookDelivery>> {
    const c = client || getDefaultClient();
    return c.get<WebhookDelivery[]>(`v1/webhooks/${id}/deliveries`, { params: buildParams(params) }) as Promise<PaginatedResponse<WebhookDelivery>>;
  },
};

// =====================================
// Analytics API
// =====================================

export interface AnalyticsSummary {
  totalRequests: number;
  uniqueCustomers: number;
  totalRevenue: number;
  avgResponseTime: number;
  errorRate: number;
  periodStart: string;
  periodEnd: string;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
}

export const analyticsAPI = {
  getSummary(params: { periodStart: string; periodEnd: string }, client?: APIClient): Promise<APIResponse<AnalyticsSummary>> {
    const c = client || getDefaultClient();
    return c.get<AnalyticsSummary>('v1/analytics/summary', { params });
  },

  getRequestsTimeSeries(params: { periodStart: string; periodEnd: string; granularity?: 'hour' | 'day' | 'week' | 'month' }, client?: APIClient): Promise<APIResponse<TimeSeriesData[]>> {
    const c = client || getDefaultClient();
    return c.get<TimeSeriesData[]>('v1/analytics/requests', { params });
  },

  getRevenueTimeSeries(params: { periodStart: string; periodEnd: string; granularity?: 'hour' | 'day' | 'week' | 'month' }, client?: APIClient): Promise<APIResponse<TimeSeriesData[]>> {
    const c = client || getDefaultClient();
    return c.get<TimeSeriesData[]>('v1/analytics/revenue', { params });
  },

  getTopCustomers(params: { periodStart: string; periodEnd: string; limit?: number }, client?: APIClient): Promise<APIResponse<Array<{ customer: Customer; requestCount: number; revenue: number }>>> {
    const c = client || getDefaultClient();
    return c.get<Array<{ customer: Customer; requestCount: number; revenue: number }>>('v1/analytics/top-customers', { params });
  },

  getTopEndpoints(params: { periodStart: string; periodEnd: string; limit?: number }, client?: APIClient): Promise<APIResponse<Array<{ endpoint: string; count: number; avgLatency: number }>>> {
    const c = client || getDefaultClient();
    return c.get<Array<{ endpoint: string; count: number; avgLatency: number }>>('v1/analytics/top-endpoints', { params });
  },
};
