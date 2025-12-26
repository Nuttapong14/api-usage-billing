'use client';

import Link from 'next/link';
import { InvoiceList } from '../../../components/invoice-list';
import { useBilling } from '../../../hooks/use-billing';

const currencyFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'THB',
  maximumFractionDigits: 2
});

export default function BillingPage() {
  const { invoices, totals, isLoading, error, isDemo, downloadInvoice } = useBilling({
    limit: 12
  });

  const outstanding = invoices.filter(
    (invoice) => invoice.status === 'pending' || invoice.status === 'overdue'
  );
  const outstandingTotal = outstanding.reduce((sum, invoice) => sum + invoice.total.amount, 0);

  return (
    <div className="page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Billing center</p>
          <h1 className="page-title">Billing history</h1>
          <p className="page-subtitle">Keep track of invoices, due dates, and receipts.</p>
        </div>
        <div className="header-actions">
          <Link className="button button--ghost" href="/">
            Back to overview
          </Link>
          <button className="button" type="button">
            Update payment method
          </button>
        </div>
      </header>

      {isDemo ? (
        <div className="banner">Demo data is shown. Connect a billing service to load live invoices.</div>
      ) : null}
      {error ? <div className="banner banner--error">{error}</div> : null}

      <section className="stats-grid">
        <div className="card stat-card reveal">
          <p className="eyebrow">Total invoices</p>
          <h2 className="stat-value">{totals.count}</h2>
          <p className="stat-meta">All time</p>
        </div>
        <div className="card stat-card reveal delay-1">
          <p className="eyebrow">Amount billed</p>
          <h2 className="stat-value">{currencyFormatter.format(totals.amount)}</h2>
          <p className="stat-meta">Across issued invoices</p>
        </div>
        <div className="card stat-card reveal delay-2">
          <p className="eyebrow">Outstanding</p>
          <h2 className="stat-value">{currencyFormatter.format(outstandingTotal)}</h2>
          <p className="stat-meta">{outstanding.length} unpaid</p>
        </div>
      </section>

      <InvoiceList invoices={invoices} onDownload={downloadInvoice} isLoading={isLoading} />
    </div>
  );
}
