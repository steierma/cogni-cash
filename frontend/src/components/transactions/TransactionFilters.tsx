import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { Filter, Search, Loader2 } from 'lucide-react';
import type { BankStatementSummary, Category } from '../../api/types';
import { fmtDate } from '../../utils/formatters';

export interface FilterState {
    search: string;
    type: 'all' | 'credit' | 'debit';
    statement: string;
    category: string;
    from: string | null;
    to: string | null;
    amountMin: string;
    amountMax: string;
}

interface TransactionFiltersProps {
    applied: FilterState;
    onApply: (filters: FilterState) => void;
    hasAppliedOnce: boolean;
    isLoading: boolean;
    statements: BankStatementSummary[];
    categories: Category[];
    minDate: string;
    maxDate: string;
}

export default function TransactionFilters({
                                               applied,
                                               onApply,
                                               hasAppliedOnce,
                                               isLoading,
                                               statements,
                                               categories,
                                               minDate,
                                               maxDate
                                           }: TransactionFiltersProps) {
    const { t } = useTranslation();
    const [localDraft, setLocalDraft] = React.useState<FilterState>(applied);
    const [prevApplied, setPrevApplied] = React.useState<FilterState>(applied);

    if (applied !== prevApplied) {
        setPrevApplied(applied);
        setLocalDraft(applied);
    }

    const isDraftDirty = JSON.stringify(localDraft) !== JSON.stringify(applied);

    const handleSubmit = (e?: React.FormEvent<HTMLFormElement>) => {
        e?.preventDefault();
        onApply(localDraft);
    };

    const isStatementsLoading = isLoading && statements.length === 0;
    const isCategoriesLoading = isLoading && categories.length === 0;

    return (
        <form onSubmit={handleSubmit}
              className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-4 space-y-3">
            <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2 text-sm font-medium text-gray-500 dark:text-gray-400">
                    <Filter size={14} />
                    {t('transactions.filters.title')}
                </div>
                {isDraftDirty && (
                    <span className="text-[10px] font-bold text-indigo-500 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded-full uppercase tracking-wider animate-pulse">
                        {t('transactions.filters.unapplied')}
                    </span>
                )}
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
                {/* Search bar now spans 3 columns to fill the gap left by the removed dropdown */}
                <div className="relative sm:col-span-2 lg:col-span-3">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500" />
                    <input
                        value={localDraft.search}
                        onChange={(e) => setLocalDraft({ ...localDraft, search: e.target.value })}
                        placeholder={t('transactions.filters.searchPlaceholder')}
                        className="w-full pl-8 pr-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                    />
                </div>

                <div className="relative">
                    <select
                        value={localDraft.statement || ""}
                        onChange={(e) => setLocalDraft({ ...localDraft, statement: e.target.value })}
                        disabled={isStatementsLoading}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 disabled:opacity-60"
                    >
                        {isStatementsLoading ? (
                            <option value={localDraft.statement || ""}>{t('common.loading', 'Loading...')}</option>
                        ) : (
                            <option value="">{t('transactions.filters.allStatements')}</option>
                        )}

                        {localDraft.statement && isStatementsLoading && (
                            <option value={localDraft.statement}>{t('common.loading', 'Loading...')}</option>
                        )}

                        {statements.map((s) => {
                            const typeMap: Record<string, string> = {
                                giro: t('bankStatements.filters.typeGiro'),
                                credit_card: t('bankStatements.filters.typeCc'),
                                extra_account: t('bankStatements.filters.typeExtra')
                            };
                            const typeLabel = typeMap[s.statement_type] || t('transactions.filters.unknown');

                            return (
                                <option key={s.id} value={s.id}>
                                    [{typeLabel}] {s.statement_no > 0 ? `No. ${s.statement_no} · ` : ''}
                                    {s.period_label} ·
                                    ···{s.iban.slice(-4)} ({fmtDate(s.start_date, 'short')} - {fmtDate(s.end_date, 'short')})
                                </option>
                            );
                        })}
                    </select>
                </div>

                <div className="relative">
                    <select
                        value={localDraft.category || "all"}
                        onChange={(e) => setLocalDraft({ ...localDraft, category: e.target.value })}
                        disabled={isCategoriesLoading}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 disabled:opacity-60"
                    >
                        {isCategoriesLoading ? (
                            <option value={localDraft.category || "all"}>{t('common.loading', 'Loading...')}</option>
                        ) : (
                            <>
                                <option value="all">{t('transactions.filters.allCategories')}</option>
                                <option value="uncategorized">{t('transactions.filters.uncategorizedOnly')}</option>
                            </>
                        )}

                        {categories.filter(c => !c.deleted_at || c.id === localDraft.category).map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                    </select>
                </div>

                <div className="flex sm:col-span-2 lg:col-span-5 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden text-sm">
                    {(['all', 'credit', 'debit'] as const).map((typeVal) => {
                        const labelMap: Record<string, string> = {
                            all: t('transactions.filters.typeAll'),
                            credit: t('transactions.filters.typeCredit'),
                            debit: t('transactions.filters.typeDebit')
                        };
                        return (
                            <button
                                key={typeVal}
                                type="button"
                                onClick={() => setLocalDraft({ ...localDraft, type: typeVal })}
                                className={`flex-1 py-2 capitalize transition-colors ${localDraft.type === typeVal
                                    ? typeVal === 'credit'
                                        ? 'bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 font-medium'
                                        : typeVal === 'debit'
                                            ? 'bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 font-medium'
                                            : 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 font-medium'
                                    : 'text-gray-500 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 bg-white dark:bg-gray-900'
                                }`}
                            >
                                {labelMap[typeVal]}
                            </button>
                        );
                    })}
                </div>
            </div>

            <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
                <div>
                    <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.from')}</label>
                    <input
                        type="date"
                        value={localDraft.from ?? minDate}
                        onChange={(e) => setLocalDraft({ ...localDraft, from: e.target.value })}
                        min={minDate}
                        max={maxDate}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 [color-scheme:light] dark:[color-scheme:dark]"
                    />
                </div>
                <div>
                    <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.to')}</label>
                    <input
                        type="date"
                        value={localDraft.to ?? maxDate}
                        onChange={(e) => setLocalDraft({ ...localDraft, to: e.target.value })}
                        min={minDate}
                        max={maxDate}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 [color-scheme:light] dark:[color-scheme:dark]"
                    />
                </div>
                <div>
                    <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.minAmount')}</label>
                    <input
                        type="number"
                        step="0.01"
                        placeholder="-∞"
                        value={localDraft.amountMin}
                        onChange={(e) => setLocalDraft({ ...localDraft, amountMin: e.target.value })}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                    />
                </div>
                <div>
                    <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.maxAmount')}</label>
                    <input
                        type="number"
                        step="0.01"
                        placeholder="∞"
                        value={localDraft.amountMax}
                        onChange={(e) => setLocalDraft({ ...localDraft, amountMax: e.target.value })}
                        className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                    />
                </div>
                <div className="flex items-end">
                    <button
                        type="submit"
                        className={`w-full py-2 px-4 rounded-lg font-medium text-sm flex items-center justify-center gap-2 transition-all ${isDraftDirty
                            ? 'bg-indigo-600 dark:bg-indigo-500 text-white shadow-md hover:bg-indigo-700 dark:hover:bg-indigo-600'
                            : hasAppliedOnce
                                ? 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500'
                                : 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/50'
                        }`}
                    >
                        {isLoading && isDraftDirty ? <Loader2 size={14} className="animate-spin" /> : <Search size={14} />}
                        {t('transactions.filters.search')}
                    </button>
                </div>
            </div>
        </form>
    );
}