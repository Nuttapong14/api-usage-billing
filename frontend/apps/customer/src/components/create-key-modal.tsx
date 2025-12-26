'use client';

import { type FormEvent, useEffect, useMemo, useState } from 'react';
import type { ApiKeyPermission, CreateApiKeyInput, CreateApiKeyResult } from '../hooks/use-api-keys';

interface CreateKeyModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (input: CreateApiKeyInput) => Promise<CreateApiKeyResult>;
}

const permissionOptions: ApiKeyPermission[] = ['read', 'write', 'admin'];

function parseList(value: string) {
  return value
    .split(/[,\n]/)
    .map((entry) => entry.trim())
    .filter(Boolean);
}

export function CreateKeyModal({ isOpen, onClose, onCreate }: CreateKeyModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [permissions, setPermissions] = useState<ApiKeyPermission[]>(['read']);
  const [scopes, setScopes] = useState('');
  const [ipWhitelist, setIpWhitelist] = useState('');
  const [allowedOrigins, setAllowedOrigins] = useState('');
  const [expiresAt, setExpiresAt] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [createdKey, setCreatedKey] = useState<string | null>(null);

  useEffect(() => {
    if (!isOpen) {
      setName('');
      setDescription('');
      setPermissions(['read']);
      setScopes('');
      setIpWhitelist('');
      setAllowedOrigins('');
      setExpiresAt('');
      setIsSubmitting(false);
      setError(null);
      setCreatedKey(null);
    }
  }, [isOpen]);

  const expiresAtValue = useMemo(() => {
    if (!expiresAt) {
      return undefined;
    }
    const parsed = new Date(expiresAt);
    if (Number.isNaN(parsed.getTime())) {
      return undefined;
    }
    return parsed.toISOString();
  }, [expiresAt]);

  if (!isOpen) {
    return null;
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const result = await onCreate({
        name: name.trim(),
        description: description.trim() || undefined,
        permissions,
        scopes: parseList(scopes),
        ipWhitelist: parseList(ipWhitelist),
        allowedOrigins: parseList(allowedOrigins),
        expiresAt: expiresAtValue
      });
      setCreatedKey(result.key);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create API key');
    } finally {
      setIsSubmitting(false);
    }
  }

  function togglePermission(permission: ApiKeyPermission) {
    setPermissions((prev) => {
      if (prev.includes(permission)) {
        const next = prev.filter((item) => item !== permission);
        return next.length === 0 ? ['read'] : next;
      }
      return [...prev, permission];
    });
  }

  async function handleCopy() {
    if (!createdKey) {
      return;
    }
    try {
      await navigator.clipboard.writeText(createdKey);
    } catch {
      setError('Unable to copy key to clipboard');
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal">
        <div className="modal-header">
          <div>
            <p className="eyebrow">New API key</p>
            <h2 className="card-title">Create API key</h2>
          </div>
          <button className="button button--ghost button--sm" type="button" onClick={onClose}>
            Close
          </button>
        </div>

        <div className="modal-body">
          {createdKey ? (
            <div className="key-display">
              <p className="eyebrow">Your new key</p>
              <p className="key-value">{createdKey}</p>
              <p className="helper-text">
                Copy this key now. You will not be able to view it again after closing this window.
              </p>
              {error ? <div className="banner banner--error">{error}</div> : null}
              <div className="modal-actions">
                <button className="button" type="button" onClick={handleCopy}>
                  Copy key
                </button>
                <button className="button button--ghost" type="button" onClick={onClose}>
                  Done
                </button>
              </div>
            </div>
          ) : (
            <form className="form-grid" onSubmit={handleSubmit}>
              <div className="field">
                <label className="label" htmlFor="api-key-name">
                  Key name
                </label>
                <input
                  id="api-key-name"
                  className="input"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Production API key"
                  required
                />
              </div>

              <div className="field">
                <label className="label" htmlFor="api-key-description">
                  Description
                </label>
                <textarea
                  id="api-key-description"
                  className="textarea"
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                  placeholder="Where will this key be used?"
                  rows={3}
                />
              </div>

              <div className="field">
                <span className="label">Permissions</span>
                <div className="permission-grid">
                  {permissionOptions.map((permission) => (
                    <label key={permission} className="permission-pill">
                      <input
                        type="checkbox"
                        checked={permissions.includes(permission)}
                        onChange={() => togglePermission(permission)}
                      />
                      <span>{permission}</span>
                    </label>
                  ))}
                </div>
              </div>

              <div className="field">
                <label className="label" htmlFor="api-key-scopes">
                  Scopes (optional)
                </label>
                <input
                  id="api-key-scopes"
                  className="input"
                  value={scopes}
                  onChange={(event) => setScopes(event.target.value)}
                  placeholder="usage:read, billing:read"
                />
                <p className="helper-text">Comma or newline separated list of API scopes.</p>
              </div>

              <div className="field">
                <label className="label" htmlFor="api-key-ip">
                  IP allowlist (optional)
                </label>
                <textarea
                  id="api-key-ip"
                  className="textarea"
                  value={ipWhitelist}
                  onChange={(event) => setIpWhitelist(event.target.value)}
                  placeholder="10.0.0.0/16, 203.0.113.4"
                  rows={2}
                />
              </div>

              <div className="field">
                <label className="label" htmlFor="api-key-origins">
                  Allowed origins (optional)
                </label>
                <textarea
                  id="api-key-origins"
                  className="textarea"
                  value={allowedOrigins}
                  onChange={(event) => setAllowedOrigins(event.target.value)}
                  placeholder="https://app.example.com"
                  rows={2}
                />
              </div>

              <div className="field">
                <label className="label" htmlFor="api-key-expires">
                  Expiration (optional)
                </label>
                <input
                  id="api-key-expires"
                  className="input"
                  type="datetime-local"
                  value={expiresAt}
                  onChange={(event) => setExpiresAt(event.target.value)}
                />
              </div>

              {error ? <div className="banner banner--error">{error}</div> : null}

              <div className="modal-actions">
                <button className="button" type="submit" disabled={isSubmitting}>
                  {isSubmitting ? 'Creating...' : 'Create key'}
                </button>
                <button className="button button--ghost" type="button" onClick={onClose}>
                  Cancel
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
}
