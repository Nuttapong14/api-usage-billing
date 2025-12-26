// Query key factory for type-safe query keys
// Based on TanStack Query best practices

export const queryKeys = {
  // Customers
  customers: {
    all: ['customers'] as const,
    lists: () => [...queryKeys.customers.all, 'list'] as const,
    list: (filters: Record<string, unknown>) =>
      [...queryKeys.customers.lists(), filters] as const,
    details: () => [...queryKeys.customers.all, 'detail'] as const,
    detail: (id: string) => [...queryKeys.customers.details(), id] as const,
    byExternalId: (externalId: string) =>
      [...queryKeys.customers.all, 'external', externalId] as const,
  },

  // API Keys
  apiKeys: {
    all: ['apiKeys'] as const,
    lists: () => [...queryKeys.apiKeys.all, 'list'] as const,
    list: (customerId: string, filters?: Record<string, unknown>) =>
      [...queryKeys.apiKeys.lists(), customerId, filters] as const,
    detail: (customerId: string, keyId: string) =>
      [...queryKeys.apiKeys.all, 'detail', customerId, keyId] as const,
  },

  // Subscriptions
  subscriptions: {
    all: ['subscriptions'] as const,
    lists: () => [...queryKeys.subscriptions.all, 'list'] as const,
    list: (customerId: string, filters?: Record<string, unknown>) =>
      [...queryKeys.subscriptions.lists(), customerId, filters] as const,
    detail: (customerId: string, subscriptionId: string) =>
      [...queryKeys.subscriptions.all, 'detail', customerId, subscriptionId] as const,
  },

  // Plans
  plans: {
    all: ['plans'] as const,
    lists: () => [...queryKeys.plans.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.plans.lists(), filters] as const,
    detail: (id: string) => [...queryKeys.plans.all, 'detail', id] as const,
  },

  // Usage
  usage: {
    all: ['usage'] as const,
    summary: (customerId: string, params: { metricName: string; periodStart: string; periodEnd: string }) =>
      [...queryKeys.usage.all, 'summary', customerId, params] as const,
    breakdown: (customerId: string, params: Record<string, unknown>) =>
      [...queryKeys.usage.all, 'breakdown', customerId, params] as const,
    current: (customerId: string) =>
      [...queryKeys.usage.all, 'current', customerId] as const,
  },

  // Invoices
  invoices: {
    all: ['invoices'] as const,
    lists: () => [...queryKeys.invoices.all, 'list'] as const,
    list: (customerId: string, filters?: Record<string, unknown>) =>
      [...queryKeys.invoices.lists(), customerId, filters] as const,
    detail: (customerId: string, invoiceId: string) =>
      [...queryKeys.invoices.all, 'detail', customerId, invoiceId] as const,
    upcoming: (customerId: string) =>
      [...queryKeys.invoices.all, 'upcoming', customerId] as const,
  },

  // Organizations
  organization: {
    all: ['organization'] as const,
    current: () => [...queryKeys.organization.all, 'current'] as const,
    settings: () => [...queryKeys.organization.all, 'settings'] as const,
  },

  // Webhooks
  webhooks: {
    all: ['webhooks'] as const,
    lists: () => [...queryKeys.webhooks.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.webhooks.lists(), filters] as const,
    detail: (id: string) => [...queryKeys.webhooks.all, 'detail', id] as const,
    deliveries: (webhookId: string, filters?: Record<string, unknown>) =>
      [...queryKeys.webhooks.all, 'deliveries', webhookId, filters] as const,
  },

  // Analytics
  analytics: {
    all: ['analytics'] as const,
    summary: (params: { periodStart: string; periodEnd: string }) =>
      [...queryKeys.analytics.all, 'summary', params] as const,
    requests: (params: Record<string, unknown>) =>
      [...queryKeys.analytics.all, 'requests', params] as const,
    revenue: (params: Record<string, unknown>) =>
      [...queryKeys.analytics.all, 'revenue', params] as const,
    topCustomers: (params: Record<string, unknown>) =>
      [...queryKeys.analytics.all, 'topCustomers', params] as const,
    topEndpoints: (params: Record<string, unknown>) =>
      [...queryKeys.analytics.all, 'topEndpoints', params] as const,
  },
} as const;

// Type helper for extracting query key types
export type QueryKeys = typeof queryKeys;
