export type DocumentType = 'tax_certificate' | 'receipt' | 'contract' | 'other';

export interface Document {
    id: string;
    user_id: string;
    type: DocumentType;
    file_name: string;
    content_hash: string;
    mime_type: string;
    metadata: Record<string, any>;
    created_at: string;
    // UI temporary fields
    document_date?: string;
}

export interface TaxYearSummary {
    year: number;
    documents: Document[];
    total_gross_income: number;
    total_net_income: number;
    total_income_tax: number;
    total_deductible: number;
}