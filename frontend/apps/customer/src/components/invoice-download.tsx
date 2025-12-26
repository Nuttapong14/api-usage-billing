'use client';

import { useState } from 'react';

interface InvoiceDownloadProps {
  invoiceId: string;
  pdfUrl?: string | null;
  onDownload?: (invoiceId: string) => Promise<string | null>;
}

export function InvoiceDownload({ invoiceId, pdfUrl, onDownload }: InvoiceDownloadProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const canDownload = Boolean(pdfUrl || onDownload);

  const handleClick = async () => {
    if (!canDownload || isLoading) {
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const url = pdfUrl ?? (await onDownload?.(invoiceId));
      if (!url) {
        setError('Unavailable');
        return;
      }
      window.open(url, '_blank', 'noopener,noreferrer');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Download failed');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="invoice-download">
      <button className="button button--ghost" onClick={handleClick} disabled={!canDownload || isLoading}>
        {isLoading ? 'Preparing...' : 'Download'}
      </button>
      {error ? <span className="helper-text">{error}</span> : null}
    </div>
  );
}
