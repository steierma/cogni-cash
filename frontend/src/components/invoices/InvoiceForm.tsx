import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Loader2, Save, Plus, Trash2, AlertCircle } from 'lucide-react';
import type { Invoice, InvoiceSplit } from "../../api/types/invoice";
import type { Category } from "../../api/types/category";

interface InvoiceFormProps {
    initialData: Partial<Invoice>;
    categories: Category[];
    onSubmit: (data: any) => void;
    isPending: boolean;
    submitLabel?: string;
}

export function InvoiceForm({ initialData, categories, onSubmit, isPending, submitLabel }: InvoiceFormProps) {
    const { t } = useTranslation();
    
    const [vendorName, setVendorName] = useState(initialData.vendor?.name || '');
    const [issuedAt, setIssuedAt] = useState(initialData.issued_at ? initialData.issued_at.slice(0, 10) : '');
    const [amount, setAmount] = useState(initialData.amount != null ? String(initialData.amount) : '');
    const [currency, setCurrency] = useState(initialData.currency || 'EUR');
    const [categoryId, setCategoryId] = useState(initialData.category_id || '');
    const [description, setDescription] = useState(initialData.description || '');
    const [splits, setSplits] = useState<Partial<InvoiceSplit>[]>(initialData.splits || []);

    const totalAmount = useMemo(() => {
        const parsed = parseFloat(amount.replace(',', '.'));
        return isNaN(parsed) ? 0 : parsed;
    }, [amount]);

    const splitTotal = useMemo(() => {
        return splits.reduce((sum, s) => sum + (Number(s.amount) || 0), 0);
    }, [splits]);

    const remainingAmount = totalAmount - splitTotal;

    const handleAddSplit = () => {
        setSplits([...splits, { category_id: '', amount: remainingAmount > 0 ? remainingAmount : 0, description: '' }]);
    };

    const handleRemoveSplit = (index: number) => {
        setSplits(splits.filter((_, i) => i !== index));
    };

    const handleUpdateSplit = (index: number, data: Partial<InvoiceSplit>) => {
        const newSplits = [...splits];
        newSplits[index] = { ...newSplits[index], ...data };
        setSplits(newSplits);
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const parsedAmount = parseFloat(amount.replace(',', '.'));
        
        // Ensure date is in ISO format for the backend
        let isoDate = initialData.issued_at || new Date().toISOString();
        if (issuedAt) {
            try {
                const date = new Date(issuedAt);
                if (!isNaN(date.getTime())) {
                    isoDate = date.toISOString();
                }
            } catch (e) {
                console.error("Invalid date", e);
            }
        }

        onSubmit({
            vendor: { id: initialData.vendor?.id ?? '', name: vendorName },
            issued_at: isoDate,
            amount: isNaN(parsedAmount) ? 0 : parsedAmount,
            currency: currency,
            category_id: categoryId === '' ? null : categoryId,
            description: description,
            splits: splits
                .filter(s => s.category_id && s.category_id !== '')
                .map(s => ({
                    ...s,
                    amount: Number(s.amount) || 0
                }))
        });
    };

    return (
        <form onSubmit={handleSubmit} className="p-4 space-y-4 flex-1 flex flex-col overflow-hidden">
            <div className="flex-1 space-y-4 overflow-y-auto pr-1">
                {/* Vendor */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.vendor')}</label>
                    <input
                        type="text"
                        value={vendorName}
                        onChange={(e) => setVendorName(e.target.value)}
                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none"
                        placeholder={t('invoices.unknownVendor')}
                    />
                </div>

                {/* Date */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.date')}</label>
                    <input
                        type="date"
                        value={issuedAt}
                        onChange={(e) => setIssuedAt(e.target.value)}
                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none [color-scheme:light] dark:[color-scheme:dark]"
                    />
                </div>

                {/* Amount + Currency */}
                <div className="grid grid-cols-3 gap-3">
                    <div className="col-span-2">
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.amount')}</label>
                        <input
                            type="text"
                            inputMode="decimal"
                            value={amount}
                            onChange={(e) => setAmount(e.target.value)}
                            className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none font-mono"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.currency')}</label>
                        <select
                            value={currency}
                            onChange={(e) => setCurrency(e.target.value)}
                            className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none font-mono"
                        >
                            <option value="EUR">EUR</option>
                            <option value="USD">USD</option>
                            <option value="GBP">GBP</option>
                            <option value="CHF">CHF</option>
                            <option value="PLN">PLN</option>
                        </select>
                    </div>
                </div>

                {/* Main Category (Fallback) */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.category')}</label>
                    <select
                        value={categoryId}
                        onChange={(e) => setCategoryId(e.target.value)}
                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none"
                    >
                        <option value="">{t('common.none')}</option>
                        {categories.filter(c => !c.deleted_at || c.id === categoryId).map((c: Category) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                    </select>
                </div>

                {/* Splits Section */}
                <div className="pt-4 border-t border-gray-100 dark:border-gray-800">
                    <div className="flex items-center justify-between mb-3">
                        <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 uppercase tracking-wider">{t('invoices.splits', 'Splits')}</h4>
                        <button
                            type="button"
                            onClick={handleAddSplit}
                            className="flex items-center gap-1 text-xs font-medium text-indigo-600 hover:text-indigo-700 dark:text-indigo-400"
                        >
                            <Plus size={14} /> {t('invoices.addSplit', 'Add Split')}
                        </button>
                    </div>

                    {splits.length > 0 && (
                        <div className="space-y-3 mb-4">
                            {splits.map((split, index) => (
                                <div key={index} className="flex gap-2 items-start bg-gray-50 dark:bg-gray-900/50 p-3 rounded-xl border border-gray-100 dark:border-gray-800">
                                    <div className="flex-1 space-y-2">
                                        <select
                                            value={split.category_id}
                                            onChange={(e) => handleUpdateSplit(index, { category_id: e.target.value })}
                                            className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-xs rounded-lg p-2 outline-none"
                                        >
                                            <option value="">{t('invoices.chooseCategory', 'Choose Category')}</option>
                                            {categories.map((c: Category) => (
                                                <option key={c.id} value={c.id}>{c.name}</option>
                                            ))}
                                        </select>
                                        <input
                                            type="text"
                                            value={split.description}
                                            onChange={(e) => handleUpdateSplit(index, { description: e.target.value })}
                                            placeholder={t('invoices.splitDescription', 'Description (optional)')}
                                            className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-xs rounded-lg p-2 outline-none"
                                        />
                                    </div>
                                    <div className="w-24">
                                        <input
                                            type="number"
                                            step="0.01"
                                            value={split.amount}
                                            onChange={(e) => handleUpdateSplit(index, { amount: parseFloat(e.target.value) || 0 })}
                                            className="w-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-xs rounded-lg p-2 outline-none font-mono"
                                        />
                                    </div>
                                    <button
                                        type="button"
                                        onClick={() => handleRemoveSplit(index)}
                                        className="p-2 text-gray-400 hover:text-red-500 transition-colors"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </div>
                            ))}
                            
                            {Math.abs(remainingAmount) > 0.001 && (
                                <div className={`flex items-center gap-2 p-2 rounded-lg text-xs font-medium ${remainingAmount > 0 ? 'text-amber-600 bg-amber-50 dark:bg-amber-900/20' : 'text-red-600 bg-red-50 dark:bg-red-900/20'}`}>
                                    <AlertCircle size={14} />
                                    {remainingAmount > 0 
                                        ? t('invoices.remainingAmount', { amount: remainingAmount.toFixed(2), currency }) 
                                        : t('invoices.overAmount', { amount: Math.abs(remainingAmount).toFixed(2), currency })
                                    }
                                </div>
                            )}
                        </div>
                    )}
                </div>

                {/* Description */}
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.description')}</label>
                    <textarea
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        rows={3}
                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none resize-none"
                        placeholder={t('invoices.emptyDescription')}
                    />
                </div>
            </div>

            <div className="pt-4 border-t border-gray-100 dark:border-gray-800 flex justify-end">
                <button
                    type="submit"
                    disabled={isPending}
                    className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-medium transition-all shadow-lg shadow-indigo-200 dark:shadow-none disabled:opacity-50 flex items-center gap-2"
                >
                    {isPending ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />}
                    {submitLabel || t('common.save')}
                </button>
            </div>
        </form>
    );
}
