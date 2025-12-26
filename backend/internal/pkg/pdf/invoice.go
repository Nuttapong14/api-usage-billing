package pdf

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var invoiceTemplates embed.FS

const (
	invoiceTemplateEN = "invoice_en.tmpl"
	invoiceTemplateTH = "invoice_th.tmpl"
)

// TemplateGenerator renders invoice PDFs using embedded templates.
type TemplateGenerator struct {
	templates *template.Template
}

// NewTemplateGenerator creates a new template-based generator.
func NewTemplateGenerator() (*TemplateGenerator, error) {
	tmpl, err := template.ParseFS(invoiceTemplates, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}
	return &TemplateGenerator{templates: tmpl}, nil
}

// GenerateInvoice renders a PDF document for an invoice.
func (g *TemplateGenerator) GenerateInvoice(_ context.Context, document InvoiceDocument) ([]byte, error) {
	templateName := invoiceTemplateTH
	if strings.ToLower(document.Language) == "en" {
		templateName = invoiceTemplateEN
	}

	var content bytes.Buffer
	if err := g.templates.ExecuteTemplate(&content, templateName, document.Invoice); err != nil {
		return nil, err
	}

	return buildSimplePDF(content.String()), nil
}

func buildSimplePDF(text string) []byte {
	lines := strings.Split(text, "\n")
	var contentBuilder strings.Builder
	contentBuilder.WriteString("BT\n/F1 12 Tf\n14 TL\n72 720 Td\n")
	for i, line := range lines {
		escaped := escapePDFText(strings.TrimSpace(line))
		if escaped == "" {
			continue
		}
		if i > 0 {
			contentBuilder.WriteString("T*\n")
		}
		contentBuilder.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
	}
	contentBuilder.WriteString("ET\n")

	stream := contentBuilder.String()

	objects := []string{
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n",
		"2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n",
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >> endobj\n",
		fmt.Sprintf("4 0 obj << /Length %d >> stream\n%s\nendstream\nendobj\n", len(stream), stream),
		"5 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for _, obj := range objects {
		offsets = append(offsets, buf.Len())
		buf.WriteString(obj)
	}

	xrefStart := buf.Len()
	buf.WriteString("xref\n0 ")
	buf.WriteString(fmt.Sprintf("%d\n", len(objects)+1))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}

	buf.WriteString("trailer << /Size ")
	buf.WriteString(fmt.Sprintf("%d /Root 1 0 R >>\n", len(objects)+1))
	buf.WriteString("startxref\n")
	buf.WriteString(fmt.Sprintf("%d\n", xrefStart))
	buf.WriteString("%%EOF\n")

	return buf.Bytes()
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}
