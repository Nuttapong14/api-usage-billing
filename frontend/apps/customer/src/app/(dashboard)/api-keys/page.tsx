'use client';

import Link from 'next/link';
import { useState } from 'react';
import { ApiKeyList } from '../../../components/api-key-list';
import { CreateKeyModal } from '../../../components/create-key-modal';
import { RotateKeyDialog } from '../../../components/rotate-key-dialog';
import { useApiKeys, type ApiKey } from '../../../hooks/use-api-keys';

export default function ApiKeysPage() {
  const { apiKeys, stats, isLoading, error, isDemo, refresh, createKey, rotateKey, revokeKey } = useApiKeys({
    limit: 20
  });
  const [createOpen, setCreateOpen] = useState(false);
  const [rotateTarget, setRotateTarget] = useState<ApiKey | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  async function handleRevoke(apiKey: ApiKey) {
    const confirmRevoke = window.confirm(`Revoke ${apiKey.name}? This cannot be undone.`);
    if (!confirmRevoke) {
      return;
    }

    const reason = window.prompt('Optional reason for revocation:', '') ?? undefined;

    try {
      setActionError(null);
      await revokeKey(apiKey.id, reason || undefined);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : 'Failed to revoke API key');
    }
  }

  return (
    <div className="page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Security center</p>
          <h1 className="page-title">API key management</h1>
          <p className="page-subtitle">
            Create, rotate, and revoke keys instantly to keep your integrations secure.
          </p>
        </div>
        <div className="header-actions">
          <button className="button button--ghost" type="button" onClick={refresh}>
            Refresh
          </button>
          <Link className="button button--ghost" href="/">
            Back to overview
          </Link>
          <button className="button" type="button" onClick={() => setCreateOpen(true)}>
            Create API key
          </button>
        </div>
      </header>

      {isDemo ? (
        <div className="banner">Demo data is shown. Connect an API service to manage live keys.</div>
      ) : null}
      {error ? <div className="banner banner--error">{error}</div> : null}
      {actionError ? <div className="banner banner--error">{actionError}</div> : null}

      <section className="stats-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Total keys</p>
          <h2 className="stat-value">{stats.total}</h2>
          <p className="stat-meta">Across environments</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">Active</p>
          <h2 className="stat-value">{stats.active}</h2>
          <p className="stat-meta">Ready for production</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Rotating</p>
          <h2 className="stat-value">{stats.rotating}</h2>
          <p className="stat-meta">Grace period in progress</p>
        </div>
        <div className="card stat-card reveal delay-3">
          <p className="eyebrow">Revoked</p>
          <h2 className="stat-value">{stats.revoked}</h2>
          <p className="stat-meta">Inactive credentials</p>
        </div>
      </section>

      <ApiKeyList
        apiKeys={apiKeys}
        isLoading={isLoading}
        onRotate={(key) => setRotateTarget(key)}
        onRevoke={handleRevoke}
      />

      <CreateKeyModal
        isOpen={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreate={createKey}
      />
      <RotateKeyDialog
        apiKey={rotateTarget}
        onClose={() => setRotateTarget(null)}
        onRotate={rotateKey}
      />
    </div>
  );
}
