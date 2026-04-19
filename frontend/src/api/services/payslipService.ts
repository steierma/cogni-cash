import { api } from '../client';
import type { Payslip, PayslipSummary } from "../types/payslip";
export const payslipService = {
    fetchPayslips: (employer?: string): Promise<Payslip[]> =>
        api.get<Payslip[]>('payslips/', { params: { employer: employer || undefined } }).then(r => r.data ?? []),

    fetchSummary: (): Promise<PayslipSummary> =>
        api.get<PayslipSummary>('payslips/summary/').then(r => r.data),

    fetchPayslip: (id: string): Promise<Payslip> =>
        api.get<Payslip>(`payslips/${id}/`).then(r => r.data),

    getPreviewUrl: async (id: string): Promise<{ url: string; mimeType: string }> => {
        const response = await api.get<Blob>(`payslips/${id}/download/`, { responseType: 'blob' });
        const mimeType = response.headers['content-type'] || 'application/pdf';
        return { url: window.URL.createObjectURL(new Blob([response.data], { type: mimeType })), mimeType };
    },

    import: async ({ file, overrides, useAI }: { file: File; overrides?: Partial<Payslip>; useAI?: boolean }) => {
        const form = new FormData();
        form.append('file', file);
        if (useAI) form.append('use_ai', 'true');
        if (overrides) {
            Object.entries(overrides).forEach(([key, val]) => {
                if (val !== undefined) form.append(key, typeof val === 'object' ? JSON.stringify(val) : String(val));
            });
        }
        return api.post('payslips/import/', form, { headers: { 'Content-Type': 'multipart/form-data' } }).then(r => r.data);
    },

    importBatch: (files: File[]): Promise<{ successful: Payslip[], failed: { filename: string, error: string }[] }> => {
        const form = new FormData();
        files.forEach(f => form.append('files', f));
        return api.post('payslips/import/batch/', form, { headers: { 'Content-Type': 'multipart/form-data' } }).then(r => r.data);
    },

    update: (id: string, data: Partial<Payslip> | FormData): Promise<Payslip> => {
        const config = data instanceof FormData ? { headers: { 'Content-Type': 'multipart/form-data' } } : undefined;
        return api.put(`payslips/${id}/`, data, config).then(r => r.data);
    },

    delete: (id: string): Promise<void> =>
        api.delete(`payslips/${id}/`).then(() => undefined),

    downloadFile: async (id: string, originalName: string): Promise<void> => {
        const response = await api.get<Blob>(`payslips/${id}/download/`, { responseType: 'blob' });
        const ct = response.headers['content-type'] as string | undefined;
        let filename = originalName || `payslip-${id}.pdf`;
        if (!originalName) {
            if (ct === 'image/jpeg') filename = `payslip-${id}.jpg`;
            else if (ct === 'image/png') filename = `payslip-${id}.png`;
        }
        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a'); link.href = url; link.setAttribute('download', filename);
        document.body.appendChild(link); link.click();
        link.parentNode?.removeChild(link); window.URL.revokeObjectURL(url);
    }
};