import type { InvoiceSummary } from '../hooks/use-billing';
import { InvoiceDownload } from './invoice-download';

interface InvoiceListProps {
  invoices: InvoiceSummary[];
  onDownload?: (invoiceId: string) => Promise<string | null>;
  isLoading?: boolean;
}

const dateFormatter = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: 'numeric',
  year: 'numeric'
});

function formatDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return dateFormatter.format(parsed);
}

function formatCurrency(amount: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    maximumFractionDigits: 2
  }).format(amount);
}

function statusClass(status: InvoiceSummary['status']) {
  switch (status) {
    case 'paid':
      return 'status-badge status-badge--paid';
    case 'pending':
      return 'status-badge status-badge--pending';
    case 'overdue':
      return 'status-badge status-badge--overdue';
    case 'cancelled':
      return 'status-badge status-badge--cancelled';
    default:
      return 'status-badge status-badge--draft';
  }
}

export function InvoiceList({ invoices, onDownload, isLoading }: InvoiceListProps) {
  return (
    <div className="card invoice-card reveal">
      <div className="card-header">
        <div>
          <p className="eyebrow">Payment history</p>
          <h2 className="card-title">Invoices</h2>
        </div>
        <span className="pill">{invoices.length} total</span>
      </div>

      {isLoading ? (
        <div className="empty-state">Loading invoices...</div>
      ) : invoices.length === 0 ? (
        <div className="empty-state">No invoices available yet.</div>
      ) : (
        <div className="invoice-table">
          <div className="invoice-row invoice-row--header">
            <span>Invoice</span>
            <span>Status</span>
            <span>Issued</span>
            <span>Due</span>
            <span>Total</span>
            <span>PDF</span>
          </div>
          {invoices.map((invoice) => (
            <div key={invoice.id} className="invoice-row">
              <div>
                <p className="invoice-id">{invoice.invoiceNumber}</p>
                <p className="muted">{invoice.id}</p>
              </div>
              <div>
                <span className={statusClass(invoice.status)}>{invoice.status}</span>
              </div>
              <div>{formatDate(invoice.issueDate)}</div>
              <div>{formatDate(invoice.dueDate)}</div>
              <div className="invoice-total">
                {formatCurrency(invoice.total.amount, invoice.total.currency)}
              </div>
              <div>
                <InvoiceDownload invoiceId={invoice.id} pdfUrl={invoice.pdfUrl} onDownload={onDownload} />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
