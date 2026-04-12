import { useTranslation } from 'react-i18next';
import { CheckSquare, ChevronDown, ChevronUp, Copy, Square, TrendingDown, TrendingUp, Unlink, MapPin, User, Zap, BarChart3, Slash } from 'lucide-react';
import type { Category, Transaction, PatternExclusion } from '../../api/types';
import { fmtCurrency, fmtDate } from '../../utils/formatters';

export type TxColKey = 'date' | 'description' | 'counterparty' | 'location' | 'reference' | 'category' | 'amount';
export type SortKey = 'booking_date' | 'description' | 'location' | 'amount';
export type SortDir = 'asc' | 'desc';

interface TransactionTableProps {
    transactions: Transaction[];
    categories: Category[];
    patternExclusions?: PatternExclusion[];
    selectedHashes: Set<string>;
    onToggleSelect: (hash: string) => void;
    onToggleSelectAll: () => void;
    sortKey: SortKey;
    sortDir: SortDir;
    onSort: (key: SortKey) => void;
    onCategoryChange: (hash: string, categoryId: string) => void;
    onMarkReviewed?: (hash: string) => void;
    onTogglePatternExclusion?: (matchTerm: string, excluded: boolean) => void;
    visibleCols: Record<TxColKey, boolean>;
}

export default function TransactionTable({
                                             transactions,
                                             categories,
                                             patternExclusions = [],
                                             selectedHashes,
                                             onToggleSelect,
                                             onToggleSelectAll,
                                             sortKey,
                                             sortDir,
                                             onSort,
                                             onCategoryChange,
                                             onMarkReviewed,
                                             onTogglePatternExclusion,
                                             visibleCols
                                         }: TransactionTableProps) {
    const { t } = useTranslation();

    const normalize = (desc: string) => desc.trim().toLowerCase().slice(0, 25);

    const isTermExcluded = (term: string) => {
        return patternExclusions.some(pe => pe.match_term === term);
    };

    const renderSortIcon = (k: SortKey) => {
        if (sortKey !== k) return null;
        return sortDir === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />;
    };

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
                                <span className="inline-flex items-center gap-1">{t('transactions.table.date', 'Date')} {renderSortIcon('booking_date')}</span>
                            </th>
                        )}

                        {(visibleCols.description || visibleCols.counterparty) && (
                            <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none align-top"
                                onClick={() => onSort('description')}
                            >
                                <div className="flex flex-col gap-0.5">
                                    <span className="inline-flex items-center gap-1">
                                        {t('transactions.table.description', 'Description')}
                                        {renderSortIcon('description')}
                                    </span>
                                    {visibleCols.counterparty && (
                                        <span className="text-xs font-normal text-gray-500 dark:text-gray-400 normal-case">
                                            {t('transactions.table.counterparty', 'Counterparty')}
                                        </span>
                                    )}
                                </div>
                            </th>
                        )}

                        {visibleCols.location && (
                            <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none" onClick={() => onSort('location')}>
                                <span className="inline-flex items-center gap-1">{t('transactions.table.location', 'Location')} {renderSortIcon('location')}</span>
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
                                <span className="inline-flex items-center gap-1 justify-end">{t('transactions.table.amount', 'Amount')} {renderSortIcon('amount')}</span>
                            </th>
                        )}
                        <th className="px-4 py-3 text-right w-24">{t('common.actions', 'Actions')}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                    {transactions.map((tx) => {
                        const isSelected = selectedHashes.has(tx.content_hash);
                        const currentCat = categories.find((c) => c.id === tx.category_id);
                        const term = normalize(tx.description);
                        const isSkipped = isTermExcluded(term);

                        return (
                            <tr key={tx.content_hash} className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors ${isSelected ? 'bg-indigo-50/30 dark:bg-indigo-900/20' : ''} ${tx.is_reconciled ? 'opacity-60' : ''} ${isSkipped ? 'opacity-70 bg-amber-50/5 dark:bg-amber-900/5' : ''} ${tx.is_prediction ? 'bg-indigo-50/10 dark:bg-indigo-900/5 border-l-2 border-l-indigo-500 dark:border-l-indigo-400' : ''}`}>
                                <td className="px-4 py-3 align-top pt-4">
                                    <button
                                        type="button"
                                        onClick={() => onToggleSelect(tx.content_hash)}
                                        className={`transition-colors ${isSelected ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                                    >
                                        {isSelected ? <CheckSquare size={16} /> : <Square size={16} />}
                                    </button>
                                </td>

                                {visibleCols.date && (
                                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap align-top pt-4">
                                        <span>{fmtDate(tx.booking_date)}</span>
                                    </td>
                                )}

                                {visibleCols.description  && (
                                    <td className="px-4 py-3 align-top pt-3.5 max-w-xs">
                                        <div className="flex flex-col gap-1">
                                            {visibleCols.description && (
                                                <span className={`flex items-center gap-1.5 text-gray-800 dark:text-gray-200 truncate`} title={tx.description}>
                                                    {tx.is_prediction && (
                                                        <span title={t('forecasting.isPrediction')} className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800/50 shrink-0 uppercase tracking-tighter">
                                                            <Zap size={9} /> {t('forecasting.isPrediction')}
                                                        </span>
                                                    )}
                                                    {tx.is_prediction && tx.description.startsWith('Variable Budget:') && (
                                                        <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-50 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400 border border-amber-200 dark:border-amber-800/50 shrink-0 uppercase tracking-tighter">
                                                            {t('forecasting.variableBudget', 'Burn Rate')}
                                                        </span>
                                                    )}
                                                    {!tx.reviewed && !tx.is_prediction && (
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
                                                    <span className="truncate">{tx.description}</span>
                                                </span>
                                            )}
                                            <span className="text-gray-500 dark:text-gray-400 text-xs truncate max-w-[12rem]"
                                                  title={[tx.counterparty_name, tx.counterparty_iban, tx.bank_transaction_code].filter(Boolean).join(' · ')}>
                                                {tx.counterparty_name ? (
                                                    <span className="flex items-center gap-1">
                                                        <User size={12} className="opacity-70 shrink-0" />
                                                        <span className="truncate">{tx.counterparty_name}</span>
                                                    </span>
                                                ) : tx.bank_transaction_code ? (
                                                    <span className="opacity-60 italic">{tx.bank_transaction_code}</span>
                                                ) : (
                                                    <span className="opacity-50">—</span>
                                                )}
                                            </span>
                                        </div>
                                    </td>
                                )}

                                {visibleCols.location && (
                                    <td className="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs max-w-[10rem] truncate align-top pt-4" title={tx.location}>
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
                                    <td className="px-4 py-3 text-gray-400 dark:text-gray-500 text-xs max-w-[10rem] truncate align-top pt-4" title={tx.reference}>
                                        <div className="flex items-center justify-between gap-2">
                                            <span className="truncate">{tx.reference || '—'}</span>
                                            <button type="button" onClick={() => navigator.clipboard.writeText(tx.content_hash)} className="text-gray-300 dark:text-gray-600 hover:text-indigo-500 dark:hover:text-indigo-400 transition-colors shrink-0">
                                                <Copy size={12} />
                                            </button>
                                        </div>
                                    </td>
                                )}

                                {visibleCols.category && (
                                    <td className="px-4 py-3 align-top pt-3.5">
                                        <select
                                            value={currentCat?.id ?? ''}
                                            onChange={(e) => onCategoryChange(tx.content_hash, e.target.value)}
                                            className="text-xs rounded-lg border border-gray-200 dark:border-gray-700 px-2 py-1 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500 max-w-[13rem] truncate transition-colors hover:border-indigo-300 dark:hover:border-indigo-500"
                                            style={currentCat ? { color: currentCat.color, borderColor: currentCat.color + '55' } : undefined}
                                        >
                                            <option value="">{t('transactions.table.unset')}</option>
                                            {categories.filter(c => !c.deleted_at || c.id === currentCat?.id).map((c) => (
                                                <option key={c.id} value={c.id}>{c.name}</option>
                                            ))}
                                        </select>
                                    </td>
                                )}

                                {visibleCols.amount && (
                                    <td className={`px-4 py-3 text-right font-mono font-medium whitespace-nowrap align-top pt-4 ${isSkipped ? 'opacity-40' : tx.amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'}`}>
                                            <span className="inline-flex items-center gap-1 justify-end w-full">
                                                {tx.amount >= 0 ? <TrendingUp size={11} className={isSkipped ? '' : "text-green-500 dark:text-green-400"} /> : <TrendingDown size={11} className={isSkipped ? '' : "text-red-400 dark:text-red-500"} />}
                                                {fmtCurrency(tx.amount, tx.currency)}
                                            </span>
                                    </td>
                                )}
                                <td className="px-4 py-3 text-right align-top pt-3.5 whitespace-nowrap">
                                    <div className="flex items-center justify-end gap-1">
                                        <button
                                            onClick={() => onTogglePatternExclusion?.(term, !isSkipped)}
                                            className={`p-1.5 rounded-lg transition-colors relative ${isSkipped
                                                ? 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/30'
                                                : 'text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20'}`}
                                            title={isSkipped ? t('transactions.includePatternInForecasting') : t('transactions.excludePatternFromForecasting')}
                                        >
                                            <div className="relative">
                                                <BarChart3 size={16} />
                                                {isSkipped && <Slash size={10} className="absolute inset-0 m-auto text-amber-600 dark:text-amber-400" />}
                                            </div>
                                        </button>

                                        {!tx.reviewed && !tx.is_prediction && (
                                            <button
                                                onClick={() => onMarkReviewed?.(tx.content_hash)}
                                                className="p-1.5 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                                                title={t('transactions.markAsReviewed', 'Mark as Reviewed')}
                                            >
                                                <CheckSquare size={16} />
                                            </button>
                                        )}
                                    </div>
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