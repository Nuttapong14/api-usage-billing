import type { UsageSeriesPoint } from '../hooks/use-usage';

interface UsageChartProps {
  title: string;
  data: UsageSeriesPoint[];
  total: number;
  subtitle?: string;
}

const compactFormatter = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1
});

export function UsageChart({ title, data, total, subtitle }: UsageChartProps) {
  const maxValue = Math.max(...data.map((point) => point.value), 1);
  const step = data.length > 1 ? 100 / (data.length - 1) : 0;
  const linePoints = data
    .map((point, index) => {
      const x = index * step;
      const y = 56 - (point.value / maxValue) * 46;
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(' ');

  const areaPoints = `0,60 ${linePoints} 100,60`;
  const startLabel = data[0]?.label ?? '';
  const endLabel = data[data.length - 1]?.label ?? '';

  return (
    <div className="card chart-card reveal">
      <div className="card-header">
        <div>
          <p className="eyebrow">{subtitle ?? 'Daily usage'}</p>
          <h2 className="card-title">{title}</h2>
        </div>
        <div className="metric-chip">
          {compactFormatter.format(total)} requests
        </div>
      </div>
      <div className="chart-wrap">
        <svg className="chart-svg" viewBox="0 0 100 60" preserveAspectRatio="none">
          <defs>
            <linearGradient id="usageGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#ff9b4a" stopOpacity="0.6" />
              <stop offset="100%" stopColor="#ff9b4a" stopOpacity="0" />
            </linearGradient>
          </defs>
          <g className="chart-grid">
            <line x1="0" y1="15" x2="100" y2="15" />
            <line x1="0" y1="30" x2="100" y2="30" />
            <line x1="0" y1="45" x2="100" y2="45" />
          </g>
          <polygon points={areaPoints} fill="url(#usageGradient)" />
          <polyline points={linePoints} fill="none" stroke="#e46a1a" strokeWidth="2" />
        </svg>
      </div>
      <div className="chart-footer">
        <span>{startLabel}</span>
        <span>{endLabel}</span>
      </div>
    </div>
  );
}
