import { APIClient, getDefaultClient } from '../client';
import { APIResponse } from '../types';
import { interpolatePath } from './runtime';

export namespace UsageApi {
  export interface CurrentUsage {
    customer_id: string;
    billing_period:   {
      start?: string;
      end?: string;
    };
    usage: UsageApi.UsageMetrics;
    quota: UsageApi.QuotaStatus;
    last_updated?: string;
  }

  export interface UsageMetrics {
    total_requests?: number;
    successful_requests?: number;
    failed_requests?: number;
    total_bandwidth_bytes?: number;
    avg_latency_ms?: number;
    p95_latency_ms?: number;
    p99_latency_ms?: number;
  }

  export interface QuotaStatus {
    requests?: UsageApi.QuotaItem;
    bandwidth_mb?: UsageApi.QuotaItem;
  }

  export interface QuotaItem {
    limit?: number;
    used?: number;
    remaining?: number;
    percentage_used?: number;
  }

  export interface UsageHistory {
    data?: UsageApi.UsageDataPoint[];
    pagination?: UsageApi.Pagination;
  }

  export interface UsageDataPoint {
    timestamp?: string;
    period?: string;
    metrics?: UsageApi.UsageMetrics;
  }

  export interface UsageBreakdown {
    total?: UsageApi.UsageMetrics;
    breakdown?: (  {
      key?: string;
      metrics?: UsageApi.UsageMetrics;
      percentage?: number;
    })[];
  }

  export interface Pagination {
    total?: number;
    limit?: number;
    offset?: number;
    has_more?: boolean;
  }

  export interface Error {
    error_code: string;
    message: string;
    details?: Record<string, unknown>;
  }

  export interface GetcurrentusageQueryParams {
    api_key_id?: string;
  }

  export interface GetusagehistoryQueryParams {
    start_date: string;
    end_date: string;
    api_key_id?: string;
    granularity?: "hourly" | "daily" | "weekly" | "monthly";
    limit?: number;
    offset?: number;
  }

  export interface GetusagebreakdownQueryParams {
    start_date: string;
    end_date: string;
    api_key_id?: string;
    group_by?: "endpoint" | "method" | "status_code" | "api_key";
  }

  export interface GetcurrentusageParams {
    query?: UsageApi.GetcurrentusageQueryParams;
  }

  export interface GetusagehistoryParams {
    query?: UsageApi.GetusagehistoryQueryParams;
  }

  export interface GetusagebreakdownParams {
    query?: UsageApi.GetusagebreakdownQueryParams;
  }

}

export const usageApiClient = {
  async getCurrentUsage(params: UsageApi.GetcurrentusageParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<UsageApi.CurrentUsage>> {
    const url = 'v1/usage/current';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< UsageApi.CurrentUsage >(url, { params: query });
  },
  async getUsageHistory(params: UsageApi.GetusagehistoryParams, client: APIClient = getDefaultClient()): Promise<APIResponse<UsageApi.UsageHistory>> {
    const url = 'v1/usage/history';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< UsageApi.UsageHistory >(url, { params: query });
  },
  async getRealtimeUsage(params: Record<string, never> = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<unknown>> {
    const url = 'v1/usage/realtime';
    const query = undefined;
    return client.get< unknown >(url, { params: query });
  },
  async getUsageBreakdown(params: UsageApi.GetusagebreakdownParams, client: APIClient = getDefaultClient()): Promise<APIResponse<UsageApi.UsageBreakdown>> {
    const url = 'v1/usage/breakdown';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< UsageApi.UsageBreakdown >(url, { params: query });
  },
};
