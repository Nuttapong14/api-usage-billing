'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';

export type ApiKeyPermission = 'read' | 'write' | 'admin';

export interface ApiKey {
  id: string;
  keyPrefix: string;
  name: string;
  description?: string | null;
  permissions: ApiKeyPermission[];
  scopes: string[];
  ipWhitelist?: string[];
  allowedOrigins?: string[];
  expiresAt?: string | null;
  lastUsedAt?: string | null;
  lastUsedIp?: string | null;
  isActive: boolean;
  isRotating: boolean;
  rotationGraceUntil?: string | null;
  createdAt: string;
  revokedAt?: string | null;
}

export interface CreateApiKeyInput {
  name: string;
  description?: string;
  permissions?: ApiKeyPermission[];
  scopes?: string[];
  ipWhitelist?: string[];
  allowedOrigins?: string[];
  expiresAt?: string | null;
}

export interface RotateApiKeyInput {
  apiKeyId: string;
  gracePeriodHours?: number;
}

export interface CreateApiKeyResult {
  key: string;
  apiKey: ApiKey;
}

export interface RotateApiKeyResult {
  key: string;
  apiKey: ApiKey;
  oldKeyId: string;
  oldKeyExpiresAt: string;
}

interface ApiKeyListResponse {
  data: Array<{
    id: string;
    key_prefix: string;
    name: string;
    description?: string | null;
    permissions: ApiKeyPermission[];
    scopes?: string[];
    ip_whitelist?: string[];
    allowed_origins?: string[];
    expires_at?: string | null;
    last_used_at?: string | null;
    last_used_ip?: string | null;
    is_active: boolean;
    is_rotating: boolean;
    rotation_grace_until?: string | null;
    created_at: string;
    revoked_at?: string | null;
  }>;
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

interface CreateApiKeyResponse {
  key: string;
  api_key: ApiKeyListResponse['data'][number];
  warning?: string;
}

interface RotateApiKeyResponse {
  new_key: string;
  new_api_key: ApiKeyListResponse['data'][number];
  old_key_id: string;
  old_key_expires_at: string;
  message?: string;
}

interface RevokeApiKeyResponse {
  id: string;
  revoked_at: string;
  message?: string;
}

interface UseApiKeysOptions {
  baseUrl?: string;
  limit?: number;
}

interface ApiKeyState {
  apiKeys: ApiKey[];
  isLoading: boolean;
  error?: string;
  isDemo: boolean;
}

const demoApiKeys: ApiKey[] = [
  {
    id: 'key-001',
    keyPrefix: 'sk_live_9fa2',
    name: 'Production Gateway',
    description: 'Primary key for the mobile API gateway.',
    permissions: ['read', 'write'],
    scopes: ['usage:read', 'subscription:read'],
    ipWhitelist: ['10.12.0.0/16'],
    allowedOrigins: ['https://app.pulseline.io'],
    expiresAt: null,
    lastUsedAt: '2025-01-24T08:14:00Z',
    lastUsedIp: '10.12.14.21',
    isActive: true,
    isRotating: false,
    rotationGraceUntil: null,
    createdAt: '2024-11-12T09:00:00Z',
    revokedAt: null
  },
  {
    id: 'key-002',
    keyPrefix: 'sk_live_3bc7',
    name: 'Staging Sandbox',
    description: 'Used by QA for sandbox testing.',
    permissions: ['read'],
    scopes: [],
    ipWhitelist: [],
    allowedOrigins: ['https://staging.pulseline.app'],
    expiresAt: '2025-02-20T00:00:00Z',
    lastUsedAt: '2025-01-23T19:20:00Z',
    lastUsedIp: '192.168.20.14',
    isActive: true,
    isRotating: true,
    rotationGraceUntil: '2025-01-25T19:20:00Z',
    createdAt: '2024-10-05T16:30:00Z',
    revokedAt: null
  },
  {
    id: 'key-003',
    keyPrefix: 'sk_live_72de',
    name: 'Legacy Partner',
    description: 'Revoked after vendor migration.',
    permissions: ['read'],
    scopes: [],
    ipWhitelist: [],
    allowedOrigins: [],
    expiresAt: null,
    lastUsedAt: '2024-12-01T10:00:00Z',
    lastUsedIp: '203.0.113.8',
    isActive: false,
    isRotating: false,
    rotationGraceUntil: null,
    createdAt: '2024-06-18T11:10:00Z',
    revokedAt: '2024-12-01T10:05:00Z'
  }
];

const baseState: ApiKeyState = {
  apiKeys: demoApiKeys,
  isLoading: false,
  isDemo: true
};

function mapApiKey(payload: ApiKeyListResponse['data'][number]): ApiKey {
  return {
    id: payload.id,
    keyPrefix: payload.key_prefix,
    name: payload.name,
    description: payload.description ?? null,
    permissions: payload.permissions ?? ['read'],
    scopes: payload.scopes ?? [],
    ipWhitelist: payload.ip_whitelist ?? [],
    allowedOrigins: payload.allowed_origins ?? [],
    expiresAt: payload.expires_at ?? null,
    lastUsedAt: payload.last_used_at ?? null,
    lastUsedIp: payload.last_used_ip ?? null,
    isActive: payload.is_active,
    isRotating: payload.is_rotating,
    rotationGraceUntil: payload.rotation_grace_until ?? null,
    createdAt: payload.created_at,
    revokedAt: payload.revoked_at ?? null
  };
}

function buildDemoKey(input: CreateApiKeyInput) {
  const randomSuffix = Math.random().toString(16).slice(2, 6);
  const key = `sk_live_${randomSuffix}${Math.random().toString(16).slice(2, 18)}`;
  const now = new Date().toISOString();

  const apiKey: ApiKey = {
    id: `key-${Date.now()}`,
    keyPrefix: key.slice(0, 12),
    name: input.name,
    description: input.description ?? null,
    permissions: input.permissions ?? ['read'],
    scopes: input.scopes ?? [],
    ipWhitelist: input.ipWhitelist ?? [],
    allowedOrigins: input.allowedOrigins ?? [],
    expiresAt: input.expiresAt ?? null,
    lastUsedAt: null,
    lastUsedIp: null,
    isActive: true,
    isRotating: false,
    rotationGraceUntil: null,
    createdAt: now,
    revokedAt: null
  };

  return { key, apiKey };
}

function normalizeInput(input: CreateApiKeyInput) {
  return {
    name: input.name,
    description: input.description ?? undefined,
    permissions: input.permissions ?? ['read'],
    scopes: input.scopes ?? [],
    ip_whitelist: input.ipWhitelist ?? [],
    allowed_origins: input.allowedOrigins ?? [],
    expires_at: input.expiresAt ?? undefined
  };
}

export function useApiKeys(options: UseApiKeysOptions = {}) {
  const baseUrl = options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL;
  const limit = options.limit ?? 20;
  const normalizedBaseUrl = baseUrl ? baseUrl.replace(/\/$/, '') : '';

  const [state, setState] = useState<ApiKeyState>(baseState);

  const fetchApiKeys = useCallback(
    async (signal?: AbortSignal) => {
      if (!normalizedBaseUrl) {
        return;
      }

      setState((prev) => ({
        ...prev,
        isLoading: true,
        error: undefined
      }));

      try {
        const params = new URLSearchParams({
          limit: String(limit),
          offset: '0'
        });

        const response = await fetch(`${normalizedBaseUrl}/v1/api-keys?${params.toString()}`, {
          headers: { Accept: 'application/json' },
          credentials: 'include',
          signal
        });

        if (!response.ok) {
          throw new Error('Unable to load API keys');
        }

        const payload = (await response.json()) as ApiKeyListResponse;
        setState({
          apiKeys: payload.data.map(mapApiKey),
          isLoading: false,
          isDemo: false
        });
      } catch (error) {
        if (signal?.aborted) {
          return;
        }

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to load API keys'
        }));
      }
    },
    [limit, normalizedBaseUrl]
  );

  useEffect(() => {
    if (!normalizedBaseUrl) {
      return;
    }

    const controller = new AbortController();
    fetchApiKeys(controller.signal);
    return () => controller.abort();
  }, [fetchApiKeys, normalizedBaseUrl]);

  const refresh = useCallback(() => {
    fetchApiKeys();
  }, [fetchApiKeys]);

  const createKey = useCallback(
    async (input: CreateApiKeyInput): Promise<CreateApiKeyResult> => {
      if (!normalizedBaseUrl) {
        const demo = buildDemoKey(input);
        setState((prev) => ({
          ...prev,
          apiKeys: [demo.apiKey, ...prev.apiKeys]
        }));
        return demo;
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/api-keys`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify(normalizeInput(input))
      });

      if (!response.ok) {
        throw new Error('Unable to create API key');
      }

      const payload = (await response.json()) as CreateApiKeyResponse;
      const apiKey = mapApiKey(payload.api_key);
      setState((prev) => ({
        ...prev,
        apiKeys: [apiKey, ...prev.apiKeys]
      }));

      return { key: payload.key, apiKey };
    },
    [normalizedBaseUrl]
  );

  const rotateKey = useCallback(
    async (input: RotateApiKeyInput): Promise<RotateApiKeyResult> => {
      const graceHours = input.gracePeriodHours ?? 24;
      if (!normalizedBaseUrl) {
        const demo = buildDemoKey({ name: 'Rotated key' });
        const graceUntil = new Date(Date.now() + graceHours * 60 * 60 * 1000).toISOString();

        setState((prev) => ({
          ...prev,
          apiKeys: [
            demo.apiKey,
            ...prev.apiKeys.map((key) =>
              key.id === input.apiKeyId
                ? { ...key, isRotating: true, rotationGraceUntil: graceUntil }
                : key
            )
          ]
        }));

        return {
          key: demo.key,
          apiKey: demo.apiKey,
          oldKeyId: input.apiKeyId,
          oldKeyExpiresAt: graceUntil
        };
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/api-keys/${input.apiKeyId}/rotate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json'
        },
        credentials: 'include',
        body: JSON.stringify({ grace_period_hours: graceHours })
      });

      if (!response.ok) {
        throw new Error('Unable to rotate API key');
      }

      const payload = (await response.json()) as RotateApiKeyResponse;
      const apiKey = mapApiKey(payload.new_api_key);
      setState((prev) => ({
        ...prev,
        apiKeys: [
          apiKey,
          ...prev.apiKeys.map((key) =>
            key.id === payload.old_key_id
              ? {
                  ...key,
                  isRotating: true,
                  rotationGraceUntil: payload.old_key_expires_at
                }
              : key
          )
        ]
      }));

      return {
        key: payload.new_key,
        apiKey,
        oldKeyId: payload.old_key_id,
        oldKeyExpiresAt: payload.old_key_expires_at
      };
    },
    [normalizedBaseUrl]
  );

  const revokeKey = useCallback(
    async (apiKeyId: string, reason?: string) => {
      const revokedAt = new Date().toISOString();
      if (!normalizedBaseUrl) {
        setState((prev) => ({
          ...prev,
          apiKeys: prev.apiKeys.map((key) =>
            key.id === apiKeyId
              ? { ...key, isActive: false, isRotating: false, revokedAt }
              : key
          )
        }));
        return;
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/api-keys/${apiKeyId}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          Accept: 'application/json'
        },
        credentials: 'include',
        body: reason ? JSON.stringify({ reason }) : undefined
      });

      if (!response.ok) {
        throw new Error('Unable to revoke API key');
      }

      const payload = (await response.json()) as RevokeApiKeyResponse;
      setState((prev) => ({
        ...prev,
        apiKeys: prev.apiKeys.map((key) =>
          key.id === payload.id
            ? { ...key, isActive: false, isRotating: false, revokedAt: payload.revoked_at }
            : key
        )
      }));
    },
    [normalizedBaseUrl]
  );

  const stats = useMemo(() => {
    const total = state.apiKeys.length;
    const active = state.apiKeys.filter((key) => key.isActive).length;
    const rotating = state.apiKeys.filter((key) => key.isRotating).length;
    const revoked = total - active;
    return { total, active, rotating, revoked };
  }, [state.apiKeys]);

  return {
    ...state,
    stats,
    refresh,
    createKey,
    rotateKey,
    revokeKey
  };
}
