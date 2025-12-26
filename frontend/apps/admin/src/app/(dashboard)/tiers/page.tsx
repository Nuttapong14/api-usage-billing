const tiers = [
  {
    name: 'Launch',
    price: '$0',
    description: 'Proof of concept and internal teams.',
    quota: '50k requests / month',
    rateLimit: '120 req/min',
    support: 'Community support'
  },
  {
    name: 'Growth',
    price: '$199',
    description: 'Scaling production traffic and external partners.',
    quota: '500k requests / month',
    rateLimit: '600 req/min',
    support: 'Priority email support'
  },
  {
    name: 'Enterprise',
    price: 'Custom',
    description: 'Dedicated SLAs and custom contract terms.',
    quota: 'Unlimited usage',
    rateLimit: 'Custom rate limits',
    support: 'Dedicated success team'
  }
];

export default function TiersPage() {
  return (
    <div className="admin-page">
      <header className="page-header reveal">
        <div>
          <p className="eyebrow">Subscription design</p>
          <h1 className="page-title">Tier management</h1>
          <p className="page-subtitle">
            Configure pricing, quota ceilings, and commercial terms for each subscription tier.
          </p>
        </div>
        <div className="header-actions">
          <button className="button button--ghost" type="button">
            View change log
          </button>
          <button className="button" type="button">
            New tier
          </button>
        </div>
      </header>

      <section className="tier-grid">
        {tiers.map((tier) => (
          <div className="card tier-card reveal" key={tier.name}>
            <div className="tier-header">
              <div>
                <p className="eyebrow">{tier.name}</p>
                <h2 className="tier-price">{tier.price}</h2>
                <p className="tier-desc">{tier.description}</p>
              </div>
              <span className="pill">{tier.quota}</span>
            </div>
            <div className="tier-details">
              <p>
                <strong>Rate limit:</strong> {tier.rateLimit}
              </p>
              <p>
                <strong>Support:</strong> {tier.support}
              </p>
            </div>
            <button className="button button--outline" type="button">
              Edit tier
            </button>
          </div>
        ))}
      </section>

      <section className="card callout-card reveal">
        <div>
          <p className="eyebrow">Usage forecasting</p>
          <h2 className="card-title">Model upcoming capacity needs</h2>
          <p className="card-subtitle">
            Use historical demand curves to set quotas that protect margins without limiting growth.
          </p>
        </div>
        <button className="button button--ghost" type="button">
          Open forecasts
        </button>
      </section>
    </div>
  );
}
