import { APIClient, getDefaultClient } from '../client';
import { APIResponse } from '../types';
import { interpolatePath } from './runtime';

export namespace WebhookApi {
  export interface Webhook {
    id: string;
    url: string;
    description?: string;
    events: string[];
    is_active: boolean;
    last_triggered_at?: string;
    consecutive_failures?: number;
    created_at: string;
    updated_at?: string;
  }

  export type WebhookWithSecret = WebhookApi.Webhook & {
    secret?: string;
  };

  export interface WebhookList {
    data?: WebhookApi.Webhook[];
    pagination?: WebhookApi.Pagination;
  }

  export interface CreateWebhookRequest {
    url: string;
    description?: string;
    events: string[];
  }

  export interface UpdateWebhookRequest {
    url?: string;
    description?: string;
    events?: string[];
    is_active?: boolean;
  }

  export interface TestWebhookResponse {
    delivery_id?: string;
    event_type?: string;
    status?: "pending" | "delivered" | "failed";
    response_status?: number;
    response_time_ms?: number;
    message?: string;
  }

  export interface Delivery {
    id?: string;
    endpoint_id?: string;
    event_type?: string;
    event_id?: string;
    payload?:   {
    };
    status?: "pending" | "delivered" | "failed";
    attempts?: number;
    max_attempts?: number;
    response_status?: number;
    response_body?: string;
    response_time_ms?: number;
    next_retry_at?: string;
    delivered_at?: string;
    failed_at?: string;
    failure_reason?: string;
    created_at?: string;
  }

  export interface DeliveryList {
    data?: WebhookApi.Delivery[];
    pagination?: WebhookApi.Pagination;
  }

  export interface EventType {
    name?: string;
    description?: string;
    category?: string;
    payload_schema?:   {
    };
  }

  export interface EventTypeList {
    events?: WebhookApi.EventType[];
  }

  export interface WebhookPayload {
    id?: string;
    type?: string;
    created_at?: string;
    data?:   {
    };
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

  export interface GetwebhookPathParams {
    id: string;
  }

  export interface UpdatewebhookPathParams {
    id: string;
  }

  export interface DeletewebhookPathParams {
    id: string;
  }

  export interface TestwebhookPathParams {
    id: string;
  }

  export interface RegeneratesecretPathParams {
    id: string;
  }

  export interface ListdeliveriesPathParams {
    id: string;
  }

  export interface GetdeliveryPathParams {
    id: string;
    delivery_id: string;
  }

  export interface RetrydeliveryPathParams {
    id: string;
    delivery_id: string;
  }

  export interface ListwebhooksQueryParams {
    is_active?: boolean;
  }

  export interface ListdeliveriesQueryParams {
    status?: "pending" | "delivered" | "failed";
    event_type?: string;
    limit?: number;
    offset?: number;
  }

  export interface ListwebhooksParams {
    query?: WebhookApi.ListwebhooksQueryParams;
  }

  export interface CreatewebhookParams {
    body: WebhookApi.CreateWebhookRequest;
  }

  export interface GetwebhookParams {
    path: WebhookApi.GetwebhookPathParams;
  }

  export interface UpdatewebhookParams {
    path: WebhookApi.UpdatewebhookPathParams;
    body: WebhookApi.UpdateWebhookRequest;
  }

  export interface DeletewebhookParams {
    path: WebhookApi.DeletewebhookPathParams;
  }

  export interface TestwebhookParams {
    path: WebhookApi.TestwebhookPathParams;
    body?: {
    event_type?: "test.webhook" | "usage.quota_warning" | "usage.quota_exceeded" | "invoice.created" | "invoice.paid" | "payment.received" | "subscription.upgraded" | "subscription.downgraded" | "subscription.cancelled" | "api_key.created" | "api_key.revoked";
  };
  }

  export interface RegeneratesecretParams {
    path: WebhookApi.RegeneratesecretPathParams;
  }

  export interface ListdeliveriesParams {
    path: WebhookApi.ListdeliveriesPathParams;
    query?: WebhookApi.ListdeliveriesQueryParams;
  }

  export interface GetdeliveryParams {
    path: WebhookApi.GetdeliveryPathParams;
  }

  export interface RetrydeliveryParams {
    path: WebhookApi.RetrydeliveryPathParams;
  }

}

export const webhookApiClient = {
  async listWebhooks(params: WebhookApi.ListwebhooksParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.WebhookList>> {
    const url = 'v1/webhooks';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< WebhookApi.WebhookList >(url, { params: query });
  },
  async createWebhook(params: WebhookApi.CreatewebhookParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.WebhookWithSecret>> {
    const url = 'v1/webhooks';
    const query = undefined;
    return client.post< WebhookApi.WebhookWithSecret >(url, params.body, { params: query });
  },
  async getWebhook(params: WebhookApi.GetwebhookParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.Webhook>> {
    const url = interpolatePath('v1/webhooks/{id}', params.path);
    const query = undefined;
    return client.get< WebhookApi.Webhook >(url, { params: query });
  },
  async updateWebhook(params: WebhookApi.UpdatewebhookParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.Webhook>> {
    const url = interpolatePath('v1/webhooks/{id}', params.path);
    const query = undefined;
    return client.patch< WebhookApi.Webhook >(url, params.body, { params: query });
  },
  async deleteWebhook(params: WebhookApi.DeletewebhookParams, client: APIClient = getDefaultClient()): Promise<APIResponse<unknown>> {
    const url = interpolatePath('v1/webhooks/{id}', params.path);
    const query = undefined;
    return client.delete< unknown >(url, { params: query });
  },
  async testWebhook(params: WebhookApi.TestwebhookParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.TestWebhookResponse>> {
    const url = interpolatePath('v1/webhooks/{id}/test', params.path);
    const query = undefined;
    return client.post< WebhookApi.TestWebhookResponse >(url, params.body, { params: query });
  },
  async regenerateSecret(params: WebhookApi.RegeneratesecretParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.WebhookWithSecret>> {
    const url = interpolatePath('v1/webhooks/{id}/secret', params.path);
    const query = undefined;
    return client.post< WebhookApi.WebhookWithSecret >(url, undefined, { params: query });
  },
  async listDeliveries(params: WebhookApi.ListdeliveriesParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.DeliveryList>> {
    const url = interpolatePath('v1/webhooks/{id}/deliveries', params.path);
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< WebhookApi.DeliveryList >(url, { params: query });
  },
  async getDelivery(params: WebhookApi.GetdeliveryParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.Delivery>> {
    const url = interpolatePath('v1/webhooks/{id}/deliveries/{delivery_id}', params.path);
    const query = undefined;
    return client.get< WebhookApi.Delivery >(url, { params: query });
  },
  async retryDelivery(params: WebhookApi.RetrydeliveryParams, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.Delivery>> {
    const url = interpolatePath('v1/webhooks/{id}/deliveries/{delivery_id}/retry', params.path);
    const query = undefined;
    return client.post< WebhookApi.Delivery >(url, undefined, { params: query });
  },
  async listEventTypes(params: Record<string, never> = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<WebhookApi.EventTypeList>> {
    const url = 'v1/webhooks/events';
    const query = undefined;
    return client.get< WebhookApi.EventTypeList >(url, { params: query });
  },
};
