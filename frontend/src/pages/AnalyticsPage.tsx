import { useState, useMemo, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { transactionService } from '../api/services/transactionService';
import { categoryService } from '../api/services/categoryService';
import { payslipService } from '../api/services/payslipService';
import { settingsService } from '../api/services/settingsService';
import { BarChart3, Filter, X, ArrowRightLeft, TrendingUp, TrendingDown, Wallet, Search, BarChart as BarChartIcon, Briefcase, Activity } from 'lucide-react';
import {
    AreaChart, Area, BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip as RechartsTooltip, ResponsiveContainer, Cell
} from 'recharts';
import type { Transaction } from "../api/types/transaction";
import type { Category } from "../api/types/category";
import type { Payslip } from "../api/types/payslip";
import { fmtCurrency } from '../utils/formatters';

interface FilterState {
    from: string;
    to: string;
    excludeIds: Set<string>;
    hideReconciled: boolean;
}

const toIsoDate = (dateStr: string): string => {
    if (!dateStr) return '';
    if (/^\d{4}-\d{2}-\d{2}/.test(dateStr)) return dateStr.slice(0, 10);
    if (/^\d{2}\.\d{2}\.\d{4}/.test(dateStr)) {
        const [d, m, y] = dateStr.slice(0, 10).split('.');
        return `${y}-${m}-${d}`;
    }
    return dateStr;
};

export default function AnalyticsPage() {
    const { t, i18n } = useTranslation();
    const [searchParams, setSearchParams] = useSearchParams();

    // --- Data Fetching ---
    const { data: allTransactions = [], isLoading: isLoadingTxns } = useQuery<Transaction[]>({
        queryKey: ['transactions', undefined, false],
        queryFn: () => transactionService.fetchTransactions(undefined, false),
    });

    const { data: categories = [] } = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: categoryService.fetchCategories,
    });

    const { data: payslips = [] } = useQuery<Payslip[], Error>({
        queryKey: ['payslips'],
        queryFn: () => payslipService.fetchPayslips(),
    });

    const { data: baseCurrency = 'EUR' } = useQuery({
        queryKey: ['settings', 'BASE_DISPLAY_CURRENCY'],
        queryFn: () => settingsService.fetchSettings().then((s) => s['BASE_DISPLAY_CURRENCY'] || 'EUR'),
    });

    // --- State: Filters ---
    const [minDate, maxDate] = useMemo((): [string, string] => {
        const days = allTransactions.map(t => toIsoDate(t.booking_date)).filter(Boolean).sort();
        if (days.length === 0) return ['', ''];
        return [days[0], days[days.length - 1]];
    }, [allTransactions]);

    const initialFilters: FilterState = useMemo(() => {
        const from = searchParams.get('from') || '';
        const to = searchParams.get('to') || '';
        const excludeStr = searchParams.get('exclude') || '';
        const hideReconciled = searchParams.get('hide_reconciled') !== 'false';
        
        return {
            from,
            to,
            excludeIds: new Set(excludeStr ? excludeStr.split(',') : []),
            hideReconciled
        };
    }, [searchParams]);

    const [draft, setDraft] = useState<FilterState>(initialFilters);
    const [applied, setApplied] = useState<FilterState>(initialFilters);

    // Default to full range if none specified in URL and transactions are loaded
    const hasSetDefaults = useRef(false);
    useEffect(() => {
        if (!hasSetDefaults.current && minDate && maxDate && !initialFilters.from && !initialFilters.to) {
            const defaultFilter = { ...initialFilters, from: minDate, to: maxDate };
            setDraft(defaultFilter);
            setApplied(defaultFilter);
            hasSetDefaults.current = true;
        }
    }, [minDate, maxDate, initialFilters]);

    // Update URL when applied filters change
    useEffect(() => {
        const next = new URLSearchParams();
        if (applied.from) next.set('from', applied.from);
        if (applied.to) next.set('to', applied.to);
        if (applied.excludeIds.size > 0) next.set('exclude', Array.from(applied.excludeIds).join(','));
        if (applied.hideReconciled === false) next.set('hide_reconciled', 'false');
        
        // Only update if it actually changed to avoid infinite loops
        const currentStr = searchParams.toString();
        const nextStr = next.toString();
        if (currentStr !== nextStr) {
            setSearchParams(next, { replace: true });
        }
    }, [applied, setSearchParams, searchParams]);

    const isDirty = JSON.stringify({ ...draft, excludeIds: Array.from(draft.excludeIds).sort() }) !==
        JSON.stringify({ ...applied, excludeIds: Array.from(applied.excludeIds).sort() });

    const toggleExcludeCategory = (id: string) => {
        const next = new Set(draft.excludeIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        setDraft({ ...draft, excludeIds: next });
    };

    const handleApply = (e?: React.FormEvent) => {
        e?.preventDefault();
        setApplied(draft);
    };

    const handleClear = () => {
        const reset = {
            from: minDate,
            to: maxDate,
            excludeIds: new Set<string>(),
            hideReconciled: true
        };
        setDraft(reset);
        setApplied(reset);
    };

    // --- Data Processing ---
    const filteredTxns = useMemo(() => {
        return allTransactions.filter(t => {
            if (applied.hideReconciled && t.is_reconciled) return false;

            if (t.category_id && applied.excludeIds.has(t.category_id)) return false;
            if (!t.category_id && applied.excludeIds.has('uncategorized')) return false;

            const day = toIsoDate(t.booking_date);
            if (applied.from && day < applied.from) return false;
            if (applied.to && day > applied.to) return false;

            return true;
        });
    }, [allTransactions, applied]);

    // --- KPIs & Bar Chart Data ---
    const { expenseBarData, totalInc, totalExp } = useMemo(() => {
        const categoryNet: Record<string, { amount: number, color: string }> = {};
        let tInc = 0;
        let tExp = 0;

        filteredTxns.forEach(t => {
            if (t.amount > 0) {
                tInc += t.amount;
            } else {
                tExp += Math.abs(t.amount);
            }

            const cat = categories.find(c => c.id === t.category_id);
            const catName = cat?.name || ('analytics.uncategorized');
            const catColor = cat?.color || '#94a3b8';

            if (!categoryNet[catName]) categoryNet[catName] = { amount: 0, color: catColor };
            categoryNet[catName].amount += t.amount;
        });

        const barData = Object.entries(categoryNet)
            .filter(([_, data]) => data.amount < 0) // Only show categories that are net expenses
            .map(([name, data]) => ({ name, value: Math.abs(data.amount), color: data.color }))
            .sort((a, b) => b.value - a.value);

        return {
            expenseBarData: barData,
            totalInc: tInc,
            totalExp: tExp
        };
    }, [filteredTxns, categories, t]);

    // --- Trend Data (Area Chart & Waterfall) ---
    const trendData = useMemo(() => {
        const monthly: Record<string, { month: string, income: number, expense: number }> = {};

        filteredTxns.forEach(t => {
            const isoDate = toIsoDate(t.booking_date);
            const month = isoDate.slice(0, 7); // YYYY-MM
            if (!monthly[month]) monthly[month] = { month, income: 0, expense: 0 };

            if (t.amount > 0) monthly[month].income += t.amount;
            else monthly[month].expense += Math.abs(t.amount);
        });

        return Object.values(monthly).sort((a, b) => a.month.localeCompare(b.month)).map(m => {
            const [y, mm] = m.month.split('-');
            const hrDoc = (payslips as Payslip[]).find(p => p.period_year === parseInt(y, 10) && p.period_month_num === parseInt(mm, 10));

            return {
                ...m,
                netIncome: hrDoc ? hrDoc.net_pay : 0,
                deductions: hrDoc ? (hrDoc.gross_pay - hrDoc.net_pay) : 0,
            };
        });
    }, [filteredTxns, payslips]);

    // --- Category Specific Trend Data ---
    const [selectedCategoryTrendId, setSelectedCategoryTrendId] = useState<string>('');

    const categoryTrendData = useMemo(() => {
        if (!selectedCategoryTrendId) return [];

        const monthly: Record<string, { month: string, amount: number }> = {};

        // Pre-fill months from the existing trendData to ensure a continuous X-axis
        trendData.forEach(td => {
            monthly[td.month] = { month: td.month, amount: 0 };
        });

        filteredTxns.forEach(t => {
            const isMatch = t.category_id === selectedCategoryTrendId ||
                (!t.category_id && selectedCategoryTrendId === 'uncategorized');

            if (isMatch) {
                const month = toIsoDate(t.booking_date).slice(0, 7);
                if (monthly[month]) {
                    monthly[month].amount += t.amount;
                } else {
                    monthly[month] = { month, amount: t.amount };
                }
            }
        });

        return Object.values(monthly).sort((a, b) => a.month.localeCompare(b.month));
    }, [filteredTxns, selectedCategoryTrendId, trendData]);

    const formatMonthAxis = (yyyyMM: string) => {
        try {
            const [y, m] = yyyyMM.split('-');
            const date = new Date(parseInt(y, 10), parseInt(m, 10) - 1);
            return date.toLocaleDateString(i18n.language, { month: 'short', year: '2-digit' });
        } catch {
            return yyyyMM;
        }
    };

    if (isLoadingTxns) return <div className="p-8 text-gray-500 animate-pulse">{t('common.loading')}</div>;

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            <div>
                <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                    <BarChart3 className="text-indigo-600 dark:text-indigo-400" /> {t('analytics.title')}
                </h1>
                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                    {t('analytics.subtitle')}
                </p>
            </div>

            <form onSubmit={handleApply} className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-4 space-y-4">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2 text-sm font-medium text-gray-500 dark:text-gray-400">
                        <Filter size={14} /> {t('analytics.filters')}
                    </div>
                    {isDirty && (
                        <span className="text-[10px] font-bold text-indigo-500 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded-full uppercase tracking-wider animate-pulse">
                            {t('analytics.unapplied')}
                        </span>
                    )}
                </div>

                <div className="flex flex-col lg:flex-row gap-6">
                    <div className="space-y-4 min-w-[280px]">
                        <div>
                            <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block uppercase tracking-wider font-bold">{t('analytics.dateRange')}</label>
                            <div className="flex items-center gap-2">
                                <input
                                    type="date"
                                    value={draft.from}
                                    onChange={(e) => setDraft({ ...draft, from: e.target.value })}
                                    min={minDate} max={maxDate}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-2 text-gray-700 dark:text-gray-300 focus:ring-2 focus:ring-indigo-300 outline-none [color-scheme:light] dark:[color-scheme:dark]"
                                />
                                <span className="text-gray-400">-</span>
                                <input
                                    type="date"
                                    value={draft.to}
                                    onChange={(e) => setDraft({ ...draft, to: e.target.value })}
                                    min={minDate} max={maxDate}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-2 text-gray-700 dark:text-gray-300 focus:ring-2 focus:ring-indigo-300 outline-none [color-scheme:light] dark:[color-scheme:dark]"
                                />
                            </div>
                        </div>

                        <label className="flex items-center gap-2 text-sm font-medium text-gray-600 dark:text-gray-300 cursor-pointer select-none">
                            <input
                                type="checkbox"
                                checked={draft.hideReconciled}
                                onChange={(e) => setDraft({ ...draft, hideReconciled: e.target.checked })}
                                className="rounded text-indigo-600 focus:ring-indigo-500 bg-gray-50 dark:bg-gray-800 border-gray-300 dark:border-gray-600 w-4 h-4 cursor-pointer"
                            />
                            {t('analytics.hideReconciled')}
                        </label>
                    </div>

                    <div className="flex-1">
                        <label className="text-xs text-gray-400 dark:text-gray-500 mb-2 block uppercase tracking-wider font-bold">
                            {t('analytics.excludeCategories')}
                        </label>
                        <div className="flex flex-wrap gap-2 max-h-32 overflow-y-auto custom-scrollbar pr-2 [color-scheme:light] dark:[color-scheme:dark]">
                            {[...categories, { id: 'uncategorized', name: t('analytics.uncategorized'), color: '#94a3b8' }].map(c => {
                                const isExcluded = draft.excludeIds.has(c.id);
                                return (
                                    <button
                                        key={c.id}
                                        type="button"
                                        onClick={() => toggleExcludeCategory(c.id)}
                                        className={`text-xs px-3 py-1.5 rounded-full border transition-all flex items-center gap-1.5 ${
                                            isExcluded
                                                ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800/50 text-red-700 dark:text-red-400'
                                                : 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-600'
                                        }`}
                                    >
                                        {c.name} {isExcluded && <X size={12} />}
                                    </button>
                                );
                            })}
                        </div>
                    </div>
                </div>

                <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-100 dark:border-gray-800">
                    <button
                        type="button"
                        onClick={handleClear}
                        className="text-sm font-medium text-gray-500 dark:text-gray-400 hover:text-red-500 dark:hover:text-red-400 px-4 py-2 rounded-lg transition-colors"
                    >
                        {t('analytics.reset')}
                    </button>
                    <button
                        type="submit"
                        className={`px-6 py-2 rounded-lg font-medium text-sm flex items-center gap-2 transition-all ${isDirty
                            ? 'bg-indigo-600 text-white shadow-md hover:bg-indigo-700'
                            : 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500'
                        }`}
                    >
                        <Search size={16} /> {t('analytics.apply')}
                    </button>
                </div>
            </form>

            {filteredTxns.length === 0 ? (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-12 text-center text-gray-500 dark:text-gray-400">
                    <Filter size={48} className="mx-auto mb-4 opacity-20" />
                    <p className="text-lg font-medium text-gray-700 dark:text-gray-300">{t('analytics.noData')}</p>
                    <p className="text-sm mt-1">{t('analytics.noDataDesc')}</p>
                </div>
            ) : (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <div className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
                            <div className="p-3 bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400 rounded-xl">
                                <TrendingUp size={20} />
                            </div>
                            <div>
                                <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('analytics.periodIncome')}</p>
                                <p className="text-xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(totalInc, baseCurrency)}</p>
                            </div>
                        </div>
                        <div className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
                            <div className="p-3 bg-red-50 dark:bg-red-900/20 text-red-500 dark:text-red-400 rounded-xl">
                                <TrendingDown size={20} />
                            </div>
                            <div>
                                <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('analytics.periodExpenses')}</p>
                                <p className="text-xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(totalExp, baseCurrency)}</p>
                            </div>
                        </div>
                        <div className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
                            <div className={`p-3 rounded-xl ${totalInc >= totalExp ? 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400' : 'bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400'}`}>
                                <Wallet size={20} />
                            </div>
                            <div>
                                <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('analytics.periodNetFlow')}</p>
                                <p className="text-xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(totalInc - totalExp, baseCurrency)}</p>
                            </div>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">

                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 min-h-[450px] flex flex-col">
                            <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                <ArrowRightLeft size={18} className="text-indigo-500 dark:text-indigo-400" />
                                {t('analytics.trendsTitle')}
                            </h2>
                            <div className="flex-1 w-full h-full min-h-[350px]">
                                {trendData.length > 0 ? (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <AreaChart data={trendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                            <defs>
                                                <linearGradient id="colorInc" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
                                                    <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
                                                </linearGradient>
                                                <linearGradient id="colorExp" x1="0" y1="0" x2="0" y2="1">
                                                    <stop offset="5%" stopColor="#f43f5e" stopOpacity={0.3}/>
                                                    <stop offset="95%" stopColor="#f43f5e" stopOpacity={0}/>
                                                </linearGradient>
                                            </defs>
                                            <XAxis
                                                dataKey="month"
                                                tickFormatter={formatMonthAxis}
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                tickMargin={10}
                                                axisLine={false}
                                                tickLine={false}
                                            />
                                            <YAxis
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                axisLine={false}
                                                tickLine={false}
                                                tickFormatter={(val) => `€${val}`}
                                            />
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#374151" opacity={0.2} />
                                            <RechartsTooltip
                                                contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', borderRadius: '8px', color: '#fff' }}
                                                labelFormatter={(label) => formatMonthAxis(label as string)}
                                                formatter={(value: any) => [fmtCurrency(Number(value) || 0, baseCurrency), '']}
                                            />
                                            <Area type="monotone" dataKey="income" name={t('dashboard.cashFlow.income')} stroke="#10b981" strokeWidth={2} fillOpacity={1} fill="url(#colorInc)" />
                                            <Area type="monotone" dataKey="expense" name={t('dashboard.cashFlow.expense')} stroke="#f43f5e" strokeWidth={2} fillOpacity={1} fill="url(#colorExp)" />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                ) : (
                                    <div className="h-full flex items-center justify-center text-gray-400">{t('dashboard.cashFlow.noData')}</div>
                                )}
                            </div>
                        </div>

                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 min-h-[450px] flex flex-col">
                            <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                <BarChartIcon size={18} className="text-indigo-500 dark:text-indigo-400" />
                                {t('analytics.expensesTitle')}
                            </h2>
                            <div className="flex-1 w-full h-full min-h-[350px]">
                                {expenseBarData.length > 0 ? (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <BarChart
                                            data={expenseBarData}
                                            layout="vertical"
                                            margin={{ top: 0, right: 30, left: 20, bottom: 0 }}
                                        >
                                            <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} stroke="#374151" opacity={0.2} />
                                            <XAxis type="number" tick={{ fontSize: 12, fill: '#9ca3af' }} tickFormatter={(val) => `€${val}`} />
                                            <YAxis dataKey="name" type="category" width={110} tick={{ fontSize: 11, fill: '#6b7280' }} />
                                            <RechartsTooltip
                                                cursor={{ fill: 'transparent' }}
                                                contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', borderRadius: '8px', color: '#fff' }}
                                                formatter={(value: any) => [fmtCurrency(Number(value) || 0, baseCurrency), t('analytics.totalSpent')]}
                                            />
                                            <Bar dataKey="value" radius={[0, 4, 4, 0]} barSize={24}>
                                                {expenseBarData.map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Bar>
                                        </BarChart>
                                    </ResponsiveContainer>
                                ) : (
                                    <div className="h-full flex items-center justify-center text-gray-400">{t('dashboard.topCategories.noData')}</div>
                                )}
                            </div>
                        </div>

                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 min-h-[450px] flex flex-col xl:col-span-2">
                            <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                <Briefcase size={18} className="text-indigo-500 dark:text-indigo-400" />
                                {t('analytics.hrTitle')}
                            </h2>
                            <div className="flex-1 w-full h-full min-h-[350px]">
                                {trendData.length > 0 && (payslips as Payslip[]).length > 0 ? (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <BarChart data={trendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#374151" opacity={0.2} />
                                            <XAxis
                                                dataKey="month"
                                                tickFormatter={formatMonthAxis}
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                axisLine={false} tickLine={false}
                                            />
                                            <YAxis
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                axisLine={false} tickLine={false}
                                                tickFormatter={(val) => `€${val}`}
                                            />
                                            <RechartsTooltip
                                                contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', borderRadius: '8px', color: '#fff' }}
                                                labelFormatter={(label) => formatMonthAxis(label as string)}
                                                formatter={(value: any) => [fmtCurrency(Number(value) || 0, baseCurrency), '']}
                                            />
                                            <Bar dataKey="netIncome" stackId="hr" fill="#10b981" name={t('analytics.hrNetIncome')} radius={[0, 0, 4, 4]} barSize={40} />
                                            <Bar dataKey="deductions" stackId="hr" fill="#fbbf24" name={t('analytics.hrDeductions')} radius={[4, 4, 0, 0]} />
                                            <Bar dataKey="expense" fill="#f43f5e" name={t('analytics.hrBankExpenses')} radius={[4, 4, 0, 0]} barSize={40} />
                                        </BarChart>
                                    </ResponsiveContainer>
                                ) : (
                                    <div className="h-full flex items-center justify-center text-gray-400">
                                        {t('analytics.hrEmpty')}
                                    </div>
                                )}
                            </div>
                        </div>

                        {/* --- Category Trend Chart --- */}
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 min-h-[450px] flex flex-col xl:col-span-2">
                            <div className="flex items-center justify-between mb-6">
                                <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-2">
                                    <Activity size={18} className="text-indigo-500 dark:text-indigo-400" />
                                    {t('analytics.categoryTrendTitle')}
                                </h2>
                                <select
                                    value={selectedCategoryTrendId}
                                    onChange={(e) => setSelectedCategoryTrendId(e.target.value)}
                                    className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-2 text-gray-700 dark:text-gray-300 focus:ring-2 focus:ring-indigo-300 outline-none max-w-xs"
                                >
                                    <option value="" disabled>{t('analytics.selectCategory')}</option>
                                    {categories.filter(c => !c.deleted_at || c.id === selectedCategoryTrendId).map(c => (
                                        <option key={c.id} value={c.id}>{c.name}</option>
                                    ))}
                                    <option value="uncategorized">{t('analytics.uncategorized')}</option>
                                </select>
                            </div>
                            <div className="flex-1 w-full h-full min-h-[350px]">
                                {!selectedCategoryTrendId ? (
                                    <div className="h-full flex items-center justify-center text-gray-400">
                                        {t('analytics.selectCategoryPrompt')}
                                    </div>
                                ) : categoryTrendData.length > 0 ? (
                                    <ResponsiveContainer width="100%" height="100%">
                                        <LineChart data={categoryTrendData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#374151" opacity={0.2} />
                                            <XAxis
                                                dataKey="month"
                                                tickFormatter={formatMonthAxis}
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                axisLine={false} tickLine={false}
                                            />
                                            <YAxis
                                                tick={{ fontSize: 12, fill: '#9ca3af' }}
                                                axisLine={false} tickLine={false}
                                                tickFormatter={(val) => `€${val}`}
                                            />
                                            <RechartsTooltip
                                                contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', borderRadius: '8px', color: '#fff' }}
                                                labelFormatter={(label) => formatMonthAxis(label as string)}
                                                formatter={(value: any) => [fmtCurrency(Number(value) || 0, baseCurrency), t('analytics.totalSpent')]}
                                            />
                                            <Line
                                                type="monotone"
                                                dataKey="amount"
                                                stroke="#8b5cf6"
                                                strokeWidth={3}
                                                dot={{ r: 4, fill: '#8b5cf6', strokeWidth: 0 }}
                                                activeDot={{ r: 6, strokeWidth: 0 }}
                                            />
                                        </LineChart>
                                    </ResponsiveContainer>
                                ) : (
                                    <div className="h-full flex items-center justify-center text-gray-400">
                                        {t('analytics.noData')}
                                    </div>
                                )}
                            </div>
                        </div>

                    </div>
                </>
            )}
        </div>
    );
}