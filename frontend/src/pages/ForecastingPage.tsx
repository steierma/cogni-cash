import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import PlannedTransactionsList from '../components/forecasting/PlannedTransactionsList';
import ForecastTable, { type ForecastColKey } from '../components/forecasting/ForecastTable';
import {
    BarChart3,
    TrendingUp,
    TrendingDown,
    Calendar,
    AlertCircle,
    Info,
    Search,
    Columns,
    Check,
    X
} from 'lucide-react';
import {
    Area,
    AreaChart,
    CartesianGrid,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
    ReferenceLine
} from 'recharts';
import { forecastingService } from '../api/services/forecastingService';
import { categoryService } from '../api/services/categoryService';
import type { CashFlowForecast, PatternExclusion } from "../api/types/transaction";
import type { Category } from "../api/types/category";
import { fmtCurrency, fmtDate } from '../utils/formatters';

interface CustomTooltipProps {
    active?: boolean;
    payload?: any[];
    label?: string;
    t: (key: string, options?: any) => string;
}

const CustomTooltip = ({ active, payload, label, t }: CustomTooltipProps) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 p-3 shadow-lg rounded-lg text-sm">
                <p className="font-bold text-gray-900 dark:text-gray-100 mb-1">{fmtDate(label ?? '')}</p>
                <p className="text-indigo-600 dark:text-indigo-400 font-medium">
                    {t('forecasting.projectedBalance')}: {fmtCurrency(payload[0].value, 'EUR')}
                </p>
                {payload[0].payload.income > 0 && (
                    <p className="text-emerald-600 dark:text-emerald-400 text-xs">
                        +{fmtCurrency(payload[0].payload.income, 'EUR')} {t('dashboard.cashFlow.income')}
                    </p>
                )}
                {payload[0].payload.expense > 0 && (
                    <p className="text-rose-600 dark:text-rose-400 text-xs">
                        -{fmtCurrency(payload[0].payload.expense, 'EUR')} {t('dashboard.cashFlow.expense')}
                    </p>
                )}
            </div>
        );
    }
    return null;
};

export default function ForecastingPage() {
    const { t, i18n } = useTranslation();
    const queryClient = useQueryClient();
    const [searchParams, setSearchParams] = useSearchParams();

    const [range, setRange] = useState<'30' | '60' | '90' | '180' | '365'>(() => {
        const r = searchParams.get('range');
        if (['30', '60', '90', '180', '365'].includes(r || '')) return r as any;
        return '30';
    });
    const [activeTab, setActiveTab] = useState<'forecast' | 'blocked'>(() => {
        const tab = searchParams.get('tab');
        if (tab === 'forecast' || tab === 'blocked') return tab;
        return 'forecast';
    });
    const [search, setSearch] = useState(searchParams.get('search') || '');
    const [showColMenu, setShowColMenu] = useState(false);

    // Update URL when state changes
    useEffect(() => {
        const next = new URLSearchParams();
        if (range !== '30') next.set('range', range);
        if (activeTab !== 'forecast') next.set('tab', activeTab);
        if (search) next.set('search', search);

        // Only update if changed
        const currentStr = searchParams.toString();
        const nextStr = next.toString();
        if (currentStr !== nextStr) {
            setSearchParams(next, { replace: true });
        }
    }, [range, activeTab, search, setSearchParams, searchParams]);

    // Columns Configuration with Local Storage
    const [visibleCols, setVisibleCols] = useState<Record<ForecastColKey, boolean>>(() => {
        const saved = localStorage.getItem('forecast_visible_cols');
        if (saved) {
            try { return JSON.parse(saved); } catch { /* fallback */ }
        }
        return { date: true, description: true, category: true, probability: true, amount: true };
    });

    useEffect(() => {
        localStorage.setItem('forecast_visible_cols', JSON.stringify(visibleCols));
    }, [visibleCols]);

    const toggleColumn = (key: ForecastColKey) => {
        setVisibleCols(prev => ({ ...prev, [key]: !prev[key] }));
    };

    // Memoize the target date string to prevent unnecessary query invalidations on every render
    const toDateStr = useMemo(() => {
        const d = new Date();
        d.setDate(d.getDate() + parseInt(range));
        return d.toISOString().split('T')[0];
    }, [range]);

    const forecastQuery = useQuery<CashFlowForecast>({
        queryKey: ['forecast', range, toDateStr],
        queryFn: () => {
            const today = new Date().toISOString().split('T')[0];
            return forecastingService.fetchForecast(today, toDateStr);
        },
        staleTime: 5 * 60 * 1000,
        refetchInterval: 60000,
    });

    const blockedPatternsQuery = useQuery<PatternExclusion[]>({
        queryKey: ['blocked-patterns'],
        queryFn: () => forecastingService.fetchPatternExclusions(),
    });

    const categoriesQuery = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: () => categoryService.fetchCategories(),
    });

    const excludeMutation = useMutation({
        mutationFn: (id: string) => forecastingService.excludeProjection(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
        }
    });

    const includeMutation = useMutation({
        mutationFn: (id: string) => forecastingService.includeProjection(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
        }
    });

    const includePatternMutation = useMutation({
        mutationFn: (term: string) => forecastingService.includePattern(term),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['forecast'] });
            queryClient.invalidateQueries({ queryKey: ['blocked-patterns'] });
        }
    });

    const forecast = forecastQuery.data;
    const isLoading = forecastQuery.isLoading;
    const isError = forecastQuery.isError;

    const chartData = forecast?.time_series ?? [];
    
    // Stable sorted and filtered predictions
    const filteredPredictions = useMemo(() => {
        let p = [...(forecast?.predictions ?? [])];
        
        if (search) {
            const s = search.toLowerCase();
            p = p.filter(tx => 
                tx.description.toLowerCase().includes(s) || 
                tx.counterparty_name?.toLowerCase().includes(s) ||
                (tx.category_id && categoriesQuery.data?.find(c => c.id === tx.category_id)?.name.toLowerCase().includes(s))
            );
        }

        return p.sort((a, b) => new Date(a.booking_date).getTime() - new Date(b.booking_date).getTime());
    }, [forecast?.predictions, search, categoriesQuery.data]);

    const currentBalance = forecast?.current_balance ?? 0;
    const projectedBalance = chartData.length > 0 ? chartData[chartData.length - 1].expected_balance : currentBalance;
    const balanceDiff = projectedBalance - currentBalance;

    const totalExpectedIncome = (forecast?.predictions ?? []).filter(p => !p.skip_forecasting && p.amount > 0).reduce((sum, p) => sum + p.amount, 0);
    const totalExpectedExpense = (forecast?.predictions ?? []).filter(p => !p.skip_forecasting && p.amount < 0).reduce((sum, p) => sum + p.amount, 0);

    const formatXAxis = (tickItem: string) => {
        const date = new Date(tickItem);
        const options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
        if (date.getMonth() === 0 && date.getDate() <= 7) {
            options.year = '2-digit';
        }
        return date.toLocaleDateString(i18n.language, options);
    };

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <BarChart3 className="text-indigo-600 dark:text-indigo-400" /> {t('forecasting.title')}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {t('forecasting.subtitle')}
                    </p>
                </div>

                <div className="flex items-center gap-2 bg-gray-100 dark:bg-gray-800 p-1 rounded-xl">
                    <button
                        onClick={() => setRange('30')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '30'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                            }`}
                    >
                        {t('forecasting.month1')}
                    </button>
                    <button
                        onClick={() => setRange('60')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '60'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                            }`}
                    >
                        {t('forecasting.months2')}
                    </button>
                    <button
                        onClick={() => setRange('90')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '90'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                            }`}
                    >
                        {t('forecasting.months3')}
                    </button>
                    <button
                        onClick={() => setRange('180')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '180'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                            }`}
                    >
                        {t('forecasting.months6')}
                    </button>
                    <button
                        onClick={() => setRange('365')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '365'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                            }`}
                    >
                        {t('forecasting.months12')}
                    </button>
                </div>
            </div>

            {isError && (
                <div className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm">
                    <AlertCircle size={16} />
                    {t('dashboard.errorLoad')}
                </div>
            )}

            {/* KPI Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                    <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('forecasting.currentBalance')}</p>
                    <p className="text-3xl font-black text-gray-900 dark:text-gray-100">
                        {isLoading ? '...' : fmtCurrency(currentBalance, 'EUR')}
                    </p>
                </div>

                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 relative overflow-hidden">
                    <div className="relative z-10">
                        <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('forecasting.projectedBalance')}</p>
                        <div className="flex items-end gap-3">
                            <p className="text-3xl font-black text-gray-900 dark:text-gray-100">
                                {isLoading ? '...' : fmtCurrency(projectedBalance, 'EUR')}
                            </p>
                            {!isLoading && (
                                <span className={`text-sm font-bold pb-1 flex items-center gap-0.5 ${balanceDiff >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}`}>
                                    {balanceDiff >= 0 ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
                                    {fmtCurrency(Math.abs(balanceDiff), 'EUR')}
                                </span>
                            )}
                        </div>
                    </div>
                    <div className="absolute right-0 top-0 bottom-0 w-32 bg-indigo-50/30 dark:bg-indigo-900/10 -mr-8 rotate-12 pointer-events-none" />
                </div>

                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                    <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('forecasting.incomeForecast')} / {t('forecasting.expenseForecast')}</p>
                    <div className="flex items-center gap-4 mt-1">
                        <div className="flex items-center gap-1.5 text-emerald-600 dark:text-emerald-400 font-bold">
                            <TrendingUp size={18} />
                            {isLoading ? '...' : fmtCurrency(totalExpectedIncome, 'EUR')}
                        </div>
                        <div className="h-4 w-px bg-gray-200 dark:bg-gray-800" />
                        <div className="flex items-center gap-1.5 text-rose-600 dark:text-rose-400 font-bold">
                            <TrendingDown size={18} />
                            {isLoading ? '...' : fmtCurrency(totalExpectedExpense, 'EUR')}
                        </div>
                    </div>
                </div>
            </div>

            {/* Main Chart */}
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                <div className="flex items-center justify-between mb-8">
                    <div className="flex items-center gap-2">
                        <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100">{t('forecasting.forecastChart')}</h2>
                        <div className="group relative">
                            <Info size={14} className="text-gray-400 cursor-help" />
                            <div className="absolute left-1/2 -translate-x-1/2 bottom-full mb-2 w-64 p-3 bg-gray-900 text-white text-[10px] rounded-lg opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50 shadow-xl leading-relaxed">
                                {t('forecasting.subtitle')}
                            </div>
                        </div>
                    </div>
                    <div className="flex items-center gap-4 text-xs font-bold uppercase tracking-tighter">
                        <div className="flex items-center gap-1.5 text-indigo-500">
                            <div className="w-3 h-3 rounded-full bg-indigo-500" />
                            {t('forecasting.balanceTrend')}
                        </div>
                    </div>
                </div>

                <div className="h-[400px] w-full">
                    {isLoading ? (
                        <div className="h-full w-full flex items-center justify-center">
                            <div className="flex flex-col items-center gap-3">
                                <div className="w-10 h-10 border-4 border-indigo-500 border-t-transparent rounded-full animate-spin" />
                                <p className="text-sm text-gray-400">{t('common.loading')}</p>
                            </div>
                        </div>
                    ) : (
                        <ResponsiveContainer width="100%" height="100%">
                            <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                                <defs>
                                    <linearGradient id="colorBalance" x1="0" y1="0" x2="0" y2="1">
                                        <stop offset="5%" stopColor="#6366f1" stopOpacity={0.2} />
                                        <stop offset="95%" stopColor="#6366f1" stopOpacity={0} />
                                    </linearGradient>
                                </defs>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e5e7eb" opacity={0.5} />
                                <XAxis
                                    dataKey="date"
                                    axisLine={false}
                                    tickLine={false}
                                    tick={{ fontSize: 10, fill: '#9ca3af' }}
                                    tickFormatter={formatXAxis}
                                    minTickGap={30}
                                />
                                <YAxis
                                    axisLine={false}
                                    tickLine={false}
                                    tick={{ fontSize: 10, fill: '#9ca3af' }}
                                    tickFormatter={(val) => `€${val}`}
                                />
                                <Tooltip content={<CustomTooltip t={t} />} />
                                <ReferenceLine y={currentBalance} stroke="#94a3b8" strokeDasharray="3 3" label={{ position: 'right', value: 'Today', fill: '#94a3b8', fontSize: 10 }} />
                                <Area
                                    type="monotone"
                                    dataKey="expected_balance"
                                    stroke="#6366f1"
                                    strokeWidth={3}
                                    fillOpacity={1}
                                    fill="url(#colorBalance)"
                                    animationDuration={1500}
                                />
                            </AreaChart>
                        </ResponsiveContainer>
                    )}
                </div>
            </div>

            {/* Planned Transactions List */}
            <PlannedTransactionsList />

            {/* Predictions List */}
            <div>
                <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-4 mb-4 border-b border-gray-200 dark:border-gray-800">
                    <div className="flex gap-6">
                        <button
                            onClick={() => setActiveTab('forecast')}
                            className={`pb-4 text-sm font-bold border-b-2 transition-colors ${
                                activeTab === 'forecast'
                                    ? 'border-indigo-500 text-indigo-600 dark:text-indigo-400'
                                    : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
                            }`}
                        >
                            <span className="flex items-center gap-2">
                                <Calendar size={16} className={activeTab === 'forecast' ? 'text-indigo-500' : 'text-gray-400'} />
                                {t('forecasting.upcomingTransactions', { days: range })}
                            </span>
                        </button>
                        <button
                            onClick={() => setActiveTab('blocked')}
                            className={`pb-4 text-sm font-bold border-b-2 transition-colors ${
                                activeTab === 'blocked'
                                    ? 'border-rose-500 text-rose-600 dark:text-rose-400'
                                    : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
                            }`}
                        >
                            <span className="flex items-center gap-2">
                                <AlertCircle size={16} className={activeTab === 'blocked' ? 'text-rose-500' : 'text-gray-400'} />
                                {t('forecasting.blockedPatterns', 'Blocked Patterns')}
                                {blockedPatternsQuery.data && blockedPatternsQuery.data.length > 0 && (
                                    <span className="bg-rose-100 dark:bg-rose-900/30 text-rose-600 dark:text-rose-400 py-0.5 px-2 rounded-full text-xs">
                                        {blockedPatternsQuery.data.length}
                                    </span>
                                )}
                            </span>
                        </button>
                    </div>

                    {activeTab === 'forecast' && (
                        <div className="flex items-center gap-2 pb-2">
                            <div className="relative group">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 group-focus-within:text-indigo-500 transition-colors" size={14} />
                                <input
                                    type="text"
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                    placeholder={t('common.search')}
                                    className="pl-9 pr-8 py-1.5 text-xs bg-gray-100 dark:bg-gray-800 border-transparent focus:bg-white dark:focus:bg-gray-900 focus:ring-2 focus:ring-indigo-500/20 rounded-lg w-full sm:w-48 transition-all"
                                />
                                {search && (
                                    <button
                                        onClick={() => setSearch('')}
                                        className="absolute right-2 top-1/2 -translate-y-1/2 p-0.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                                    >
                                        <X size={12} />
                                    </button>
                                )}
                            </div>

                            <div className="relative">
                                <button
                                    onClick={() => setShowColMenu(!showColMenu)}
                                    className={`p-1.5 rounded-lg border transition-all ${showColMenu 
                                        ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-200 dark:border-indigo-800 text-indigo-600 dark:text-indigo-400' 
                                        : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}
                                    title={t('transactions.columns', 'Columns')}
                                >
                                    <Columns size={14} />
                                </button>
                                
                                {showColMenu && (
                                    <>
                                        <div className="fixed inset-0 z-10" onClick={() => setShowColMenu(false)} />
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
                                                    onClick={() => toggleColumn(key as ForecastColKey)}
                                                    className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
                                                >
                                                    {label} {visibleCols[key as ForecastColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
                                                </button>
                                            ))}
                                        </div>
                                    </>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                {activeTab === 'forecast' && (
                    <ForecastTable 
                        predictions={filteredPredictions}
                        isLoading={isLoading}
                        visibleCols={visibleCols}
                        onInclude={(id) => includeMutation.mutate(id)}
                        onExclude={(id) => excludeMutation.mutate(id)}
                    />
                )}

                {activeTab === 'blocked' && (
                    <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                        {blockedPatternsQuery.isLoading ? (
                            <div className="p-8 flex flex-col gap-4">
                                {[1, 2].map(i => (
                                    <div key={i} className="h-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl animate-pulse" />
                                ))}
                            </div>
                        ) : !blockedPatternsQuery.data || blockedPatternsQuery.data.length === 0 ? (
                            <div className="p-12 flex flex-col items-center justify-center text-center">
                                <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-full mb-4">
                                    <AlertCircle size={32} className="text-gray-300 dark:text-gray-600" />
                                </div>
                                <h3 className="text-gray-900 dark:text-gray-100 font-semibold">{t('forecasting.noBlockedPatterns', 'No Blocked Patterns')}</h3>
                                <p className="text-sm text-gray-500 dark:text-gray-400 mt-1 max-w-xs">
                                    {t('forecasting.noBlockedPatternsDesc', 'You have not muted any recurring patterns. You can mute patterns directly from the Transactions page.')}
                                </p>
                            </div>
                        ) : (
                            <div className="overflow-x-auto">
                                <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50">
                                    <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 font-bold tracking-wider">
                                        <tr>
                                            <th className="px-6 py-4 text-left">{t('forecasting.patternMatchTerm', 'Match Term')}</th>
                                            <th className="px-6 py-4 text-left">{t('common.date', 'Date Added')}</th>
                                            <th className="px-6 py-4 text-right">{t('common.actions', 'Actions')}</th>
                                        </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                                        {blockedPatternsQuery.data.map((pattern) => (
                                            <tr key={pattern.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="flex items-center gap-2 text-sm font-medium text-gray-900 dark:text-gray-100">
                                                        <AlertCircle size={16} className="text-rose-500" />
                                                        "{pattern.match_term}"
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                                                    {fmtDate(pattern.created_at)}
                                                </td>
                                                <td className="px-6 py-4 text-right">
                                                    <button
                                                        onClick={() => includePatternMutation.mutate(pattern.match_term)}
                                                        className="px-3 py-1.5 text-xs font-bold bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors"
                                                    >
                                                        {t('common.restore', 'Restore')}
                                                    </button>
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                        <div className="bg-gray-50 dark:bg-gray-800/50 px-6 py-4">
                            <p className="text-xs text-gray-400 dark:text-gray-500 flex items-center gap-1.5">
                                <Info size={14} /> {t('forecasting.blockedPatternsInfo', 'Muted patterns are completely ignored by the forecasting engine.')}
                            </p>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
