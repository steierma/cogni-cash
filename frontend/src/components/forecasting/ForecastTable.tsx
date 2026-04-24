import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Zap, TrendingDown, Star, Info, ChevronRight, Search, ArrowLeftRight, ChevronDown, ChevronUp, Loader2, Users } from 'lucide-react';
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
}

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
function BurnRateExpansion({
    prediction,
    colSpan,
    baseCurrency,
}: {
    prediction: PredictedTransaction;
    colSpan: number;
    baseCurrency?: string;
}) {
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
}

export default function ForecastTable({
    predictions,
    isLoading,
    visibleCols,
}: ForecastTableProps) {
    const { t } = useTranslation();
    const [expandedId, setExpandedId] = useState<string | null>(null);

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
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm">
                    <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 font-bold tracking-wider">
                        <tr>
                            {visibleCols.date && <th className="px-6 py-4 text-left">{t('dashboard.recentTxns.date')}</th>}
                            {visibleCols.description && <th className="px-6 py-4 text-left">{t('dashboard.recentTxns.description')}</th>}
                            {visibleCols.category && <th className="px-6 py-4 text-left">{t('dashboard.recentTxns.category')}</th>}
                            {visibleCols.probability && <th className="px-6 py-4 text-center">{t('forecasting.probability')}</th>}
                            {visibleCols.amount && <th className="px-6 py-4 text-right">{t('dashboard.recentTxns.amount')}</th>}
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                        {predictions.map((p) => {
                            const isVariable = p.description.startsWith('Variable Budget:');
                            const isBonus = p.description.startsWith('Bonus:');
                            const isPlanned = p.description.startsWith('Planned:');
                            const cleanDescription = isVariable
                                ? p.description.replace('Variable Budget:', '').trim() 
                                : isBonus 
                                    ? p.description.replace('Bonus:', '').trim()
                                    : isPlanned
                                        ? p.description.replace('Planned:', '').trim()
                                        : p.description;

                            let icon = <Zap size={14} />;
                            let iconBg = 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400';
                            
                            if (isVariable) {
                                icon = <TrendingDown size={14} />;
                                iconBg = 'bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-400';
                            } else if (isBonus) {
                                icon = <Star size={14} />;
                                iconBg = 'bg-emerald-100 dark:bg-emerald-900/40 text-emerald-600 dark:text-emerald-400';
                            }

                            const isExpanded = expandedId === p.id;

                            // Only the first occurrence per category (current month burn rates)
                            const now = new Date();
                            const bookingDate = new Date(p.booking_date);
                            const isCurrentMonth = bookingDate.getMonth() === now.getMonth() && bookingDate.getFullYear() === now.getFullYear();
                            const canExpand = isVariable && isCurrentMonth;

                            return (
                                <>
                                    <tr
                                        key={p.id}
                                        onClick={canExpand ? () => setExpandedId(isExpanded ? null : p.id) : undefined}
                                        className={`transition-colors group ${isVariable ? 'bg-amber-50/20 dark:bg-amber-900/5' : isBonus ? 'bg-emerald-50/20 dark:bg-emerald-900/5' : ''} ${canExpand ? 'cursor-pointer hover:bg-amber-50/60 dark:hover:bg-amber-900/10' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}
                                    >
                                        {visibleCols.date && (
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex flex-col">
                                                    <span className="text-sm font-bold text-gray-900 dark:text-gray-100">{fmtDate(p.booking_date)}</span>
                                                    <span className="text-[10px] text-gray-400 uppercase font-mono">ESTIMATED</span>
                                                </div>
                                            </td>
                                        )}
                                        {visibleCols.description && (
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${iconBg}`}>
                                                        {icon}
                                                    </div>
                                                    <div className="flex flex-col">
                                                        <div className="flex items-center gap-2">
                                                            {p.subscription_id ? (
                                                                <Link
                                                                    to={`/subscriptions/${p.subscription_id}`}
                                                                    className="text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:underline"
                                                                    title={t('forecasting.viewSubscription', 'View Subscription')}
                                                                    onClick={e => e.stopPropagation()}
                                                                >
                                                                    {cleanDescription}
                                                                </Link>
                                                            ) : (
                                                                <span className="text-sm font-medium text-gray-800 dark:text-gray-200" title={p.description}>
                                                                    {cleanDescription}
                                                                </span>
                                                            )}
                                                            {isVariable && (
                                                                <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 border border-amber-200 dark:border-amber-800/50 uppercase tracking-tighter">
                                                                    {t('forecasting.variableBudget', 'Burn Rate')}
                                                                </span>
                                                            )}
                                                            {isBonus && (
                                                                <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300 border border-emerald-200 dark:border-emerald-800/50 uppercase tracking-tighter">
                                                                    {t('payslips.modals.bonuses', 'Bonus')}
                                                                </span>
                                                            )}
                                                            {isPlanned && (
                                                                <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 border border-blue-200 dark:border-blue-800/50 uppercase tracking-tighter">
                                                                    {t('forecasting.planned', 'Planned')}
                                                                </span>
                                                            )}
                                                            {p.is_shared && (
                                                                <span className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-bold bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-800/50 uppercase tracking-tighter">
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
                                                        {p.subscription_id && (
                                                            <span className="text-[10px] text-indigo-500 dark:text-indigo-400/70 font-medium">
                                                                {t('forecasting.fromSubscription', 'From subscription')}
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>
                                            </td>
                                        )}
                                        {visibleCols.category && (
                                            <td className="px-6 py-4">
                                                <CategoryBadge category={p.category_id ?? undefined} />
                                            </td>
                                        )}
                                        {visibleCols.probability && (
                                            <td className="px-6 py-4 text-center">
                                                <div className="flex items-center justify-center gap-2">
                                                    <div className="w-16 bg-gray-100 dark:bg-gray-800 h-1.5 rounded-full overflow-hidden">
                                                        <div
                                                            className={`h-full rounded-full ${isVariable ? 'bg-amber-500' : 'bg-indigo-500'}`}
                                                            style={{ width: `${Math.round(p.probability * 100)}%` }}
                                                        />
                                                    </div>
                                                    <span className={`text-xs font-bold ${isVariable ? 'text-amber-600 dark:text-amber-400' : 'text-indigo-600 dark:text-indigo-400'}`}>{Math.round(p.probability * 100)}%</span>
                                                </div>
                                            </td>
                                        )}
                                        {visibleCols.amount && (
                                            <td className={`px-6 py-4 text-right font-mono font-bold ${p.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-gray-100'}`}>
                                                <div className="flex flex-col items-end">
                                                    <div className="flex items-center gap-2">
                                                        <span>{fmtCurrency(p.amount, p.currency)}</span>
                                                        {canExpand && (
                                                            <span className="text-amber-400 dark:text-amber-500">
                                                                {isExpanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
                                                            </span>
                                                        )}
                                                    </div>
                                                    {p.base_currency && p.base_currency !== p.currency && p.base_amount !== 0 && (
                                                        <span className="text-[10px] text-gray-400 dark:text-gray-500 font-normal">
                                                            {fmtCurrency(p.base_amount, p.base_currency)}
                                                        </span>
                                                    )}
                                                </div>
                                            </td>
                                        )}
                                    </tr>
                                    {canExpand && isExpanded && (
                                        <BurnRateExpansion
                                            key={`${p.id}-expansion`}
                                            prediction={p}
                                            colSpan={colCount}
                                        />
                                    )}
                                </>
                            );
                        })}
                    </tbody>
                </table>
            </div>
            <div className="bg-gray-50 dark:bg-gray-800/50 px-6 py-4 flex justify-between items-center">
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
