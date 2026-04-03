import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { Save, Plus, X, Loader2 } from 'lucide-react';
import type { Payslip } from '../../api/types';

interface PayslipFormProps {
    initialData: Partial<Payslip>;
    onSubmit: (data: Partial<Payslip>) => void;
    isPending?: boolean;
    submitLabel?: string;
    showSubmitIcon?: boolean;
}

export function PayslipForm({ initialData, onSubmit, isPending, submitLabel, showSubmitIcon = true }: PayslipFormProps) {
    const { t } = useTranslation();
    const [editBonuses, setEditBonuses] = useState<{ description: string; amount: string }[]>([]);

    useEffect(() => {
        if (initialData.bonuses) {
            setEditBonuses(initialData.bonuses.map(sz => ({ description: sz.description, amount: sz.amount.toString() })));
        } else {
            setEditBonuses([]);
        }
    }, [initialData.bonuses]);

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const data: Partial<Payslip> = {
            period_month_num: Number(formData.get('period_month_num')),
            period_year: Number(formData.get('period_year')),
            employer_name: formData.get('employer_name') as string,
            tax_class: formData.get('tax_class') as string,
            tax_id: formData.get('tax_id') as string,
            gross_pay: Number(formData.get('gross_pay')),
            net_pay: Number(formData.get('net_pay')),
            payout_amount: Number(formData.get('payout_amount')),
            custom_deductions: Number(formData.get('custom_deductions')),
        };
        data.bonuses = editBonuses
            .filter(b => b.description.trim() && b.amount.trim())
            .map(b => ({ description: b.description, amount: Number(b.amount) }));
        
        onSubmit(data);
    };

    return (
        <form onSubmit={handleSubmit} className="flex flex-col h-full">
            <div className="flex-1 overflow-y-auto p-4 space-y-6">
                <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2">
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.employer')}</label>
                            <input name="employer_name" defaultValue={initialData.employer_name} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" required />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.month')}</label>
                            <input name="period_month_num" type="number" min="1" max="12" defaultValue={initialData.period_month_num} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" required />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.year')}</label>
                            <input name="period_year" type="number" defaultValue={initialData.period_year} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" required />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.taxClass')}</label>
                            <input name="tax_class" defaultValue={initialData.tax_class} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.taxId')}</label>
                            <input name="tax_id" defaultValue={initialData.tax_id} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" />
                        </div>
                    </div>

                    <hr className="border-gray-100 dark:border-gray-800" />

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.gross')}</label>
                            <input name="gross_pay" type="number" step="0.01" defaultValue={initialData.gross_pay} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none font-mono" required />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.net')}</label>
                            <input name="net_pay" type="number" step="0.01" defaultValue={initialData.net_pay} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none font-mono" required />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.leasing')}</label>
                            <input name="custom_deductions" type="number" step="0.01" defaultValue={initialData.custom_deductions} className="w-full px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none font-mono" />
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1 text-indigo-600 dark:text-indigo-400">{t('payslips.modals.payout')}</label>
                            <input name="payout_amount" type="number" step="0.01" defaultValue={initialData.payout_amount} className="w-full px-3 py-2 text-sm border-2 border-indigo-100 dark:border-indigo-900/50 rounded-lg dark:bg-gray-800 focus:ring-2 focus:ring-indigo-500 outline-none font-mono font-bold text-indigo-600 dark:text-indigo-400" required />
                        </div>
                    </div>

                    <hr className="border-gray-100 dark:border-gray-800" />

                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('payslips.modals.bonuses')}</label>
                            <button type="button" onClick={() => setEditBonuses(prev => [...prev, { description: '', amount: '' }])} className="text-xs text-indigo-600 hover:text-indigo-700 dark:text-indigo-400 flex items-center gap-1 font-medium bg-indigo-50 dark:bg-indigo-900/20 px-2 py-1 rounded-md transition-colors">
                                <Plus size={14} /> {t('common.add')}
                            </button>
                        </div>
                        <div className="space-y-3">
                            {editBonuses.map((b, idx) => (
                                <div key={idx} className="flex gap-2 items-start animate-in slide-in-from-right-2 duration-200">
                                    <input placeholder={t('payslips.modals.bonusDesc')} value={b.description} onChange={e => { const next = [...editBonuses]; next[idx].description = e.target.value; setEditBonuses(next); }} className="flex-1 px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none" />
                                    <input type="number" step="0.01" placeholder="0.00" value={b.amount} onChange={e => { const next = [...editBonuses]; next[idx].amount = e.target.value; setEditBonuses(next); }} className="w-24 px-3 py-2 text-sm border rounded-lg dark:bg-gray-800 dark:border-gray-700 focus:ring-2 focus:ring-indigo-500 outline-none font-mono" />
                                    <button type="button" onClick={() => setEditBonuses(prev => prev.filter((_, i) => i !== idx))} className="p-2 text-gray-400 hover:text-red-500 transition-colors"><X size={16} /></button>
                                </div>
                            ))}
                            {editBonuses.length === 0 && (
                                <p className="text-xs text-center text-gray-400 py-4 bg-gray-50 dark:bg-gray-800/30 rounded-xl border border-dashed border-gray-200 dark:border-gray-800">
                                    {t('payslips.chart.noBonuses')}
                                </p>
                            )}
                        </div>
                    </div>
                </div>
            </div>

            <div className="p-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 flex justify-end">
                <button type="submit" disabled={isPending} className="w-full sm:w-auto px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-medium shadow-lg shadow-indigo-200 dark:shadow-none transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed transform active:scale-[0.98]">
                    {isPending ? <Loader2 className="animate-spin" size={18} /> : (showSubmitIcon && <Save size={18} />)}
                    {submitLabel || t('common.saveChanges')}
                </button>
            </div>
        </form>
    );
}
