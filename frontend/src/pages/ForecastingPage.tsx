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
import { settingsService } from '../api/services/settingsService';
import { useEffectiveSettings } from '../hooks/useEffectiveSettings';
import { getNamespacedKey } from '../api/utils/settingsHelper';
import type { CashFlowForecast } from "../api/types/transaction";
import type { Category } from "../api/types/category";
import { fmtCurrency, fmtDate, getLocalISODate } from '../utils/formatters';

interface CustomTooltipProps {
    active?: boolean;
    payload?: {
        value: number;
        payload: {
            income: number;
            expense: number;
        };
    }[];
    label?: string;
    t: (key: string, options?: Record<string, unknown>) => string;
    baseCurrency: string;
}

const CustomTooltip = ({ active, payload, label, t, baseCurrency }: CustomTooltipProps) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 p-3 shadow-lg rounded-lg text-sm">
                <p className="font-bold text-gray-900 dark:text-gray-100 mb-1">{fmtDate(label ?? '')}</p>
                <p className="text-indigo-600 dark:text-indigo-400 font-medium">
                    {t('forecasting.projectedBalance')}: {fmtCurrency(payload[0].value, baseCurrency)}
                </p>
                {payload[0].payload.income > 0 && (
                    <p className="text-emerald-600 dark:text-emerald-400 text-xs">
                        +{fmtCurrency(payload[0].payload.income, baseCurrency)} {t('dashboard.cashFlow.income')}
                    </p>
                )}
                {payload[0].payload.expense > 0 && (
                    <p className="text-rose-600 dark:text-rose-400 text-xs">
                        -{fmtCurrency(payload[0].payload.expense, baseCurrency)} {t('dashboard.cashFlow.expense')}
                    </p>
                )}
            </div>
        );
    }
    return null;
};

export default function ForecastingPage() {
    const { t, i18n } = useTranslation();
    const [searchParams, setSearchParams] = useSearchParams();
    const qc = useQueryClient();

    const [range, setRange] = useState<'30' | '60' | '90' | '180' | '365' | '730'>(() => {
        const r = searchParams.get('range');
        const validRanges = ['30', '60', '90', '180', '365', '730'];
        if (r && validRanges.includes(r)) return r as '30' | '60' | '90' | '180' | '365' | '730';
        return '90';
    });
    const [activeTab, setActiveTab] = useState<'forecast' | 'blocked'>(() => {
        const tab = searchParams.get('tab');
        if (tab === 'forecast' || tab === 'blocked') return tab;
        return 'forecast';
    });
    const [search, setSearch] = useState(searchParams.get('search') || '');
    const [showColMenu, setShowColMenu] = useState(false);

    const { data: settings } = useEffectiveSettings();

    const updateSettingsMut = useMutation({
        mutationFn: (data: Record<string, string>) => settingsService.updateSettings(data),
        onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] })
    });

    // Update URL when state changes
    useEffect(() => {
        const next = new URLSearchParams();
        if (range !== '90') next.set('range', range);
        if (activeTab !== 'forecast') next.set('tab', activeTab);
        if (search) next.set('search', search);

        // Only update if changed
        const currentStr = searchParams.toString();
        const nextStr = next.toString();
        if (currentStr !== nextStr) {
            setSearchParams(next, { replace: true });
        }
    }, [range, activeTab, search, setSearchParams, searchParams]);

    // Columns Configuration
    const [visibleCols, setVisibleCols] = useState<Record<ForecastColKey, boolean>>({
        date: true, description: true, category: true, probability: true, amount: true
    });

    useEffect(() => {
        if (settings?.forecast_visible_cols) {
            try {
                setVisibleCols(JSON.parse(settings.forecast_visible_cols));
            } catch (e) {
                console.error("Failed to parse forecast column settings", e);
            }
        }
    }, [settings?.forecast_visible_cols]);

    const toggleColumn = (key: ForecastColKey) => {
        setVisibleCols(prev => {
            const next = { ...prev, [key]: !prev[key] };
            const nsKey = getNamespacedKey('forecast_visible_cols', true);
            updateSettingsMut.mutate({ [nsKey]: JSON.stringify(next) });
            return next;
        });
    };

    // Memoize the target date string to prevent unnecessary query invalidations on every render
    const toDateStr = useMemo(() => {
        const d = new Date();
        d.setDate(d.getDate() + parseInt(range));
        return getLocalISODate(d);
    }, [range]);

    const forecastQuery = useQuery<CashFlowForecast>({
        queryKey: ['forecast', range, toDateStr],
        queryFn: () => {
            const today = getLocalISODate();
            return forecastingService.fetchForecast(today, toDateStr);
        },
        staleTime: 5 * 60 * 1000,
        refetchInterval: 60000,
    });

    const categoriesQuery = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: () => categoryService.fetchCategories(),
    });

    const baseCurrency = settings?.['BASE_DISPLAY_CURRENCY'] || 'EUR';

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

    const totalExpectedIncome = (forecast?.predictions ?? []).filter(p => p.amount > 0).reduce((sum, p) => sum + p.amount, 0);
    const totalExpectedExpense = (forecast?.predictions ?? []).filter(p => p.amount < 0).reduce((sum, p) => sum + p.amount, 0);

    const formatXAxis = (tickItem: string) => {
        const date = new Date(tickItem);
        const options: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric' };
        if (date.getMonth() === 0 && date.getDate() <= 7) {
            options.year = '2-digit';
        }
        return date.toLocaleDateString(i18n.language, options);
    };

    return (
        <div className="space-y-6 pb-20 animate-in fade-in duration-300">
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

                {/* Mobile Range Dropdown */}
                <div className="sm:hidden w-full">
                    <select
                        value={range}
                        onChange={(e) => setRange(e.target.value as any)}
                        className="w-full px-4 py-2.5 bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 border-transparent focus:ring-2 focus:ring-indigo-500/20 rounded-xl text-sm font-bold"
                    >
                        <option value="30">{t('forecasting.month1')}</option>
                        <option value="60">{t('forecasting.months2')}</option>
                        <option value="90">{t('forecasting.months3')}</option>
                        <option value="180">{t('forecasting.months6')}</option>
                        <option value="365">{t('forecasting.months12')}</option>
                        <option value="730">{t('forecasting.months24')}</option>
                    </select>
                </div>

                {/* Desktop Range Buttons */}
                <div className="hidden sm:flex items-center gap-2 bg-gray-100 dark:bg-gray-800 p-1 rounded-xl">
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
                    <button
                        onClick={() => setRange('730')}
                        className={`px-4 py-1.5 text-xs font-bold rounded-lg transition-all ${range === '730'
                            ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-300 shadow-sm'
                            : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
                        }`}
                    >
                        {t('forecasting.months24')}
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
                        {isLoading ? '...' : fmtCurrency(currentBalance, baseCurrency)}
                    </p>
                </div>

                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 relative overflow-hidden">
                    <div className="relative z-10">
                        <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('forecasting.projectedBalance')}</p>
                        <div className="flex items-end gap-3">
                            <p className="text-3xl font-black text-gray-900 dark:text-gray-100">
                                {isLoading ? '...' : fmtCurrency(projectedBalance, baseCurrency)}
                            </p>
                            {!isLoading && (
                                <span className={`text-sm font-bold pb-1 flex items-center gap-0.5 ${balanceDiff >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'}`}>
                                    {balanceDiff >= 0 ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
                                    {fmtCurrency(Math.abs(balanceDiff), baseCurrency)}
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
                            {isLoading ? '...' : fmtCurrency(totalExpectedIncome, baseCurrency)}
                        </div>
                        <div className="h-4 w-px bg-gray-200 dark:bg-gray-800" />
                        <div className="flex items-center gap-1.5 text-rose-600 dark:text-rose-400 font-bold">
                            <TrendingDown size={18} />
                            {isLoading ? '...' : fmtCurrency(totalExpectedExpense, baseCurrency)}
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
                                    tickFormatter={(val) => `${baseCurrency === 'EUR' ? '€' : baseCurrency}${val}`}
                                />
                                <Tooltip content={<CustomTooltip t={t} baseCurrency={baseCurrency} />} />
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
                        </div>
                    )}
                </div>

                {activeTab === 'forecast' && (
                    <ForecastTable
                        predictions={filteredPredictions}
                        isLoading={isLoading}
                        visibleCols={visibleCols}
                        showColMenu={showColMenu}
                        onToggleColMenu={() => setShowColMenu(!showColMenu)}
                        onToggleColumn={toggleColumn}
                    />
                )}
            </div>
        </div>
    );
}