package billing

import (
	"github.com/shopspring/decimal"

	"github.com/Nuttapong14/api-usage-billing/internal/pkg/currency"
)

// CalculateInvoiceAmounts calculates totals for the given line items and discount.
func CalculateInvoiceAmounts(lineItems []LineItem, discount decimal.Decimal, currencyCode string) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	subtotal := decimal.Zero
	for _, item := range lineItems {
		subtotal = subtotal.Add(decimal.NewFromFloat(item.Amount.Amount))
	}

	if discount.IsNegative() {
		discount = decimal.Zero
	}
	taxable := subtotal.Sub(discount)
	if taxable.IsNegative() {
		taxable = decimal.Zero
	}

	tax := currency.CalculateVAT(taxable)
	total := taxable.Add(tax)
	return subtotal, tax, total
}

// BuildInvoiceAmounts builds the response amounts for an invoice.
func BuildInvoiceAmounts(subtotal, discount, tax, total decimal.Decimal, currencyCode string) InvoiceAmounts {
	taxRate, _ := currency.VATRate().Float64()
	return InvoiceAmounts{
		Subtotal:  moneyFromDecimal(subtotal, currencyCode),
		Discount:  moneyFromDecimal(discount, currencyCode),
		TaxRate:   taxRate,
		TaxAmount: moneyFromDecimal(tax, currencyCode),
		Total:     moneyFromDecimal(total, currencyCode),
	}
}
