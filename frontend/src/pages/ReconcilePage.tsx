import { memo, useCallback, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation, Trans } from 'react-i18next';
import type { TFunction } from 'i18next';
import axios from 'axios';
import {
    AlertCircle,
    CheckCircle2,
    CheckSquare,
    CreditCard,
    Info,
    Landmark,
    Link2,
    Search,
    Settings2,
    Sparkles,
    Square,
    Trash2,
    X,
    Filter,
    History,
    HandMetal,
    ChevronLeft,
    ChevronRight,
    Wallet
} from 'lucide-react';
import { reconciliationService } from '../api/services/reconciliationService';
import { transactionService } from '../api/services/transactionService';
import type { ReconciliationPairSuggestion, Reconciliation } from "../api/types/transaction";
import { fmtCurrency, fmtDate, getLocalISODate } from '../utils/formatters';

// Helper: translate a statement_type string into a localized label with icon
function AccountBadge({ type, t }: { type?: string; t: TFunction }) {
    const label = type
        ? t(`reconcile.account.${type}`)
        : t('reconcile.account.unknown');

    const icon = type === 'giro'
        ? <Landmark size={12} />
        : type === 'credit_card'
            ? <CreditCard size={12} />
            : <Wallet size={12} />;

    const colors = type === 'giro'
        ? 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300'
        : type === 'credit_card'
            ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
            : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300';

    return (
        <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${colors}`}>
            {icon} {label}
        </span>
    );
}

interface ReconciliationRowProps {
    suggestion: ReconciliationPairSuggestion;
    isSelected: boolean;
    onToggle: (hash: string) => void;
    t: TFunction;
}

const ReconciliationRow = memo(({ suggestion, isSelected, onToggle, t }: ReconciliationRowProps) => {
    const source = suggestion.source_transaction;
    const target = suggestion.target_transaction;

    return (
        <div className={`border-b border-gray-100 dark:border-gray-800 last:border-b-0 transition-colors ${isSelected ? 'bg-indigo-50/30 dark:bg-indigo-900/10' : ''}`}>
            <div className="p-5 flex flex-col md:flex-row gap-6 items-center">
                <button
                    onClick={() => onToggle(source.content_hash)}
                    className={`shrink-0 ${isSelected ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500 hover:text-indigo-500'}`}
                >
                    {isSelected ? <CheckSquare size={24} /> : <Square size={24} />}
                </button>

                <div className="flex-1 min-w-0 space-y-2 opacity-90 hover:opacity-100 transition-opacity">
                    <div className="flex items-center gap-2 text-xs font-bold text-indigo-600 dark:text-indigo-400 uppercase tracking-wider mb-1">
                        <Landmark size={14} /> {t('reconcile.sourceGiro')}
                    </div>
                    <AccountBadge type={source.statement_type} t={t} />
                    <div className="font-mono text-lg font-bold text-gray-900 dark:text-gray-100">
                        {fmtCurrency(source.amount, source.currency)}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">
                        <p><span className="font-medium">{t('reconcile.date')}</span> {fmtDate(source.booking_date, 'short')}</p>
                        <p className="truncate mt-1" title={[source.counterparty_name, source.description].filter(Boolean).join(' · ')}>
                            {source.counterparty_name && <span className="font-bold text-gray-900 dark:text-gray-100 mr-1.5">{source.counterparty_name}</span>}
                            <span className="font-medium text-gray-500">{t('reconcile.ref')}</span> {source.description}
                        </p>
                    </div>
                </div>

                <div className="hidden md:flex flex-col items-center justify-center text-gray-300 dark:text-gray-600 px-4">
                    <Link2 size={24} className={isSelected ? 'text-indigo-500 dark:text-indigo-400' : 'opacity-50'} />
                    {suggestion.match_score >= 0.9 && (
                        <span className="text-[10px] uppercase font-bold text-green-500 mt-1 flex items-center gap-1">
                            <Sparkles size={10} /> {t('reconcile.perfectMatch')}
                        </span>
                    )}
                </div>

                <div className="flex-1 min-w-0 space-y-2 opacity-90 hover:opacity-100 transition-opacity">
                    <div className="flex items-center gap-2 text-xs font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider mb-1">
                        <CreditCard size={14} /> {t('reconcile.targetCredit')}
                    </div>
                    <AccountBadge type={target.statement_type} t={t} />
                    <div className="font-mono text-lg font-bold text-green-600 dark:text-green-400">
                        +{fmtCurrency(target.amount, target.currency)}
                    </div>
                    <div className="text-sm text-gray-600 dark:text-gray-400">
                        <p><span className="font-medium">{t('reconcile.date')}</span> {fmtDate(target.booking_date, 'short')}</p>
                        <p className="truncate mt-1" title={[target.counterparty_name, target.description].filter(Boolean).join(' · ')}>
                            {target.counterparty_name && <span className="font-bold text-gray-900 dark:text-gray-100 mr-1.5">{target.counterparty_name}</span>}
                            <span className="font-medium text-gray-500">{t('reconcile.ref')}</span> {target.description}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
});
ReconciliationRow.displayName = 'ReconciliationRow';

export default function ReconcilePage() {
    const { t } = useTranslation();
    const qc = useQueryClient();

    const [activeTab, setActiveTab] = useState<'suggestions' | 'manual' | 'history'>('suggestions');
    const [successMessage, setSuccessMessage] = useState<string | null>(null);
    const [errorMessage, setErrorMessage] = useState<string | null>(null);

    // --- Suggestions Tab State ---
    const [matchWindowDays, setMatchWindowDays] = useState<number>(7);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [deselectedHashes, setDeselectedHashes] = useState<Record<string, boolean>>({});

    // --- Manual Tab State ---
    const [manualSearchInput, setManualSearchInput] = useState('');
    const [manualDateFrom, setManualDateFrom] = useState('');
    const [manualDateTo, setManualDateTo] = useState('');
    const [appliedManualFilters, setAppliedManualFilters] = useState({ search: '', dateFrom: '', dateTo: '' });
    const [selectedManualSettlement, setSelectedManualSettlement] = useState<string | null>(null);
    const [selectedManualTarget, setSelectedManualTarget] = useState<string | null>(null);

    // --- Queries ---
    const { data: suggestions, isLoading: isLoadingSuggestions, isError: isErrorSuggestions } = useQuery({
        queryKey: ['reconciliation-suggestions', matchWindowDays],
        queryFn: () => reconciliationService.fetchSuggestions(matchWindowDays),
        enabled: activeTab === 'suggestions'
    });

    const { data: unreconciledTxns, isLoading: isLoadingManual } = useQuery({
        queryKey: ['transactions', 'unreconciled'],
        queryFn: () => transactionService.fetchTransactions(undefined, true),
        enabled: activeTab === 'manual'
    });

    const { data: historyList, isLoading: isLoadingHistory } = useQuery({
        queryKey: ['reconciliations'],
        queryFn: () => reconciliationService.fetchReconciliations(),
        enabled: activeTab === 'history'
    });

    // --- Mutations ---
    const reconcileBatchMutation = useMutation({
        mutationFn: async (pairs: { settlementHash: string, targetHash: string }[]) => {
            const promises = pairs.map(p => reconciliationService.create(p.settlementHash, p.targetHash));
            return Promise.all(promises);
        },
        onSuccess: () => {
            handleSuccess(t('reconcile.successLinked'));
            setSelectedManualSettlement(null);
            setSelectedManualTarget(null);
            setDeselectedHashes({});
        },
        onError: (err: unknown) => {
            const msg = axios.isAxiosError(err) ? (err.response?.data?.error || t('reconcile.errorFetch')) : t('reconcile.errorFetch');
            setErrorMessage(msg);
            setTimeout(() => setErrorMessage(null), 5000);
        }
    });

    const deleteRecMutation = useMutation({
        mutationFn: (id: string) => reconciliationService.delete(id),
        onSuccess: () => {
            handleSuccess(t('reconcile.successDeleted'));
        }
    });

    const handleSuccess = (msg: string) => {
        setSuccessMessage(msg);
        qc.invalidateQueries({ queryKey: ['reconciliation-suggestions'] });
        qc.invalidateQueries({ queryKey: ['transactions'] });
        qc.invalidateQueries({ queryKey: ['reconciliations'] });
        qc.invalidateQueries({ queryKey: ['analytics'] });
        setTimeout(() => setSuccessMessage(null), 4000);
    };

    // --- Suggestions Logic ---
    const filteredSuggestions = useMemo(() => {
        const list = suggestions || [];
        if (!searchTerm.trim()) return list;
        const lowerTerm = searchTerm.toLowerCase();
        return list.filter(sugg => {
            const sourceDesc = (sugg.source_transaction.description || '').toLowerCase();
            const sourceCp = (sugg.source_transaction.counterparty_name || '').toLowerCase();
            const targetDesc = (sugg.target_transaction.description || '').toLowerCase();
            const targetCp = (sugg.target_transaction.counterparty_name || '').toLowerCase();
            const sourceRef = (sugg.source_transaction.reference || '').toLowerCase();
            const targetRef = (sugg.target_transaction.reference || '').toLowerCase();
            const sourceAmt = Math.abs(sugg.source_transaction.amount).toString();
            return sourceDesc.includes(lowerTerm) || targetDesc.includes(lowerTerm) ||
                sourceCp.includes(lowerTerm) || targetCp.includes(lowerTerm) ||
                sourceRef.includes(lowerTerm) || targetRef.includes(lowerTerm) || sourceAmt.includes(lowerTerm);
        });
    }, [suggestions, searchTerm]);

    const pairsToLink = useMemo(() => {
        const list = suggestions || [];
        return list.reduce((acc, sugg) => {
            if (!deselectedHashes[sugg.source_transaction.content_hash]) {
                acc.push({ settlementHash: sugg.source_transaction.content_hash, targetHash: sugg.target_transaction.content_hash });
            }
            return acc;
        }, [] as { settlementHash: string, targetHash: string }[]);
    }, [suggestions, deselectedHashes]);

    const visibleSelectedCount = useMemo(() => {
        return filteredSuggestions.filter(sugg => !deselectedHashes[sugg.source_transaction.content_hash]).length;
    }, [filteredSuggestions, deselectedHashes]);

    const handleToggleAll = useCallback(() => {
        const allCurrentlySelected = filteredSuggestions.every(sugg => !deselectedHashes[sugg.source_transaction.content_hash]);
        const newState = { ...deselectedHashes };
        filteredSuggestions.forEach(sugg => {
            newState[sugg.source_transaction.content_hash] = allCurrentlySelected;
        });
        setDeselectedHashes(newState);
    }, [filteredSuggestions, deselectedHashes]);

    // --- Manual Logic ---
    const applyManualFilters = () => {
        setAppliedManualFilters({ search: manualSearchInput.toLowerCase(), dateFrom: manualDateFrom, dateTo: manualDateTo });
    };

    const shiftDateRange = (direction: -1 | 1) => {
        let fromDate = new Date(manualDateFrom);
        let toDate = new Date(manualDateTo);
        if (isNaN(fromDate.getTime()) || isNaN(toDate.getTime())) {
            toDate = new Date();
            fromDate = new Date();
            fromDate.setDate(toDate.getDate() - 30);
        }
        const diffTime = toDate.getTime() - fromDate.getTime();
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
        const shiftAmount = (diffDays === 0 ? 1 : diffDays) * direction;
        fromDate.setDate(fromDate.getDate() + shiftAmount);
        toDate.setDate(toDate.getDate() + shiftAmount);
        const format = (d: Date) => getLocalISODate(d);
        const newFrom = format(fromDate);
        const newTo = format(toDate);
        setManualDateFrom(newFrom);
        setManualDateTo(newTo);
        setAppliedManualFilters(prev => ({ ...prev, dateFrom: newFrom, dateTo: newTo }));
    };

    const manuallyFilteredTxns = useMemo(() => {
        if (!unreconciledTxns) return [];
        return unreconciledTxns.filter(tx => {
            if (appliedManualFilters.dateFrom && tx.booking_date < appliedManualFilters.dateFrom) return false;
            if (appliedManualFilters.dateTo && tx.booking_date > appliedManualFilters.dateTo) return false;
            if (appliedManualFilters.search) {
                const term = appliedManualFilters.search;
                const desc = (tx.description || '').toLowerCase();
                const cp = (tx.counterparty_name || '').toLowerCase();
                const ref = (tx.reference || '').toLowerCase();
                const amt = Math.abs(tx.amount).toString();
                if (!desc.includes(term) && !cp.includes(term) && !ref.includes(term) && !amt.includes(term)) return false;
            }
            return true;
        });
    }, [unreconciledTxns, appliedManualFilters]);

    const manualDebits = manuallyFilteredTxns.filter(tx => tx.amount < 0);
    const manualCredits = manuallyFilteredTxns.filter(tx => tx.amount > 0);

    const linkManualPair = () => {
        if (selectedManualSettlement && selectedManualTarget) {
            reconcileBatchMutation.mutate([{ settlementHash: selectedManualSettlement, targetHash: selectedManualTarget }]);
        }
    };

    return (
        <div className="space-y-6 pb-20 animate-in fade-in duration-300">
            <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Link2 className="text-indigo-600 dark:text-indigo-400" /> {t('reconcile.title')}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {t('reconcile.subtitle')}
                    </p>
                </div>
                <div className="flex bg-gray-100 dark:bg-gray-800 p-1 rounded-xl shrink-0">
                    <button onClick={() => setActiveTab('suggestions')} className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'suggestions' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}`}>
                        <Sparkles size={16} /> {t('reconcile.tabs.suggestions')}
                    </button>
                    <button onClick={() => setActiveTab('manual')} className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'manual' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}`}>
                        <HandMetal size={16} /> {t('reconcile.tabs.manual')}
                    </button>
                    <button onClick={() => setActiveTab('history')} className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'history' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}`}>
                        <History size={16} /> {t('reconcile.tabs.history')}
                    </button>
                </div>
            </div>

            {successMessage && (
                <div className="flex items-center gap-3 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/50 rounded-xl text-green-700 dark:text-green-400">
                    <CheckCircle2 size={20} />
                    <p className="text-sm font-medium">{successMessage}</p>
                </div>
            )}

            {errorMessage && (
                <div className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400">
                    <AlertCircle size={20} />
                    <p className="text-sm font-medium">{errorMessage}</p>
                </div>
            )}

            {/* TAB 1: SUGGESTIONS */}
            {activeTab === 'suggestions' && (
                <div className="space-y-6 animate-in fade-in">
                    <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
                        <div className="flex items-center gap-3 w-full sm:w-auto">
                            <button onClick={handleToggleAll} className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800">
                                <CheckSquare size={16} /> {t('reconcile.toggleAll')}
                            </button>
                        </div>
                        <div className="flex flex-wrap items-center gap-3 w-full sm:w-auto">
                            <div className="relative flex-1 sm:min-w-[250px]">
                                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"><Search size={16} className="text-gray-400" /></div>
                                <input type="text" value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} placeholder={t('reconcile.searchPlaceholder')} className="block w-full pl-9 pr-10 py-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:outline-none placeholder:text-gray-400" />
                                {searchTerm && (<button onClick={() => setSearchTerm('')} className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X size={14} /></button>)}
                            </div>
                            <div className="flex items-center gap-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-2 shadow-sm">
                                <Settings2 size={16} className="text-gray-500" />
                                <span className="text-sm font-medium text-gray-500">±</span>
                                <input type="number" min="1" max="90" value={matchWindowDays} onChange={(e) => setMatchWindowDays(Number(e.target.value) || 7)} className="w-12 bg-transparent border-none p-0 text-sm text-center focus:ring-0 font-semibold text-gray-900 dark:text-gray-100" />
                                <span className="text-sm font-medium text-gray-500">{t('reconcile.days')}</span>
                            </div>
                        </div>
                    </div>

                    <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800/50 rounded-xl p-4 flex gap-3 text-blue-800 dark:text-blue-300 shadow-sm">
                        <Info className="shrink-0 mt-0.5" size={20} />
                        <div className="text-sm">
                            <p className="font-bold mb-1">{t('reconcile.infoTitle1to1')}</p>
                            <p><Trans i18nKey="reconcile.infoDesc1to1" /></p>
                        </div>
                    </div>

                    {isErrorSuggestions && (
                        <div className="flex items-center gap-3 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400">
                            <AlertCircle size={20} /><p className="text-sm font-medium">{t('reconcile.errorFetch')}</p>
                        </div>
                    )}

                    {isLoadingSuggestions ? (
                        <div className="h-32 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm animate-pulse" />
                    ) : (suggestions?.length || 0) === 0 ? (
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-8 flex flex-col items-center text-center">
                            <div className="w-12 h-12 bg-green-100 dark:bg-green-900/20 text-green-600 dark:text-green-400 rounded-full flex items-center justify-center mb-4"><CheckCircle2 size={24} /></div>
                            <h3 className="text-gray-900 dark:text-gray-100 font-semibold mb-1">{t('reconcile.noMatches1to1')}</h3>
                            <p className="text-sm text-gray-500 dark:text-gray-400 max-w-md">{t('reconcile.noMatchesDesc1to1')}</p>
                        </div>
                    ) : filteredSuggestions.length === 0 ? (
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-8 text-center text-gray-500 dark:text-gray-400">{t('reconcile.noSearchResults')}</div>
                    ) : (
                        <div className="space-y-6">
                            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm">
                                {filteredSuggestions.map((sugg) => (
                                    <ReconciliationRow key={`${sugg.source_transaction.content_hash}-${sugg.target_transaction.content_hash}`} suggestion={sugg} isSelected={!deselectedHashes[sugg.source_transaction.content_hash]} onToggle={(hash) => setDeselectedHashes(prev => ({ ...prev, [hash]: !prev[hash] }))} t={t} />
                                ))}
                            </div>

                            {/* Floating Bulk Link Button */}
                            {pairsToLink.length > 0 && (
                                <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4 duration-300 w-[95%] sm:w-auto">
                                    <button
                                        onClick={() => reconcileBatchMutation.mutate(pairsToLink)}
                                        disabled={reconcileBatchMutation.isPending}
                                        className="w-full flex justify-center items-center gap-2 text-sm font-bold px-8 py-4 rounded-2xl bg-indigo-600 text-white hover:bg-indigo-700 disabled:opacity-40 disabled:cursor-not-allowed transition-all shadow-2xl shadow-indigo-500/20 border border-indigo-500/50 hover:-translate-y-1 active:scale-95"
                                    >
                                        <Link2 size={18} />
                                        {reconcileBatchMutation.isPending ? t('reconcile.updating') : t('reconcile.linkCount', { count: pairsToLink.length })}
                                        {searchTerm && pairsToLink.length > visibleSelectedCount && (<span className="opacity-75 text-xs ml-2">({t('reconcile.includesHidden')})</span>)}
                                    </button>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            {/* TAB 2: MANUAL MATCH */}
            {activeTab === 'manual' && (
                <div className="space-y-6 animate-in fade-in">
                    {/* Manual Filter Bar */}
                    <div className="bg-white dark:bg-gray-900 p-4 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col md:flex-row gap-4 items-end">
                        <div className="flex-1 w-full space-y-1">
                            <label className="text-xs font-semibold text-gray-500 uppercase">{t('reconcile.manual.search')}</label>
                            <input
                                type="text"
                                value={manualSearchInput}
                                onChange={e => setManualSearchInput(e.target.value)}
                                placeholder={t('reconcile.searchPlaceholder')}
                                className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 outline-none text-gray-900 dark:text-gray-100"
                            />
                        </div>

                        {/* Date Pagination and Inputs */}
                        <div className="flex items-end gap-2 w-full md:w-auto">
                            <button
                                onClick={() => shiftDateRange(-1)}
                                className="mb-[2px] p-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-500 hover:text-indigo-600 hover:border-indigo-300 transition-colors"
                                title={t('reconcile.manual.prevPeriod')}
                            >
                                <ChevronLeft size={18} />
                            </button>

                            <div className="flex gap-2">
                                <div className="space-y-1">
                                    <label className="text-xs font-semibold text-gray-500 uppercase">{t('reconcile.manual.dateFrom')}</label>
                                    <input
                                        type="date"
                                        value={manualDateFrom}
                                        onChange={e => setManualDateFrom(e.target.value)}
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-gray-100"
                                    />
                                </div>
                                <div className="space-y-1">
                                    <label className="text-xs font-semibold text-gray-500 uppercase">{t('reconcile.manual.dateTo')}</label>
                                    <input
                                        type="date"
                                        value={manualDateTo}
                                        onChange={e => setManualDateTo(e.target.value)}
                                        className="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-gray-100"
                                    />
                                </div>
                            </div>

                            <button
                                onClick={() => shiftDateRange(1)}
                                className="mb-[2px] p-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-500 hover:text-indigo-600 hover:border-indigo-300 transition-colors"
                                title={t('reconcile.manual.nextPeriod')}
                            >
                                <ChevronRight size={18} />
                            </button>
                        </div>

                        <button
                            onClick={applyManualFilters}
                            className="w-full md:w-auto px-6 py-2 bg-indigo-600 text-white font-medium rounded-lg text-sm hover:bg-indigo-700 flex justify-center items-center gap-2"
                        >
                            <Filter size={16} /> {t('common.filters')}
                        </button>
                    </div>

                    {isLoadingManual ? (
                        <div className="h-64 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 animate-pulse" />
                    ) : (
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                            {/* Debits Column */}
                            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden flex flex-col h-[500px]">
                                <div className="p-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                                    <h3 className="font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2"><Landmark className="text-indigo-500" size={18} /> {t('reconcile.manual.selectDebit')}</h3>
                                </div>
                                <div className="flex-1 overflow-y-auto p-2 space-y-2">
                                    {manualDebits.length === 0 ? (
                                        <div className="p-4 text-center text-sm text-gray-500">{t('reconcile.manual.noDebits')}</div>
                                    ) : manualDebits.map(tx => (
                                        <div key={tx.content_hash} onClick={() => setSelectedManualSettlement(tx.content_hash)} className={`p-3 rounded-lg border cursor-pointer transition-colors ${selectedManualSettlement === tx.content_hash ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20' : 'border-gray-100 dark:border-gray-800 hover:border-indigo-300'}`}>
                                            <div className="flex justify-between items-start mb-1">
                                                <span className="font-mono font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(tx.amount, tx.currency)}</span>
                                                <span className="text-xs text-gray-500">{fmtDate(tx.booking_date, 'short')}</span>
                                            </div>
                                            <div className="flex items-center gap-2 mb-2"><AccountBadge type={tx.statement_type} t={t} /></div>
                                            <div>
                                                {tx.counterparty_name && (
                                                    <div className="text-xs font-bold text-gray-900 dark:text-gray-100 truncate mb-0.5" title={tx.counterparty_name}>
                                                        {tx.counterparty_name}
                                                    </div>
                                                )}
                                                <p className="text-xs text-gray-600 dark:text-gray-400 line-clamp-2" title={tx.description}>{tx.description}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            {/* Credits Column */}
                            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden flex flex-col h-[500px]">
                                <div className="p-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                                    <h3 className="font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2"><CreditCard className="text-emerald-500" size={18} /> {t('reconcile.manual.selectCredit')}</h3>
                                </div>
                                <div className="flex-1 overflow-y-auto p-2 space-y-2">
                                    {manualCredits.length === 0 ? (
                                        <div className="p-4 text-center text-sm text-gray-500">{t('reconcile.manual.noCredits')}</div>
                                    ) : manualCredits.map(tx => (
                                        <div key={tx.content_hash} onClick={() => setSelectedManualTarget(tx.content_hash)} className={`p-3 rounded-lg border cursor-pointer transition-colors ${selectedManualTarget === tx.content_hash ? 'border-emerald-500 bg-emerald-50 dark:bg-emerald-900/20' : 'border-gray-100 dark:border-gray-800 hover:border-emerald-300'}`}>
                                            <div className="flex justify-between items-start mb-1">
                                                <span className="font-mono font-bold text-emerald-600 dark:text-emerald-400">+{fmtCurrency(tx.amount, tx.currency)}</span>
                                                <span className="text-xs text-gray-500">{fmtDate(tx.booking_date, 'short')}</span>
                                            </div>
                                            <div className="flex items-center gap-2 mb-2"><AccountBadge type={tx.statement_type} t={t} /></div>
                                            <div>
                                                {tx.counterparty_name && (
                                                    <div className="text-xs font-bold text-gray-900 dark:text-gray-100 truncate mb-0.5" title={tx.counterparty_name}>
                                                        {tx.counterparty_name}
                                                    </div>
                                                )}
                                                <p className="text-xs text-gray-600 dark:text-gray-400 line-clamp-2" title={tx.description}>{tx.description}</p>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    <div className="bg-white dark:bg-gray-900 p-4 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center justify-between">
                        <div className="text-sm">
                            {selectedManualSettlement && selectedManualTarget ? (
                                <span className="text-indigo-600 dark:text-indigo-400 font-medium">{t('reconcile.manual.readyToLink')}</span>
                            ) : (
                                <span className="text-gray-500">{t('reconcile.manual.selectBoth')}</span>
                            )}
                        </div>
                        <button
                            onClick={linkManualPair}
                            disabled={!selectedManualSettlement || !selectedManualTarget || reconcileBatchMutation.isPending}
                            className="px-6 py-2 bg-indigo-600 text-white font-medium rounded-lg text-sm hover:bg-indigo-700 disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2"
                        >
                            <Link2 size={16} /> {t('reconcile.manual.link')}
                        </button>
                    </div>
                </div>
            )}

            {/* TAB 3: HISTORY / LINKED */}
            {activeTab === 'history' && (
                <div className="space-y-4 animate-in fade-in">
                    {isLoadingHistory ? (
                        <div className="h-64 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 animate-pulse" />
                    ) : !historyList || historyList.length === 0 ? (
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-8 text-center text-gray-500">{t('reconcile.history.empty')}</div>
                    ) : (
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                            <div className="overflow-x-auto">
                                <table className="w-full text-left text-sm">
                                    <thead className="bg-gray-50 dark:bg-gray-800/50 text-gray-500 dark:text-gray-400 font-medium border-b border-gray-200 dark:border-gray-800">
                                    <tr>
                                        <th className="px-5 py-3">{t('reconcile.history.linkedAt')}</th>
                                        <th className="px-5 py-3">{t('reconcile.history.amount')}</th>
                                        <th className="px-5 py-3">{t('reconcile.history.details')}</th>
                                        <th className="px-5 py-3 text-right">{t('common.actions')}</th>
                                    </tr>
                                    </thead>
                                    <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                                    {historyList.map((rec: Reconciliation) => (
                                        <tr key={rec.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                                            <td className="px-5 py-4 text-gray-900 dark:text-gray-100 whitespace-nowrap">{fmtDate(rec.reconciled_at, 'short')}</td>
                                            <td className="px-5 py-4 font-mono font-bold text-gray-900 dark:text-gray-100 whitespace-nowrap">{fmtCurrency(rec.amount, 'EUR')}</td>
                                            <td className="px-5 py-4 text-sm text-gray-600 dark:text-gray-400">
                                                {/* Source row */}
                                                <div className="flex items-center gap-2 mb-2">
                                                    <span className="font-semibold text-indigo-500 text-xs w-4">S</span>
                                                    <AccountBadge type={rec.settlement_statement_type} t={t} />
                                                    <span className="text-xs text-gray-500 whitespace-nowrap">{rec.settlement_booking_date ? fmtDate(rec.settlement_booking_date, 'short') : '—'}</span>
                                                    <span className="truncate max-w-[250px]" title={rec.settlement_transaction_description || rec.settlement_transaction_hash}>
                                                        {rec.settlement_transaction_description || rec.settlement_transaction_hash}
                                                    </span>
                                                </div>
                                                {/* Target row */}
                                                <div className="flex items-center gap-2">
                                                    <span className="font-semibold text-emerald-500 text-xs w-4">T</span>
                                                    <AccountBadge type={rec.target_statement_type} t={t} />
                                                    <span className="text-xs text-gray-500 whitespace-nowrap">{rec.target_booking_date ? fmtDate(rec.target_booking_date, 'short') : '—'}</span>
                                                    <span className="truncate max-w-[250px]" title={rec.target_transaction_description || rec.target_transaction_hash}>
                                                        {rec.target_transaction_description || rec.target_transaction_hash}
                                                    </span>
                                                </div>
                                            </td>
                                            <td className="px-5 py-4 text-right">
                                                <button onClick={() => deleteRecMutation.mutate(rec.id)} className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors" title={t('reconcile.history.delete')}>
                                                    <Trash2 size={18} />
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                    </tbody>
                                </table>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}