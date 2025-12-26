import type { RevenueReport } from '../hooks/use-analytics';

interface RevenueChartProps {
  title: string;
  report: RevenueReport;
  isLoading?: boolean;
}

function formatCurrency(amount: number, currency: string) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0
  }).format(amount);
}

function formatDate(value?: string) {
  if (!value) {
    return '';
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return parsed.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function RevenueChart({ title, report, isLoading }: RevenueChartProps) {
  const data = report.data ?? [];
  const currency = report.total.currency || 'USD';
  const totalLabel = formatCurrency(report.total.amount, currency);

  const width = 640;
  const height = 220;
  const padding = 24;

  const maxValue = Math.max(...data.map((point) => point.amount.amount), 1);
  const points = data.map((point, index) => {
    const x =
      data.length === 1
        ? width / 2
        : padding + (index / (data.length - 1)) * (width - padding * 2);
    const y = height - padding - (point.amount.amount / maxValue) * (height - padding * 2);
    return { x, y, value: point.amount.amount };
  });

  const linePath = points.map((point) => `${point.x},${point.y}`).join(' ');
  const areaPath =
    points.length > 1
      ? `M ${points[0].x} ${height - padding} L ${linePath.replace(/,/g, ' ')} L ${
          points[points.length - 1].x
        } ${height - padding} Z`
      : '';

  return (
    <section className="card chart-card reveal">
      <header className="card-header">
        <div>
          <p className="eyebrow">Revenue trend</p>
          <h2 className="card-title">{title}</h2>
          <p className="card-subtitle">
            {formatDate(report.periodStart)} to {formatDate(report.periodEnd)} - {report.granularity}
          </p>
        </div>
        <div className="chart-meta">
          <p className="chart-total">{totalLabel}</p>
          <p className="chart-label">{data.length} data points</p>
        </div>
      </header>

      <div className="chart-body">
        {isLoading ? <div className="chart-loading">Loading revenue data...</div> : null}
        {!isLoading && data.length === 0 ? (
          <div className="chart-loading">No revenue data for this range.</div>
        ) : null}
        {!isLoading && data.length > 0 ? (
          <svg className="chart-svg" viewBox={`0 0 ${width} ${height}`} role="img" aria-label="Revenue chart">
            <defs>
              <linearGradient id="revenueGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="rgba(229, 107, 66, 0.4)" />
                <stop offset="100%" stopColor="rgba(229, 107, 66, 0)" />
              </linearGradient>
            </defs>
            {areaPath ? <path d={areaPath} fill="url(#revenueGradient)" /> : null}
            <polyline
              points={linePath}
              fill="none"
              stroke="rgba(229, 107, 66, 0.9)"
              strokeWidth="3"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {points.map((point, index) => (
              <circle key={index} cx={point.x} cy={point.y} r="4.5" fill="#0f2d2e" />
            ))}
          </svg>
        ) : null}
      </div>
    </section>
  );
}
