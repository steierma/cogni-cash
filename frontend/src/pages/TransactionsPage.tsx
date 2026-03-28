import * as React from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import axios from 'axios';
import {
    BarChart3, Database, Layers, Loader2, Search, MapPin,
    Sparkles, TrendingDown, TrendingUp, Trophy, Unlink, X, ArrowLeftRight, Columns, Check
} from 'lucide-react';

import {
    cancelAutoCategorize, fetchBankStatements, fetchCategories,
    fetchTransactions, getAutoCategorizeStatus, startAutoCategorize,
    updateTransactionCategory, fetchSettings, updateSettings,
    markTransactionReviewed
} from '../api/client';
import type { BankStatementSummary, Category, JobState, Transaction } from '../api/types';
import { fmtCurrency, fmtDate } from '../utils/formatters';

import TransactionFilters, { type FilterState } from '../components/transactions/TransactionFilters';
import TransactionTable, { type TxColKey, type SortKey, type SortDir } from '../components/transactions/TransactionTable';

function formatChartMonth(yyyyMM: string, locale: string) {
    try {
        const [year, month] = yyyyMM.split('-');
        const date = new Date(parseInt(year), parseInt(month) - 1);
        return date.toLocaleDateString(locale, { month: 'short', year: '2-digit' });
    } catch {
        return yyyyMM;
    }
}

export default function TransactionsPage() {
    const { t, i18n } = useTranslation();
    const [searchParams] = useSearchParams();
    const preselectedStatement = searchParams.get('statement') ?? '';

    const initialFilters: FilterState = {
        search: '',
        type: 'all',
        statement: preselectedStatement,
        category: 'all',
        from: null,
        to: null,
        amountMin: '',
        amountMax: '',
    };

    const [applied, setApplied] = React.useState<FilterState>(initialFilters);
    const [hasAppliedOnce, setHasAppliedOnce] = React.useState(false);

    const [hideReconciled, setHideReconciled] = React.useState(() => {
        const hr = searchParams.get('hide_reconciled');
        if (hr !== null) return hr === 'true';
        return preselectedStatement ? false : true;
    });

    const [selectedHashes, setSelectedHashes] = React.useState<Set<string>>(new Set());
    const [sortKey, setSortKey] = React.useState<SortKey>('booking_date');
    const [sortDir, setSortDir] = React.useState<SortDir>('desc');

    const [showVisuals, setShowVisuals] = React.useState(true);
    const [topHitsCount, setTopHitsCount] = React.useState<number>(3);
    const [toastMessage, setToastMessage] = React.useState<string | null>(null);

    // Columns Configuration
    const [showColMenu, setShowColMenu] = React.useState(false);
    const [visibleCols, setVisibleCols] = React.useState<Record<TxColKey, boolean>>({
        date: true, description: true, location: true, reference: false, category: true, amount: true
    });

    const qc = useQueryClient();

    const { data: statements = [], isLoading: isLoadingStatements } = useQuery<BankStatementSummary[]>({
        queryKey: ['bank-statements'],
        queryFn: fetchBankStatements,
    });

    const { data: allTxns = [], isLoading: isLoadingTxns, isError } = useQuery<Transaction[]>({
        queryKey: ['transactions', applied.statement, hideReconciled],
        queryFn: () => fetchTransactions(applied.statement || undefined, hideReconciled),
        enabled: hasAppliedOnce,
    });

    const { data: categories = [] } = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: fetchCategories,
        staleTime: 5 * 60 * 1000,
    });

    const { data: settings } = useQuery({
        queryKey: ['settings'],
        queryFn: fetchSettings
    });

    const { data: jobStatus, refetch: refetchJobStatus } = useQuery<JobState>({
        queryKey: ['auto-categorize-status'],
        queryFn: getAutoCategorizeStatus,
        refetchInterval: (query) => query.state?.data?.is_running ? 1500 : false,
    });

    const updateSettingsMut = useMutation({
        mutationFn: updateSettings,
        onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] })
    });

    const markReviewedMutation = useMutation({
        mutationFn: (hash: string) => markTransactionReviewed(hash),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['transactions'] });
            qc.invalidateQueries({ queryKey: ['analytics'] });
        },
    });

    const batchMarkReviewedMutation = useMutation({
        mutationFn: async (hashes: string[]) => {
            await Promise.all(hashes.map(hash => markTransactionReviewed(hash)));
        },
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['transactions'] });
            qc.invalidateQueries({ queryKey: ['analytics'] });
            setSelectedHashes(new Set());
        },
    });

    React.useEffect(() => {
        if (settings?.transactions_visible_cols) {
            try {
                setVisibleCols(JSON.parse(settings.transactions_visible_cols));
            } catch (e) {
                console.error("Failed to parse column settings", e);
            }
        }
    }, [settings?.transactions_visible_cols]);

    const toggleColumn = (col: TxColKey) => {
        setVisibleCols(prev => {
            const next = { ...prev, [col]: !prev[col] };
            updateSettingsMut.mutate({ transactions_visible_cols: JSON.stringify(next) });
            return next;
        });
    };

    const wasRunning = React.useRef(false);
    React.useEffect(() => {
        if (jobStatus) {
            if (jobStatus.is_running) {
                wasRunning.current = true;
            } else if (wasRunning.current) {
                wasRunning.current = false;
                if (jobStatus.status === 'completed') {
                    setToastMessage(`Successfully categorized ${jobStatus.processed} transactions!`);
                } else if (jobStatus.status === 'cancelled') {
                    setToastMessage(`Categorization cancelled after processing ${jobStatus.processed} items.`);
                }
                setTimeout(() => setToastMessage(null), 5000);
                qc.invalidateQueries({ queryKey: ['transactions'] });
                qc.invalidateQueries({ queryKey: ['analytics'] });
            }
        }
    }, [jobStatus, qc]);

    const isLoading = isLoadingStatements || (hasAppliedOnce && isLoadingTxns);

    const toDay = (date: string): string => {
        if (!date) return '';
        if (/^\d{4}-\d{2}-\d{2}T/.test(date)) return date.slice(0, 10);
        if (/^\d{4}-\d{2}-\d{2}$/.test(date)) return date;
        if (/^\d{2}\.\d{2}\.\d{4}$/.test(date)) {
            const [d, m, y] = date.split('.');
            return `${y}-${m}-${d}`;
        }
        return '';
    };

    const [minDate, maxDate] = React.useMemo((): [string, string] => {
        const days = allTxns.map((t) => toDay(t.booking_date)).filter(Boolean).sort() as string[];
        if (days.length === 0) return ['', ''];
        return [days[0], days[days.length - 1]];
    }, [allTxns]);

    const handleApply = (newFilters: FilterState) => {
        setApplied(newFilters);
        setHasAppliedOnce(true);
        setSelectedHashes(new Set());
    };

    const handleClear = () => {
        setApplied(initialFilters);
        setHasAppliedOnce(false);
        setSelectedHashes(new Set());
    };

    const filtered: Transaction[] = React.useMemo(() => {
        if (!hasAppliedOnce) return [];
        let rows = allTxns;

        if (hideReconciled) {
            rows = rows.filter((t) => !t.is_reconciled);
        }

        if (applied.category !== 'all') {
            if (applied.category === 'uncategorized') rows = rows.filter((t) => !t.category_id);
            else rows = rows.filter((t) => t.category_id === applied.category);
        }

        if (applied.type !== 'all') {
            rows = rows.filter((t) => applied.type === 'credit' ? t.amount >= 0 : t.amount < 0);
        }

        if (applied.search.trim()) {
            const q = applied.search.toLowerCase();
            rows = rows.filter((t) =>
                t.description.toLowerCase().includes(q) ||
                (t.reference && t.reference.toLowerCase().includes(q)) ||
                (t.location && t.location.toLowerCase().includes(q))
            );
        }

        const fromDate = applied.from ?? minDate;
        const toDate = applied.to ?? maxDate;
        const bookingDay = (d: string) => d.length > 10 ? d.slice(0, 10) : d;

        if (fromDate) rows = rows.filter((t) => bookingDay(t.booking_date) >= fromDate);
        if (toDate) rows = rows.filter((t) => bookingDay(t.booking_date) <= toDate);

        const parseAmount = (val: string) => parseFloat(val.replace(',', '.'));

        if (applied.amountMin !== '') rows = rows.filter((t) => t.amount >= parseAmount(applied.amountMin));
        if (applied.amountMax !== '') rows = rows.filter((t) => t.amount <= parseAmount(applied.amountMax));

        rows = [...rows].sort((a, b) => {
            let cmp = 0;
            if (sortKey === 'booking_date') cmp = a.booking_date.localeCompare(b.booking_date);
            else if (sortKey === 'description') cmp = a.description.localeCompare(b.description);
            else if (sortKey === 'location') cmp = (a.location || '').localeCompare(b.location || '');
            else if (sortKey === 'amount') cmp = a.amount - b.amount;
            return sortDir === 'asc' ? cmp : -cmp;
        });

        return rows;
    }, [allTxns, applied, minDate, maxDate, sortKey, sortDir, hasAppliedOnce, hideReconciled]);

    const totalCredit = filtered.filter((t) => t.amount > 0).reduce((s, t) => s + t.amount, 0);
    const totalDebit = filtered.filter((t) => t.amount < 0).reduce((s, t) => s + t.amount, 0);
    const hasAppliedFilters = JSON.stringify(applied) !== JSON.stringify(initialFilters);

    const toggleSort = (key: SortKey) => {
        if (sortKey === key) setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
        else {
            setSortKey(key);
            setSortDir('desc');
        }
    };

    const chartData = React.useMemo(() => {
        const monthly: Record<string, { inc: number; exp: number }> = {};
        const cats: Record<string, number> = {};

        filtered.forEach(t => {
            const period = t.booking_date.slice(0, 7);
            if (!monthly[period]) monthly[period] = { inc: 0, exp: 0 };

            if (t.amount >= 0) {
                monthly[period].inc += t.amount;
            } else {
                monthly[period].exp += Math.abs(t.amount);
                const catInfo = categories.find(c => c.id === t.category_id);
                const c = catInfo ? catInfo.name : 'Uncategorized';
                cats[c] = (cats[c] || 0) + Math.abs(t.amount);
            }
        });

        return {
            monthly: Object.entries(monthly).sort().slice(-12),
            cats: Object.entries(cats).sort((a, b) => b[1] - a[1]).slice(0, 6)
        };
    }, [filtered, categories]);

    const maxFlowVal = chartData.monthly.length ? Math.max(...chartData.monthly.flatMap(d => [d[1].inc, d[1].exp])) || 1 : 1;
    const maxCatVal = chartData.cats.length ? Math.max(...chartData.cats.map(c => c[1])) || 1 : 1;

    const { topIncomes, topSpends } = React.useMemo(() => {
        const credits = filtered.filter(t => t.amount > 0).sort((a, b) => b.amount - a.amount);
        const debits = filtered.filter(t => t.amount < 0).sort((a, b) => a.amount - b.amount);
        return {
            topIncomes: credits.slice(0, topHitsCount),
            topSpends: debits.slice(0, topHitsCount)
        };
    }, [filtered, topHitsCount]);

    const startJobMutation = useMutation({
        mutationFn: startAutoCategorize,
        onSuccess: () => refetchJobStatus(),
        onError: (err: unknown) => {
            if (axios.isAxiosError(err)) {
                if (err.response?.status === 404) setToastMessage("No uncategorized transactions found.");
                else if (err.response?.status === 409) refetchJobStatus();
                else setToastMessage("Failed to start. Is the local AI service running?");
            } else {
                setToastMessage("An unknown error occurred.");
            }
            setTimeout(() => setToastMessage(null), 4000);
        }
    });

    const cancelJobMutation = useMutation({
        mutationFn: cancelAutoCategorize,
        onSuccess: () => refetchJobStatus()
    });

    const catMutation = useMutation({
        mutationFn: ({ hash, categoryId }: { hash: string; categoryId: string }) => updateTransactionCategory(hash, categoryId),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['transactions'] });
            qc.invalidateQueries({ queryKey: ['bank-statements'] });
            qc.invalidateQueries({ queryKey: ['analytics'] });
        },
    });

    const batchCatMutation = useMutation({
        mutationFn: async ({ hashes, categoryId }: { hashes: string[]; categoryId: string }) => {
            await Promise.all(hashes.map(hash => updateTransactionCategory(hash, categoryId)));
        },
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['transactions'] });
            qc.invalidateQueries({ queryKey: ['bank-statements'] });
            qc.invalidateQueries({ queryKey: ['analytics'] });
            setSelectedHashes(new Set());
        },
    });

    const toggleSelect = (hash: string) => {
        const next = new Set(selectedHashes);
        if (next.has(hash)) next.delete(hash);
        else next.add(hash);
        setSelectedHashes(next);
    };

    const toggleSelectAll = () => {
        if (selectedHashes.size === filtered.length && filtered.length > 0) setSelectedHashes(new Set());
        else setSelectedHashes(new Set(filtered.map(t => t.content_hash)));
    };

    const handleReviewAll = () => {
        const unreviewedHashes = filtered.filter(t => !t.reviewed).map(t => t.content_hash);
        if (unreviewedHashes.length > 0) {
            batchMarkReviewedMutation.mutate(unreviewedHashes);
        }
    };

    const unreviewedCount = React.useMemo(() => filtered.filter(t => !t.reviewed).length, [filtered]);

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <ArrowLeftRight className="text-indigo-600 dark:text-indigo-400" /> {t('transactions.title')}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {!hasAppliedOnce
                            ? t('transactions.ready')
                            : isLoading ? 'Loading…' : t('transactions.countOf', { filtered: filtered.length.toLocaleString(i18n.language), total: allTxns.length.toLocaleString(i18n.language) })}
                    </p>
                </div>

                <div className="flex items-center gap-3">
                    {toastMessage && (
                        <span className="text-sm font-medium text-indigo-600 dark:text-indigo-300 bg-indigo-50 dark:bg-indigo-900/30 px-3 py-1.5 rounded-lg border border-indigo-100 dark:border-indigo-800/50 animate-in fade-in duration-300">
                            {toastMessage}
                        </span>
                    )}

                    {jobStatus?.is_running ? (
                        <div className="flex items-center gap-3 bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-800/50 px-4 py-2 rounded-xl shadow-sm animate-in fade-in">
                            <Loader2 size={16} className="animate-spin text-indigo-600 dark:text-indigo-400" />
                            <div className="flex flex-col">
                                <span className="text-xs text-indigo-400 dark:text-indigo-500 font-medium uppercase tracking-wider">{t('transactions.categorizing')}</span>
                                <span className="text-sm font-bold text-indigo-700 dark:text-indigo-300 leading-tight">
                                    {jobStatus.processed} / {jobStatus.total}
                                </span>
                            </div>
                            <button
                                onClick={() => cancelJobMutation.mutate()}
                                disabled={cancelJobMutation.isPending}
                                className="ml-2 p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                                title={t('transactions.stopBatch')}
                            >
                                <X size={16} />
                            </button>
                        </div>
                    ) : (
                        <button
                            onClick={() => startJobMutation.mutate()}
                            disabled={startJobMutation.isPending}
                            className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 dark:bg-indigo-500 dark:hover:bg-indigo-600 text-white px-4 py-2 rounded-xl text-sm font-medium transition-all shadow-sm disabled:opacity-70 disabled:cursor-not-allowed"
                        >
                            <Sparkles size={16} />
                            {t('transactions.autoCategorize')}
                        </button>
                    )}
                </div>
            </div>

            <div className="flex items-center justify-end gap-2 flex-wrap">
                {hasAppliedOnce && filtered.length > 0 && (
                    <>
                        <div className="relative">
                            <button
                                onClick={() => setShowColMenu(!showColMenu)}
                                className="flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg border transition-colors bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50"
                            >
                                <Columns size={14} /> {t('transactions.columns', 'Columns')}
                            </button>
                            {showColMenu && (
                                <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg z-10 overflow-hidden p-2 space-y-1">
                                    {[
                                        { key: 'date', label: t('transactions.cols.date', 'Date') },
                                        { key: 'description', label: t('transactions.cols.description', 'Description') },
                                        { key: 'location', label: t('transactions.cols.location', 'Location') },
                                        { key: 'reference', label: t('transactions.cols.reference', 'Reference') },
                                        { key: 'category', label: t('transactions.cols.category', 'Category') },
                                        { key: 'amount', label: t('transactions.cols.amount', 'Amount') },
                                    ].map(({ key, label }) => (
                                        <button
                                            key={key}
                                            onClick={() => toggleColumn(key as TxColKey)}
                                            className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
                                        >
                                            {label} {visibleCols[key as TxColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
                                        </button>
                                    ))}
                                </div>
                            )}
                        </div>

                        <button
                            onClick={() => setShowVisuals(!showVisuals)}
                            className={`flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg border transition-colors ${showVisuals
                                ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-100 dark:border-indigo-800/50 text-indigo-600 dark:text-indigo-400'
                                : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                            }`}
                        >
                            <BarChart3 size={14} /> {showVisuals ? t('transactions.hideCharts') : t('transactions.showCharts')}
                        </button>
                    </>
                )}
                {hasAppliedOnce && (
                    <button
                        onClick={() => setHideReconciled(h => !h)}
                        className={`flex items-center gap-1.5 text-sm px-3 py-1.5 rounded-lg border transition-colors ${hideReconciled
                            ? 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800/50 text-amber-700 dark:text-amber-400'
                            : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                        }`}
                    >
                        <Unlink size={14} /> {hideReconciled ? t('transactions.showReconciled') : t('transactions.hideReconciled')}
                    </button>
                )}
                {hasAppliedFilters && (
                    <button
                        onClick={handleClear}
                        className="flex items-center gap-1.5 text-sm text-gray-500 dark:text-gray-400 hover:text-red-500 dark:hover:text-red-400 px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-800 hover:border-red-200 dark:hover:border-red-800 transition-colors bg-white dark:bg-gray-900"
                    >
                        <X size={13} /> {t('transactions.clearFilters')}
                    </button>
                )}
            </div>

            {isError && (
                <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm">
                    Failed to load transactions. Is the backend running?
                </div>
            )}

            <TransactionFilters
                applied={applied}
                onApply={handleApply}
                hasAppliedOnce={hasAppliedOnce}
                isLoading={isLoading}
                statements={statements}
                categories={categories}
                minDate={minDate}
                maxDate={maxDate}
            />

            {isLoading && (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden animate-pulse">
                    <div className="divide-y divide-gray-100 dark:divide-gray-800">
                        {Array.from({ length: 8 }).map((_, i) => (
                            <div key={i} className="flex items-center gap-4 px-4 py-3">
                                <div className="h-3 w-20 bg-gray-100 dark:bg-gray-800 rounded" />
                                <div className="h-3 flex-1 bg-gray-100 dark:bg-gray-800 rounded" />
                                <div className="h-3 w-28 bg-gray-100 dark:bg-gray-800 rounded" />
                                <div className="h-3 w-24 bg-gray-100 dark:bg-gray-800 rounded" />
                                <div className="h-3 w-16 bg-gray-100 dark:bg-gray-800 rounded ml-auto" />
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {!isLoading && !hasAppliedOnce && (
                <div className="flex flex-col items-center justify-center py-32 bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-200 dark:border-gray-800 text-gray-400 dark:text-gray-500">
                    <div className="bg-indigo-50 dark:bg-indigo-900/20 p-4 rounded-full mb-4">
                        <Database size={32} className="text-indigo-400 dark:text-indigo-500" />
                    </div>
                    <p className="text-base font-medium text-gray-600 dark:text-gray-300">{t('common.noData')}</p>
                    <p className="text-sm text-gray-400 dark:text-gray-500 mt-1 max-w-xs text-center">
                        Configure your filters above and click "Search" to fetch your transactions.
                    </p>
                </div>
            )}

            {!isLoading && hasAppliedOnce && filtered.length > 0 && (
                <div className="space-y-4 animate-in fade-in duration-300">
                    <div className="grid grid-cols-3 gap-4">
                        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 text-center shadow-sm">
                            <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('transactions.filteredRows')}</p>
                            <p className="text-xl font-bold text-gray-900 dark:text-gray-100">{filtered.length.toLocaleString(i18n.language)}</p>
                        </div>
                        <div className="bg-green-50 dark:bg-green-900/20 rounded-xl border border-green-100 dark:border-green-800/50 p-4 text-center shadow-sm">
                            <p className="text-xs text-green-600 dark:text-green-400 uppercase tracking-wide mb-1">{t('transactions.income')}</p>
                            <p className="text-xl font-bold text-green-700 dark:text-green-300">{fmtCurrency(totalCredit)}</p>
                        </div>
                        <div className="bg-red-50 dark:bg-red-900/20 rounded-xl border border-red-100 dark:border-red-800/50 p-4 text-center shadow-sm">
                            <p className="text-xs text-red-500 dark:text-red-400 uppercase tracking-wide mb-1">{t('transactions.expenses')}</p>
                            <p className="text-xl font-bold text-red-600 dark:text-red-400">{fmtCurrency(totalDebit)}</p>
                        </div>
                    </div>

                    {showVisuals && (
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 flex flex-col">
                                <div className="flex items-center justify-between mb-2">
                                    <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-2">
                                        <BarChart3 size={16} className="text-indigo-500 dark:text-indigo-400" /> {t('transactions.filteredCashFlow')}
                                    </h3>
                                    <div className="flex gap-3 text-[10px] uppercase font-bold text-gray-400 dark:text-gray-500">
                                        <span className="flex items-center gap-1.5"><div className="w-2 h-2 bg-emerald-400 dark:bg-emerald-500 rounded-full" /> {t('transactions.income')}</span>
                                        <span className="flex items-center gap-1.5"><div className="w-2 h-2 bg-rose-400 dark:bg-rose-500 rounded-full" /> {t('transactions.expenses')}</span>
                                    </div>
                                </div>
                                {chartData.monthly.length > 0 ? (
                                    <div className="flex-1 relative mt-4 min-h-[160px]">
                                        <div className="absolute inset-0 flex flex-col justify-between pointer-events-none opacity-20 dark:opacity-40 z-0">
                                            <div className="border-t border-gray-300 dark:border-gray-700 w-full" />
                                            <div className="border-t border-gray-300 dark:border-gray-700 w-full" />
                                            <div className="border-t border-gray-300 dark:border-gray-700 w-full" />
                                            <div className="border-t border-gray-300 dark:border-gray-700 w-full" />
                                            <div className="border-t border-gray-300 dark:border-gray-700 w-full mt-auto" />
                                        </div>
                                        <div className="relative h-full flex items-end justify-between gap-1 z-10 pb-1">
                                            {chartData.monthly.map(([period, val]) => {
                                                const incPct = maxFlowVal ? (val.inc / maxFlowVal) * 100 : 0;
                                                const expPct = maxFlowVal ? (val.exp / maxFlowVal) * 100 : 0;
                                                const monthLabel = formatChartMonth(period, i18n.language);

                                                return (
                                                    <div key={period} className="flex-1 flex flex-col items-center gap-2 group h-full justify-end">
                                                        <div className="w-full flex items-end justify-center gap-1 h-full pb-1">
                                                            <div
                                                                className="w-full max-w-[14px] bg-emerald-400 dark:bg-emerald-500 rounded-t-sm transition-all group-hover:bg-emerald-500 dark:group-hover:bg-emerald-400 opacity-80 group-hover:opacity-100"
                                                                style={{ height: `${Math.max(incPct, val.inc > 0 ? 2 : 0)}%` }}
                                                                title={`${t('transactions.income')}: ${fmtCurrency(val.inc)}`}
                                                            />
                                                            <div
                                                                className="w-full max-w-[14px] bg-rose-400 dark:bg-rose-500 rounded-t-sm transition-all group-hover:bg-rose-500 dark:group-hover:bg-rose-400 opacity-80 group-hover:opacity-100"
                                                                style={{ height: `${Math.max(expPct, val.exp > 0 ? 2 : 0)}%` }}
                                                                title={`${t('transactions.expenses')}: ${fmtCurrency(val.exp)}`}
                                                            />
                                                        </div>
                                                        <span className="text-[10px] text-gray-400 dark:text-gray-500 font-mono whitespace-nowrap">{monthLabel}</span>
                                                    </div>
                                                );
                                            })}
                                        </div>
                                    </div>
                                ) : (
                                    <div className="flex-1 flex items-center justify-center text-sm text-gray-400 dark:text-gray-500">{t('common.noTimeline')}</div>
                                )}
                            </div>

                            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                                <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                    {t('transactions.topExpenses')}
                                </h3>
                                <div className="space-y-4">
                                    {chartData.cats.length > 0 ? chartData.cats.map(([name, amt]) => {
                                        const catInfo = categories.find(c => c.name === name);
                                        return (
                                            <div key={name} className="space-y-1.5">
                                                <div className="flex justify-between text-xs font-medium text-gray-600 dark:text-gray-400">
                                                    <span>{name}</span>
                                                    <span className="font-mono">{fmtCurrency(amt)}</span>
                                                </div>
                                                <div className="w-full bg-gray-100 dark:bg-gray-800 h-1.5 rounded-full overflow-hidden">
                                                    <div
                                                        className="h-full rounded-full transition-all duration-500"
                                                        style={{ width: `${(amt / maxCatVal) * 100}%`, backgroundColor: catInfo?.color || '#cbd5e1' }}
                                                    />
                                                </div>
                                            </div>
                                        );
                                    }) : (
                                        <div className="h-full flex items-center justify-center pb-8 text-sm text-gray-400 dark:text-gray-500">
                                            No expense categories mapped in current filter.
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}

            {!isLoading && hasAppliedOnce && showVisuals && (topIncomes.length > 0 || topSpends.length > 0) && (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6 animate-in fade-in duration-300">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-2">
                            <Trophy size={16} className="text-indigo-500 dark:text-indigo-400" />
                            {t('transactions.topTransactions')}
                        </h3>
                        <div className="flex items-center gap-2">
                            <span className="text-xs text-gray-500 dark:text-gray-400">{t('transactions.show')}</span>
                            <select
                                value={topHitsCount}
                                onChange={(e) => setTopHitsCount(Number(e.target.value))}
                                className="text-xs rounded-lg border border-gray-200 dark:border-gray-700 py-1.5 px-2 bg-gray-50 dark:bg-gray-800 focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 outline-none text-gray-700 dark:text-gray-300 cursor-pointer"
                            >
                                <option value={3}>{t('transactions.top3')}</option>
                                <option value={5}>{t('transactions.top5')}</option>
                                <option value={10}>{t('transactions.top10')}</option>
                            </select>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <h4 className="text-[11px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-3 border-b border-gray-100 dark:border-gray-800 pb-2 flex items-center gap-1.5">
                                <TrendingDown size={14} className="text-red-400 dark:text-red-500" /> {t('transactions.largestSpends')}
                            </h4>
                            <div className="space-y-2">
                                {topSpends.map(tx => (
                                    <div key={tx.content_hash} className="flex justify-between items-center bg-red-50/40 dark:bg-red-900/10 hover:bg-red-50/80 dark:hover:bg-red-900/20 transition-colors p-2.5 rounded-lg border border-red-100/50 dark:border-red-900/30">
                                        <div className="min-w-0 flex-1 pr-3">
                                            <p className="text-sm font-medium text-gray-800 dark:text-gray-200 truncate" title={tx.description}>{tx.description}</p>
                                            <p className="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5 flex items-center gap-1.5 flex-wrap">
                                                <span>{fmtDate(tx.booking_date)}</span>
                                                {tx.location && (
                                                    <>
                                                        <span className="opacity-50">•</span>
                                                        <span className="flex items-center gap-0.5"><MapPin size={10} className="inline opacity-70" /> {tx.location}</span>
                                                    </>
                                                )}
                                                <span className="opacity-50">•</span>
                                                <span>{categories.find(c => c.id === tx.category_id)?.name || 'Uncategorized'}</span>
                                            </p>
                                        </div>
                                        <span className="font-mono text-sm font-bold text-red-600 dark:text-red-400 whitespace-nowrap">
                                            {fmtCurrency(tx.amount, tx.currency)}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div>
                            <h4 className="text-[11px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-3 border-b border-gray-100 dark:border-gray-800 pb-2 flex items-center gap-1.5">
                                <TrendingUp size={14} className="text-emerald-500 dark:text-emerald-400" /> {t('transactions.largestIncomes')}
                            </h4>
                            <div className="space-y-2">
                                {topIncomes.map(tx => (
                                    <div key={tx.content_hash} className="flex justify-between items-center bg-emerald-50/40 dark:bg-emerald-900/10 hover:bg-emerald-50/80 dark:hover:bg-emerald-900/20 transition-colors p-2.5 rounded-lg border border-emerald-100/50 dark:border-emerald-900/30">
                                        <div className="min-w-0 flex-1 pr-3">
                                            <p className="text-sm font-medium text-gray-800 dark:text-gray-200 truncate" title={tx.description}>{tx.description}</p>
                                            <p className="text-[10px] text-gray-500 dark:text-gray-400 mt-0.5 flex items-center gap-1.5 flex-wrap">
                                                <span>{fmtDate(tx.booking_date)}</span>
                                                {tx.location && (
                                                    <>
                                                        <span className="opacity-50">•</span>
                                                        <span className="flex items-center gap-0.5"><MapPin size={10} className="inline opacity-70" /> {tx.location}</span>
                                                    </>
                                                )}
                                                <span className="opacity-50">•</span>
                                                <span>{categories.find(c => c.id === tx.category_id)?.name || 'Uncategorized'}</span>
                                            </p>
                                        </div>
                                        <span className="font-mono text-sm font-bold text-emerald-700 dark:text-emerald-400 whitespace-nowrap">
                                            {fmtCurrency(tx.amount, tx.currency)}
                                        </span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {!isLoading && hasAppliedOnce && filtered.length === 0 && (
                <div className="flex flex-col items-center justify-center py-20 bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-200 dark:border-gray-800 text-gray-400 dark:text-gray-500">
                    <Search size={36} className="mb-3 opacity-30 dark:opacity-20" />
                    <p className="text-sm">{t('transactions.noMatches')}</p>
                    {hasAppliedFilters && (
                        <button onClick={handleClear} className="mt-2 text-sm text-indigo-500 dark:text-indigo-400 hover:underline">
                            {t('transactions.clearFilters')}
                        </button>
                    )}
                </div>
            )}

            {!isLoading && hasAppliedOnce && filtered.length > 0 && (
                <TransactionTable
                    transactions={filtered}
                    categories={categories}
                    selectedHashes={selectedHashes}
                    onToggleSelect={toggleSelect}
                    onToggleSelectAll={toggleSelectAll}
                    sortKey={sortKey}
                    sortDir={sortDir}
                    onSort={toggleSort}
                    onCategoryChange={(hash, categoryId) => catMutation.mutate({ hash, categoryId })}
                    onMarkReviewed={(hash) => markReviewedMutation.mutate(hash)}
                    visibleCols={visibleCols}
                />
            )}

            {/* Sticky Bulk Action Bar */}
            {selectedHashes.size > 0 && (
                <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4 duration-300 w-[95%] sm:w-auto">
                    <div className="bg-gray-900 dark:bg-gray-800 text-white rounded-2xl shadow-2xl px-6 py-4 flex flex-wrap sm:flex-nowrap items-center justify-center gap-4 sm:gap-6 border border-white/10 dark:border-gray-700">
                        <div className="flex items-center gap-2 pr-0 sm:pr-4 border-r-0 sm:border-r border-gray-700 dark:border-gray-600">
                            <Layers size={18} className="text-indigo-400" />
                            <span className="font-bold text-sm whitespace-nowrap">{t('transactions.selected', { count: selectedHashes.size })}</span>
                        </div>

                        <div className="flex items-center gap-3">
                            <span className="hidden sm:inline-block text-xs text-gray-400 font-medium uppercase tracking-wider">{t('transactions.actions', 'Actions')}</span>
                            <button
                                onClick={() => batchMarkReviewedMutation.mutate(Array.from(selectedHashes))}
                                disabled={batchMarkReviewedMutation.isPending}
                                className="flex items-center gap-1.5 bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold px-3 py-1.5 rounded-lg transition-colors shadow-sm disabled:opacity-50 whitespace-nowrap"
                            >
                                <Check size={14} /> {t('transactions.markReviewed', 'Mark Reviewed')}
                            </button>
                        </div>

                        <div className="flex items-center gap-3">
                            <span className="hidden sm:inline-block text-xs text-gray-400 font-medium uppercase tracking-wider">{t('transactions.setCategory')}</span>
                            <select
                                className="bg-gray-800 dark:bg-gray-900 text-sm rounded-lg border border-gray-700 dark:border-gray-600 px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 min-w-[140px]"
                                defaultValue="placeholder"
                                onChange={(e) => {
                                    const val = e.target.value;
                                    if (val === 'placeholder') return;

                                    batchCatMutation.mutate({
                                        hashes: Array.from(selectedHashes),
                                        categoryId: val === 'unset' ? '' : val
                                    });
                                    e.target.value = 'placeholder';
                                }}
                                disabled={batchCatMutation.isPending}
                            >
                                <option value="placeholder" disabled hidden>{t('transactions.choose')}</option>
                                <option value="unset">{t('transactions.unset')}</option>
                                {categories.map((c) => (
                                    <option key={c.id} value={c.id}>{c.name}</option>
                                ))}
                            </select>
                        </div>

                        <button onClick={() => setSelectedHashes(new Set())} className="absolute -top-2 -right-2 sm:relative sm:top-auto sm:right-auto bg-gray-900 sm:bg-transparent rounded-full p-1 sm:p-0 text-gray-400 hover:text-white transition-colors border sm:border-0 border-gray-700">
                            <X size={20} />
                        </button>
                    </div>
                </div>
            )}

            {/* Floating "Review All" Button */}
            {unreviewedCount > 0 && hasAppliedOnce && (
                <div className={`fixed z-40 transition-all duration-300 animate-in fade-in slide-in-from-bottom-4 
                    ${selectedHashes.size > 0 ? 'bottom-32 sm:bottom-6 right-4 sm:right-6' : 'bottom-6 right-4 sm:right-6'}
                `}>
                    <button
                        onClick={handleReviewAll}
                        disabled={batchMarkReviewedMutation.isPending}
                        className="flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white shadow-xl shadow-indigo-500/20 border border-indigo-500/50 px-5 py-3 rounded-2xl text-sm font-bold transition-transform hover:-translate-y-1 active:scale-95 disabled:opacity-50 disabled:hover:translate-y-0"
                    >
                        <Check size={18} />
                        {t('transactions.reviewAll')} ({unreviewedCount})
                    </button>
                </div>
            )}
        </div>
    );
}