import { useTranslation } from 'react-i18next';
import { Zap, TrendingDown, Star, RotateCcw, Trash2, Info, ChevronRight, Search, ArrowLeftRight } from 'lucide-react';
import { Link } from 'react-router-dom';
import { fmtCurrency, fmtDate } from '../../utils/formatters';
import CategoryBadge from '../CategoryBadge';
import type { PredictedTransaction } from "../../api/types/transaction";
export type ForecastColKey = 'date' | 'description' | 'category' | 'probability' | 'amount';

interface ForecastTableProps {
    predictions: PredictedTransaction[];
    isLoading: boolean;
    visibleCols: Record<ForecastColKey, boolean>;
    onInclude: (id: string) => void;
    onExclude: (id: string) => void;
}

export default function ForecastTable({
    predictions,
    isLoading,
    visibleCols,
    onInclude,
    onExclude
}: ForecastTableProps) {
    const { t } = useTranslation();

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
                            <th className="px-6 py-4 text-right">{t('common.actions')}</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                        {predictions.map((p) => {
                            const isVariable = p.description.startsWith('Variable Budget:');
                            const isBonus = p.description.startsWith('Bonus:');
                            const cleanDescription = isVariable 
                                ? p.description.replace('Variable Budget:', '').trim() 
                                : isBonus 
                                    ? p.description.replace('Bonus:', '').trim()
                                    : p.description;
                            const isExcluded = p.skip_forecasting === true;

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
                                <tr key={p.id} className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group ${isVariable ? 'bg-amber-50/20 dark:bg-amber-900/5' : isBonus ? 'bg-emerald-50/20 dark:bg-emerald-900/5' : ''} ${isExcluded ? 'opacity-50 grayscale select-none' : ''}`}>
                                    {visibleCols.date && (
                                        <td className="px-6 py-4 whitespace-nowrap">
                                            <div className="flex flex-col">
                                                <span className={`text-sm font-bold text-gray-900 dark:text-gray-100 ${isExcluded ? 'line-through' : ''}`}>{fmtDate(p.booking_date)}</span>
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
                                                        <span className={`text-sm font-medium text-gray-800 dark:text-gray-200 ${isExcluded ? 'line-through' : ''}`} title={p.description}>
                                                            {cleanDescription}
                                                        </span>
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
                                                    </div>
                                                    {isVariable && (
                                                        <span className="text-[10px] text-amber-600 dark:text-amber-400/70 font-medium">
                                                            Based on historical average
                                                        </span>
                                                    )}
                                                    {isBonus && (
                                                        <span className="text-[10px] text-emerald-600 dark:text-emerald-400/70 font-medium">
                                                            Seasonal payout detected
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
                                        <td className={`px-6 py-4 text-right font-mono font-bold ${isExcluded ? 'text-gray-400 line-through' : p.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-gray-100'}`}>
                                            {fmtCurrency(p.amount, p.currency)}
                                        </td>
                                    )}
                                    <td className="px-6 py-4 text-right">
                                        {isExcluded ? (
                                            <button
                                                onClick={() => onInclude(p.id)}
                                                className="p-1.5 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                                                title={t('forecasting.includeProjection')}
                                            >
                                                <RotateCcw size={16} />
                                            </button>
                                        ) : (
                                            <button
                                                onClick={() => onExclude(p.id)}
                                                className="p-1.5 text-gray-400 hover:text-rose-600 dark:hover:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-900/20 rounded-lg transition-colors"
                                                title={t('forecasting.excludeProjection')}
                                            >
                                                <Trash2 size={16} />
                                            </button>
                                        )}
                                    </td>
                                </tr>
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
