import type { ApiKey } from '../hooks/use-api-keys';

interface ApiKeyListProps {
  apiKeys: ApiKey[];
  isLoading?: boolean;
  onRotate: (apiKey: ApiKey) => void;
  onRevoke: (apiKey: ApiKey) => void;
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit'
});

function formatDate(value?: string | null) {
  if (!value) {
    return 'Never';
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return dateFormatter.format(parsed);
}

function statusClass(key: ApiKey) {
  if (!key.isActive) {
    return 'status-badge status-badge--revoked';
  }
  if (key.isRotating) {
    return 'status-badge status-badge--rotating';
  }
  return 'status-badge status-badge--active';
}

function statusLabel(key: ApiKey) {
  if (!key.isActive) {
    return 'revoked';
  }
  if (key.isRotating) {
    return 'rotating';
  }
  return 'active';
}

export function ApiKeyList({ apiKeys, isLoading, onRotate, onRevoke }: ApiKeyListProps) {
  return (
    <div className="card api-key-card reveal">
      <div className="card-header">
        <div>
          <p className="eyebrow">Credentials</p>
          <h2 className="card-title">API keys</h2>
        </div>
        <span className="pill">{apiKeys.length} total</span>
      </div>

      {isLoading ? (
        <div className="empty-state">Loading API keys...</div>
      ) : apiKeys.length === 0 ? (
        <div className="empty-state">No API keys yet. Create your first key to get started.</div>
      ) : (
        <div className="api-key-table">
          <div className="api-key-row api-key-row--header">
            <span>Key</span>
            <span>Permissions</span>
            <span>Last used</span>
            <span>Status</span>
            <span>Expires</span>
            <span>Actions</span>
          </div>
          {apiKeys.map((apiKey) => (
            <div key={apiKey.id} className="api-key-row">
              <div>
                <p className="api-key-name">{apiKey.name}</p>
                <p className="muted">Prefix {apiKey.keyPrefix}</p>
              </div>
              <div className="tag-list">
                {(apiKey.permissions.length ? apiKey.permissions : ['read']).map((permission) => (
                  <span key={permission} className="tag">
                    {permission}
                  </span>
                ))}
              </div>
              <div>
                <p className="api-key-meta">{formatDate(apiKey.lastUsedAt)}</p>
                <p className="muted">{apiKey.lastUsedIp ?? 'No IP logged'}</p>
              </div>
              <div>
                <span className={statusClass(apiKey)}>{statusLabel(apiKey)}</span>
                {apiKey.isRotating && apiKey.rotationGraceUntil ? (
                  <p className="muted">Grace until {formatDate(apiKey.rotationGraceUntil)}</p>
                ) : null}
              </div>
              <div>
                <p className="api-key-meta">{formatDate(apiKey.expiresAt)}</p>
              </div>
              <div className="api-key-actions">
                <button
                  className="button button--ghost button--sm"
                  type="button"
                  disabled={!apiKey.isActive || apiKey.isRotating}
                  onClick={() => onRotate(apiKey)}
                >
                  Rotate
                </button>
                <button
                  className="button button--ghost button--sm button--danger"
                  type="button"
                  disabled={!apiKey.isActive}
                  onClick={() => onRevoke(apiKey)}
                >
                  Revoke
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
