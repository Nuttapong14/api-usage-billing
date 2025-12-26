// API Response Types
export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: APIError;
  meta?: ResponseMeta;
}

export interface APIError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
  requestId?: string;
}

export interface ResponseMeta {
  requestId: string;
  timestamp: string;
  pagination?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface PaginatedResponse<T> extends APIResponse<T[]> {
  meta: ResponseMeta & {
    pagination: PaginationMeta;
  };
}

// Request Types
export interface PaginationParams {
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface SearchParams extends PaginationParams {
  q?: string;
  filters?: Record<string, string | string[]>;
}

// Client Configuration
export interface APIClientConfig {
  baseUrl: string;
  timeout?: number;
  retries?: number;
  getAccessToken?: () => string | null | Promise<string | null>;
  onUnauthorized?: () => void;
  onError?: (error: APIError) => void;
  requestIdHeader?: string;
  organizationIdHeader?: string;
  getOrganizationId?: () => string | null;
}

// HTTP Method Types
export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export interface RequestOptions {
  headers?: Record<string, string>;
  params?: Record<string, string | number | boolean | undefined>;
  timeout?: number;
  signal?: AbortSignal;
  skipAuth?: boolean;
}

// Resource Types (for type-safe endpoints)
export interface Customer {
  id: string;
  organizationId: string;
  externalId?: string;
  name: string;
  email: string;
  status: 'active' | 'inactive' | 'suspended';
  metadata?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface APIKey {
  id: string;
  customerId: string;
  name: string;
  prefix: string;
  lastUsedAt?: string;
  expiresAt?: string;
  status: 'active' | 'revoked' | 'expired';
  scopes: string[];
  createdAt: string;
}

export interface Subscription {
  id: string;
  customerId: string;
  planId: string;
  status: 'active' | 'canceled' | 'past_due' | 'trialing';
  currentPeriodStart: string;
  currentPeriodEnd: string;
  cancelAtPeriodEnd: boolean;
  metadata?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

export interface Plan {
  id: string;
  name: string;
  description?: string;
  isActive: boolean;
  features: PlanFeature[];
  basePrice: number;
  currency: string;
  billingPeriod: 'monthly' | 'yearly';
  createdAt: string;
  updatedAt: string;
}

export interface PlanFeature {
  id: string;
  name: string;
  key: string;
  type: 'boolean' | 'limit' | 'usage';
  value?: number;
  unit?: string;
}

export interface UsageRecord {
  id: string;
  customerId: string;
  metricName: string;
  quantity: number;
  timestamp: string;
  properties?: Record<string, unknown>;
}

export interface UsageSummary {
  metricName: string;
  total: number;
  periodStart: string;
  periodEnd: string;
  breakdown?: UsageBreakdown[];
}

export interface UsageBreakdown {
  date: string;
  quantity: number;
}

export interface Invoice {
  id: string;
  customerId: string;
  subscriptionId?: string;
  status: 'draft' | 'open' | 'paid' | 'void' | 'uncollectible';
  amountDue: number;
  amountPaid: number;
  currency: string;
  dueDate: string;
  paidAt?: string;
  lineItems: InvoiceLineItem[];
  createdAt: string;
}

export interface InvoiceLineItem {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  amount: number;
  type: 'subscription' | 'usage' | 'one_time';
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  settings?: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

// Webhook Types
export interface Webhook {
  id: string;
  url: string;
  events: string[];
  status: 'active' | 'disabled';
  secret?: string;
  createdAt: string;
}

export interface WebhookDelivery {
  id: string;
  webhookId: string;
  event: string;
  statusCode: number;
  success: boolean;
  requestBody: string;
  responseBody?: string;
  deliveredAt: string;
}
