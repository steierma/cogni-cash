export interface Vendor {
    id: string;
    name: string;
}

export interface InvoiceSplit {
    id: string;
    invoice_id: string;
    category_id: string;
    amount: number;
    base_amount: number;
    description: string;
}

export interface Invoice {
    id: string;
    user_id: string;
    vendor: Vendor;
    category_id: string | null;
    amount: number;
    currency: string;
    base_amount: number;
    base_currency: string;
    issued_at: string;
    description: string;
    is_shared: boolean;
    shared_with?: string[];
    owner_id: string;
    splits?: InvoiceSplit[];
}