import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { FileText, Trash2 } from 'lucide-react';
import { deleteInvoice, fetchInvoices } from '../api/client';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import CategoryBadge from '../components/CategoryBadge';
import type { Invoice } from '../api/types';

export default function InvoicesPage() {
    const { t } = useTranslation();
    const qc = useQueryClient();
    const { data: invoices = [], isLoading } = useQuery({
        queryKey: ['invoices'],
        queryFn: fetchInvoices,
    });

    const deleteMutation = useMutation({
        mutationFn: deleteInvoice,
        onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    });

    return (
        <div className="max-w-6xl mx-auto space-y-6 animate-in fade-in duration-300">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('invoices.title')}</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('invoices.subtitle')}</p>
                </div>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm text-left">
                        <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide">
                        <tr>
                            <th className="px-4 py-3 font-medium">{t('invoices.vendor')}</th>
                            <th className="px-4 py-3 font-medium">{t('invoices.category')}</th>
                            <th className="px-4 py-3 font-medium">{t('invoices.date')}</th>
                            <th className="px-4 py-3 font-medium">{t('invoices.description')}</th>
                            <th className="px-4 py-3 font-medium text-right">{t('invoices.amount')}</th>
                            <th className="px-4 py-3 font-medium text-right">{t('invoices.actions')}</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                        {invoices.map((inv: Invoice) => (
                            <tr key={inv.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group">
                                <td className="px-4 py-4 font-medium text-gray-900 dark:text-gray-100 whitespace-nowrap">{inv.vendor?.name || t('invoices.unknownVendor')}</td>
                                <td className="px-4 py-4"><CategoryBadge category={inv.category_id ?? undefined} /></td>
                                <td className="px-4 py-4 text-gray-500 dark:text-gray-400 whitespace-nowrap">{fmtDate(inv.issued_at, 'short')}</td>
                                <td className="px-4 py-4 text-gray-600 dark:text-gray-400 max-w-xs truncate" title={inv.description}>{inv.description || t('invoices.emptyDescription')}</td>
                                <td className="px-4 py-4 text-right font-mono font-medium text-gray-900 dark:text-gray-100">{fmtCurrency(inv.amount, inv.currency)}</td>
                                <td className="px-4 py-4 text-right">
                                    <button
                                        onClick={() => deleteMutation.mutate(inv.id)}
                                        disabled={deleteMutation.isPending}
                                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors disabled:opacity-50"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </td>
                            </tr>
                        ))}
                        {!isLoading && invoices.length === 0 && (
                            <tr>
                                <td colSpan={6} className="px-4 py-16 text-center text-gray-500 dark:text-gray-400">
                                    <FileText size={40} className="mx-auto mb-3 opacity-20 dark:opacity-10" />
                                    {t('invoices.noInvoices')}
                                </td>
                            </tr>
                        )}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}