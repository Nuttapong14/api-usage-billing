package currency

import "github.com/shopspring/decimal"

var (
	defaultVATRate = decimal.NewFromInt(7).Div(decimal.NewFromInt(100))
)

// VATRate returns the default VAT rate (7%).
func VATRate() decimal.Decimal {
	return defaultVATRate
}

// CalculateVAT returns the VAT amount for the given subtotal.
func CalculateVAT(subtotal decimal.Decimal) decimal.Decimal {
	return subtotal.Mul(defaultVATRate).Round(2)
}

// ApplyVAT returns the tax amount and total for the given subtotal.
func ApplyVAT(subtotal decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	tax := CalculateVAT(subtotal)
	return tax, subtotal.Add(tax)
}
