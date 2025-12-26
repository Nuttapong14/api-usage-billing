package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *ServiceImpl) revenueReport(
	ctx context.Context,
	orgID uuid.UUID,
	start, end time.Time,
	granularity string,
) (*RevenueReport, error) {
	periodExpr := "date_trunc('day', issue_date)"
	labelFormat := "Jan 2"
	if granularity == "month" {
		periodExpr = "date_trunc('month', issue_date)"
		labelFormat = "Jan 2006"
	}

	query := fmt.Sprintf(`
SELECT
  %s AS period,
  COALESCE(SUM(total), 0)::float8 AS total,
  COUNT(*) AS invoice_count,
  COALESCE(MAX(currency), 'THB') AS currency
FROM invoices
WHERE organization_id = ?
  AND issue_date >= ?
  AND issue_date <= ?
  AND status IN ?
GROUP BY period
ORDER BY period`, periodExpr)

	rows := []struct {
		Period       time.Time
		Total        sql.NullFloat64
		InvoiceCount int64
		Currency     sql.NullString
	}{}

	if err := s.db.WithContext(ctx).Raw(query, orgID, start, end, revenueStatuses).Scan(&rows).Error; err != nil {
		return nil, err
	}

	currency := "THB"
	points := make([]RevenuePoint, 0, len(rows))
	var totalAmount float64

	for _, row := range rows {
		amount := 0.0
		if row.Total.Valid {
			amount = row.Total.Float64
		}
		if row.Currency.Valid && row.Currency.String != "" {
			currency = row.Currency.String
		}

		points = append(points, RevenuePoint{
			Period:       row.Period.Format(labelFormat),
			Date:         row.Period,
			Amount:       Money{Amount: amount, Currency: currency},
			InvoiceCount: row.InvoiceCount,
		})
		totalAmount += amount
	}

	return &RevenueReport{
		Total: Money{
			Amount:   totalAmount,
			Currency: currency,
		},
		PeriodStart: start,
		PeriodEnd:   end,
		Granularity: granularity,
		Data:        points,
	}, nil
}
