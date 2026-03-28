import { useTranslation } from 'react-i18next';
import { CheckSquare, ChevronDown, ChevronUp, Copy, Square, TrendingDown, TrendingUp, Unlink, MapPin } from 'lucide-react';
import type { Category, Transaction } from '../../api/types';
import { fmtCurrency, fmtDate } from '../../utils/formatters';

export type TxColKey = 'date' | 'description' | 'location' | 'reference' | 'category' | 'amount';
export type SortKey = 'booking_date' | 'description' | 'location' | 'amount';
export type SortDir = 'asc' | 'desc';

interface TransactionTableProps {
    transactions: Transaction[];
    categories: Category[];
    selectedHashes: Set<string>;
    onToggleSelect: (hash: string) => void;
    onToggleSelectAll: () => void;
    sortKey: SortKey;
    sortDir: SortDir;
    onSort: (key: SortKey) => void;
    onCategoryChange: (hash: string, categoryId: string) => void;
    onMarkReviewed?: (hash: string) => void;
    visibleCols: Record<TxColKey, boolean>;
}

export default function TransactionTable({
                                             transactions,
                                             categories,
                                             selectedHashes,
                                             onToggleSelect,
                                             onToggleSelectAll,
                                             sortKey,
                                             sortDir,
                                             onSort,
                                             onCategoryChange,
                                             onMarkReviewed,
                                             visibleCols
                                         }: TransactionTableProps) {
    const { t } = useTranslation();

    const SortIcon = ({ k }: { k: SortKey }) =>
        sortKey === k ? (sortDir === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />) : null;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden relative">
            <div className="overflow-x-auto [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700">
                <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800 text-sm">
                    <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide">
                    <tr>
                        <th className="px-4 py-3 text-left w-10">
                            <button
                                type="button"
                                onClick={onToggleSelectAll}
                                className={`transition-colors ${selectedHashes.size === transactions.length && transactions.length > 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                            >
                                {selectedHashes.size === transactions.length && transactions.length > 0 ? <CheckSquare size={16} /> : <Square size={16} />}
                            </button>
                        </th>

                        {visibleCols.date && (
                            <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap" onClick={() => onSort('booking_date')}>
                                <span className="inline-flex items-center gap-1">{t('transactions.table.date', 'Date')} <SortIcon k="booking_date" /></span>
                            </th>
                        )}

                        {visibleCols.description && (
                            <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none" onClick={() => onSort('description')}>
                                <span className="inline-flex items-center gap-1">{t('transactions.table.description', 'Description')} <SortIcon k="description" /></span>
                            </th>
                        )}

                        {visibleCols.location && (
                            <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none" onClick={() => onSort('location')}>
                                <span className="inline-flex items-center gap-1">{t('transactions.table.location', 'Location')} <SortIcon k="location" /></span>
                            </th>
                        )}

                        {visibleCols.reference && (
                            <th className="px-4 py-3 text-left">{t('transactions.table.reference', 'Reference')}</th>
                        )}

                        {visibleCols.category && (
                            <th className="px-4 py-3 text-left">{t('transactions.table.category', 'Category')}</th>
                        )}

                        {visibleCols.amount && (
                            <th className="px-4 py-3 text-right cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none" onClick={() => onSort('amount')}>
                                <span className="inline-flex items-center gap-1 justify-end">{t('transactions.table.amount', 'Amount')} <SortIcon k="amount" /></span>
                            </th>
                        )}
                        <th className="px-4 py-3 text-right w-20">{t('common.actions', 'Actions')}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                    {transactions.map((tx) => {
                        const isSelected = selectedHashes.has(tx.content_hash);
                        const currentCat = categories.find((c) => c.id === tx.category_id);

                        return (
                            <tr key={tx.content_hash} className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors ${isSelected ? 'bg-indigo-50/30 dark:bg-indigo-900/20' : ''} ${tx.is_reconciled ? 'opacity-60' : ''}`}>
                                <td className="px-4 py-3">
                                    <button
                                        type="button"
                                        onClick={() => onToggleSelect(tx.content_hash)}
                                        className={`transition-colors ${isSelected ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                                    >
                                        {isSelected ? <CheckSquare size={16} /> : <Square size={16} />}
                                    </button>
                                </td>

                                {visibleCols.date && (
                                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                                        {fmtDate(tx.booking_date)}
                                    </td>
                                )}

                                {visibleCols.description && (
                                    <td className="px-4 py-3 text-gray-800 dark:text-gray-200 max-w-xs truncate" title={tx.description}>
                                            <span className="flex items-center gap-1.5">
                                                {!tx.reviewed && (
                                                    <div className="w-2 h-2 rounded-full bg-indigo-500 shrink-0 animate-pulse" title={t('transactions.table.unreviewed', 'New / Unreviewed')} />
                                                )}
                                                {tx.is_reconciled && (
                                                    <span title={t('transactions.table.reconciled')} className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-medium bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50 shrink-0">
                                                        <Unlink size={9} /> {t('transactions.table.reconciled')}
                                                    </span>
                                                )}
                                                {!tx.bank_statement_id && (
                                                    <span title={t('transactions.table.liveFeed')} className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800/50 shrink-0">
                                                        {t('transactions.table.liveFeed')}
                                                    </span>
                                                )}
                                                {tx.description}
                                            </span>
                                    </td>
                                )}

                                {visibleCols.location && (
                                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs max-w-[10rem] truncate" title={tx.location}>
                                        {tx.location ? (
                                            <span className="flex items-center gap-1">
                                                    <MapPin size={12} className="opacity-70" /> {tx.location}
                                                </span>
                                        ) : (
                                            <span className="opacity-50">—</span>
                                        )}
                                    </td>
                                )}

                                {visibleCols.reference && (
                                    <td className="px-4 py-3 text-gray-400 dark:text-gray-500 text-xs max-w-[10rem] truncate" title={tx.reference}>
                                        <div className="flex items-center justify-between gap-2">
                                            <span className="truncate">{tx.reference || '—'}</span>
                                            <button type="button" onClick={() => navigator.clipboard.writeText(tx.content_hash)} className="text-gray-300 dark:text-gray-600 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors shrink-0">
                                                <Copy size={12} />
                                            </button>
                                        </div>
                                    </td>
                                )}

                                {visibleCols.category && (
                                    <td className="px-4 py-3">
                                        <select
                                            value={currentCat?.id ?? ''}
                                            onChange={(e) => onCategoryChange(tx.content_hash, e.target.value)}
                                            className="text-xs rounded-lg border border-gray-200 dark:border-gray-700 px-2 py-1 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500 max-w-[13rem] truncate transition-colors hover:border-indigo-300 dark:hover:border-indigo-500"
                                            style={currentCat ? { color: currentCat.color, borderColor: currentCat.color + '55' } : undefined}
                                        >
                                            <option value="">{t('transactions.table.unset')}</option>
                                            {categories.map((c) => (
                                                <option key={c.id} value={c.id}>{c.name}</option>
                                            ))}
                                        </select>
                                    </td>
                                )}

                                {visibleCols.amount && (
                                    <td className={`px-4 py-3 text-right font-mono font-medium whitespace-nowrap ${tx.amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'}`}>
                                            <span className="inline-flex items-center gap-1 justify-end w-full">
                                                {tx.amount >= 0 ? <TrendingUp size={11} className="text-green-500 dark:text-green-400" /> : <TrendingDown size={11} className="text-red-400 dark:text-red-500" />}
                                                {fmtCurrency(tx.amount, tx.currency)}
                                            </span>
                                    </td>
                                )}
                                <td className="px-4 py-3 text-right">
                                    {!tx.reviewed && (
                                        <button
                                            onClick={() => onMarkReviewed?.(tx.content_hash)}
                                            className="p-1.5 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                                            title={t('transactions.markAsReviewed', 'Mark as Reviewed')}
                                        >
                                            <CheckSquare size={16} />
                                        </button>
                                    )}
                                </td>
                            </tr>
                        );
                    })}
                    </tbody>
                </table>
            </div>
            <div className="px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 dark:text-gray-500 text-right">
                {t('transactions.table.showing', { count: transactions.length })}
            </div>
        </div>
    );
}
