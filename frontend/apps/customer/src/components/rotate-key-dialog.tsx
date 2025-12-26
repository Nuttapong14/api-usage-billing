'use client';

import { type FormEvent, useEffect, useState } from 'react';
import type { ApiKey, RotateApiKeyInput, RotateApiKeyResult } from '../hooks/use-api-keys';

interface RotateKeyDialogProps {
  apiKey: ApiKey | null;
  onClose: () => void;
  onRotate: (input: RotateApiKeyInput) => Promise<RotateApiKeyResult>;
}

export function RotateKeyDialog({ apiKey, onClose, onRotate }: RotateKeyDialogProps) {
  const [graceHours, setGraceHours] = useState(24);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newKey, setNewKey] = useState<string | null>(null);
  const [expiresAt, setExpiresAt] = useState<string | null>(null);

  useEffect(() => {
    setGraceHours(24);
    setIsSubmitting(false);
    setError(null);
    setNewKey(null);
    setExpiresAt(null);
  }, [apiKey]);

  if (!apiKey) {
    return null;
  }

  async function handleRotate(event: FormEvent) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);

    try {
      const result = await onRotate({ apiKeyId: apiKey.id, gracePeriodHours: graceHours });
      setNewKey(result.key);
      setExpiresAt(result.oldKeyExpiresAt);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to rotate API key');
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleCopy() {
    if (!newKey) {
      return;
    }
    try {
      await navigator.clipboard.writeText(newKey);
    } catch {
      setError('Unable to copy key to clipboard');
    }
  }

  return (
    <div className="modal-backdrop">
      <div className="modal">
        <div className="modal-header">
          <div>
            <p className="eyebrow">Rotate key</p>
            <h2 className="card-title">Rotate {apiKey.name}</h2>
          </div>
          <button className="button button--ghost button--sm" type="button" onClick={onClose}>
            Close
          </button>
        </div>

        <div className="modal-body">
          {newKey ? (
            <div className="key-display">
              <p className="eyebrow">New API key</p>
              <p className="key-value">{newKey}</p>
              <p className="helper-text">
                The old key stays active until {expiresAt ?? 'the end of the grace period'}.
              </p>
              {error ? <div className="banner banner--error">{error}</div> : null}
              <div className="modal-actions">
                <button className="button" type="button" onClick={handleCopy}>
                  Copy new key
                </button>
                <button className="button button--ghost" type="button" onClick={onClose}>
                  Done
                </button>
              </div>
            </div>
          ) : (
            <form className="form-grid" onSubmit={handleRotate}>
              <div className="card note-card">
                <p className="eyebrow">Rotation window</p>
                <p className="page-subtitle">
                  The old key remains valid during the grace period so you can deploy the new key
                  without downtime.
                </p>
              </div>

              <div className="field">
                <label className="label" htmlFor="grace-period">
                  Grace period (hours)
                </label>
                <input
                  id="grace-period"
                  className="input"
                  type="number"
                  min={1}
                  max={168}
                  value={graceHours}
                  onChange={(event) => setGraceHours(Number(event.target.value))}
                />
                <p className="helper-text">Default is 24 hours. Maximum is 168 hours.</p>
              </div>

              {error ? <div className="banner banner--error">{error}</div> : null}

              <div className="modal-actions">
                <button className="button" type="submit" disabled={isSubmitting}>
                  {isSubmitting ? 'Rotating...' : 'Rotate key'}
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
