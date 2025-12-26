import { APIClient, getDefaultClient } from '../client';
import { APIResponse } from '../types';
import { interpolatePath } from './runtime';

export namespace ApikeyApi {
  export interface ApiKey {
    id: string;
    key_prefix: string;
    name: string;
    description?: string;
    permissions: ("read" | "write" | "admin")[];
    scopes?: string[];
    ip_whitelist?: string[];
    allowed_origins?: string[];
    expires_at?: string;
    last_used_at?: string;
    last_used_ip?: string;
    is_active: boolean;
    is_rotating?: boolean;
    rotation_grace_until?: string;
    created_at: string;
    revoked_at?: string;
  }

  export interface ApiKeyList {
    data?: ApikeyApi.ApiKey[];
    pagination?: ApikeyApi.Pagination;
  }

  export interface CreateApiKeyRequest {
    name: string;
    description?: string;
    permissions?: ("read" | "write" | "admin")[];
    scopes?: string[];
    ip_whitelist?: string[];
    allowed_origins?: string[];
    expires_at?: string;
  }

  export interface CreateApiKeyResponse {
    key: string;
    api_key: ApikeyApi.ApiKey;
    warning?: string;
  }

  export interface UpdateApiKeyRequest {
    name?: string;
    description?: string;
    permissions?: ("read" | "write" | "admin")[];
    scopes?: string[];
    ip_whitelist?: string[];
    allowed_origins?: string[];
    expires_at?: string;
  }

  export interface RotateApiKeyResponse {
    new_key: string;
    new_api_key: ApikeyApi.ApiKey;
    old_key_id?: string;
    old_key_expires_at: string;
    message?: string;
  }

  export interface ApiKeyUsage {
    api_key_id?: string;
    period?:   {
      start?: string;
      end?: string;
    };
    total_requests?: number;
    successful_requests?: number;
    failed_requests?: number;
    total_bandwidth_bytes?: number;
    avg_latency_ms?: number;
    top_endpoints?: (  {
      endpoint?: string;
      count?: number;
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

  export interface GetapikeyPathParams {
    id: string;
  }

  export interface UpdateapikeyPathParams {
    id: string;
  }

  export interface RevokeapikeyPathParams {
    id: string;
  }

  export interface RotateapikeyPathParams {
    id: string;
  }

  export interface GetapikeyusagePathParams {
    id: string;
  }

  export interface ListapikeysQueryParams {
    is_active?: boolean;
    limit?: number;
    offset?: number;
  }

  export interface GetapikeyusageQueryParams {
    start_date?: string;
    end_date?: string;
  }

  export interface ListapikeysParams {
    query?: ApikeyApi.ListapikeysQueryParams;
  }

  export interface CreateapikeyParams {
    body: ApikeyApi.CreateApiKeyRequest;
  }

  export interface GetapikeyParams {
    path: ApikeyApi.GetapikeyPathParams;
  }

  export interface UpdateapikeyParams {
    path: ApikeyApi.UpdateapikeyPathParams;
    body: ApikeyApi.UpdateApiKeyRequest;
  }

  export interface RevokeapikeyParams {
    path: ApikeyApi.RevokeapikeyPathParams;
    body?: {
    reason?: string;
  };
  }

  export interface RotateapikeyParams {
    path: ApikeyApi.RotateapikeyPathParams;
    body?: {
    grace_period_hours?: number;
  };
  }

  export interface GetapikeyusageParams {
    path: ApikeyApi.GetapikeyusagePathParams;
    query?: ApikeyApi.GetapikeyusageQueryParams;
  }

}

export const apikeyApiClient = {
  async listApiKeys(params: ApikeyApi.ListapikeysParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.ApiKeyList>> {
    const url = 'v1/api-keys';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< ApikeyApi.ApiKeyList >(url, { params: query });
  },
  async createApiKey(params: ApikeyApi.CreateapikeyParams, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.CreateApiKeyResponse>> {
    const url = 'v1/api-keys';
    const query = undefined;
    return client.post< ApikeyApi.CreateApiKeyResponse >(url, params.body, { params: query });
  },
  async getApiKey(params: ApikeyApi.GetapikeyParams, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.ApiKey>> {
    const url = interpolatePath('v1/api-keys/{id}', params.path);
    const query = undefined;
    return client.get< ApikeyApi.ApiKey >(url, { params: query });
  },
  async updateApiKey(params: ApikeyApi.UpdateapikeyParams, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.ApiKey>> {
    const url = interpolatePath('v1/api-keys/{id}', params.path);
    const query = undefined;
    return client.patch< ApikeyApi.ApiKey >(url, params.body, { params: query });
  },
  async revokeApiKey(params: ApikeyApi.RevokeapikeyParams, client: APIClient = getDefaultClient()): Promise<APIResponse<{
  id?: string;
  revoked_at?: string;
  message?: string;
}>> {
    const url = interpolatePath('v1/api-keys/{id}', params.path);
    const query = undefined;
    return client.delete< {
  id?: string;
  revoked_at?: string;
  message?: string;
} >(url, { params: query });
  },
  async rotateApiKey(params: ApikeyApi.RotateapikeyParams, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.RotateApiKeyResponse>> {
    const url = interpolatePath('v1/api-keys/{id}/rotate', params.path);
    const query = undefined;
    return client.post< ApikeyApi.RotateApiKeyResponse >(url, params.body, { params: query });
  },
  async getApiKeyUsage(params: ApikeyApi.GetapikeyusageParams, client: APIClient = getDefaultClient()): Promise<APIResponse<ApikeyApi.ApiKeyUsage>> {
    const url = interpolatePath('v1/api-keys/{id}/usage', params.path);
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< ApikeyApi.ApiKeyUsage >(url, { params: query });
  },
};
