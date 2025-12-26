import { APIClient, getDefaultClient } from '../client';
import { APIResponse } from '../types';
import { interpolatePath } from './runtime';

export namespace SubscriptionApi {
  export interface Subscription {
    id: string;
    customer_id: string;
    tier: SubscriptionApi.Tier;
    status: "active" | "paused" | "cancelled" | "expired";
    billing_cycle: "monthly" | "yearly";
    current_period:   {
      start?: string;
      end?: string;
    };
    usage?: SubscriptionApi.UsageSummary;
    pending_change?: SubscriptionApi.PendingChange;
    trial_ends_at?: string;
    cancelled_at?: string;
    created_at?: string;
  }

  export interface Tier {
    id: string;
    name: string;
    slug?: string;
    description?: string;
    pricing: SubscriptionApi.TierPricing;
    quotas: SubscriptionApi.TierQuotas;
    rate_limits?: SubscriptionApi.TierRateLimits;
    features?: Record<string, unknown>;
    sla_percentage?: number;
    is_current?: boolean;
  }

  export interface TierPricing {
    monthly?: SubscriptionApi.Money;
    yearly?: SubscriptionApi.Money;
    overage?:   {
      per_request?: SubscriptionApi.Money;
      per_mb?: SubscriptionApi.Money;
    };
  }

  export interface TierQuotas {
    requests?: number;
    bandwidth_mb?: number;
    compute_seconds?: number;
  }

  export interface TierRateLimits {
    per_second?: number;
    per_minute?: number;
    burst?: number;
  }

  export interface TierList {
    tiers?: SubscriptionApi.Tier[];
  }

  export interface UsageSummary {
    requests?:   {
      used?: number;
      limit?: number;
      percentage?: number;
    };
    bandwidth_mb?:   {
      used?: number;
      limit?: number;
      percentage?: number;
    };
  }

  export interface PendingChange {
    type?: "upgrade" | "downgrade" | "cancel";
    new_tier?: SubscriptionApi.Tier;
    effective_date?: string;
    created_at?: string;
  }

  export interface UpgradeRequest {
    tier_id: string;
    billing_cycle?: "monthly" | "yearly";
  }

  export interface DowngradeRequest {
    tier_id: string;
    billing_cycle?: "monthly" | "yearly";
  }

  export interface CancelRequest {
    reason?: string;
    feedback?: string;
  }

  export interface SubscriptionChangeResponse {
    subscription?: SubscriptionApi.Subscription;
    change_type?: "immediate" | "scheduled";
    effective_date?: string;
    proration?: SubscriptionApi.Proration;
    message?: string;
  }

  export interface Proration {
    credit?: SubscriptionApi.Money;
    charge?: SubscriptionApi.Money;
    net?: SubscriptionApi.Money;
  }

  export interface ChangePreview {
    current_tier?: SubscriptionApi.Tier;
    new_tier?: SubscriptionApi.Tier;
    change_type?: "upgrade" | "downgrade";
    effective?: "immediate" | "next_period";
    effective_date?: string;
    proration?: SubscriptionApi.Proration;
    new_billing_amount?: SubscriptionApi.Money;
  }

  export interface PaymentRequiredError {
    error_code?: string;
    message?: string;
    amount_due?: SubscriptionApi.Money;
    payment_url?: string;
  }

  export interface Money {
    amount?: number;
    currency?: string;
  }

  export interface Error {
    error_code: string;
    message: string;
    details?: Record<string, unknown>;
  }

  export interface GettierPathParams {
    id: string;
  }

  export interface ListavailabletiersQueryParams {
    include_current?: boolean;
  }

  export interface ListavailabletiersParams {
    query?: SubscriptionApi.ListavailabletiersQueryParams;
  }

  export interface GettierParams {
    path: SubscriptionApi.GettierPathParams;
  }

  export interface UpgradesubscriptionParams {
    body: SubscriptionApi.UpgradeRequest;
  }

  export interface DowngradesubscriptionParams {
    body: SubscriptionApi.DowngradeRequest;
  }

  export interface CancelsubscriptionParams {
    body?: SubscriptionApi.CancelRequest;
  }

  export interface PreviewsubscriptionchangeParams {
    body: {
    tier_id: string;
    billing_cycle?: "monthly" | "yearly";
  };
  }

}

export const subscriptionApiClient = {
  async getCurrentSubscription(params: Record<string, never> = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.Subscription>> {
    const url = 'v1/subscription';
    const query = undefined;
    return client.get< SubscriptionApi.Subscription >(url, { params: query });
  },
  async listAvailableTiers(params: SubscriptionApi.ListavailabletiersParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.TierList>> {
    const url = 'v1/subscription/tiers';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< SubscriptionApi.TierList >(url, { params: query });
  },
  async getTier(params: SubscriptionApi.GettierParams, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.Tier>> {
    const url = interpolatePath('v1/subscription/tiers/{id}', params.path);
    const query = undefined;
    return client.get< SubscriptionApi.Tier >(url, { params: query });
  },
  async upgradeSubscription(params: SubscriptionApi.UpgradesubscriptionParams, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.SubscriptionChangeResponse>> {
    const url = 'v1/subscription/upgrade';
    const query = undefined;
    return client.post< SubscriptionApi.SubscriptionChangeResponse >(url, params.body, { params: query });
  },
  async downgradeSubscription(params: SubscriptionApi.DowngradesubscriptionParams, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.SubscriptionChangeResponse>> {
    const url = 'v1/subscription/downgrade';
    const query = undefined;
    return client.post< SubscriptionApi.SubscriptionChangeResponse >(url, params.body, { params: query });
  },
  async cancelSubscription(params: SubscriptionApi.CancelsubscriptionParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.SubscriptionChangeResponse>> {
    const url = 'v1/subscription/cancel';
    const query = undefined;
    return client.post< SubscriptionApi.SubscriptionChangeResponse >(url, params.body, { params: query });
  },
  async reactivateSubscription(params: Record<string, never> = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.Subscription>> {
    const url = 'v1/subscription/reactivate';
    const query = undefined;
    return client.post< SubscriptionApi.Subscription >(url, undefined, { params: query });
  },
  async previewSubscriptionChange(params: SubscriptionApi.PreviewsubscriptionchangeParams, client: APIClient = getDefaultClient()): Promise<APIResponse<SubscriptionApi.ChangePreview>> {
    const url = 'v1/subscription/preview-change';
    const query = undefined;
    return client.post< SubscriptionApi.ChangePreview >(url, params.body, { params: query });
  },
};
