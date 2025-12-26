interface BillingPeriodCardProps {
  start: string;
  end: string;
  planName?: string;
  nextInvoiceDate?: string;
  status?: string;
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

export function BillingPeriodCard({ start, end, planName, nextInvoiceDate, status }: BillingPeriodCardProps) {
  return (
    <div className="card billing-card reveal delay-2">
      <div className="card-header">
        <div>
          <p className="eyebrow">Billing period</p>
          <h2 className="card-title">{formatDate(start)} - {formatDate(end)}</h2>
        </div>
        {status ? <span className="pill pill--solid">{status}</span> : null}
      </div>
      <div className="billing-details">
        <div>
          <p className="label">Plan</p>
          <p className="value">{planName ?? 'Growth Tier'}</p>
        </div>
        <div>
          <p className="label">Next invoice</p>
          <p className="value">{nextInvoiceDate ? formatDate(nextInvoiceDate) : 'On renewal'}</p>
        </div>
      </div>
    </div>
  );
}
