import axios, {type AxiosResponse, type InternalAxiosRequestConfig} from 'axios';
import type {
    BankStatement,
    BankStatementSummary,
    Category,
    ImportBatchResponse,
    Invoice,
    JobState,
    Payslip,
    Reconciliation, ReconciliationPairSuggestion,
    SystemInfo,
    Transaction,
    TransactionAnalytics,
    User
} from './types';

const api = axios.create({baseURL: '/api/v1'});

// ── Authentication Interceptors ───────────────────────────────────────────────

api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

api.interceptors.response.use(
    (response: AxiosResponse) => response,
    (error: unknown) => {
        if (axios.isAxiosError(error) && error.response?.status === 401) {
            localStorage.removeItem('auth_token');
            if (window.location.pathname !== '/login') {
                window.location.href = '/login';
            }
        }
        return Promise.reject(error);
    }
);

// ── System & Settings ─────────────────────────────────────────────────────────

export const fetchSystemInfo = (): Promise<SystemInfo> =>
    api.get<SystemInfo>('/system/info').then((r: AxiosResponse<SystemInfo>) => r.data);

export const fetchSettings = (): Promise<Record<string, string>> =>
    api.get<Record<string, string>>('/settings/').then((r: AxiosResponse<Record<string, string>>) => r.data ?? {});

export const updateSettings = (settings: Record<string, string>): Promise<void> =>
    api.patch('/settings/', settings).then(() => undefined);

// ── Auth ──────────────────────────────────────────────────────────────────────

export const login = (username: string, password: string): Promise<{ token: string }> =>
    api.post<{ token: string }>('/login', {username, password}).then((r: AxiosResponse<{ token: string }>) => r.data);

export const changePassword = (oldPassword: string, newPassword: string): Promise<void> =>
    api.post('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword
    }).then(() => undefined);

// ── Users ─────────────────────────────────────────────────────────────────────

export const fetchUsers = (search?: string): Promise<User[]> =>
    api.get<User[]>('/users', {params: {q: search}}).then((r: AxiosResponse<User[]>) => r.data ?? []);

export const fetchUser = (id: string): Promise<User> =>
    api.get<User>(`/users/${id}`).then((r: AxiosResponse<User>) => r.data);

export const createUser = (data: Partial<User> & { password?: string }): Promise<User> =>
    api.post<User>('/users', data).then((r: AxiosResponse<User>) => r.data);

export const updateUser = (id: string, data: Partial<User>): Promise<User> =>
    api.put<User>(`/users/${id}`, data).then((r: AxiosResponse<User>) => r.data);

export const fetchMe = (): Promise<User> =>
    api.get<User>('/auth/me').then((r: AxiosResponse<User>) => r.data);

export const deleteUser = (id: string): Promise<void> =>
    api.delete(`/users/${id}`).then(() => undefined);

// ── Invoices ──────────────────────────────────────────────────────────────────

export const fetchInvoices = (): Promise<Invoice[]> =>
    api.get<Invoice[]>('/invoices/').then((r: AxiosResponse<Invoice[]>) => r.data ?? []);

export const fetchInvoice = (id: string): Promise<Invoice> =>
    api.get<Invoice>(`/invoices/${id}`).then((r: AxiosResponse<Invoice>) => r.data);

export const importInvoice = (file: File): Promise<Invoice> => {
    const form = new FormData();
    form.append('file', file);
    return api.post<Invoice>('/invoices/import', form, {
        headers: { 'Content-Type': 'multipart/form-data' }
    }).then((r: AxiosResponse<Invoice>) => r.data);
};

export interface InvoiceUpdatePayload {
    vendor?: { id: string; name: string };
    category_id?: string | null;
    amount?: number;
    currency?: string;
    issued_at?: string;
    description?: string;
}

export const updateInvoice = (id: string, data: InvoiceUpdatePayload): Promise<Invoice> => {
    // Map frontend shape → backend updateInvoiceRequest shape
    const body: Record<string, unknown> = {};
    if (data.vendor?.name !== undefined) body.vendor_name = data.vendor.name;
    if ('category_id' in data) body.category_id = data.category_id ?? null;
    if (data.amount !== undefined) body.amount = data.amount;
    if (data.currency !== undefined) body.currency = data.currency;
    if (data.issued_at !== undefined) body.issued_at = data.issued_at ? new Date(data.issued_at).toISOString() : null;
    if (data.description !== undefined) body.description = data.description;
    return api.put<Invoice>(`/invoices/${id}`, body).then((r: AxiosResponse<Invoice>) => r.data);
};

export const deleteInvoice = (id: string): Promise<void> =>
    api.delete(`/invoices/${id}`).then(() => undefined);

export const getInvoicePreviewUrl = async (id: string): Promise<string> => {
    const response = await api.get<Blob>(`/invoices/${id}/download`, {
        responseType: 'blob',
    });
    const blob = new Blob([response.data], { type: 'application/pdf' });
    return window.URL.createObjectURL(blob);
};

export const downloadInvoiceFile = async (id: string, vendorName?: string): Promise<void> => {
    const response = await api.get<Blob>(`/invoices/${id}/download`, {
        responseType: 'blob',
    });

    const disposition = response.headers['content-disposition'] as string | undefined;
    let filename = vendorName ? `invoice-${vendorName.replace(/[^a-z0-9]/gi, '_')}` : `invoice-${id}`;

    if (disposition && disposition.indexOf('filename=') !== -1) {
        const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(disposition);
        if (matches != null && matches[1]) {
            filename = matches[1].replace(/['"]/g, '');
        }
    } else {
        const contentType = response.headers['content-type'] as string | undefined;
        if (contentType === 'application/pdf') filename += '.pdf';
        else filename += '.pdf'; // Fallback
    }

    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();

    link.parentNode?.removeChild(link);
    window.URL.revokeObjectURL(url);
};

// ── Bank Statements ───────────────────────────────────────────────────────────

export const fetchBankStatements = (): Promise<BankStatementSummary[]> =>
    api.get<BankStatementSummary[]>('/bank-statements/').then((r: AxiosResponse<BankStatementSummary[]>) => r.data ?? []);

export const fetchBankStatement = (id: string): Promise<BankStatement> =>
    api.get<BankStatement>(`/bank-statements/${id}`).then((r: AxiosResponse<BankStatement>) => r.data);

export const fetchBankStatementBlob = async (id: string): Promise<Blob> => {
    const response = await api.get<Blob>(`/bank-statements/${id}/download`, {
        responseType: 'blob',
    });
    return response.data;
};

export const importBankStatement = (
    files: File[],
    useAI: boolean = false,
    statementType: string = 'auto'
): Promise<ImportBatchResponse> => {
    const form = new FormData();
    files.forEach(file => form.append('files', file));
    form.append('use_ai', useAI.toString());

    if (statementType !== 'auto') {
        form.append('statement_type', statementType);
    }

    return api
        .post<ImportBatchResponse>('/bank-statements/import', form, {
            headers: {'Content-Type': 'multipart/form-data'},
        })
        .then((r: AxiosResponse<ImportBatchResponse>) => r.data);
};

export const downloadBankStatement = async (id: string): Promise<void> => {
    const response = await api.get<Blob>(`/bank-statements/${id}/download`, {
        responseType: 'blob',
    });

    const disposition = response.headers['content-disposition'] as string | undefined;
    let filename = `statement-${id}`;

    if (disposition && disposition.indexOf('filename=') !== -1) {
        const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(disposition);
        if (matches != null && matches[1]) {
            filename = matches[1].replace(/['"]/g, '');
        }
    } else {
        const contentType = response.headers['content-type'] as string | undefined;
        if (contentType === 'application/pdf') filename += '.pdf';
        else if (contentType === 'text/csv') filename += '.csv';
        else if (contentType === 'application/vnd.ms-excel') filename += '.xls';
        else if (contentType === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') filename += '.xlsx';
    }

    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();

    link.parentNode?.removeChild(link);
    window.URL.revokeObjectURL(url);
};

export const deleteBankStatement = (id: string): Promise<void> =>
    api.delete(`/bank-statements/${id}`).then(() => undefined);

// ── Transactions ──────────────────────────────────────────────────────────────

export const fetchTransactions = (
    statementId?: string,
    hideReconciled: boolean = true,
    categoryId?: string
): Promise<Transaction[]> =>
    api.get<Transaction[]>('/transactions/', {
        params: {
            statement_id: statementId || undefined,
            hide_reconciled: hideReconciled,
            category_id: categoryId || undefined
        }
    }).then((r: AxiosResponse<Transaction[]>) => r.data ?? []);

export const fetchAnalytics = (hideReconciled: boolean = true): Promise<TransactionAnalytics> =>
    api.get<TransactionAnalytics>('/transactions/analytics', {
        params: {hide_reconciled: hideReconciled}
    }).then((r: AxiosResponse<TransactionAnalytics>) => r.data);

export const updateTransactionCategory = (contentHash: string, categoryId: string): Promise<void> =>
    api.patch(`/transactions/${contentHash}/category`, {category_id: categoryId}).then(() => undefined);

export const startAutoCategorize = (): Promise<{ message: string }> =>
    api.post<{ message: string }>('/transactions/auto-categorize/start').then(r => r.data);

export const getAutoCategorizeStatus = (): Promise<JobState> =>
    api.get<JobState>('/transactions/auto-categorize/status').then(r => r.data);

export const cancelAutoCategorize = (): Promise<{ message: string }> =>
    api.post<{ message: string }>('/transactions/auto-categorize/cancel').then(r => r.data);

// ── Categories ────────────────────────────────────────────────────────────────

export const fetchCategories = (): Promise<Category[]> =>
    api.get<Category[]>('/categories/').then((r: AxiosResponse<Category[]>) => r.data ?? []);

export const createCategory = (name: string, color: string): Promise<Category> =>
    api.post<Category>('/categories/', {name, color}).then((r: AxiosResponse<Category>) => r.data);

export const updateCategory = (id: string, name: string, color: string): Promise<Category> =>
    api.put<Category>(`/categories/${id}`, {name, color}).then((r: AxiosResponse<Category>) => r.data);

export const deleteCategory = (id: string): Promise<void> =>
    api.delete(`/categories/${id}`).then(() => undefined);

// ── Health ────────────────────────────────────────────────────────────────────

export const fetchHealth = (): Promise<{ status: string }> =>
    axios.get<{ status: string }>('/health').then((r: AxiosResponse<{ status: string }>) => r.data);

// ── Reconciliations ───────────────────────────────────────────────────────────

export const createReconciliation = (
    settlementTxHash: string,
    targetTxHash: string
): Promise<Reconciliation> =>
    api
        .post<Reconciliation>('/reconciliations/', {
            settlement_tx_hash: settlementTxHash,
            target_tx_hash: targetTxHash,
        })
        .then((r) => r.data);

export const fetchReconciliationSuggestions = async (windowDays: number = 7): Promise<ReconciliationPairSuggestion[]> => {
    const response = await api.get(`/reconciliations/suggestions?window=${windowDays}`);
    return response.data;
};

export const fetchReconciliations = (): Promise<Reconciliation[]> =>
    api.get<Reconciliation[]>('/reconciliations/').then((r) => r.data ?? []);

export const deleteReconciliation = (id: string): Promise<void> =>
    api.delete(`/reconciliations/${id}`).then(() => undefined);

// ── Payslips ──────────────────────────────────────────────────────────────────

export const fetchPayslips = (): Promise<Payslip[]> =>
    api.get<Payslip[]>('/payslips/').then((r: AxiosResponse<Payslip[]>) => r.data ?? []);

export const fetchPayslip = (id: string): Promise<Payslip> =>
    api.get<Payslip>(`/payslips/${id}`).then((r: AxiosResponse<Payslip>) => r.data);

export const getPayslipPreviewUrl = async (id: string): Promise<string> => {
    const response = await api.get<Blob>(`/payslips/${id}/download`, {
        responseType: 'blob',
    });

    const blob = new Blob([response.data], {type: 'application/pdf'});
    return window.URL.createObjectURL(blob);
};

export const importPayslip = async ({file, overrides, useAI}: {
    file: File;
    overrides?: Partial<Payslip>;
    useAI?: boolean
}) => {
    const form = new FormData();
    form.append('file', file);

    if (useAI) {
        form.append('use_ai', 'true');
    }

    if (overrides) {
        if (overrides.period_month_num) form.append('period_month_num', overrides.period_month_num.toString());
        if (overrides.period_year) form.append('period_year', overrides.period_year.toString());
        if (overrides.employee_name) form.append('employee_name', overrides.employee_name);
        if (overrides.gross_pay) form.append('gross_pay', overrides.gross_pay.toString());
        if (overrides.net_pay) form.append('net_pay', overrides.net_pay.toString());
        if (overrides.payout_amount) form.append('payout_amount', overrides.payout_amount.toString());
        if (overrides.custom_deductions) form.append('custom_deductions', overrides.custom_deductions.toString());
        if (overrides.tax_class) form.append('tax_class', overrides.tax_class);
        if (overrides.tax_id) form.append('tax_id', overrides.tax_id);
        if (overrides.bonuses) {
            form.append('bonuses', JSON.stringify(overrides.bonuses));
        }
    }
    return api.post('/payslips/import', form, {headers: {'Content-Type': 'multipart/form-data'}}).then(r => r.data);
};

export const updatePayslip = (id: string, data: Partial<Payslip> | FormData): Promise<Payslip> => {
    const isFormData = data instanceof FormData;
    const config = isFormData ? {headers: {'Content-Type': 'multipart/form-data'}} : undefined;

    return api.put(`/payslips/${id}`, data, config).then(r => r.data);
};

export const deletePayslip = (id: string): Promise<void> =>
    api.delete(`/payslips/${id}`).then(() => undefined);

export const downloadPayslipFile = async (id: string, originalName: string): Promise<void> => {
    const response = await api.get<Blob>(`/payslips/${id}/download`, {
        responseType: 'blob',
    });

    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', originalName || `payslip-${id}.pdf`);
    document.body.appendChild(link);
    link.click();

    link.parentNode?.removeChild(link);
    window.URL.revokeObjectURL(url);
};