import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Plus, Edit2, Trash2, CheckCircle2, Clock, AlertCircle, Repeat } from 'lucide-react';
import { forecastingService } from '../../api/services/forecastingService';
import type { PlannedTransaction, CreatePlannedTransactionRequest, UpdatePlannedTransactionRequest } from "../../api/types/transaction";
import { fmtCurrency, fmtDate } from '../../utils/formatters';
import CategoryBadge from '../CategoryBadge';
import PlannedTransactionModal from './PlannedTransactionModal';

export default function PlannedTransactionsList() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingTx, setEditingTx] = useState<PlannedTransaction | null>(null);

    const { data: transactions = [], isLoading } = useQuery({
        queryKey: ['planned-transactions'],
        queryFn: forecastingService.fetchPlannedTransactions,
    });

    const createMutation = useMutation({
        mutationFn: (data: CreatePlannedTransactionRequest) => forecastingService.createPlannedTransaction(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['planned-transactions'] });
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
        },
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: UpdatePlannedTransactionRequest }) => forecastingService.updatePlannedTransaction(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['planned-transactions'] });
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
        },
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => forecastingService.deletePlannedTransaction(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['planned-transactions'] });
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
        },
    });

    const handleSave = async (data: CreatePlannedTransactionRequest | UpdatePlannedTransactionRequest) => {
        if (editingTx) {
            await updateMutation.mutateAsync({ id: editingTx.id, data: data as UpdatePlannedTransactionRequest });
        } else {
            await createMutation.mutateAsync(data as CreatePlannedTransactionRequest);
        }
    };

    const handleEdit = (tx: PlannedTransaction) => {
        setEditingTx(tx);
        setIsModalOpen(true);
    };

    const handleDelete = (id: string) => {
        if (window.confirm(t('common.confirmDelete'))) {
            deleteMutation.mutate(id);
        }
    };

    const StatusIcon = ({ status }: { status: string }) => {
        switch (status) {
            case 'matched': return <CheckCircle2 size={16} className="text-emerald-500" />;
            case 'expired': return <AlertCircle size={16} className="text-rose-500" />;
            default: return <Clock size={16} className="text-amber-500" />;
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                    {t('forecasting.plannedTransactions')}
                </h2>
                <button
                    onClick={() => { setEditingTx(null); setIsModalOpen(true); }}
                    className="px-3 py-1.5 flex items-center gap-1.5 text-sm font-medium text-white bg-indigo-600 rounded-lg hover:bg-indigo-700 transition-colors"
                >
                    <Plus size={16} />
                    {t('forecasting.addPlanned')}
                </button>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                {isLoading ? (
                    <div className="p-8 flex flex-col gap-4">
                        {[1, 2].map(i => (
                            <div key={i} className="h-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl animate-pulse" />
                        ))}
                    </div>
                ) : transactions.length === 0 ? (
                    <div className="p-8 text-center text-gray-500 dark:text-gray-400 text-sm">
                        {t('forecasting.noPlannedTransactions')}
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50">
                            <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 font-bold tracking-wider">
                                <tr>
                                    <th className="px-6 py-4 text-left">{t('transactions.form.date')}</th>
                                    <th className="px-6 py-4 text-left">{t('transactions.form.description')}</th>
                                    <th className="px-6 py-4 text-left">{t('transactions.form.category')}</th>
                                    <th className="px-6 py-4 text-center">{t('common.status')}</th>
                                    <th className="px-6 py-4 text-right">{t('transactions.form.amount')}</th>
                                    <th className="px-6 py-4 text-right">{t('common.actions')}</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                                {transactions.map(tx => (
                                    <tr key={tx.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-gray-100">
                                            {fmtDate(tx.date)}
                                        </td>
                                        <td className="px-6 py-4 text-sm font-medium text-gray-900 dark:text-gray-100">
                                            <div className="flex flex-col">
                                                <span>{tx.description}</span>
                                                {tx.interval_months > 0 && (
                                                    <span className="text-[10px] text-indigo-500 dark:text-indigo-400 flex items-center gap-1 uppercase tracking-wider mt-0.5">
                                                        <Repeat size={10} />
                                                        {tx.interval_months === 1 ? t('forecasting.interval.monthly') :
                                                         tx.interval_months === 3 ? t('forecasting.interval.quarterly') :
                                                         tx.interval_months === 6 ? t('forecasting.interval.halfYearly') :
                                                         tx.interval_months === 12 ? t('forecasting.interval.yearly') :
                                                         `${tx.interval_months} ${t('common.months')}`}
                                                        {tx.end_date && ` • ${t('forecasting.until')} ${fmtDate(tx.end_date)}`}
                                                    </span>
                                                )}
                                                {tx.is_superseded && (
                                                    <span className="text-[10px] text-amber-500 flex items-center gap-1 uppercase tracking-wider mt-0.5">
                                                        <AlertCircle size={10} />
                                                        {t('forecasting.superseded', 'Superseded by auto-forecast')}
                                                    </span>
                                                )}
                                            </div>
                                        </td>
                                        <td className="px-6 py-4">
                                            <CategoryBadge category={tx.category_id} />
                                        </td>
                                        <td className="px-6 py-4 text-center">
                                            <div className="flex items-center justify-center gap-1.5" title={t(`forecasting.status_${tx.status}`)}>
                                                <StatusIcon status={tx.status} />
                                                <span className="text-xs font-medium capitalize text-gray-600 dark:text-gray-400">
                                                    {t(`forecasting.status_${tx.status}`)}
                                                </span>
                                            </div>
                                        </td>
                                        <td className={`px-6 py-4 text-right font-mono font-bold ${tx.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-gray-100'}`}>
                                            {fmtCurrency(tx.amount, 'EUR')}
                                        </td>
                                        <td className="px-6 py-4 text-right space-x-2">
                                            <button onClick={() => handleEdit(tx)} className="text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
                                                <Edit2 size={16} />
                                            </button>
                                            <button onClick={() => handleDelete(tx.id)} className="text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors">
                                                <Trash2 size={16} />
                                            </button>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            <PlannedTransactionModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSave={handleSave}
                initialData={editingTx}
            />
        </div>
    );
}