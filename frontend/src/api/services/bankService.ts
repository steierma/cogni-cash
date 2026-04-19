import { api } from '../client';
import type { BankStatementSummary, BankStatement, BankInstitution, BankConnection } from "../types/bank";
import type { ImportBatchResponse } from "../types/common";
export const bankService = {
    // Statements
    fetchStatements: (): Promise<BankStatementSummary[]> =>
        api.get<BankStatementSummary[]>('bank-statements/').then(r => r.data ?? []),

    fetchStatement: (id: string): Promise<BankStatement> =>
        api.get<BankStatement>(`bank-statements/${id}/`).then(r => r.data),

    fetchStatementBlob: (id: string): Promise<Blob> =>
        api.get<Blob>(`bank-statements/${id}/download/`, { responseType: 'blob' }).then(r => r.data),

    importStatement: (files: File[], useAI: boolean = false, statementType: string = 'auto'): Promise<ImportBatchResponse> => {
        const form = new FormData();
        files.forEach(f => form.append('files', f));
        form.append('use_ai', useAI.toString());
        if (statementType !== 'auto') form.append('statement_type', statementType);
        return api.post<ImportBatchResponse>('bank-statements/import/', form, { headers: { 'Content-Type': 'multipart/form-data' } }).then(r => r.data);
    },

    downloadStatement: async (id: string): Promise<void> => {
        const response = await api.get<Blob>(`bank-statements/${id}/download/`, { responseType: 'blob' });
        const disposition = response.headers['content-disposition'] as string | undefined;
        let filename = `statement-${id}`;
        if (disposition && disposition.indexOf('filename=') !== -1) {
            const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(disposition);
            if (matches != null && matches[1]) filename = matches[1].replace(/['"]/g, '');
        } else {
            const ct = response.headers['content-type'] as string | undefined;
            if (ct === 'application/pdf') filename += '.pdf';
            else if (ct === 'text/csv') filename += '.csv';
            else if (ct === 'application/vnd.ms-excel') filename += '.xls';
            else if (ct === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') filename += '.xlsx';
        }
        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a');
        link.href = url; link.setAttribute('download', filename);
        document.body.appendChild(link); link.click();
        link.parentNode?.removeChild(link); window.URL.revokeObjectURL(url);
    },

    deleteStatement: (id: string): Promise<void> =>
        api.delete(`bank-statements/${id}/`).then(() => undefined),

    // Connections & Accounts
    fetchInstitutions: (country: string = 'DE', sandbox: boolean = false): Promise<BankInstitution[]> =>
        api.get<BankInstitution[]>('bank/institutions/', { params: { country, sandbox } }).then(r => r.data ?? []),

    createConnection: (institutionId: string, institutionName: string, country: string, redirectUrl: string, sandbox: boolean = false): Promise<BankConnection> =>
        api.post<BankConnection>('bank/connections/', { institution_id: institutionId, institution_name: institutionName, country, redirect_url: redirectUrl, sandbox }).then(r => r.data),

    finishConnection: (requisitionId: string, code?: string): Promise<void> =>
        api.post('bank/connections/finish/', { requisition_id: requisitionId, code }).then(() => undefined),

    fetchConnections: (): Promise<BankConnection[]> =>
        api.get<BankConnection[]>('bank/connections/').then(r => r.data ?? []),

    deleteConnection: (id: string): Promise<void> =>
        api.delete(`bank/connections/${id}/`).then(() => undefined),

    updateAccountType: (accountId: string, accountType: string): Promise<void> =>
        api.put(`bank/accounts/${accountId}/type/`, { account_type: accountType }).then(() => undefined),

    syncAccounts: (): Promise<{ message: string }> =>
        api.post<{ message: string }>('bank/sync/').then(r => r.data)
};