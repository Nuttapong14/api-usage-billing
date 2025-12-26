package analytics

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

func (s *ServiceImpl) customerRanking(
	ctx context.Context,
	orgID uuid.UUID,
	start, end time.Time,
	limit int,
) (*CustomerRanking, error) {
	revenueRows := []struct {
		CustomerID   uuid.UUID
		Revenue      sql.NullFloat64
		InvoiceCount int64
		LastInvoice  sql.NullTime
		Currency     sql.NullString
	}{}

	revenueQuery := `
SELECT
  customer_id,
  COALESCE(SUM(total), 0)::float8 AS revenue,
  COUNT(*) AS invoice_count,
  MAX(issue_date) AS last_invoice,
  COALESCE(MAX(currency), 'THB') AS currency
FROM invoices
WHERE organization_id = ?
  AND issue_date >= ?
  AND issue_date <= ?
  AND status IN ?
GROUP BY customer_id
ORDER BY revenue DESC
LIMIT ?`

	if err := s.db.WithContext(ctx).Raw(revenueQuery, orgID, start, end, revenueStatuses, limit).Scan(&revenueRows).Error; err != nil {
		return nil, err
	}

	requestRows := []struct {
		CustomerID   uuid.UUID
		RequestCount int64
	}{}

	endExclusive := end.AddDate(0, 0, 1)
	requestQuery := `
SELECT
  customer_id,
  COUNT(*) AS request_count
FROM usage_records
WHERE organization_id = ?
  AND recorded_at >= ?
  AND recorded_at < ?
GROUP BY customer_id
ORDER BY request_count DESC
LIMIT ?`

	if err := s.db.WithContext(ctx).Raw(requestQuery, orgID, start, endExclusive, limit).Scan(&requestRows).Error; err != nil {
		return nil, err
	}

	revenueMap := make(map[uuid.UUID]struct {
		amount       float64
		invoiceCount int64
		lastInvoice  *time.Time
		currency     string
	})
	currency := "THB"
	for _, row := range revenueRows {
		amount := 0.0
		if row.Revenue.Valid {
			amount = row.Revenue.Float64
		}
		rowCurrency := currency
		if row.Currency.Valid && row.Currency.String != "" {
			rowCurrency = row.Currency.String
			currency = rowCurrency
		}
		var lastInvoice *time.Time
		if row.LastInvoice.Valid {
			parsed := row.LastInvoice.Time
			lastInvoice = &parsed
		}
		revenueMap[row.CustomerID] = struct {
			amount       float64
			invoiceCount int64
			lastInvoice  *time.Time
			currency     string
		}{
			amount:       amount,
			invoiceCount: row.InvoiceCount,
			lastInvoice:  lastInvoice,
			currency:     rowCurrency,
		}
	}

	requestMap := make(map[uuid.UUID]int64)
	for _, row := range requestRows {
		requestMap[row.CustomerID] = row.RequestCount
	}

	rankedIDs := make([]uuid.UUID, 0, limit)
	if len(revenueRows) > 0 {
		for _, row := range revenueRows {
			rankedIDs = append(rankedIDs, row.CustomerID)
		}
	} else {
		for _, row := range requestRows {
			rankedIDs = append(rankedIDs, row.CustomerID)
		}
	}

	customers := []struct {
		ID          uuid.UUID
		DisplayName string
		CompanyName *string
		Email       string
	}{}
	if len(rankedIDs) > 0 {
		if err := s.db.WithContext(ctx).
			Table("customers").
			Select("id, display_name, company_name, email").
			Where("id IN ?", rankedIDs).
			Scan(&customers).Error; err != nil {
			return nil, err
		}
	}

	customerMap := make(map[uuid.UUID]struct {
		name  string
		email string
	})
	for _, c := range customers {
		name := c.DisplayName
		if c.CompanyName != nil && *c.CompanyName != "" {
			name = *c.CompanyName
		}
		customerMap[c.ID] = struct {
			name  string
			email string
		}{
			name:  name,
			email: c.Email,
		}
	}

	metrics := make([]CustomerMetric, 0, len(rankedIDs))
	for _, id := range rankedIDs {
		identity, ok := customerMap[id]
		if !ok {
			continue
		}

		revenueEntry := revenueMap[id]
		if revenueEntry.currency == "" {
			revenueEntry.currency = currency
		}
		requests := requestMap[id]
		metrics = append(metrics, CustomerMetric{
			CustomerID: id,
			Name:       identity.name,
			Email:      identity.email,
			Revenue: Money{
				Amount:   revenueEntry.amount,
				Currency: revenueEntry.currency,
			},
			RequestCount: requests,
			InvoiceCount: revenueEntry.invoiceCount,
			LastInvoiceAt: revenueEntry.lastInvoice,
		})
	}

	return &CustomerRanking{
		PeriodStart: start,
		PeriodEnd:   end,
		Data:        metrics,
	}, nil
}
