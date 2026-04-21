import {useState} from 'react';
import {useQuery} from '@tanstack/react-query';
import {Link} from 'react-router-dom';
import {useTranslation} from 'react-i18next';
import {
    AlertCircle, ArrowRight, Calendar, ChevronRight, Landmark,
    TrendingDown, TrendingUp, Wallet, BarChart3, PieChart,
    PiggyBank, LayoutDashboard, Unlink, Zap, RefreshCcw
} from 'lucide-react';
import { bankService } from '../api/services/bankService';
import { transactionService } from '../api/services/transactionService';
import { payslipService } from '../api/services/payslipService';
import { forecastingService } from '../api/services/forecastingService';
import { subscriptionService } from '../api/services/subscriptionService';
import { settingsService } from '../api/services/settingsService';
import type { BankStatementSummary } from "../api/types/bank";
import type { Transaction, TransactionAnalytics, CashFlowForecast } from "../api/types/transaction";
import type { Payslip } from "../api/types/payslip";
import {fmtCurrency, fmtDate} from '../utils/formatters';
import CategoryBadge from '../components/CategoryBadge';

// ── helpers ──────────────────────────────────────────────────────────────────

function formatChartMonth(yyyyMM: string, locale: string) {
    try {
        const [year, month] = yyyyMM.split('-');
        const date = new Date(parseInt(year), parseInt(month) - 1);
        return date.toLocaleDateString(locale, {month: 'short', year: '2-digit'});
    } catch {
        return yyyyMM;
    }
}

function getGreeting(t: any) {
    const hour = new Date().getHours();
    if (hour < 12) return t('dashboard.greeting.morning');
    if (hour < 18) return t('dashboard.greeting.afternoon');
    return t('dashboard.greeting.evening');
}

// ── sub-components ───────────────────────────────────────────────────────────

function TrueSavingsCard({analytics, payslips, baseCurrency}: { analytics?: TransactionAnalytics, payslips?: Payslip[], baseCurrency: string }) {
    const {t} = useTranslation();
    if (!analytics || !payslips || payslips.length === 0) return null;

    const latestPayslip = payslips[0];
    const trueNetIncome = latestPayslip.base_net_pay || latestPayslip.net_pay;

    // Match the payslip's period to the analytics time_series data
    const payslipPeriod = `${latestPayslip.period_year}-${String(latestPayslip.period_month_num).padStart(2, '0')}`;
    const currentMonthData = analytics.time_series?.find(ts => ts.date === payslipPeriod);
    const currentMonthExpenses = currentMonthData ? currentMonthData.expense : 0;

    const trueSavings = trueNetIncome - currentMonthExpenses;
    const savingsRate = trueNetIncome > 0 ? ((trueSavings / trueNetIncome) * 100).toFixed(1) : '0.0';

    return (
        <div
            className="bg-gradient-to-br from-indigo-50 to-white dark:from-indigo-950/30 dark:to-gray-900 p-6 rounded-2xl border border-indigo-100 dark:border-indigo-900/50 shadow-sm mb-6">
            <h3 className="text-sm font-bold uppercase tracking-wider text-indigo-800 dark:text-indigo-300 mb-4">
                {t('dashboard.trueSavings.title')}
            </h3>

            <div className="flex items-end gap-4">
                <p className="text-4xl font-black text-gray-900 dark:text-gray-100">
                    {fmtCurrency(trueSavings, baseCurrency)}
                </p>
                <div
                    className={`flex items-center gap-1 text-sm font-bold pb-1 ${trueSavings >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'}`}>
                    {trueSavings >= 0 ? <PiggyBank size={18}/> : <TrendingDown size={18}/>}
                    {savingsRate}% {t('dashboard.trueSavings.saved')}
                </div>
            </div>

            <div className="mt-5 h-2.5 w-full bg-gray-200 dark:bg-gray-800 rounded-full overflow-hidden flex">
                <div
                    className="bg-red-500 h-full transition-all duration-1000"
                    style={{width: `${Math.min((currentMonthExpenses / trueNetIncome) * 100, 100)}%`}}
                    title={t('dashboard.cashFlow.expense')}
                />
                <div
                    className="bg-emerald-500 h-full transition-all duration-1000"
                    style={{width: `${Math.max(100 - (currentMonthExpenses / trueNetIncome) * 100, 0)}%`}}
                    title={t('dashboard.trueSavings.saved')}
                />
            </div>
            <div className="flex justify-between text-xs font-medium text-gray-500 dark:text-gray-400 mt-2">
                <span>{t('dashboard.trueSavings.spent')} {fmtCurrency(currentMonthExpenses, baseCurrency)}</span>
                <span>{t('dashboard.trueSavings.actualNet')} {fmtCurrency(trueNetIncome, baseCurrency)}</span>
            </div>
        </div>
    );
}

function KpiCard({label, value, sub, color, icon}: {
    label: string;
    value: string;
    sub?: string;
    color: 'indigo' | 'green' | 'red' | 'gray' | 'blue';
    icon: React.ReactNode
}) {
    const bg: Record<string, string> = {
        indigo: 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400',
        green: 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400',
        red: 'bg-red-50 dark:bg-red-900/20 text-red-500 dark:text-red-400',
        gray: 'bg-gray-100 dark:bg-gray-800/50 text-gray-500 dark:text-gray-400',
        blue: 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400',
    };
    return (
        <div
            className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-5 flex items-start gap-4 transition-transform hover:-translate-y-0.5 duration-200">
            <div className={`p-3 rounded-xl ${bg[color]}`}>{icon}</div>
            <div className="min-w-0">
                <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{label}</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 truncate">{value}</p>
                {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{sub}</p>}
            </div>
        </div>
    );
}

function StatementCard({stmt}: { stmt: BankStatementSummary }) {
    const {t} = useTranslation();
    return (
        <Link
            to={`/transactions?statement=${stmt.id}`}
            className="flex-none w-72 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-5 hover:border-indigo-300 dark:hover:border-indigo-500 hover:shadow-md transition-all group block"
        >
            <div className="flex justify-between items-start mb-3">
                <p className="text-xs text-gray-400 dark:text-gray-500 font-mono truncate mr-2" title={stmt.iban}>
                    {stmt.iban}
                </p>
                <span
                    className="flex items-center gap-1 text-[10px] font-bold bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 px-2 py-0.5 rounded uppercase whitespace-nowrap">
                    <Calendar size={10}/>
                    {stmt.period_label}
                </span>
            </div>

            <div className="flex items-end justify-between mt-4">
                <div>
                    <p className="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-tight">{t('dashboard.statementCard.closingBalance')}</p>
                    <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
                        {fmtCurrency(stmt.new_balance, stmt.currency)}
                    </p>
                </div>
                <span
                    className="text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 px-2 py-1 rounded-lg border border-gray-100 dark:border-gray-700">
                    {stmt.transaction_count} {t('dashboard.statementCard.txns')}
                </span>
            </div>
        </Link>
    );
}

function SubscriptionSummaryCard({baseCurrency}: { baseCurrency: string }) {
    const {t} = useTranslation();
    const {data: subscriptions = [], isLoading} = useQuery({
        queryKey: ['subscriptions'],
        queryFn: subscriptionService.fetchSubscriptions,
    });

    const activeSubs = subscriptions.filter(s => s.status === 'active');
    const totalMonthly = activeSubs.reduce((acc, s) => {
        const amount = s.amount;
        if (s.billing_cycle === 'yearly') return acc + (amount / 12);
        return acc + (amount / (s.billing_interval || 1));
    }, 0);

    if (isLoading || subscriptions.length === 0) return null;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2 text-gray-800 dark:text-gray-200">
                    <RefreshCcw size={18} className="text-indigo-500 dark:text-indigo-400"/>
                    <h2 className="text-base font-semibold">{t('layout.subscriptions')}</h2>
                </div>
                <Link
                    to="/subscriptions"
                    className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
                >
                    {t('dashboard.bankStatements.viewAll')} <ChevronRight size={14}/>
                </Link>
            </div>

            <div className="flex items-end justify-between">
                <div>
                    <p className="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-tight">{t('subscriptions.monthlySpend')}</p>
                    <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
                        {fmtCurrency(totalMonthly, baseCurrency)}
                    </p>
                </div>
                <div className="text-right">
                    <p className="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-tight">Active Services</p>
                    <p className="text-lg font-bold text-indigo-600 dark:text-indigo-400">{activeSubs.length}</p>
                </div>
            </div>
        </div>
    );
}

function ForecastingWidget() {
    const {t} = useTranslation();
    const {data: forecast, isLoading} = useQuery<CashFlowForecast>({
        queryKey: ['forecast', '30'],
        queryFn: () => {
            const toDate = new Date();
            toDate.setDate(toDate.getDate() + 30);
            return forecastingService.fetchForecast(undefined, toDate.toISOString().split('T')[0]);
        },
    });

    if (isLoading || !forecast || !forecast.predictions || forecast.predictions.length === 0) return null;

    const upcoming = forecast.predictions.slice(0, 4);

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 mb-6">
            <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                    <Zap size={18} className="text-indigo-500 dark:text-indigo-400"/>
                    <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200">{t('dashboard.forecasting.title')}</h2>
                </div>
                <Link
                    to="/forecasting"
                    className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
                >
                    {t('dashboard.forecasting.viewAll')} <ChevronRight size={14}/>
                </Link>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {upcoming.sort((a, b) => new Date(a.booking_date).getTime() - new Date(b.booking_date).getTime()).map((p, idx) => (
                    <div key={idx}
                         className="p-4 rounded-xl bg-gray-50 dark:bg-gray-800/50 border border-gray-100 dark:border-gray-700/50 hover:border-indigo-200 dark:hover:border-indigo-900/50 transition-colors group">
                        <div className="flex justify-between items-start mb-2">
                            <span
                                className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-tight">{fmtDate(p.booking_date)}</span>
                            <span
                                className={`text-xs font-bold ${p.amount >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-900 dark:text-gray-100'}`}>
                                {fmtCurrency(p.amount, p.currency)}
                            </span>
                        </div>
                        <p className="text-sm font-medium text-gray-700 dark:text-gray-300 truncate leading-tight"
                           title={p.description}>
                            {p.description}
                        </p>
                        <div className="mt-2 flex items-center gap-1.5">
                            <div className="w-1.5 h-1.5 rounded-full bg-indigo-500"/>
                            <span className="text-[10px] font-bold text-indigo-500/80 dark:text-indigo-400/80 uppercase">Expected</span>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

// ── page ─────────────────────────────────────────────────────────────────────

export default function DashboardPage() {
    const {t, i18n} = useTranslation();
    const [hideReconciled, setHideReconciled] = useState(true);

    const payslipsQuery = useQuery<Payslip[], Error>({
        queryKey: ['payslips'],
        queryFn: () => payslipService.fetchPayslips(),
    });

    const statementsQuery = useQuery<BankStatementSummary[]>({
        queryKey: ['bank-statements'],
        queryFn: bankService.fetchStatements,
    });

    const transactionsQuery = useQuery<Transaction[]>({
        queryKey: ['transactions', hideReconciled],
        queryFn: () => transactionService.fetchTransactions(undefined, hideReconciled),
    });

    const analyticsQuery = useQuery<TransactionAnalytics>({
        queryKey: ['analytics', hideReconciled],
        queryFn: () => transactionService.fetchAnalytics(hideReconciled),
    });

    const { data: baseCurrency = 'EUR' } = useQuery({
        queryKey: ['settings', 'BASE_DISPLAY_CURRENCY'],
        queryFn: () => settingsService.fetchSettings().then((s) => s['BASE_DISPLAY_CURRENCY'] || 'EUR'),
    });

    const isLoading = statementsQuery.isLoading || transactionsQuery.isLoading || analyticsQuery.isLoading;
    const isError = statementsQuery.isError || transactionsQuery.isError || analyticsQuery.isError;

    const statements = statementsQuery.data ?? [];
    const transactions = transactionsQuery.data ?? [];
    const analytics = analyticsQuery.data;

    const recentTxns = transactions.slice(0, 8);

    const chartData = analytics?.time_series ?? [];

    const maxFlowValue = chartData.length
        ? Math.max(...chartData.flatMap(d => [d.income || 0, d.expense || 0, 1]))
        : 1;

    const maxCategoryValue = (analytics?.category_totals || []).length
        ? Math.max(...(analytics?.category_totals || []).map(c => c.amount || 0))
        : 1;

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <LayoutDashboard className="text-indigo-600 dark:text-indigo-400"/> {getGreeting(t)}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {t('dashboard.subtitle')}
                    </p>
                </div>

                <div className="flex items-center gap-3">
                    <button
                        onClick={() => setHideReconciled(h => !h)}
                        className={`flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg border transition-colors shadow-sm ${hideReconciled
                            ? 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800/50 text-amber-700 dark:text-amber-400'
                            : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                        }`}
                    >
                        <Unlink
                            size={14}/> {hideReconciled ? t('transactions.showReconciled') : t('transactions.hideReconciled')}
                    </button>
                </div>
            </div>

            {isError && (
                <div
                    className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm">
                    <AlertCircle size={16}/>
                    {t('dashboard.errorLoad')}
                </div>
            )}

            {/* True Savings Card */}
            {!isLoading && <TrueSavingsCard analytics={analytics} payslips={payslipsQuery.data} baseCurrency={baseCurrency}/>}

            {/* KPI metrics */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <KpiCard
                    label={t('dashboard.kpi.statements')}
                    value={isLoading ? '…' : String(statements.length)}
                    sub={t('dashboard.kpi.statementsSub')}
                    color="gray"
                    icon={<Landmark size={20}/>}
                />
                <KpiCard
                    label={t('dashboard.kpi.totalIncome')}
                    value={isLoading ? '…' : (analytics ? fmtCurrency(analytics.total_income, baseCurrency) : '—')}
                    sub={t('dashboard.kpi.totalIncomeSub')}
                    color="green"
                    icon={<TrendingUp size={20}/>}
                />
                <KpiCard
                    label={t('dashboard.kpi.totalExpenses')}
                    value={isLoading ? '…' : (analytics ? fmtCurrency(analytics.total_expense, baseCurrency) : '—')}
                    sub={t('dashboard.kpi.totalExpensesSub')}
                    color="red"
                    icon={<TrendingDown size={20}/>}
                />
                <KpiCard
                    label={t('dashboard.kpi.netSavings')}
                    value={isLoading ? '…' : (analytics ? fmtCurrency(analytics.net_savings, baseCurrency) : '—')}
                    sub={t('dashboard.kpi.netSavingsSub')}
                    color={analytics && analytics.net_savings >= 0 ? 'blue' : 'red'}
                    icon={<Wallet size={20}/>}
                />
            </div>

            {/* Charts Section */}
            {!isLoading && analytics && (
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

                    {/* Cash Flow Chart */}
                    <div
                        className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 flex flex-col min-h-[340px]">
                        <div className="flex items-center justify-between mb-6">
                            <div className="flex items-center gap-2 text-gray-800 dark:text-gray-200">
                                <BarChart3 size={18} className="text-indigo-500 dark:text-indigo-400"/>
                                <h2 className="text-base font-semibold">{t('dashboard.cashFlow.title')}</h2>
                            </div>
                            <div
                                className="flex gap-3 text-[10px] uppercase font-bold text-gray-400 dark:text-gray-500">
                                <span className="flex items-center gap-1.5"><div
                                    className="w-2 h-2 bg-emerald-400 dark:bg-emerald-500 rounded-full"/>
                                    {t('dashboard.cashFlow.income')}</span>
                                <span className="flex items-center gap-1.5"><div
                                    className="w-2 h-2 bg-rose-400 dark:bg-rose-500 rounded-full"/>
                                    {t('dashboard.cashFlow.expense')}</span>
                            </div>
                        </div>

                        {chartData.length > 0 ? (
                            <div className="flex-1 relative flex flex-col">
                                <div
                                    className="absolute inset-0 flex flex-col justify-between pointer-events-none opacity-20 dark:opacity-40 z-0 pb-6">
                                    <div className="border-t border-gray-400 dark:border-gray-600 w-full"/>
                                    <div className="border-t border-gray-400 dark:border-gray-600 w-full"/>
                                    <div className="border-t border-gray-400 dark:border-gray-600 w-full"/>
                                    <div className="border-t border-gray-400 dark:border-gray-600 w-full"/>
                                    <div className="border-t border-gray-400 dark:border-gray-600 w-full"/>
                                </div>

                                <div
                                    className="flex-1 relative z-10 flex flex-row-reverse items-end gap-2 overflow-x-auto [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
                                    {[...chartData].reverse().map((pt) => {
                                        const monthLabel = formatChartMonth(pt.date, i18n.language);
                                        const incPct = maxFlowValue ? (pt.income / maxFlowValue) * 100 : 0;
                                        const expPct = maxFlowValue ? (pt.expense / maxFlowValue) * 100 : 0;

                                        return (
                                            <div key={pt.date}
                                                 className="group relative h-full flex flex-col justify-end min-w-[48px]">
                                                <div className="w-full flex items-end justify-center gap-1 h-full pb-6">
                                                    <div
                                                        className="w-4 bg-emerald-400 dark:bg-emerald-500 rounded-t-md transition-all group-hover:opacity-80"
                                                        style={{height: `${Math.max(incPct, pt.income > 0 ? 2 : 0)}%`}}
                                                        title={`${monthLabel} ${t('dashboard.cashFlow.income')}: ${fmtCurrency(pt.income, baseCurrency)}`}
                                                    />
                                                    <div
                                                        className="w-4 bg-rose-400 dark:bg-rose-500 rounded-t-md transition-all group-hover:opacity-80"
                                                        style={{height: `${Math.max(expPct, pt.expense > 0 ? 2 : 0)}%`}}
                                                        title={`${monthLabel} ${t('dashboard.cashFlow.expense')}: ${fmtCurrency(pt.expense, baseCurrency)}`}
                                                    />
                                                </div>
                                                <span
                                                    className="absolute bottom-0 left-1/2 -translate-x-1/2 text-[10px] font-medium text-gray-400 dark:text-gray-500 whitespace-nowrap">
                                                    {monthLabel}
                                                </span>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        ) : (
                            <div
                                className="flex-1 flex items-center justify-center text-sm text-gray-400 dark:text-gray-600">
                                {t('dashboard.cashFlow.noData')}
                            </div>
                        )}
                    </div>

                    {/* Top Categories */}
                    <div
                        className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 flex flex-col min-h-[340px]">
                        <div className="flex items-center gap-2 mb-6 text-gray-800 dark:text-gray-200">
                            <PieChart size={18} className="text-indigo-500 dark:text-indigo-400"/>
                            <h2 className="text-base font-semibold">{t('dashboard.topCategories.title')}</h2>
                        </div>

                        {(analytics?.category_totals || []).length > 0 ? (
                            <div className="space-y-4 flex-1 flex flex-col justify-center">
                                {(analytics?.category_totals || []).filter(c => c.type === 'expense').slice(0, 5).map((cat) => (
                                    <div key={cat.category}>
                                        <div className="flex justify-between items-end mb-1.5">
                                            <span
                                                className="text-sm font-medium text-gray-700 dark:text-gray-300 truncate pr-2">{cat.category || t('dashboard.topCategories.uncategorized')}</span>
                                            <span
                                                className="text-sm text-gray-500 dark:text-gray-400 font-mono whitespace-nowrap">{fmtCurrency(cat.amount, baseCurrency)}</span>
                                        </div>
                                        <div
                                            className="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-2.5 overflow-hidden">
                                            <div
                                                className="h-full rounded-full transition-all duration-500"
                                                style={{
                                                    width: `${(cat.amount / maxCategoryValue) * 100}%`,
                                                    backgroundColor: cat.color || '#94a3b8'
                                                }}
                                            />
                                        </div>
                                    </div>
                                ))}
                            </div>
                        ) : (
                            <div
                                className="flex-1 flex items-center justify-center text-sm text-gray-400 dark:text-gray-600">
                                {t('dashboard.topCategories.noData')}
                            </div>
                        )}
                    </div>
                </div>
            )}

            {/* Forecasting Widget */}
            {!isLoading && <ForecastingWidget />}

            {/* Subscriptions Widget */}
            {!isLoading && <SubscriptionSummaryCard baseCurrency={baseCurrency} />}

            {/* Swipeable Statements Section */}
            <div>
                <div className="flex items-center justify-between mb-4 mt-2">
                    <div className="flex items-center gap-2">
                        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200">{t('dashboard.bankStatements.title')}</h2>
                        {!isLoading && statements.length > 0 && (
                            <span
                                className="text-[10px] bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 px-1.5 py-0.5 rounded-full font-bold">
                                {statements.length}
                            </span>
                        )}
                    </div>
                    <Link
                        to="/bank-statements"
                        className="text-xs font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1 bg-indigo-50 dark:bg-indigo-900/30 px-3 py-1.5 rounded-md transition-colors"
                    >
                        {t('dashboard.bankStatements.viewAll')} <ChevronRight size={14}/>
                    </Link>
                </div>

                {isLoading ? (
                    <div className="flex gap-4 overflow-hidden pb-4">
                        {[1, 2, 3].map((i) => (
                            <div key={i}
                                 className="flex-none w-72 bg-white dark:bg-gray-900 rounded-2xl border border-gray-100 dark:border-gray-800 h-32 animate-pulse"/>
                        ))}
                    </div>
                ) : statements.length === 0 ? (
                    <div
                        className="flex flex-col items-center justify-center py-12 text-gray-400 dark:text-gray-600 bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-200 dark:border-gray-800">
                        <Landmark size={40} className="mb-3 opacity-30"/>
                        <p className="text-sm">{t('dashboard.bankStatements.noStatements')}</p>
                        <Link to="/import"
                              className="mt-3 text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-500">
                            {t('dashboard.bankStatements.importFirst')}
                        </Link>
                    </div>
                ) : (
                    <div className="relative group">
                        {/* 👇 Replaced 'scrollbar-hide' with cross-browser hidden scrollbar classes */}
                        <div
                            className="flex gap-4 overflow-x-auto pb-4 snap-x snap-mandatory [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
                            {statements.map((s: BankStatementSummary) => (
                                <div key={s.id} className="snap-start">
                                    <StatementCard stmt={s}/>
                                </div>
                            ))}
                            <div className="flex-none w-4"/>
                        </div>
                        <div
                            className="absolute right-0 top-0 bottom-4 w-12 bg-gradient-to-l from-gray-50/80 dark:from-gray-950/80 to-transparent pointer-events-none hidden md:block"/>
                    </div>
                )}
            </div>

            {/* Recent transactions */}
            {recentTxns.length > 0 && (
                <div>
                    <div className="flex items-center justify-between mb-4">
                        <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200">{t('dashboard.recentTxns.title')}</h2>
                        <Link
                            to="/transactions"
                            className="text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
                        >
                            {t('dashboard.recentTxns.viewAll')} <ArrowRight size={13}/>
                        </Link>
                    </div>
                    <div
                        className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col">
                        <div className="overflow-x-auto">
                            <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm">
                                <thead
                                    className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide">
                                <tr>
                                    <th className="px-5 py-3.5 text-left font-medium">{t('dashboard.recentTxns.date')}</th>
                                    <th className="px-5 py-3.5 text-left font-medium">{t('dashboard.recentTxns.description')}</th>
                                    <th className="px-5 py-3.5 text-left font-medium">{t('dashboard.recentTxns.category')}</th>
                                    <th className="px-5 py-3.5 text-right font-medium">{t('dashboard.recentTxns.amount')}</th>
                                </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                                {recentTxns.map((tx: Transaction) => (
                                    <tr key={tx.content_hash}
                                        className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                        <td className="px-5 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                                            {fmtDate(tx.booking_date)}
                                        </td>
                                        <td className="px-5 py-3 text-gray-800 dark:text-gray-200 max-w-sm truncate"
                                            title={tx.description}>
                                            {tx.description}
                                        </td>
                                        <td className="px-5 py-3">
                                            <CategoryBadge category={tx.category_id ?? undefined}/>
                                        </td>
                                        <td className={`px-5 py-3 text-right font-mono font-medium whitespace-nowrap ${
                                            tx.amount >= 0 ? 'text-green-600 dark:text-green-400' : 'text-gray-900 dark:text-gray-100'
                                        }`}
                                        >
                                            <span className="inline-flex items-center justify-end gap-1.5 w-full">
                                                {tx.amount >= 0 ? <TrendingUp size={14}
                                                                              className="text-green-500 dark:text-green-400"/> :
                                                    <TrendingDown size={14}
                                                                  className="text-gray-400 dark:text-gray-500"/>}
                                                {fmtCurrency(tx.amount, tx.currency)}
                                            </span>
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}