import ky, { KyInstance, Options, HTTPError } from 'ky';
import { APIClientConfig, APIError, APIResponse, RequestOptions } from './types';

export class APIClient {
  private client: KyInstance;
  private config: APIClientConfig;

  constructor(config: APIClientConfig) {
    this.config = {
      timeout: 30000,
      retries: 2,
      requestIdHeader: 'X-Request-ID',
      organizationIdHeader: 'X-Organization-ID',
      ...config,
    };

    this.client = ky.create({
      prefixUrl: this.config.baseUrl,
      timeout: this.config.timeout,
      retry: {
        limit: this.config.retries,
        methods: ['get', 'put', 'delete'],
        statusCodes: [408, 429, 500, 502, 503, 504],
      },
      hooks: {
        beforeRequest: [
          async (request) => {
            // Add auth header
            if (!request.headers.get('Authorization') && this.config.getAccessToken) {
              const token = await this.config.getAccessToken();
              if (token) {
                request.headers.set('Authorization', `Bearer ${token}`);
              }
            }

            // Add organization header
            if (this.config.getOrganizationId) {
              const orgId = this.config.getOrganizationId();
              if (orgId && this.config.organizationIdHeader) {
                request.headers.set(this.config.organizationIdHeader, orgId);
              }
            }

            // Add request ID
            if (this.config.requestIdHeader) {
              const requestId = crypto.randomUUID();
              request.headers.set(this.config.requestIdHeader, requestId);
            }
          },
        ],
        afterResponse: [
          async (_request, _options, response) => {
            if (response.status === 401 && this.config.onUnauthorized) {
              this.config.onUnauthorized();
            }
            return response;
          },
        ],
      },
    });
  }

  private async request<T>(
    method: 'get' | 'post' | 'put' | 'patch' | 'delete',
    path: string,
    data?: unknown,
    options?: RequestOptions
  ): Promise<APIResponse<T>> {
    const kyOptions: Options = {
      searchParams: options?.params as Record<string, string | number | boolean>,
      timeout: options?.timeout,
      signal: options?.signal,
      headers: options?.headers,
    };

    if (data && method !== 'get') {
      kyOptions.json = data;
    }

    try {
      const response = await this.client[method](path, kyOptions);
      const result = await response.json<APIResponse<T>>();
      return result;
    } catch (error) {
      if (error instanceof HTTPError) {
        const apiError = await this.parseError(error);
        if (this.config.onError) {
          this.config.onError(apiError);
        }
        return {
          success: false,
          error: apiError,
        };
      }

      const unknownError: APIError = {
        code: 'NETWORK_ERROR',
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      };

      if (this.config.onError) {
        this.config.onError(unknownError);
      }

      return {
        success: false,
        error: unknownError,
      };
    }
  }

  private async parseError(error: HTTPError): Promise<APIError> {
    try {
      const body = await error.response.json<APIResponse<never>>();
      if (body.error) {
        return body.error;
      }
    } catch {
      // Response is not JSON
    }

    const statusErrors: Record<number, string> = {
      400: 'BAD_REQUEST',
      401: 'UNAUTHORIZED',
      403: 'FORBIDDEN',
      404: 'NOT_FOUND',
      409: 'CONFLICT',
      422: 'VALIDATION_ERROR',
      429: 'RATE_LIMITED',
      500: 'INTERNAL_ERROR',
      502: 'BAD_GATEWAY',
      503: 'SERVICE_UNAVAILABLE',
      504: 'GATEWAY_TIMEOUT',
    };

    return {
      code: statusErrors[error.response.status] || 'UNKNOWN_ERROR',
      message: error.message || `HTTP ${error.response.status}`,
    };
  }

  // HTTP Methods
  async get<T>(path: string, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('get', path, undefined, options);
  }

  async post<T>(path: string, data?: unknown, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('post', path, data, options);
  }

  async put<T>(path: string, data?: unknown, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('put', path, data, options);
  }

  async patch<T>(path: string, data?: unknown, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('patch', path, data, options);
  }

  async delete<T>(path: string, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('delete', path, undefined, options);
  }
}

// Singleton instance management
let defaultClient: APIClient | null = null;

export function createAPIClient(config: APIClientConfig): APIClient {
  return new APIClient(config);
}

export function setDefaultClient(client: APIClient): void {
  defaultClient = client;
}

export function getDefaultClient(): APIClient {
  if (!defaultClient) {
    throw new Error('Default API client not initialized. Call setDefaultClient first.');
  }
  return defaultClient;
}
