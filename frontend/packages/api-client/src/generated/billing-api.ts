import { APIClient, getDefaultClient } from '../client';
import { APIResponse } from '../types';
import { interpolatePath } from './runtime';

export namespace BillingApi {
  export interface Invoice {
    id: string;
    invoice_number: string;
    customer_id: string;
    subscription_id?: string;
    billing_period:   {
      start?: string;
      end?: string;
    };
    amounts: BillingApi.InvoiceAmounts;
    line_items?: BillingApi.LineItem[];
    status: "draft" | "pending" | "paid" | "overdue" | "cancelled" | "refunded";
    dates:   {
      issue_date?: string;
      due_date?: string;
      paid_at?: string;
    };
    pdf_url?: string;
    notes?: string;
    created_at?: string;
    updated_at?: string;
  }

  export interface InvoiceAmounts {
    subtotal?: BillingApi.Money;
    discount?: BillingApi.Money;
    tax_rate?: number;
    tax_amount?: BillingApi.Money;
    total?: BillingApi.Money;
  }

  export interface LineItem {
    type?: "subscription" | "overage" | "discount" | "adjustment";
    description?: string;
    quantity?: number;
    unit_price?: BillingApi.Money;
    amount?: BillingApi.Money;
  }

  export interface Money {
    amount?: number;
    currency?: string;
  }

  export interface InvoiceList {
    data?: BillingApi.InvoiceSummary[];
    pagination?: BillingApi.Pagination;
  }

  export interface InvoiceSummary {
    id?: string;
    invoice_number?: string;
    total?: BillingApi.Money;
    status?: string;
    issue_date?: string;
    due_date?: string;
    pdf_url?: string;
  }

  export interface PaymentRequest {
    payment_method: "credit_card" | "bank_transfer" | "promptpay";
    return_url?: string;
    payment_source?:   {
      token?: string;
    };
  }

  export interface PaymentResponse {
    payment_id?: string;
    status?: "pending" | "requires_action" | "completed";
    redirect_url?: string;
    qr_code?: string;
  }

  export interface Payment {
    id?: string;
    invoice_id?: string;
    amount?: BillingApi.Money;
    payment_method?: string;
    payment_gateway?: string;
    gateway_transaction_id?: string;
    status?: "pending" | "completed" | "failed" | "refunded";
    paid_at?: string;
    created_at?: string;
  }

  export interface PaymentList {
    data?: BillingApi.Payment[];
    pagination?: BillingApi.Pagination;
  }

  export interface CreditNote {
    id?: string;
    credit_note_number?: string;
    invoice_id?: string;
    type?: "credit" | "debit";
    amounts?:   {
      amount?: BillingApi.Money;
      tax_amount?: BillingApi.Money;
      total?: BillingApi.Money;
    };
    reason?: string;
    line_items?: BillingApi.LineItem[];
    status?: "issued" | "applied" | "cancelled";
    pdf_url?: string;
    created_at?: string;
  }

  export interface CreditNoteList {
    data?: BillingApi.CreditNote[];
    pagination?: BillingApi.Pagination;
  }

  export interface Pagination {
    total?: number;
    limit?: number;
    offset?: number;
    has_more?: boolean;
  }

  export interface Error {
    error_code: string;
    message: string;
    details?: Record<string, unknown>;
  }

  export interface GetinvoicePathParams {
    id: string;
  }

  export interface DownloadinvoicepdfPathParams {
    id: string;
  }

  export interface PayinvoicePathParams {
    id: string;
  }

  export interface GetpaymentPathParams {
    id: string;
  }

  export interface GetcreditnotePathParams {
    id: string;
  }

  export interface ListinvoicesQueryParams {
    status?: "draft" | "pending" | "paid" | "overdue" | "cancelled";
    start_date?: string;
    end_date?: string;
    limit?: number;
    offset?: number;
    sort_by?: "issue_date" | "due_date" | "total" | "status";
  }

  export interface DownloadinvoicepdfQueryParams {
    language?: "th" | "en";
  }

  export interface ListpaymentsQueryParams {
    invoice_id?: string;
    status?: "pending" | "completed" | "failed" | "refunded";
    limit?: number;
    offset?: number;
  }

  export interface ListcreditnotesQueryParams {
    invoice_id?: string;
    type?: "credit" | "debit";
    limit?: number;
    offset?: number;
  }

  export interface ListinvoicesParams {
    query?: BillingApi.ListinvoicesQueryParams;
  }

  export interface GetinvoiceParams {
    path: BillingApi.GetinvoicePathParams;
  }

  export interface DownloadinvoicepdfParams {
    path: BillingApi.DownloadinvoicepdfPathParams;
    query?: BillingApi.DownloadinvoicepdfQueryParams;
  }

  export interface PayinvoiceParams {
    path: BillingApi.PayinvoicePathParams;
    body: BillingApi.PaymentRequest;
  }

  export interface ListpaymentsParams {
    query?: BillingApi.ListpaymentsQueryParams;
  }

  export interface GetpaymentParams {
    path: BillingApi.GetpaymentPathParams;
  }

  export interface ListcreditnotesParams {
    query?: BillingApi.ListcreditnotesQueryParams;
  }

  export interface GetcreditnoteParams {
    path: BillingApi.GetcreditnotePathParams;
  }

}

export const billingApiClient = {
  async listInvoices(params: BillingApi.ListinvoicesParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.InvoiceList>> {
    const url = 'v1/invoices';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< BillingApi.InvoiceList >(url, { params: query });
  },
  async getInvoice(params: BillingApi.GetinvoiceParams, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.Invoice>> {
    const url = interpolatePath('v1/invoices/{id}', params.path);
    const query = undefined;
    return client.get< BillingApi.Invoice >(url, { params: query });
  },
  async downloadInvoicePdf(params: BillingApi.DownloadinvoicepdfParams, client: APIClient = getDefaultClient()): Promise<APIResponse<{
  message?: string;
  retry_after?: number;
}>> {
    const url = interpolatePath('v1/invoices/{id}/pdf', params.path);
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< {
  message?: string;
  retry_after?: number;
} >(url, { params: query });
  },
  async payInvoice(params: BillingApi.PayinvoiceParams, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.PaymentResponse>> {
    const url = interpolatePath('v1/invoices/{id}/pay', params.path);
    const query = undefined;
    return client.post< BillingApi.PaymentResponse >(url, params.body, { params: query });
  },
  async listPayments(params: BillingApi.ListpaymentsParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.PaymentList>> {
    const url = 'v1/payments';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< BillingApi.PaymentList >(url, { params: query });
  },
  async getPayment(params: BillingApi.GetpaymentParams, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.Payment>> {
    const url = interpolatePath('v1/payments/{id}', params.path);
    const query = undefined;
    return client.get< BillingApi.Payment >(url, { params: query });
  },
  async listCreditNotes(params: BillingApi.ListcreditnotesParams = {}, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.CreditNoteList>> {
    const url = 'v1/credit-notes';
    const query = params.query as Record<string, string | number | boolean> | undefined;
    return client.get< BillingApi.CreditNoteList >(url, { params: query });
  },
  async getCreditNote(params: BillingApi.GetcreditnoteParams, client: APIClient = getDefaultClient()): Promise<APIResponse<BillingApi.CreditNote>> {
    const url = interpolatePath('v1/credit-notes/{id}', params.path);
    const query = undefined;
    return client.get< BillingApi.CreditNote >(url, { params: query });
  },
};
