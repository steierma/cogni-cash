import { api } from '../client';
import type { Invoice, InvoiceSplit } from "../types/invoice";
import type { AxiosResponse } from 'axios';

export interface InvoiceUpdatePayload {
    vendor?: { id?: string; name: string };
    description?: string;
    category_id?: string | null;
    amount?: number;
    currency?: string;
    issued_at?: string;
    metadata?: Record<string, unknown>;
    splits?: Partial<InvoiceSplit>[];
}

export const invoiceService = {
    fetchInvoices: (source?: string): Promise<Invoice[]> =>
        api.get<Invoice[]>('invoices/', {
            params: { source }
        }).then((r: AxiosResponse<Invoice[]>) => r.data ?? []),

    fetchInvoice: (id: string): Promise<Invoice> =>
        api.get<Invoice>(`invoices/${id}/`).then((r: AxiosResponse<Invoice>) => r.data),

    import: (file: File, overrides?: InvoiceUpdatePayload): Promise<Invoice> => {
        const form = new FormData();
        form.append('file', file);
        if (overrides) {
            if (overrides.vendor?.name) form.append('vendor_name', overrides.vendor.name);
            if (overrides.amount) form.append('amount', String(overrides.amount));
            if (overrides.currency) form.append('currency', overrides.currency);
            if (overrides.issued_at) form.append('issued_at', overrides.issued_at);
            if (overrides.category_id) form.append('category_id', overrides.category_id);
            if (overrides.splits && overrides.splits.length > 0) {
                form.append('splits', JSON.stringify(overrides.splits));
            }
        }
        return api.post<Invoice>('invoices/import/', form, {
            headers: { 'Content-Type': 'multipart/form-data' }
        }).then((r: AxiosResponse<Invoice>) => r.data);
    },

    manualImport: (data: InvoiceUpdatePayload): Promise<Invoice> => {
        const payload = {
            vendor_name: data.vendor?.name,
            description: data.description,
            category_id: data.category_id,
            amount: data.amount,
            currency: data.currency,
            issued_at: data.issued_at,
            splits: data.splits
        };
        return api.post<Invoice>('invoices/', payload).then((r: AxiosResponse<Invoice>) => r.data);
    },

    update: (id: string, data: InvoiceUpdatePayload): Promise<Invoice> => {
        // Map frontend shape → backend updateInvoiceRequest shape if necessary
        const payload = {
            vendor_id: data.vendor?.id || undefined,
            vendor_name: data.vendor?.name,
            description: data.description,
            category_id: data.category_id,
            amount: data.amount,
            currency: data.currency,
            issued_at: data.issued_at,
            metadata: data.metadata,
            splits: data.splits
        };
        return api.put<Invoice>(`invoices/${id}/`, payload).then((r: AxiosResponse<Invoice>) => r.data);
    },

    delete: (id: string): Promise<void> =>
        api.delete(`invoices/${id}/`).then(() => undefined),

    updateCategoryBulk: (ids: string[], categoryId: string | null): Promise<void> =>
        api.patch('invoices/bulk-category/', { ids, category_id: categoryId }).then(() => undefined),

    downloadFile: async (id: string, vendorName?: string): Promise<void> =>
 {
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
