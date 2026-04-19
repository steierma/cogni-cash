import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Loader2, Save } from 'lucide-react';
import type { Invoice } from "../../api/types/invoice";
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

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        const parsedAmount = parseFloat(amount.replace(',', '.'));
        onSubmit({
            vendor: { id: initialData.vendor?.id ?? '', name: vendorName },
            issued_at: issuedAt,
            amount: isNaN(parsedAmount) ? initialData.amount : parsedAmount,
            currency: currency,
            category_id: categoryId || null,
            description: description
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
                        <input
                            type="text"
                            maxLength={3}
                            value={currency}
                            onChange={(e) => setCurrency(e.target.value.toUpperCase())}
                            className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none font-mono uppercase"
                        />
                    </div>
                </div>

                {/* Category */}
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
