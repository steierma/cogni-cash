import { api } from '../client';
import type { Invoice } from "../types/invoice";
import type { AxiosResponse } from 'axios';

export interface InvoiceUpdatePayload {
    vendor?: { id?: string; name: string };
    description?: string;
    category_id?: string | null;
    amount?: number;
    currency?: string;
    issued_at?: string;
    metadata?: Record<string, any>;
}

export const invoiceService = {
    fetchInvoices: (source?: string): Promise<Invoice[]> =>
        api.get<Invoice[]>('invoices/', {
            params: { source }
        }).then((r: AxiosResponse<Invoice[]>) => r.data ?? []),

    fetchInvoice: (id: string): Promise<Invoice> =>
        api.get<Invoice>(`invoices/${id}/`).then((r: AxiosResponse<Invoice>) => r.data),

    import: (file: File): Promise<Invoice> => {
        const form = new FormData();
        form.append('file', file);
        return api.post<Invoice>('invoices/import/', form, {
            headers: { 'Content-Type': 'multipart/form-data' }
        }).then((r: AxiosResponse<Invoice>) => r.data);
    },

    update: (id: string, data: InvoiceUpdatePayload): Promise<Invoice> => {
        // Map frontend shape → backend updateInvoiceRequest shape if necessary
        // The backend expects: { vendor_id, vendor_name, description, category_id, amount, currency, issued_at }
        const payload = {
            vendor_id: data.vendor?.id || undefined,
            vendor_name: data.vendor?.name,
            description: data.description,
            category_id: data.category_id,
            amount: data.amount,
            currency: data.currency,
            issued_at: data.issued_at,
            metadata: data.metadata
        };
        return api.put<Invoice>(`invoices/${id}/`, payload).then((r: AxiosResponse<Invoice>) => r.data);
    },

    delete: (id: string): Promise<void> =>
        api.delete(`invoices/${id}/`).then(() => undefined),

    downloadFile: async (id: string, vendorName?: string): Promise<void> => {
        const response = await api.get<Blob>(`invoices/${id}/download/`, {
            responseType: 'blob',
        });
        const blob = new Blob([response.data], { type: response.headers['content-type'] });
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        const filename = vendorName ? `Invoice-${vendorName.replace(/\s+/g, '_')}-${id.slice(0, 8)}.pdf` : `Invoice-${id}.pdf`;
        link.setAttribute('download', filename);
        document.body.appendChild(link);
        link.click();
        link.parentNode?.removeChild(link);
        window.URL.revokeObjectURL(url);
    },

    getPreviewUrl: async (id: string): Promise<{ url: string; mimeType: string }> => {
        const response = await api.get<Blob>(`invoices/${id}/download/`, {
            responseType: 'blob',
        });
        const mimeType = response.headers['content-type'] || 'application/pdf';
        return {
            url: URL.createObjectURL(response.data),
            mimeType,
        };
    },

    share: (id: string, userId: string, permission: 'view' | 'edit' = 'view'): Promise<void> =>
        api.post(`invoices/${id}/share/`, { user_id: userId, permission }).then(() => undefined),

    revokeShare: (id: string, userId: string): Promise<void> =>
        api.delete(`invoices/${id}/share/${userId}/`).then(() => undefined),

    fetchShares: (id: string): Promise<string[]> =>
        api.get<string[]>(`invoices/${id}/shares/`).then((r: AxiosResponse<string[]>) => r.data ?? [])
};
