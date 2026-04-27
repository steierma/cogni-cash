import * as React from 'react';
import { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Zap, TrendingDown, Star, Info, ChevronRight, ArrowLeftRight, ChevronDown, ChevronUp, Loader2, Users, Search, Columns, Check } from 'lucide-react';
import { Link } from 'react-router-dom';
import { fmtCurrency, fmtDate, getLocalISODate } from '../../utils/formatters';
import CategoryBadge from '../CategoryBadge';
import type { PredictedTransaction, Transaction } from "../../api/types/transaction";
import { transactionService } from '../../api/services/transactionService';

export type ForecastColKey = 'date' | 'description' | 'category' | 'probability' | 'amount';

interface ForecastTableProps {
    predictions: PredictedTransaction[];
    isLoading: boolean;
    visibleCols: Record<ForecastColKey, boolean>;
    showColMenu: boolean;
    onToggleColMenu: () => void;
    onToggleColumn: (key: ForecastColKey) => void;
}

type DisplayItem = 
    | { type: 'row'; prediction: PredictedTransaction; canExpand: boolean; isExpanded: boolean }
    | { type: 'expansion'; prediction: PredictedTransaction };

/** Fetches + filters actual transactions for a category within the current month */
function useBurnRateActuals(categoryId: string | null, enabled: boolean) {
    const now = new Date();
    const monthStart = getLocalISODate(new Date(now.getFullYear(), now.getMonth(), 1));
    const monthEnd = getLocalISODate(new Date(now.getFullYear(), now.getMonth() + 1, 0));

    return useQuery<Transaction[]>({
        queryKey: ['burn-rate-actuals', categoryId, monthStart, monthEnd],
        queryFn: async () => {
            const txns = await transactionService.fetchTransactions(
                undefined, true, categoryId ?? undefined, undefined, undefined, false, true
            );
            return txns.filter(tx => {
                const d = tx.booking_date.slice(0, 10);
                return d >= monthStart && d <= monthEnd;
            });
        },
        enabled: enabled && !!categoryId,
        staleTime: 2 * 60 * 1000,
    });
}

/** Expandable sub-row for a burn rate prediction */
const BurnRateExpansion = React.memo(({
    prediction,
    colSpan,
    baseCurrency,
}: {
    prediction: PredictedTransaction;
    colSpan: number;
    baseCurrency?: string;
}) => {
    const { t } = useTranslation();
    const { data: actuals = [], isLoading } = useBurnRateActuals(prediction.category_id, true);

    const actualTotal = actuals.reduce((sum, tx) => sum + tx.amount, 0);
    const burnRateAmount = prediction.amount; // negative = expense
    const remaining = burnRateAmount - actualTotal; // how much budget remains (both negative)
    const overBudget = actualTotal < burnRateAmount; // spent more than burn rate estimate

    return (
        <tr className="bg-amber-50/40 dark:bg-amber-900/10">
            <td colSpan={colSpan} className="px-6 py-4">
                <div className="space-y-3">
                    {/* Summary bar */}
                    <div className="flex flex-wrap items-center gap-4 text-xs">
                        <span className="font-bold text-amber-700 dark:text-amber-300 uppercase tracking-wider">
                            {t('forecasting.burnRateActuals', 'This Month – Actual vs. Burn Rate')}
                        </span>
                        <div className="flex items-center gap-3 ml-auto">
                            <div className="flex flex-col items-end">
                                <span className="text-gray-400 uppercase tracking-wider">{t('forecasting.burnRateEstimate', 'Burn Rate Est.')}</span>
                                <span className="font-bold text-amber-600 dark:text-amber-400 font-mono">{fmtCurrency(burnRateAmount, baseCurrency ?? 'EUR')}</span>
                            </div>
                            <div className="h-8 w-px bg-amber-200 dark:bg-amber-800" />
                            <div className="flex flex-col items-end">
                                <span className="text-gray-400 uppercase tracking-wider">{t('forecasting.actualSpent', 'Actual Spent')}</span>
                                <span className={`font-bold font-mono ${overBudget ? 'text-rose-600 dark:text-rose-400' : 'text-gray-800 dark:text-gray-200'}`}>
                                    {fmtCurrency(actualTotal, baseCurrency ?? 'EUR')}
                                </span>
                            </div>
                            <div className="h-8 w-px bg-amber-200 dark:bg-amber-800" />
                            <div className="flex flex-col items-end">
                                <span className="text-gray-400 uppercase tracking-wider">{overBudget ? t('forecasting.overBudget', 'Over Budget') : t('forecasting.remaining', 'Remaining')}</span>
                                <span className={`font-bold font-mono ${overBudget ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'}`}>
                                    {overBudget ? '+' : ''}{fmtCurrency(Math.abs(remaining), baseCurrency ?? 'EUR')}
                                </span>
                            </div>
                        </div>
                    </div>

                    {/* Progress bar */}
                    {burnRateAmount !== 0 && (
                        <div className="w-full h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                            <div
                                className={`h-full rounded-full transition-all ${overBudget ? 'bg-rose-500' : 'bg-amber-500'}`}
                                style={{ width: `${Math.min(100, Math.round((actualTotal / burnRateAmount) * 100))}%` }}
                            />
                        </div>
                    )}

                    {/* Actual transactions table */}
                    {isLoading ? (
                        <div className="flex items-center gap-2 text-xs text-gray-400 py-2">
                            <Loader2 size={12} className="animate-spin" />
                            {t('common.loading')}
                        </div>
                    ) : actuals.length === 0 ? (
                        <p className="text-xs text-gray-400 italic py-1">{t('forecasting.noActualsThisMonth', 'No transactions in this category yet this month.')}</p>
                    ) : (
                        <div className="rounded-xl overflow-hidden border border-amber-200 dark:border-amber-800/50">
                            <table className="min-w-full text-xs divide-y divide-amber-100 dark:divide-amber-800/30">
                                <thead className="bg-amber-100/60 dark:bg-amber-900/20 text-amber-700 dark:text-amber-400 uppercase tracking-wider font-bold">
                                    <tr>
                                        <th className="px-4 py-2 text-left">{t('dashboard.recentTxns.date')}</th>
                                        <th className="px-4 py-2 text-left">{t('dashboard.recentTxns.description')}</th>
                                        <th className="px-4 py-2 text-right">{t('dashboard.recentTxns.amount')}</th>
                                    </tr>
                                </thead>
                                <tbody className="bg-white/60 dark:bg-gray-900/40 divide-y divide-amber-50 dark:divide-amber-900/20">
                                    {actuals.map(tx => (
                                        <tr key={tx.id} className="hover:bg-amber-50 dark:hover:bg-amber-900/10 transition-colors">
                                            <td className="px-4 py-2 whitespace-nowrap text-gray-600 dark:text-gray-400 font-mono">{fmtDate(tx.booking_date)}</td>
                                            <td className="px-4 py-2 text-gray-700 dark:text-gray-300 truncate max-w-xs">
                                                {tx.counterparty_name || tx.description}
                                            </td>
                                            <td className={`px-4 py-2 text-right font-mono font-bold ${tx.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-800 dark:text-gray-200'}`}>
                                                {fmtCurrency(tx.amount, tx.currency)}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </td>
        </tr>
    );
});

const ForecastRow = React.memo(({
    prediction,
    visibleCols,
    canExpand,
    isExpanded,
    onToggleExpand,
    t
}: {
    prediction: PredictedTransaction;
    visibleCols: Record<ForecastColKey, boolean>;
    canExpand: boolean;
    isExpanded: boolean;
    onToggleExpand: (id: string) => void;
    t: any;
}) => {
    const isVariable = prediction.description.startsWith('Variable Budget:');
    const isBonus = prediction.description.startsWith('Bonus:');
    const isPlanned = prediction.description.startsWith('Planned:');
    const cleanDescription = isVariable
        ? prediction.description.replace('Variable Budget:', '').trim() 
        : isBonus 
            ? prediction.description.replace('Bonus:', '').trim()
            : isPlanned
                ? prediction.description.replace('Planned:', '').trim()
                : prediction.description;

    let icon = <Zap size={14} />;
    let iconBg = 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400';
    
    if (isVariable) {
        icon = <TrendingDown size={14} />;
        iconBg = 'bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-400';
    } else if (isBonus) {
        icon = <Star size={14} />;
        iconBg = 'bg-emerald-100 dark:bg-emerald-900/40 text-emerald-600 dark:text-emerald-400';
    }

    return (
        <tr
            onClick={canExpand ? () => onToggleExpand(prediction.id) : undefined}
            className={`transition-colors group border-b border-gray-100 dark:border-gray-800/50 ${isVariable ? 'bg-amber-50/20 dark:bg-amber-900/5' : isBonus ? 'bg-emerald-50/20 dark:bg-emerald-900/5' : ''} ${canExpand ? 'cursor-pointer hover:bg-amber-50/60 dark:hover:bg-amber-900/10' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}
        >
            {visibleCols.date && (
                <td className="px-6 py-4 whitespace-nowrap align-top pt-5">
                    <div className="flex flex-col">
                        <span className="text-sm font-bold text-gray-900 dark:text-gray-100">{fmtDate(prediction.booking_date)}</span>
                        <span className="text-[10px] text-gray-400 uppercase font-mono">ESTIMATED</span>
                    </div>
                </td>
            )}
            {visibleCols.description && (
                <td className="px-6 py-4 align-top pt-4">
                    <div className="flex items-center gap-2">
                        <div className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 ${iconBg}`}>
                            {icon}
                        </div>
                        <div className="flex flex-col min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                                {prediction.subscription_id ? (
                                    <Link
                                        to={`/subscriptions/${prediction.subscription_id}`}
                                        className="text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:underline truncate"
                                        title={t('forecasting.viewSubscription', 'View Subscription')}
                                        onClick={e => e.stopPropagation()}
                                    >
                                        {cleanDescription}
                                    </Link>
                                ) : (
                                    <span className="text-sm font-medium text-gray-800 dark:text-gray-200 truncate" title={prediction.description}>
                                        {cleanDescription}
                                    </span>
                                )}
                                {isVariable && (
                                    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 border border-amber-200 dark:border-amber-800/50 uppercase tracking-tighter shrink-0">
                                        {t('forecasting.variableBudget', 'Burn Rate')}
                                    </span>
                                )}
                                {isBonus && (
                                    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800/50 uppercase tracking-tighter shrink-0">
                                        {t('payslips.modals.bonuses', 'Bonus')}
                                    </span>
                                )}
                                {isPlanned && (
                                    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-800/50 uppercase tracking-tighter shrink-0">
                                        {t('forecasting.planned', 'Planned')}
                                    </span>
                                )}
                                {prediction.is_shared && (
                                    <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-800/50 uppercase tracking-tighter shrink-0">
                                        <Users size={10} /> {t('common.shared', 'Shared')}
                                    </span>
                                )}
                            </div>
                            {isVariable && (
                                <span className="text-[10px] text-amber-600 dark:text-amber-400/70 font-medium">
                                    {canExpand
                                        ? t('forecasting.clickToSeeActuals', 'Click to compare with actual transactions')
                                        : 'Based on historical average'}
                                </span>
                            )}
                            {isBonus && (
                                <span className="text-[10px] text-emerald-600 dark:text-emerald-400/70 font-medium">
                                    Seasonal payout detected
                                </span>
                            )}
                            {prediction.subscription_id && (
                                <span className="text-[10px] text-indigo-500 dark:text-indigo-400/70 font-medium">
                                    {t('forecasting.fromSubscription', 'From subscription')}
                                </span>
                            )}
                        </div>
                    </div>
                </td>
            )}
            {visibleCols.category && (
                <td className="px-6 py-4 align-top pt-5">
                    <CategoryBadge category={prediction.category_id ?? undefined} />
                </td>
            )}
            {visibleCols.probability && (
                <td className="px-6 py-4 text-center align-top pt-5">
                    <div className="flex items-center justify-center gap-2">
                        <div className="w-16 bg-gray-100 dark:bg-gray-800 h-1.5 rounded-full overflow-hidden">
                            <div
                                className={`h-full rounded-full ${isVariable ? 'bg-amber-500' : 'bg-indigo-500'}`}
                                style={{ width: `${Math.round(prediction.probability * 100)}%` }}
                            />
                        </div>
                        <span className={`text-xs font-bold ${isVariable ? 'text-amber-600 dark:text-amber-400' : 'text-indigo-600 dark:text-indigo-400'}`}>{Math.round(prediction.probability * 100)}%</span>
                    </div>
                </td>
            )}
            {visibleCols.amount && (
                <td className={`px-6 py-4 text-right font-mono font-bold align-top pt-5 ${prediction.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-gray-100'}`}>
                    <div className="flex flex-col items-end">
                        <div className="flex items-center gap-2">
                            <span>{fmtCurrency(prediction.amount, prediction.currency)}</span>
                            {canExpand && (
                                <span className="text-amber-400 dark:text-amber-500">
                                    {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                                </span>
                            )}
                        </div>
                        {prediction.base_currency && prediction.base_currency !== prediction.currency && prediction.base_amount !== 0 && (
                            <span className="text-[10px] text-gray-400 dark:text-gray-500 font-normal">
                                {fmtCurrency(prediction.base_amount, prediction.base_currency)}
                            </span>
                        )}
                    </div>
                </td>
            )}
        </tr>
    );
});

export default function ForecastTable({
    predictions,
    isLoading,
    visibleCols,
    showColMenu,
    onToggleColMenu,
    onToggleColumn
}: ForecastTableProps) {
    const { t } = useTranslation();
    const [expandedId, setExpandedId] = useState<string | null>(null);
    const parentRef = React.useRef<HTMLDivElement>(null);

    const displayItems = useMemo<DisplayItem[]>(() => {
        const now = new Date();
        const items: DisplayItem[] = [];
        predictions.forEach(p => {
            const bookingDate = new Date(p.booking_date);
            const isCurrentMonth = bookingDate.getMonth() === now.getMonth() && bookingDate.getFullYear() === now.getFullYear();
            const canExpand = p.description.startsWith('Variable Budget:') && isCurrentMonth;
            const isExpanded = expandedId === p.id;
            
            items.push({ type: 'row', prediction: p, canExpand, isExpanded });
            if (canExpand && isExpanded) {
                items.push({ type: 'expansion', prediction: p });
            }
        });
        return items;
    }, [predictions, expandedId]);

    const rowVirtualizer = useVirtualizer({
        count: displayItems.length,
        getScrollElement: () => parentRef.current,
        estimateSize: (index) => displayItems[index].type === 'expansion' ? 220 : 72,
        overscan: 10,
    });

    const virtualRows = rowVirtualizer.getVirtualItems();
    const totalSize = rowVirtualizer.getTotalSize();

    const paddingTop = virtualRows.length > 0 ? virtualRows[0].start : 0;
    const paddingBottom = virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1].end : 0;

    const toggleExpand = React.useCallback((id: string) => {
        setExpandedId(prev => prev === id ? null : id);
    }, []);

    if (isLoading) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden p-8 flex flex-col gap-4">
                {[1, 2, 3, 4].map(i => (
                    <div key={i} className="h-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl animate-pulse" />
                ))}
            </div>
        );
    }

    if (predictions.length === 0) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden p-12 flex flex-col items-center justify-center text-center">
                <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-full mb-4">
                    <Search size={32} className="text-gray-300 dark:text-gray-600" />
                </div>
                <h3 className="text-gray-900 dark:text-gray-100 font-semibold">{t('forecasting.noForecasts')}</h3>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1 max-w-xs">
                    {t('forecasting.noForecastsDesc', 'Forecasts are automatically generated as you import more historical transaction data or based on your filters.')}
                </p>
            </div>
        );
    }

    const colCount = Object.values(visibleCols).filter(Boolean).length;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col">
            <div 
                ref={parentRef}
                className="overflow-x-auto overflow-y-auto max-h-[calc(100vh-320px)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700"
            >
                <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800 text-sm border-separate border-spacing-0">
                    <thead className="bg-gray-50 dark:bg-gray-800/90 backdrop-blur-sm text-xs uppercase text-gray-400 dark:text-gray-500 font-bold tracking-wider sticky top-0 z-20 shadow-sm">
                        <tr>
                            {visibleCols.date && <th className="px-6 py-4 text-left bg-inherit border-b border-gray-100 dark:border-gray-800">{t('dashboard.recentTxns.date')}</th>}
                            {visibleCols.description && <th className="px-6 py-4 text-left bg-inherit border-b border-gray-100 dark:border-gray-800">{t('dashboard.recentTxns.description')}</th>}
                            {visibleCols.category && <th className="px-6 py-4 text-left bg-inherit border-b border-gray-100 dark:border-gray-800">{t('dashboard.recentTxns.category')}</th>}
                            {visibleCols.probability && <th className="px-6 py-4 text-center bg-inherit border-b border-gray-100 dark:border-gray-800">{t('forecasting.probability')}</th>}
                            {visibleCols.amount && (
                                <th className="px-6 py-4 text-right bg-inherit border-b border-gray-100 dark:border-gray-800">
                                    <div className="flex items-center justify-end gap-3">
                                        <span>{t('dashboard.recentTxns.amount')}</span>
                                        <div className="relative normal-case tracking-normal font-medium">
                                            <button
                                                onClick={onToggleColMenu}
                                                className={`p-1.5 rounded-lg border transition-all ${showColMenu
                                                    ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-200 dark:border-indigo-800 text-indigo-600 dark:text-indigo-400'
                                                    : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}
                                                title={t('transactions.columns', 'Columns')}
                                            >
                                                <Columns size={12} />
                                            </button>

                                            {showColMenu && (
                                                <>
                                                    <div className="fixed inset-0 z-10" onClick={onToggleColMenu} />
                                                    <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-xl shadow-xl border border-gray-200 dark:border-gray-700 p-2 z-20 animate-in fade-in zoom-in duration-150 origin-top-right">
                                                        {[
                                                            { key: 'date', label: t('dashboard.recentTxns.date') },
                                                            { key: 'description', label: t('dashboard.recentTxns.description') },
                                                            { key: 'category', label: t('dashboard.recentTxns.category') },
                                                            { key: 'probability', label: t('forecasting.probability') },
                                                            { key: 'amount', label: t('dashboard.recentTxns.amount') },
                                                        ].map(({ key, label }) => (
                                                            <button
                                                                key={key}
                                                                onClick={() => onToggleColumn(key as ForecastColKey)}
                                                                className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors text-left"
                                                            >
                                                                {label} {visibleCols[key as ForecastColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
                                                            </button>
                                                        ))}
                                                    </div>
                                                </>
                                            )}
                                        </div>
                                    </div>
                                </th>
                            )}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                        {paddingTop > 0 && (
                            <tr>
                                <td style={{ height: `${paddingTop}px` }} colSpan={colCount} />
                            </tr>
                        )}
                        {virtualRows.map((virtualRow) => {
                            const item = displayItems[virtualRow.index];
                            if (item.type === 'row') {
                                return (
                                    <ForecastRow
                                        key={item.prediction.id}
                                        prediction={item.prediction}
                                        visibleCols={visibleCols}
                                        canExpand={item.canExpand}
                                        isExpanded={item.isExpanded}
                                        onToggleExpand={toggleExpand}
                                        t={t}
                                    />
                                );
                            }
                            return (
                                <BurnRateExpansion
                                    key={`${item.prediction.id}-expansion`}
                                    prediction={item.prediction}
                                    colSpan={colCount}
                                />
                            );
                        })}
                        {paddingBottom > 0 && (
                            <tr>
                                <td style={{ height: `${paddingBottom}px` }} colSpan={colCount} />
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
            <div className="bg-gray-50 dark:bg-gray-800/50 px-6 py-4 flex justify-between items-center border-t border-gray-100 dark:border-gray-800">
                <p className="text-xs text-gray-400 dark:text-gray-500 flex items-center gap-1.5">
                    <Info size={14} /> {t('forecasting.subtitle')}
                </p>
                <Link 
                    to="/transactions" 
                    className="text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:underline flex items-center gap-1"
                >
                    <ArrowLeftRight size={14} /> {t('layout.transactions')} <ChevronRight size={14} />
                </Link>
            </div>
        </div>
    );
}
