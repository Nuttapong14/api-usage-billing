'use client';

import { useQuery, useMutation, useQueryClient, UseQueryOptions, UseMutationOptions } from '@tanstack/react-query';
import {
  customersAPI,
  apiKeysAPI,
  subscriptionsAPI,
  plansAPI,
  usageAPI,
  invoicesAPI,
  organizationsAPI,
  webhooksAPI,
  analyticsAPI,
  APIResponse,
  PaginatedResponse,
  Customer,
  APIKey,
  Subscription,
  Plan,
  UsageSummary,
  Invoice,
  Organization,
  Webhook,
  WebhookDelivery,
  SearchParams,
  PaginationParams,
} from '@api-billing/api-client';
import { queryKeys } from './keys';

// =====================================
// Customer Hooks
// =====================================

export function useCustomers(
  params?: SearchParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<Customer>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.customers.list(params || {}),
    queryFn: () => customersAPI.list(params),
    ...options,
  });
}

export function useCustomer(
  id: string,
  options?: Omit<UseQueryOptions<APIResponse<Customer>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.customers.detail(id),
    queryFn: () => customersAPI.get(id),
    enabled: !!id,
    ...options,
  });
}

export function useCustomerByExternalId(
  externalId: string,
  options?: Omit<UseQueryOptions<APIResponse<Customer>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.customers.byExternalId(externalId),
    queryFn: () => customersAPI.getByExternalId(externalId),
    enabled: !!externalId,
    ...options,
  });
}

export function useCreateCustomer(
  options?: UseMutationOptions<APIResponse<Customer>, Error, Partial<Customer>>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Customer>) => customersAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.customers.all });
    },
    ...options,
  });
}

export function useUpdateCustomer(
  options?: UseMutationOptions<APIResponse<Customer>, Error, { id: string; data: Partial<Customer> }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }) => customersAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.customers.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.customers.lists() });
    },
    ...options,
  });
}

export function useDeleteCustomer(
  options?: UseMutationOptions<APIResponse<void>, Error, string>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => customersAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.customers.all });
    },
    ...options,
  });
}

// =====================================
// API Key Hooks
// =====================================

export function useAPIKeys(
  customerId: string,
  params?: PaginationParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<APIKey>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.apiKeys.list(customerId, params),
    queryFn: () => apiKeysAPI.list(customerId, params),
    enabled: !!customerId,
    ...options,
  });
}

export function useCreateAPIKey(
  options?: UseMutationOptions<
    APIResponse<APIKey & { key: string }>,
    Error,
    { customerId: string; data: { name: string; scopes?: string[]; expiresAt?: string } }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ customerId, data }) => apiKeysAPI.create(customerId, data),
    onSuccess: (_, { customerId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.list(customerId) });
    },
    ...options,
  });
}

export function useRevokeAPIKey(
  options?: UseMutationOptions<APIResponse<void>, Error, { customerId: string; keyId: string }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ customerId, keyId }) => apiKeysAPI.revoke(customerId, keyId),
    onSuccess: (_, { customerId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeys.list(customerId) });
    },
    ...options,
  });
}

// =====================================
// Subscription Hooks
// =====================================

export function useSubscriptions(
  customerId: string,
  params?: PaginationParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<Subscription>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.subscriptions.list(customerId, params),
    queryFn: () => subscriptionsAPI.list(customerId, params),
    enabled: !!customerId,
    ...options,
  });
}

export function useSubscription(
  customerId: string,
  subscriptionId: string,
  options?: Omit<UseQueryOptions<APIResponse<Subscription>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.subscriptions.detail(customerId, subscriptionId),
    queryFn: () => subscriptionsAPI.get(customerId, subscriptionId),
    enabled: !!customerId && !!subscriptionId,
    ...options,
  });
}

export function useCreateSubscription(
  options?: UseMutationOptions<
    APIResponse<Subscription>,
    Error,
    { customerId: string; data: { planId: string; metadata?: Record<string, unknown> } }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ customerId, data }) => subscriptionsAPI.create(customerId, data),
    onSuccess: (_, { customerId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.subscriptions.list(customerId) });
    },
    ...options,
  });
}

export function useCancelSubscription(
  options?: UseMutationOptions<
    APIResponse<Subscription>,
    Error,
    { customerId: string; subscriptionId: string; atPeriodEnd?: boolean }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ customerId, subscriptionId, atPeriodEnd }) =>
      subscriptionsAPI.cancel(customerId, subscriptionId, atPeriodEnd),
    onSuccess: (_, { customerId, subscriptionId }) => {
      queryClient.invalidateQueries({
        queryKey: queryKeys.subscriptions.detail(customerId, subscriptionId),
      });
      queryClient.invalidateQueries({ queryKey: queryKeys.subscriptions.list(customerId) });
    },
    ...options,
  });
}

// =====================================
// Plan Hooks
// =====================================

export function usePlans(
  params?: SearchParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<Plan>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.plans.list(params),
    queryFn: () => plansAPI.list(params),
    ...options,
  });
}

export function usePlan(
  id: string,
  options?: Omit<UseQueryOptions<APIResponse<Plan>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.plans.detail(id),
    queryFn: () => plansAPI.get(id),
    enabled: !!id,
    ...options,
  });
}

// =====================================
// Usage Hooks
// =====================================

export function useUsageSummary(
  customerId: string,
  params: { metricName: string; periodStart: string; periodEnd: string },
  options?: Omit<UseQueryOptions<APIResponse<UsageSummary>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.usage.summary(customerId, params),
    queryFn: () => usageAPI.getSummary(customerId, params),
    enabled: !!customerId && !!params.metricName,
    ...options,
  });
}

export function useCurrentUsage(
  customerId: string,
  options?: Omit<UseQueryOptions<APIResponse<UsageSummary[]>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.usage.current(customerId),
    queryFn: () => usageAPI.getCurrentUsage(customerId),
    enabled: !!customerId,
    ...options,
  });
}

export function useRecordUsage(
  options?: UseMutationOptions<
    APIResponse<{ recorded: number }>,
    Error,
    {
      customerId: string;
      records: Array<{
        metricName: string;
        quantity: number;
        timestamp?: string;
        properties?: Record<string, unknown>;
      }>;
    }
  >
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ customerId, records }) => usageAPI.recordBatch(customerId, records),
    onSuccess: (_, { customerId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.usage.current(customerId) });
    },
    ...options,
  });
}

// =====================================
// Invoice Hooks
// =====================================

export function useInvoices(
  customerId: string,
  params?: PaginationParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<Invoice>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.invoices.list(customerId, params),
    queryFn: () => invoicesAPI.list(customerId, params),
    enabled: !!customerId,
    ...options,
  });
}

export function useInvoice(
  customerId: string,
  invoiceId: string,
  options?: Omit<UseQueryOptions<APIResponse<Invoice>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.invoices.detail(customerId, invoiceId),
    queryFn: () => invoicesAPI.get(customerId, invoiceId),
    enabled: !!customerId && !!invoiceId,
    ...options,
  });
}

export function useUpcomingInvoice(
  customerId: string,
  options?: Omit<UseQueryOptions<APIResponse<Invoice>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.invoices.upcoming(customerId),
    queryFn: () => invoicesAPI.getUpcoming(customerId),
    enabled: !!customerId,
    ...options,
  });
}

// =====================================
// Organization Hooks
// =====================================

export function useOrganization(
  options?: Omit<UseQueryOptions<APIResponse<Organization>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.organization.current(),
    queryFn: () => organizationsAPI.getCurrent(),
    ...options,
  });
}

export function useOrganizationSettings(
  options?: Omit<UseQueryOptions<APIResponse<Record<string, unknown>>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.organization.settings(),
    queryFn: () => organizationsAPI.getSettings(),
    ...options,
  });
}

export function useUpdateOrganization(
  options?: UseMutationOptions<APIResponse<Organization>, Error, Partial<Organization>>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data) => organizationsAPI.update(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.organization.current() });
    },
    ...options,
  });
}

// =====================================
// Webhook Hooks
// =====================================

export function useWebhooks(
  params?: PaginationParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<Webhook>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.webhooks.list(params),
    queryFn: () => webhooksAPI.list(params),
    ...options,
  });
}

export function useWebhook(
  id: string,
  options?: Omit<UseQueryOptions<APIResponse<Webhook>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.webhooks.detail(id),
    queryFn: () => webhooksAPI.get(id),
    enabled: !!id,
    ...options,
  });
}

export function useWebhookDeliveries(
  webhookId: string,
  params?: PaginationParams,
  options?: Omit<UseQueryOptions<PaginatedResponse<WebhookDelivery>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.webhooks.deliveries(webhookId, params),
    queryFn: () => webhooksAPI.getDeliveries(webhookId, params),
    enabled: !!webhookId,
    ...options,
  });
}

export function useCreateWebhook(
  options?: UseMutationOptions<APIResponse<Webhook>, Error, { url: string; events: string[] }>
) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data) => webhooksAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.webhooks.all });
    },
    ...options,
  });
}

export function useTestWebhook(
  options?: UseMutationOptions<APIResponse<WebhookDelivery>, Error, string>
) {
  return useMutation({
    mutationFn: (id) => webhooksAPI.test(id),
    ...options,
  });
}

// =====================================
// Analytics Hooks
// =====================================

export function useAnalyticsSummary(
  params: { periodStart: string; periodEnd: string },
  options?: Omit<UseQueryOptions<APIResponse<ReturnType<typeof analyticsAPI.getSummary> extends Promise<infer R> ? R extends APIResponse<infer T> ? T : never : never>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.analytics.summary(params),
    queryFn: () => analyticsAPI.getSummary(params),
    enabled: !!params.periodStart && !!params.periodEnd,
    ...options,
  });
}

export function useRequestsTimeSeries(
  params: {
    periodStart: string;
    periodEnd: string;
    granularity?: 'hour' | 'day' | 'week' | 'month';
  },
  options?: Omit<UseQueryOptions<APIResponse<Array<{ timestamp: string; value: number }>>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.analytics.requests(params),
    queryFn: () => analyticsAPI.getRequestsTimeSeries(params),
    enabled: !!params.periodStart && !!params.periodEnd,
    ...options,
  });
}

export function useRevenueTimeSeries(
  params: {
    periodStart: string;
    periodEnd: string;
    granularity?: 'hour' | 'day' | 'week' | 'month';
  },
  options?: Omit<UseQueryOptions<APIResponse<Array<{ timestamp: string; value: number }>>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.analytics.revenue(params),
    queryFn: () => analyticsAPI.getRevenueTimeSeries(params),
    enabled: !!params.periodStart && !!params.periodEnd,
    ...options,
  });
}

export function useTopCustomers(
  params: { periodStart: string; periodEnd: string; limit?: number },
  options?: Omit<UseQueryOptions<APIResponse<Array<{ customer: Customer; requestCount: number; revenue: number }>>>, 'queryKey' | 'queryFn'>
) {
  return useQuery({
    queryKey: queryKeys.analytics.topCustomers(params),
    queryFn: () => analyticsAPI.getTopCustomers(params),
    enabled: !!params.periodStart && !!params.periodEnd,
    ...options,
  });
}
