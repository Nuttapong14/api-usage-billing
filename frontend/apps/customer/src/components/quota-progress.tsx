import type { QuotaProgressItem } from '../hooks/use-usage';

interface QuotaProgressProps {
  items: QuotaProgressItem[];
  updatedAt?: string;
  title?: string;
}

const numberFormatter = new Intl.NumberFormat('en-US', {
  notation: 'compact',
  maximumFractionDigits: 1
});

export function QuotaProgress({ items, updatedAt, title = 'Quota status' }: QuotaProgressProps) {
  return (
    <div className="card quota-card reveal delay-1">
      <div className="card-header">
        <div>
          <p className="eyebrow">{updatedAt ? `Updated ${updatedAt}` : 'Live usage'}</p>
          <h2 className="card-title">{title}</h2>
        </div>
        <span className="pill">Auto alert</span>
      </div>
      <div className="quota-list">
        {items.map((item) => {
          const percent = Math.min(item.percentUsed, 100);
          const level = percent >= 95 ? 'danger' : percent >= 80 ? 'warn' : 'safe';
          const limitLabel =
            item.limit === null ? 'Unlimited' : numberFormatter.format(item.limit);

          return (
            <div key={item.label} className="quota-item">
              <div className="quota-item__row">
                <span className="quota-item__label">{item.label}</span>
                <span className="quota-item__value">
                  {numberFormatter.format(item.used)} / {limitLabel} {item.unit}
                </span>
              </div>
              <div className="quota-bar">
                <div
                  className="quota-bar__fill"
                  data-level={level}
                  style={{ width: `${percent}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
