import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { CheckCircle, Sparkles } from 'lucide-react';
import { categorizeDocument } from '../api/client';
import type { Invoice } from '../api/types';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import CategoryBadge from '../components/CategoryBadge';

export default function CategorizePage() {
    const { t } = useTranslation();
    const [rawText, setRawText] = useState('');
    const qc = useQueryClient();

    const mutation = useMutation({
        mutationFn: categorizeDocument,
        onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (rawText.trim()) mutation.mutate(rawText.trim());
    };

    const result: Invoice | undefined = mutation.data;

    return (
        <div className="max-w-2xl mx-auto animate-in fade-in duration-300">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">{t('categorize.title')}</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
                {t('categorize.subtitle')}
            </p>

            <form onSubmit={handleSubmit} className="space-y-4">
                <textarea
                    value={rawText}
                    onChange={(e) => setRawText(e.target.value)}
                    rows={10}
                    placeholder={t('categorize.placeholder')}
                    className="w-full rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 p-4 text-sm font-mono focus:ring-2 focus:ring-indigo-300 focus:outline-none transition-shadow resize-y"
                />
                <button
                    type="submit"
                    disabled={!rawText.trim() || mutation.isPending}
                    className="flex items-center gap-2 px-6 py-2.5 bg-indigo-600 text-white font-medium text-sm rounded-xl hover:bg-indigo-700 disabled:opacity-50 transition-colors shadow-sm"
                >
                    <Sparkles size={16} />
                    {mutation.isPending ? t('categorize.processing') : t('categorize.submit')}
                </button>
            </form>

            {mutation.isError && (
                <div className="mt-4 p-4 rounded-xl bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 text-red-700 dark:text-red-400 text-sm">
                    {(mutation.error as Error)?.message ?? t('categorize.errorDefault')}
                </div>
            )}

            {result && (
                <div className="mt-6 p-5 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 shadow-sm space-y-3">
                    <div className="flex items-center gap-2 text-green-600 dark:text-green-400 font-medium text-sm mb-1">
                        <CheckCircle size={16} />
                        {t('categorize.savedSuccess')}
                    </div>
                    <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                        <Detail label={t('categorize.vendor')} value={result.vendor?.name} emptyValue={t('categorize.emptyValue')} />
                        <div>
                            <span className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide">{t('categorize.category')}</span>
                            <div className="mt-0.5"><CategoryBadge category={result.category_id ?? undefined} /></div>
                        </div>
                        <Detail label={t('categorize.amount')} value={fmtCurrency(result.amount, result.currency || 'EUR')} emptyValue={t('categorize.emptyValue')} />
                        <Detail label={t('categorize.date')} value={fmtDate(result.issued_at, 'short')} emptyValue={t('categorize.emptyValue')} />
                        <Detail label={t('categorize.description')} value={result.description} className="col-span-2" emptyValue={t('categorize.emptyValue')} />
                    </div>
                </div>
            )}
        </div>
    );
}

function Detail({ label, value, emptyValue, className = '' }: { label: string; value?: string; emptyValue: string; className?: string }) {
    return (
        <div className={className}>
            <span className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide">{label}</span>
            <p className="font-medium text-gray-900 dark:text-gray-100 mt-0.5">{value || emptyValue}</p>
        </div>
    );
}