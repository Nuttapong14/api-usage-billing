'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';

export type InvoiceStatus = 'draft' | 'pending' | 'paid' | 'overdue' | 'cancelled';

export interface InvoiceSummary {
  id: string;
  invoiceNumber: string;
  total: {
    amount: number;
    currency: string;
  };
  status: InvoiceStatus;
  issueDate: string;
  dueDate: string;
  pdfUrl?: string | null;
}

interface InvoiceListResponse {
  data: Array<{
    id: string;
    invoice_number: string;
    total: {
      amount: number;
      currency: string;
    };
    status: InvoiceStatus;
    issue_date: string;
    due_date: string;
    pdf_url?: string | null;
  }>;
  pagination: {
    total: number;
    limit: number;
    offset: number;
    has_more: boolean;
  };
}

interface UseBillingOptions {
  baseUrl?: string;
  limit?: number;
}

interface BillingState {
  invoices: InvoiceSummary[];
  isLoading: boolean;
  error?: string;
  isDemo: boolean;
}

const demoInvoices: InvoiceSummary[] = [
  {
    id: 'inv-2025-0012',
    invoiceNumber: 'INV-2025-0012',
    total: { amount: 4820.45, currency: 'THB' },
    status: 'paid',
    issueDate: '2025-01-01',
    dueDate: '2025-01-10',
    pdfUrl: null
  },
  {
    id: 'inv-2025-0013',
    invoiceNumber: 'INV-2025-0013',
    total: { amount: 5165.9, currency: 'THB' },
    status: 'paid',
    issueDate: '2025-02-01',
    dueDate: '2025-02-10',
    pdfUrl: null
  },
  {
    id: 'inv-2025-0014',
    invoiceNumber: 'INV-2025-0014',
    total: { amount: 5380.12, currency: 'THB' },
    status: 'pending',
    issueDate: '2025-03-01',
    dueDate: '2025-03-10',
    pdfUrl: null
  },
  {
    id: 'inv-2025-0015',
    invoiceNumber: 'INV-2025-0015',
    total: { amount: 5714.88, currency: 'THB' },
    status: 'overdue',
    issueDate: '2025-04-01',
    dueDate: '2025-04-10',
    pdfUrl: null
  }
];

const baseState: BillingState = {
  invoices: demoInvoices,
  isLoading: false,
  isDemo: true
};

function mapInvoices(payload: InvoiceListResponse): InvoiceSummary[] {
  return payload.data.map((invoice) => ({
    id: invoice.id,
    invoiceNumber: invoice.invoice_number,
    total: invoice.total,
    status: invoice.status,
    issueDate: invoice.issue_date,
    dueDate: invoice.due_date,
    pdfUrl: invoice.pdf_url ?? null
  }));
}

export function useBilling(options: UseBillingOptions = {}) {
  const baseUrl = options.baseUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL;
  const limit = options.limit ?? 10;
  const normalizedBaseUrl = baseUrl ? baseUrl.replace(/\/$/, '') : '';

  const [state, setState] = useState<BillingState>(baseState);

  useEffect(() => {
    if (!normalizedBaseUrl) {
      return;
    }

    const controller = new AbortController();

    async function loadInvoices() {
      setState((prev) => ({
        ...prev,
        isLoading: true,
        error: undefined
      }));

      try {
        const params = new URLSearchParams({
          limit: String(limit),
          offset: '0',
          sort_by: 'issue_date',
          sort_dir: 'desc'
        });

        const response = await fetch(`${normalizedBaseUrl}/v1/invoices?${params.toString()}`, {
          headers: { Accept: 'application/json' },
          credentials: 'include',
          signal: controller.signal
        });

        if (!response.ok) {
          throw new Error('Unable to load invoices');
        }

        const payload = (await response.json()) as InvoiceListResponse;
        setState({
          invoices: mapInvoices(payload),
          isLoading: false,
          isDemo: false
        });
      } catch (error) {
        if (controller.signal.aborted) {
          return;
        }

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: error instanceof Error ? error.message : 'Failed to load invoices'
        }));
      }
    }

    loadInvoices();

    return () => controller.abort();
  }, [normalizedBaseUrl, limit]);

  const downloadInvoice = useCallback(
    async (invoiceId: string) => {
      const localInvoice = state.invoices.find((invoice) => invoice.id === invoiceId);
      if (!normalizedBaseUrl) {
        return localInvoice?.pdfUrl ?? null;
      }

      const response = await fetch(`${normalizedBaseUrl}/v1/invoices/${invoiceId}/pdf?language=en`, {
        headers: { Accept: 'application/pdf' },
        credentials: 'include'
      });

      if (response.status === 202) {
        throw new Error('Invoice PDF is being generated');
      }

      if (!response.ok) {
        throw new Error('Unable to download invoice');
      }

      const blob = await response.blob();
      const url = URL.createObjectURL(blob);
      setTimeout(() => URL.revokeObjectURL(url), 10000);
      return url;
    },
    [normalizedBaseUrl, state.invoices]
  );

  const totals = useMemo(() => {
    return state.invoices.reduce(
      (acc, invoice) => {
        acc.amount += invoice.total.amount;
        acc.count += 1;
        return acc;
      },
      { amount: 0, count: 0 }
    );
  }, [state.invoices]);

  return {
    ...state,
    totals,
    downloadInvoice
  };
}
