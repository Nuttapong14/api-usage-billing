package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"

	billingdomain "github.com/your-org/api-usage-billing/backend/internal/domain/billing"
	subscriptionmodel "github.com/your-org/api-usage-billing/backend/internal/domain/subscription"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/pdf"
	"github.com/your-org/api-usage-billing/backend/internal/pkg/storage"
	"github.com/your-org/api-usage-billing/backend/internal/repository"
)

// ServiceImpl implements billing operations.
type ServiceImpl struct {
	invoices      repository.InvoiceRepository
	payments      repository.PaymentRepository
	subscriptions repository.SubscriptionRepository
	tiers         repository.TierRepository
	usage         repository.UsageRepository
	pdfGenerator  pdf.Generator
	uploader      storage.Uploader
	clock         func() time.Time
}

// NewService creates a new billing service.
func NewService(
	invoices repository.InvoiceRepository,
	payments repository.PaymentRepository,
	subscriptions repository.SubscriptionRepository,
	tiers repository.TierRepository,
	usage repository.UsageRepository,
	pdfGenerator pdf.Generator,
	uploader storage.Uploader,
) *ServiceImpl {
	return &ServiceImpl{
		invoices:      invoices,
		payments:      payments,
		subscriptions: subscriptions,
		tiers:         tiers,
		usage:         usage,
		pdfGenerator:  pdfGenerator,
		uploader:      uploader,
		clock:         time.Now,
	}
}

// ListInvoices returns invoices for a customer.
func (s *ServiceImpl) ListInvoices(ctx context.Context, params ListInvoicesParams) (*InvoiceList, error) {
	filters, err := toInvoiceFilters(params)
	if err != nil {
		return nil, err
	}
	if err := s.validateInvoiceFilters(filters); err != nil {
		return nil, err
	}

	invoices, total, err := s.invoices.ListForCustomer(ctx, filters)
	if err != nil {
		return nil, err
	}

	data := make([]InvoiceSummary, 0, len(invoices))
	for i := range invoices {
		data = append(data, toInvoiceSummary(&invoices[i]))
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filters.Offset
	if offset < 0 {
		offset = 0
	}
	return &InvoiceList{
		Data: data,
		Pagination: Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: int64(offset+limit) < total,
		},
	}, nil
}

// GetInvoice returns a single invoice.
func (s *ServiceImpl) GetInvoice(ctx context.Context, params GetInvoiceParams) (*Invoice, error) {
	invoice, err := s.getInvoiceDomain(ctx, params.OrganizationID, params.CustomerID, params.InvoiceID)
	if err != nil {
		return nil, err
	}

	return toInvoiceResponse(invoice)
}

// GetInvoicePDF generates or retrieves the invoice PDF.
func (s *ServiceImpl) GetInvoicePDF(ctx context.Context, params GetInvoicePDFParams) (*InvoicePDF, error) {
	if s.pdfGenerator == nil {
		return nil, ErrPDFNotAvailable
	}

	invoice, err := s.getInvoiceDomain(ctx, params.OrganizationID, params.CustomerID, params.InvoiceID)
	if err != nil {
		return nil, err
	}

	response, err := toInvoiceResponse(invoice)
	if err != nil {
		return nil, err
	}

	language := normalizeLanguage(params.Language)
	pdfData, err := s.pdfGenerator.GenerateInvoice(ctx, pdf.InvoiceDocument{
		Invoice:  toPDFInvoice(*response),
		Language: language,
	})
	if err != nil {
		return nil, err
	}

	var pdfURL *string
	if s.uploader != nil {
		objectName := fmt.Sprintf("invoices/%s/%s.pdf", invoice.CustomerID, invoice.ID)
		url, err := s.uploader.Upload(ctx, objectName, pdfData, "application/pdf")
		if err == nil {
			pdfURL = &url
			invoice.PDFURL = pdfURL
			now := s.clock().UTC()
			invoice.PDFGeneratedAt = &now
			_ = s.invoices.Update(ctx, invoice)
		}
	}

	return &InvoicePDF{
		Data:        pdfData,
		ContentType: "application/pdf",
		URL:         pdfURL,
	}, nil
}

// PayInvoice records a payment initiation for an invoice.
func (s *ServiceImpl) PayInvoice(ctx context.Context, params PayInvoiceParams) (*PaymentResponse, error) {
	invoice, err := s.getInvoiceDomain(ctx, params.OrganizationID, params.CustomerID, params.InvoiceID)
	if err != nil {
		return nil, err
	}

	if invoice.Status == billingdomain.InvoiceStatusPaid || invoice.Status == billingdomain.InvoiceStatusCancelled || invoice.Status == billingdomain.InvoiceStatusRefunded {
		return nil, ErrInvoiceNotPayable
	}

	if invoice.Status == billingdomain.InvoiceStatusDraft {
		_ = s.invoices.UpdateStatus(ctx, invoice.ID, billingdomain.InvoiceStatusPending, nil)
	}

	method := params.Request.PaymentMethod
	payment := &billingdomain.Payment{
		InvoiceID:      invoice.ID,
		OrganizationID: invoice.OrganizationID,
		Amount:         invoice.Total,
		Currency:       invoice.Currency,
		PaymentMethod:  &method,
		Status:         billingdomain.PaymentStatusPending,
		GatewayResponse: datatypes.JSON([]byte("{}")),
		CreatedAt:      s.clock().UTC(),
		UpdatedAt:      s.clock().UTC(),
	}

	if err := s.payments.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &PaymentResponse{
		PaymentID: payment.ID,
		Status:    "pending",
	}, nil
}

// GenerateInvoice creates a new invoice for a billing period.
func (s *ServiceImpl) GenerateInvoice(ctx context.Context, params GenerateInvoiceParams) (*Invoice, error) {
	if s.subscriptions == nil || s.tiers == nil || s.usage == nil {
		return nil, ErrInvoiceGenerationUnavailable
	}
	if err := s.validateInvoiceDates(params); err != nil {
		return nil, err
	}

	sub, err := s.subscriptions.GetByCustomerID(ctx, params.CustomerID)
	if err != nil {
		return nil, err
	}

	tier, err := s.tiers.GetByID(ctx, sub.TierID)
	if err != nil {
		return nil, err
	}
	if params.OrganizationID != uuid.Nil && tier.OrganizationID != params.OrganizationID {
		return nil, ErrInvalidInvoiceFilter
	}

	billingCycle := string(sub.BillingCycle)
	price := tierPriceForCycle(tier, billingCycle)
		lineItems := []LineItem{
			{
				Type:        "subscription",
				Description: fmt.Sprintf("%s Plan - %s", tier.Name, formatBillingCycle(billingCycle)),
				Quantity:    1,
				UnitPrice:   moneyFromDecimal(price, tier.Currency),
				Amount:      moneyFromDecimal(price, tier.Currency),
			},
		}

	metrics, err := s.usage.GetUsageSummary(ctx, tier.OrganizationID, params.CustomerID, nil, params.PeriodStart, params.PeriodEnd.AddDate(0, 0, 1))
	if err != nil {
		return nil, err
	}

	overageItems, _ := CalculateOverage(metrics, tier)
	lineItems = append(lineItems, overageItems...)

	discount := decimal.Zero
	subtotal, tax, total := CalculateInvoiceAmounts(lineItems, discount, tier.Currency)

	issueDate := params.IssueDate
	if issueDate.IsZero() {
		issueDate = s.clock().UTC()
	}
	dueDate := params.DueDate
	if dueDate.IsZero() {
		dueDate = issueDate.AddDate(0, 0, 30)
	}

	lineItemsJSON, err := encodeLineItems(lineItems)
	if err != nil {
		return nil, err
	}

	invoice := &billingdomain.Invoice{
		InvoiceNumber:      generateInvoiceNumber(issueDate),
		CustomerID:         params.CustomerID,
		SubscriptionID:     resolveSubscriptionID(sub, params.SubscriptionID),
		OrganizationID:     tier.OrganizationID,
		BillingPeriodStart: params.PeriodStart,
		BillingPeriodEnd:   params.PeriodEnd,
		Subtotal:           subtotal,
		DiscountAmount:     discount,
		TaxRate:            decimal.NewFromFloat(0.07),
		TaxAmount:          tax,
		Total:              total,
		Currency:           tier.Currency,
		Status:             billingdomain.InvoiceStatusPending,
		IssueDate:          issueDate,
		DueDate:            dueDate,
		LineItems:          lineItemsJSON,
		Metadata:           datatypes.JSON([]byte("{}")),
		CreatedAt:          s.clock().UTC(),
		UpdatedAt:          s.clock().UTC(),
	}

	if err := s.invoices.Create(ctx, invoice); err != nil {
		return nil, err
	}

	return toInvoiceResponse(invoice)
}

func (s *ServiceImpl) getInvoiceDomain(ctx context.Context, orgID, customerID, invoiceID uuid.UUID) (*billingdomain.Invoice, error) {
	invoice, err := s.invoices.GetByIDForCustomer(ctx, customerID, invoiceID)
	if err != nil {
		return nil, err
	}
	if orgID != uuid.Nil && invoice.OrganizationID != orgID {
		return nil, repository.ErrInvoiceNotFound
	}
	return invoice, nil
}

func toInvoiceFilters(params ListInvoicesParams) (repository.InvoiceFilters, error) {
	filters := repository.InvoiceFilters{
		OrganizationID: params.OrganizationID,
		CustomerID:     params.CustomerID,
		StartDate:      params.StartDate,
		EndDate:        params.EndDate,
		Limit:          params.Limit,
		Offset:         params.Offset,
		SortBy:         params.SortBy,
		SortDir:        params.SortDir,
	}

	if params.Status != "" {
		status := billingdomain.InvoiceStatus(strings.ToLower(params.Status))
		switch status {
		case billingdomain.InvoiceStatusDraft, billingdomain.InvoiceStatusPending, billingdomain.InvoiceStatusPaid,
			billingdomain.InvoiceStatusOverdue, billingdomain.InvoiceStatusCancelled, billingdomain.InvoiceStatusRefunded:
			filters.Status = &status
		default:
			return repository.InvoiceFilters{}, ErrInvalidInvoiceFilter
		}
	}

	return filters, nil
}

func toInvoiceResponse(invoice *billingdomain.Invoice) (*Invoice, error) {
	if invoice == nil {
		return nil, repository.ErrInvoiceNotFound
	}

	items, err := decodeLineItems(invoice.LineItems)
	if err != nil {
		return nil, err
	}

	taxRate, _ := invoice.TaxRate.Float64()
	amounts := InvoiceAmounts{
		Subtotal:  moneyFromDecimal(invoice.Subtotal, invoice.Currency),
		Discount:  moneyFromDecimal(invoice.DiscountAmount, invoice.Currency),
		TaxRate:   taxRate,
		TaxAmount: moneyFromDecimal(invoice.TaxAmount, invoice.Currency),
		Total:     moneyFromDecimal(invoice.Total, invoice.Currency),
	}

	return &Invoice{
		ID:             invoice.ID,
		InvoiceNumber:  invoice.InvoiceNumber,
		CustomerID:     invoice.CustomerID,
		SubscriptionID: invoice.SubscriptionID,
		BillingPeriod: BillingPeriod{
			Start: invoice.BillingPeriodStart,
			End:   invoice.BillingPeriodEnd,
		},
		Amounts:   amounts,
		LineItems: items,
		Status:    string(invoice.Status),
		Dates: InvoiceDates{
			IssueDate: invoice.IssueDate,
			DueDate:   invoice.DueDate,
			PaidAt:    invoice.PaidAt,
		},
		PDFURL:    invoice.PDFURL,
		Notes:     invoice.Notes,
		CreatedAt: invoice.CreatedAt,
		UpdatedAt: invoice.UpdatedAt,
	}, nil
}

func toInvoiceSummary(invoice *billingdomain.Invoice) InvoiceSummary {
	return InvoiceSummary{
		ID:            invoice.ID,
		InvoiceNumber: invoice.InvoiceNumber,
		Total:         moneyFromDecimal(invoice.Total, invoice.Currency),
		Status:        string(invoice.Status),
		IssueDate:     invoice.IssueDate,
		DueDate:       invoice.DueDate,
		PDFURL:        invoice.PDFURL,
	}
}

type lineItemsPayload struct {
	Items []LineItem `json:"items"`
}

func encodeLineItems(items []LineItem) (datatypes.JSON, error) {
	payload := lineItemsPayload{Items: items}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(data), nil
}

func decodeLineItems(data datatypes.JSON) ([]LineItem, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var payload lineItemsPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload.Items, nil
}

func generateInvoiceNumber(issueDate time.Time) string {
	suffix := uuid.NewString()[:8]
	return fmt.Sprintf("INV-%s-%s", issueDate.UTC().Format("2006"), strings.ToUpper(suffix))
}

func normalizeLanguage(language string) string {
	if strings.ToLower(language) == "en" {
		return "en"
	}
	return "th"
}

func formatBillingCycle(value string) string {
	switch strings.ToLower(value) {
	case "yearly":
		return "Yearly"
	case "monthly":
		return "Monthly"
	default:
		return value
	}
}

func moneyFromDecimal(amount decimal.Decimal, currency string) Money {
	value, _ := amount.Float64()
	return Money{Amount: value, Currency: currency}
}

func toPDFInvoice(invoice Invoice) pdf.Invoice {
	lineItems := make([]pdf.LineItem, 0, len(invoice.LineItems))
	for _, item := range invoice.LineItems {
		lineItems = append(lineItems, pdf.LineItem{
			Type:        item.Type,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   pdf.Money{Amount: item.UnitPrice.Amount, Currency: item.UnitPrice.Currency},
			Amount:      pdf.Money{Amount: item.Amount.Amount, Currency: item.Amount.Currency},
		})
	}

	return pdf.Invoice{
		InvoiceNumber:     invoice.InvoiceNumber,
		CustomerID:        invoice.CustomerID.String(),
		BillingPeriodStart: invoice.BillingPeriod.Start.Format("2006-01-02"),
		BillingPeriodEnd:   invoice.BillingPeriod.End.Format("2006-01-02"),
		Status:            invoice.Status,
		IssueDate:         invoice.Dates.IssueDate.Format("2006-01-02"),
		DueDate:           invoice.Dates.DueDate.Format("2006-01-02"),
		PaidAt:            formatOptionalTime(invoice.Dates.PaidAt),
		Amounts: pdf.InvoiceAmounts{
			Subtotal:  pdf.Money{Amount: invoice.Amounts.Subtotal.Amount, Currency: invoice.Amounts.Subtotal.Currency},
			Discount:  pdf.Money{Amount: invoice.Amounts.Discount.Amount, Currency: invoice.Amounts.Discount.Currency},
			TaxRate:   invoice.Amounts.TaxRate,
			TaxAmount: pdf.Money{Amount: invoice.Amounts.TaxAmount.Amount, Currency: invoice.Amounts.TaxAmount.Currency},
			Total:     pdf.Money{Amount: invoice.Amounts.Total.Amount, Currency: invoice.Amounts.Total.Currency},
		},
		LineItems: lineItems,
	}
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format("2006-01-02 15:04:05")
}

func resolveSubscriptionID(sub *subscriptionmodel.Subscription, override *uuid.UUID) *uuid.UUID {
	if override != nil {
		return override
	}
	if sub == nil {
		return nil
	}
	return &sub.ID
}

func tierPriceForCycle(tier *subscriptionmodel.SubscriptionTier, cycle string) decimal.Decimal {
	if tier == nil {
		return decimal.Zero
	}
	if cycle == string(subscriptionmodel.BillingCycleYearly) && tier.PriceYearly != nil {
		return *tier.PriceYearly
	}
	return tier.PriceMonthly
}

func isInvoiceStatusPayable(status billingdomain.InvoiceStatus) bool {
	return status == billingdomain.InvoiceStatusPending || status == billingdomain.InvoiceStatusOverdue || status == billingdomain.InvoiceStatusDraft
}

func (s *ServiceImpl) ensureInvoicePayable(invoice *billingdomain.Invoice) error {
	if invoice == nil {
		return repository.ErrInvoiceNotFound
	}
	if !isInvoiceStatusPayable(invoice.Status) {
		return ErrInvoiceNotPayable
	}
	return nil
}

func (s *ServiceImpl) validateInvoiceFilters(filters repository.InvoiceFilters) error {
	if filters.StartDate != nil && filters.EndDate != nil {
		if filters.EndDate.Before(*filters.StartDate) {
			return ErrInvalidInvoiceFilter
		}
	}
	return nil
}

func (s *ServiceImpl) validateInvoiceForOrganization(invoice *billingdomain.Invoice, orgID uuid.UUID) error {
	if invoice == nil {
		return repository.ErrInvoiceNotFound
	}
	if orgID != uuid.Nil && invoice.OrganizationID != orgID {
		return repository.ErrInvoiceNotFound
	}
	return nil
}

func (s *ServiceImpl) validateInvoiceForCustomer(invoice *billingdomain.Invoice, customerID uuid.UUID) error {
	if invoice == nil {
		return repository.ErrInvoiceNotFound
	}
	if customerID != uuid.Nil && invoice.CustomerID != customerID {
		return repository.ErrInvoiceNotFound
	}
	return nil
}

func (s *ServiceImpl) validateStatusForPayment(invoice *billingdomain.Invoice) error {
	if invoice == nil {
		return repository.ErrInvoiceNotFound
	}
	if invoice.Status == billingdomain.InvoiceStatusPaid || invoice.Status == billingdomain.InvoiceStatusCancelled || invoice.Status == billingdomain.InvoiceStatusRefunded {
		return ErrInvoiceNotPayable
	}
	return nil
}

func (s *ServiceImpl) validatePaymentRequest(request PaymentRequest) error {
	if strings.TrimSpace(request.PaymentMethod) == "" {
		return ErrInvoiceNotPayable
	}
	return nil
}

func (s *ServiceImpl) validateInvoiceDates(params GenerateInvoiceParams) error {
	if params.PeriodEnd.Before(params.PeriodStart) {
		return ErrInvalidInvoiceFilter
	}
	return nil
}

func (s *ServiceImpl) ensureDependencies() error {
	if s.subscriptions == nil || s.tiers == nil || s.usage == nil {
		return ErrInvoiceGenerationUnavailable
	}
	return nil
}

func (s *ServiceImpl) validateInvoiceState(invoice *billingdomain.Invoice, orgID, customerID uuid.UUID) error {
	if err := s.validateInvoiceForOrganization(invoice, orgID); err != nil {
		return err
	}
	if err := s.validateInvoiceForCustomer(invoice, customerID); err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) validateAndNormalizeListParams(params ListInvoicesParams) (repository.InvoiceFilters, error) {
	filters, err := toInvoiceFilters(params)
	if err != nil {
		return repository.InvoiceFilters{}, err
	}
	if err := s.validateInvoiceFilters(filters); err != nil {
		return repository.InvoiceFilters{}, err
	}
	return filters, nil
}

func (s *ServiceImpl) validateAndNormalizeGeneration(params GenerateInvoiceParams) error {
	if err := s.ensureDependencies(); err != nil {
		return err
	}
	if err := s.validateInvoiceDates(params); err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) validateAndNormalizePayment(invoice *billingdomain.Invoice, request PaymentRequest) error {
	if err := s.validateStatusForPayment(invoice); err != nil {
		return err
	}
	if err := s.validatePaymentRequest(request); err != nil {
		return err
	}
	return nil
}

func (s *ServiceImpl) validateAndNormalizeInvoice(ctx context.Context, orgID, customerID, invoiceID uuid.UUID) (*billingdomain.Invoice, error) {
	invoice, err := s.getInvoiceDomain(ctx, orgID, customerID, invoiceID)
	if err != nil {
		return nil, err
	}
	if err := s.validateInvoiceState(invoice, orgID, customerID); err != nil {
		return nil, err
	}
	return invoice, nil
}

var _ Service = (*ServiceImpl)(nil)
