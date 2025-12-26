import type { EndpointStat } from '../hooks/use-analytics';

interface TopApisProps {
  endpoints: EndpointStat[];
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 }).format(value);
}

function formatPercent(value: number) {
  return `${value.toFixed(1)}%`;
}

export function TopApis({ endpoints }: TopApisProps) {
  const maxRequests = Math.max(...endpoints.map((endpoint) => endpoint.requests), 1);

  return (
    <section className="card api-card reveal delay-1">
      <header className="card-header">
        <div>
          <p className="eyebrow">Top APIs</p>
          <h2 className="card-title">Most requested endpoints</h2>
          <p className="card-subtitle">Usage intensity, latency, and error rates.</p>
        </div>
        <span className="pill pill--dark">High traffic</span>
      </header>

      <div className="api-list">
        {endpoints.map((endpoint) => (
          <div className="api-row" key={`${endpoint.method}-${endpoint.endpoint}`}>
            <div className="api-label">
              <span className="api-method">{endpoint.method}</span>
              <span className="api-path">{endpoint.endpoint}</span>
            </div>
            <div className="api-stats">
              <span>{formatNumber(endpoint.requests)} calls</span>
              <span>{formatPercent(endpoint.errorRate)} errors</span>
              <span>{Math.round(endpoint.avgLatencyMs)} ms avg</span>
            </div>
            <div className="api-bar">
              <span
                className="api-bar__fill"
                style={{ width: `${(endpoint.requests / maxRequests) * 100}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
