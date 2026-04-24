import { api } from '../client';
import type { Document, DocumentType, TaxYearSummary } from '../types';
import type { AxiosResponse } from 'axios';

export interface DocumentUpdatePayload {
    file_name?: string;
    type?: DocumentType;
    document_date?: string;
    metadata?: Record<string, unknown>;
}

export const documentService = {
    list: (type?: DocumentType, search?: string): Promise<Document[]> =>
        api.get<Document[]>('documents/', {
            params: { type, search }
        }).then((r: AxiosResponse<Document[]>) => r.data ?? []),

    upload: (file: File, type: DocumentType = 'other', skipAi: boolean = false, date?: string, fileName?: string): Promise<Document> => {
        const form = new FormData();
        form.append('file', file);
        form.append('document_type', type);
        form.append('skip_ai', skipAi.toString());
        if (date) form.append('document_date', date);
        if (fileName) form.append('file_name', fileName);

        return api.post<Document>('documents/upload/', form, {
            headers: { 'Content-Type': 'multipart/form-data' }
        }).then((r: AxiosResponse<Document>) => r.data);
    },

    getDetail: (id: string): Promise<Document> =>
        api.get<Document>(`documents/${id}/`).then((r: AxiosResponse<Document>) => r.data),

    update: (id: string, data: DocumentUpdatePayload): Promise<Document> =>
        api.put<Document>(`documents/${id}/`, data).then((r: AxiosResponse<Document>) => r.data),

    delete: (id: string): Promise<void> =>
        api.delete(`documents/${id}/`).then(() => undefined),

    getDownloadUrl: (id: string): string => `/api/v1/documents/${id}/download/`,

    getPreview: async (id: string): Promise<{ url: string; mimeType: string }> => {
        const response = await api.get<Blob>(`documents/${id}/download/`, {
            responseType: 'blob',
        });

        const mimeType = response.headers['content-type'] || 'application/pdf';
        return {
            url: URL.createObjectURL(response.data),
            mimeType,
        };
    },

    getTaxSummary: (year: number): Promise<TaxYearSummary> =>
        api.get<TaxYearSummary>(`documents/tax-summary/${year}/`).then((r: AxiosResponse<TaxYearSummary>) => r.data)
};