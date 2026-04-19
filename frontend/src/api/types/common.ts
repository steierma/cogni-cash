export interface ImportResult {
    filename: string;
    status: 'imported' | 'duplicate' | 'error';
    error?: string;
    id?: string;
}

export interface ImportBatchResponse {
    summary: {
        total: number;
        imported: number;
    };
    results: ImportResult[];
}