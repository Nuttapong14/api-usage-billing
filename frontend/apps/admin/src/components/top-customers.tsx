import type { CustomerMetric } from '../hooks/use-analytics';

interface TopCustomersProps {
  customers: CustomerMetric[];
  title?: string;
  subtitle?: string;
}

function formatCurrency(amount: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0
  }).format(amount);
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 }).format(value);
}

function formatDate(value?: string) {
  if (!value) {
    return 'No invoice';
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return parsed.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function TopCustomers({ customers, title, subtitle }: TopCustomersProps) {
  return (
    <section className="card table-card reveal">
      <header className="card-header">
        <div>
          <p className="eyebrow">Top customers</p>
          <h2 className="card-title">{title ?? 'Revenue leaders'}</h2>
          <p className="card-subtitle">{subtitle ?? 'Ranked by billed revenue and API activity.'}</p>
        </div>
        <span className="pill">Updated live</span>
      </header>

      <div className="table-wrap">
        <table className="table">
          <thead>
            <tr>
              <th>Customer</th>
              <th>Revenue</th>
              <th>Requests</th>
              <th>Invoices</th>
              <th>Last invoice</th>
            </tr>
          </thead>
          <tbody>
            {customers.map((customer, index) => (
              <tr key={customer.customerId}>
                <td>
                  <div className="customer-cell">
                    <span className="customer-rank">#{index + 1}</span>
                    <div>
                      <div className="customer-name">{customer.name}</div>
                      <div className="customer-email">{customer.email}</div>
                    </div>
                  </div>
                </td>
                <td>{formatCurrency(customer.revenue.amount, customer.revenue.currency)}</td>
                <td>{formatNumber(customer.requestCount)}</td>
                <td>{customer.invoiceCount}</td>
                <td>{formatDate(customer.lastInvoiceAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
