import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { X, Save } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { categoryService } from '../../api/services/categoryService';
import type { Category } from "../../api/types/category";
import type { PlannedTransaction, CreatePlannedTransactionRequest, UpdatePlannedTransactionRequest } from "../../api/types/transaction";

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onSave: (data: CreatePlannedTransactionRequest | UpdatePlannedTransactionRequest) => Promise<void>;
    initialData?: PlannedTransaction | null;
}

export default function PlannedTransactionModal({ isOpen, onClose, onSave, initialData }: Props) {
    const { t } = useTranslation();
    const [amount, setAmount] = useState<string>('');
    const [currency, setCurrency] = useState<string>('EUR');
    const [date, setDate] = useState<string>('');
    const [description, setDescription] = useState<string>('');
    const [categoryId, setCategoryId] = useState<string>('');
    const [intervalMonths, setIntervalMonths] = useState<number>(0);
    const [endDate, setEndDate] = useState<string>('');
    const [isSaving, setIsSaving] = useState(false);
    
    const { data: categories = [] } = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: categoryService.fetchCategories,
        staleTime: 5 * 60 * 1000,
    });

    useEffect(() => {
        if (isOpen) {
            if (initialData) {
                setAmount(initialData.amount.toString());
                setCurrency(initialData.currency || 'EUR');
                setDate(initialData.date.split('T')[0]);
                setDescription(initialData.description);
                setCategoryId(initialData.category_id || '');
                setIntervalMonths(initialData.interval_months || 0);
                setEndDate(initialData.end_date ? initialData.end_date.split('T')[0] : '');
            } else {
                setAmount('');
                setCurrency('EUR');
                setDate(new Date().toISOString().split('T')[0]);
                setDescription('');
                setCategoryId('');
                setIntervalMonths(0);
                setEndDate('');
            }
        }
    }, [isOpen, initialData]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSaving(true);
        try {
            const payload: any = {
                amount: parseFloat(amount),
                currency,
                date: new Date(date).toISOString(),
                description,
                category_id: categoryId || null,
                interval_months: intervalMonths,
                end_date: intervalMonths > 0 && endDate ? new Date(endDate).toISOString() : undefined,
            };
            if (initialData) {
                payload.status = initialData.status;
            }
            await onSave(payload);
            onClose();
        } finally {
            setIsSaving(false);
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div className="fixed inset-0 bg-black/30 backdrop-blur-sm" aria-hidden="true" onClick={onClose} />
            <div className="relative z-10 w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-xl overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">
                        {initialData ? t('forecasting.editPlanned') : t('forecasting.addPlanned')}
                    </h3>
                    <button type="button" onClick={onClose} className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300">
                        <X size={20} />
                    </button>
                </div>

                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {t('transactions.form.description')} <span className="text-red-500">*</span>
                        </label>
                        <input
                            type="text"
                            required
                            value={description}
                            onChange={e => setDescription(e.target.value)}
                            className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white"
                            placeholder={t('transactions.form.descriptionPlaceholder')}
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                {t('transactions.form.amount')} <span className="text-red-500">*</span>
                            </label>
                            <div className="flex gap-2">
                                <input
                                    type="number"
                                    required
                                    step="0.01"
                                    value={amount}
                                    onChange={e => setAmount(e.target.value)}
                                    className="flex-1 min-w-0 px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white"
                                />
                                <select
                                    value={currency}
                                    onChange={e => setCurrency(e.target.value)}
                                    className="w-24 px-2 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white font-mono text-sm"
                                >
                                    <option value="EUR">EUR</option>
                                    <option value="USD">USD</option>
                                    <option value="GBP">GBP</option>
                                    <option value="CHF">CHF</option>
                                    <option value="PLN">PLN</option>
                                </select>
                            </div>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                {t('transactions.form.date')} <span className="text-red-500">*</span>
                            </label>
                            <input
                                type="date"
                                required
                                value={date}
                                onChange={e => setDate(e.target.value)}
                                className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white"
                            />
                        </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                                {t('forecasting.frequency', 'Frequency')}
                            </label>
                            <select
                                value={intervalMonths}
                                onChange={e => setIntervalMonths(parseInt(e.target.value))}
                                className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white"
                            >
                                <option value={0}>{t('forecasting.interval.once', 'Once')}</option>
                                <option value={1}>{t('forecasting.interval.monthly', 'Monthly')}</option>
                                <option value={3}>{t('forecasting.interval.quarterly', 'Quarterly')}</option>
                                <option value={6}>{t('forecasting.interval.halfYearly', 'Half-Yearly')}</option>
                                <option value={12}>{t('forecasting.interval.yearly', 'Yearly')}</option>
                            </select>
                        </div>
                        <div>
                            <label className={`block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 ${intervalMonths === 0 ? 'opacity-50' : ''}`}>
                                {t('forecasting.endDate', 'End Date')}
                            </label>
                            <input
                                type="date"
                                disabled={intervalMonths === 0}
                                value={endDate}
                                onChange={e => setEndDate(e.target.value)}
                                className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white disabled:opacity-50 disabled:cursor-not-allowed"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                            {t('transactions.form.category')}
                        </label>
                        <select
                            value={categoryId}
                            onChange={e => setCategoryId(e.target.value)}
                            className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 text-gray-900 dark:text-white"
                        >
                            <option value="">-- {t('transactions.form.noCategory')} --</option>
                            {categories.filter(c => !c.deleted_at || c.id === categoryId).map(c => (
                                <option key={c.id} value={c.id}>{c.name}</option>
                            ))}
                        </select>
                    </div>

                    <div className="pt-4 flex justify-end gap-3">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700"
                        >
                            {t('common.cancel')}
                        </button>
                        <button
                            type="submit"
                            disabled={isSaving}
                            className="px-4 py-2 flex items-center gap-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-lg hover:bg-indigo-700 disabled:opacity-50"
                        >
                            <Save size={16} />
                            {isSaving ? t('common.saving') : t('common.save')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}