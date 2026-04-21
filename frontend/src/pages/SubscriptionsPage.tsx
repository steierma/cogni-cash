import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { 
    RefreshCcw, 
    Inbox, 
    CalendarClock, 
    CheckCircle2, 
    XCircle, 
    ChevronRight,
    Search,
    TrendingUp,
    Clock,
    RotateCcw,
    Trash2,
    Sparkles,
    User,
    Settings2,
    Save,
    History
} from 'lucide-react';
import { subscriptionService } from '../api/services/subscriptionService';
import { settingsService } from '../api/services/settingsService';
import { categoryService } from '../api/services/categoryService';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import type { Category } from '../api/types/category';
import type { Subscription, SuggestedSubscription, DiscoveryFeedback, BaseTransaction } from '../api/types/subscription';

export default function SubscriptionsPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const navigate = useNavigate();
    const [search, setSearch] = useState('');
    const [toastMessage, setToastMessage] = useState<string | null>(null);

    const { data: subscriptions = [], isLoading: isLoadingSubs } = useQuery<Subscription[]>({
        queryKey: ['subscriptions'],
        queryFn: subscriptionService.fetchSubscriptions
    });

    const { data: suggested = [], isLoading: isLoadingSuggested } = useQuery<SuggestedSubscription[]>({
        queryKey: ['suggestedSubscriptions'],
        queryFn: subscriptionService.fetchSuggestedSubscriptions
    });

    const { data: feedback = [] } = useQuery<DiscoveryFeedback[]>({
        queryKey: ['discoveryFeedback'],
        queryFn: subscriptionService.fetchDiscoveryFeedback
    });

    const declined = feedback.filter(f => f.status === 'DECLINED' || f.status === 'AI_REJECTED');

    const { data: categories = [] } = useQuery({
        queryKey: ['categories'],
        queryFn: () => categoryService.fetchCategories()
    });

    const approveMutation = useMutation({
        mutationFn: subscriptionService.approveSubscription,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['suggestedSubscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
            setToastMessage(t('subscriptions.approveSuccessAsync'));
            setTimeout(() => setToastMessage(null), 5000);
        }
    });

    const declineMutation = useMutation({
        mutationFn: (merchantName: string) => subscriptionService.declineSuggestion(merchantName),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['suggestedSubscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['discoveryFeedback'] });
        }
    });

    const undeclineMutation = useMutation({
        mutationFn: (merchantName: string) => subscriptionService.removeDiscoveryFeedback(merchantName),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['suggestedSubscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['discoveryFeedback'] });
        }
    });

    const enrichMutation = useMutation({
        mutationFn: subscriptionService.enrichSubscription,
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
            navigate(`/subscriptions/${data.id}?enriched=true`);
        }
    });

    const deleteMutation = useMutation({
        mutationFn: subscriptionService.deleteSubscription,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
        }
    });

    const handleDelete = (e: React.MouseEvent, id: string, name: string) => {
        e.stopPropagation();
        if (window.confirm(t('subscriptions.deleteConfirm', { name }))) {
            deleteMutation.mutate(id);
        }
    };

    const totalMonthly = subscriptions
        .filter(s => s.status === 'active' || s.status === 'cancellation_pending')
        .reduce((acc, s) => {
            const amount = s.amount;
            const interval = s.billing_interval || 1;
            
            if (s.billing_cycle === 'yearly') {
                // Paid every X years -> (amount / 12) / X
                return acc + (amount / (12 * interval));
            }
            // Paid every X months (e.g. 3 for quarterly) -> amount / X
            return acc + (amount / interval);
        }, 0);

    const filteredSubs = subscriptions.filter(s => 
        s.merchant_name.toLowerCase().includes(search.toLowerCase())
    );

    const activeGroup = filteredSubs.filter(s => s.status !== 'cancelled');
    const cancelledGroup = filteredSubs.filter(s => s.status === 'cancelled');

    return (
        <div className="space-y-8 animate-in fade-in duration-500">
            {/* Header & Stats */}
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
                <div className="space-y-1">
                    <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                        {t('layout.subscriptions')}
                    </h1>
                    <p className="text-gray-500 dark:text-gray-400">
                        {t('subscriptions.subtitle')}
                    </p>
                </div>

                <div className="flex items-center gap-4">
                    {toastMessage && (
                        <div className="flex items-center gap-2 px-4 py-2 bg-indigo-50 dark:bg-indigo-900/30 border border-indigo-100 dark:border-indigo-800/50 text-indigo-700 dark:text-indigo-300 text-sm font-bold rounded-2xl shadow-sm animate-in fade-in slide-in-from-right-4 duration-300">
                            <Sparkles size={16} className="animate-pulse" />
                            {toastMessage}
                        </div>
                    )}
                    
                    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm flex items-center gap-4 min-w-[200px]">
                        <div className="h-12 w-12 rounded-xl bg-indigo-50 dark:bg-indigo-900/30 flex items-center justify-center text-indigo-600 dark:text-indigo-400 shrink-0">
                            <TrendingUp size={24} />
                        </div>
                        <div>
                            <p className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">{t('subscriptions.monthlySpend')}</p>
                            <p className="text-2xl font-bold text-gray-900 dark:text-white">{fmtCurrency(totalMonthly)}</p>
                        </div>
                    </div>
                </div>
            </div>

            <DiscoverySettings />

            {/* Discovery Inbox */}
            {isLoadingSuggested && (
                <div className="bg-indigo-600 rounded-3xl p-6 md:p-8 text-white shadow-xl shadow-indigo-500/20 relative overflow-hidden">
                    <div className="absolute top-0 right-0 -mt-8 -mr-8 h-64 w-64 bg-white/10 rounded-full blur-3xl pointer-events-none" />
                    <div className="relative z-10 flex flex-col items-center justify-center py-12 space-y-4">
                        <div className="h-12 w-12 border-4 border-white/20 border-t-white rounded-full animate-spin" />
                        <div className="text-center space-y-1">
                            <p className="text-xl font-bold">{t('subscriptions.discoveryLoading')}</p>
                            <p className="text-indigo-100 text-sm">{t('subscriptions.discoverySubtitle')}</p>
                        </div>
                    </div>
                </div>
            )}

            {!isLoadingSuggested && suggested.length > 0 && (
                <div className="bg-indigo-600 rounded-3xl p-6 md:p-8 text-white shadow-xl shadow-indigo-500/20 relative">
                    <div className="absolute top-0 right-0 -mt-8 -mr-8 h-64 w-64 bg-white/10 rounded-full blur-3xl pointer-events-none" />
                    
                    <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
                        <div className="space-y-2">
                            <div className="flex items-center gap-2 text-indigo-100 font-bold uppercase tracking-widest text-xs">
                                <Inbox size={16} />
                                {t('subscriptions.discoveryInbox')}
                            </div>
                            <h2 className="text-2xl md:text-3xl font-bold">
                                {t('subscriptions.discoveryTitle', { count: suggested.length })}
                            </h2>
                            <p className="text-indigo-100 max-w-xl">
                                {t('subscriptions.discoverySubtitle')}
                            </p>
                        </div>
                        
                        <div className="flex -space-x-4">
                            {suggested.slice(0, 3).map((s, i) => (
                                <div key={i} className="h-12 w-12 rounded-full bg-white/20 backdrop-blur-md border-2 border-white/30 flex items-center justify-center text-lg font-bold">
                                    {s.merchant_name.charAt(0)}
                                </div>
                            ))}
                            {suggested.length > 3 && (
                                <div className="h-12 w-12 rounded-full bg-white/20 backdrop-blur-md border-2 border-white/30 flex items-center justify-center text-sm font-bold">
                                    +{suggested.length - 3}
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="mt-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                        {suggested.map((s, idx) => (
                            <div key={idx} className="bg-white/10 backdrop-blur-md border border-white/20 rounded-2xl p-5 hover:bg-white/20 transition-all group relative hover:z-20">
                                <div className="flex justify-between items-start mb-4">
                                    <div className="space-y-1">
                                        <p className="font-bold text-lg leading-tight">{s.merchant_name}</p>
                                        <p className="text-indigo-200 text-sm">{fmtCurrency(Math.abs(s.estimated_amount))} / {s.billing_cycle}</p>
                                    </div>
                                    <div className="group/tooltip relative h-10 w-10 rounded-xl bg-white/10 flex items-center justify-center cursor-help">
                                        <CalendarClock size={20} />
                                        
                                        {/* Tooltip */}
                                        <div className="absolute bottom-full right-0 mb-3 w-48 bg-gray-900 border border-white/10 rounded-2xl p-4 shadow-2xl opacity-0 group-hover/tooltip:opacity-100 transition-all pointer-events-none z-50 transform translate-y-2 group-hover/tooltip:translate-y-0">
                                            <p className="text-[10px] font-bold text-indigo-400 uppercase tracking-widest mb-3 border-b border-white/5 pb-2">{t('subscriptions.baseTransactions')}</p>
                                            <div className="space-y-2.5">
                                                {s.base_transactions?.slice(-3).reverse().map((bt: BaseTransaction, i: number) => (
                                                    <div key={i} className="flex justify-between items-center gap-2">
                                                        <span className="text-[11px] text-gray-400">{fmtDate(bt.date)}</span>
                                                        <span className="text-[11px] font-bold text-white">{fmtCurrency(Math.abs(bt.amount))}</span>
                                                    </div>
                                                ))}
                                            </div>
                                            <div className="absolute -bottom-1 right-4 w-2 h-2 bg-gray-900 rotate-45" />
                                        </div>
                                    </div>
                                </div>
                                <div className="flex items-center justify-between gap-3 mt-4">
                                    <button 
                                        onClick={() => approveMutation.mutate(s)}
                                        disabled={approveMutation.isPending}
                                        className="flex-1 bg-white text-indigo-600 py-2.5 rounded-xl text-sm font-bold hover:bg-indigo-50 transition-colors disabled:opacity-50"
                                    >
                                        {t('subscriptions.approve')}
                                    </button>
                                    <button 
                                        onClick={() => declineMutation.mutate(s.merchant_name)}
                                        disabled={declineMutation.isPending}
                                        className="px-3 py-2.5 rounded-xl border border-white/20 hover:bg-white/10 transition-colors text-white/70 hover:text-white disabled:opacity-50"
                                        title={t('subscriptions.declineSuggestion')}
                                    >
                                        <XCircle size={18} />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            {/* Active Subscriptions List */}
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl shadow-sm overflow-hidden">
                <div className="p-6 border-b border-gray-100 dark:border-gray-800 flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <h3 className="text-xl font-bold">{t('subscriptions.activeAndPausedSubscriptions')}</h3>
                    
                    <div className="relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                        <input 
                            type="text" 
                            placeholder={t('subscriptions.searchPlaceholder')}
                            value={search}
                            onChange={(e) => setSearch(e.target.value)}
                            className="pl-10 pr-4 py-2 bg-gray-50 dark:bg-gray-800 border-none rounded-xl focus:ring-2 focus:ring-indigo-500 w-full md:w-64 transition-all"
                        />
                    </div>
                </div>

                <div className="overflow-x-auto">
                    <table className="w-full text-left">
                        <thead>
                            <tr className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest border-b border-gray-100 dark:border-gray-800">
                                <th className="px-6 py-4">{t('subscriptions.merchantService')}</th>
                                <th className="px-6 py-4">{t('subscriptions.statusHeader')}</th>
                                <th className="px-6 py-4">{t('subscriptions.billing')}</th>
                                <th className="px-6 py-4">{t('subscriptions.amountHeader')}</th>
                                <th className="px-6 py-4">{t('subscriptions.nextPayment')}</th>
                                <th className="px-6 py-4 text-right">{t('common.actions')}</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                            {activeGroup.map((s) => (
                                <tr 
                                    key={s.id} 
                                    onClick={() => navigate(`/subscriptions/${s.id}`)}
                                    className="group hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
                                >
                                    <td className="px-6 py-5">
                                        <div className="flex items-center gap-4">
                                            <div className="h-12 w-12 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center font-bold text-gray-500 shrink-0 group-hover:bg-indigo-50 dark:group-hover:bg-indigo-900/30 group-hover:text-indigo-600 transition-colors">
                                                {s.merchant_name.charAt(0)}
                                            </div>
                                            <div className="space-y-0.5">
                                                <div className="font-bold text-gray-900 dark:text-white flex items-center gap-2">
                                                    {s.merchant_name}
                                                    {s.is_trial && (
                                                        <span className="bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 text-[10px] px-2 py-0.5 rounded-full uppercase tracking-tighter font-black">{t('subscriptions.trial')}</span>
                                                    )}
                                                </div>
                                                <p className="text-sm text-gray-500 dark:text-gray-400">
                                                    {categories.find((c: Category) => c.id === s.category_id)?.name || t('common.uncategorized')}
                                                </p>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-5">
                                        <StatusBadge status={s.status} />
                                    </td>
                                    <td className="px-6 py-5">
                                        <p className="text-sm font-medium capitalize">
                                            {s.billing_cycle}
                                            {s.billing_interval > 1 && ` ${t('subscriptions.everyXMonths', { count: s.billing_interval })}`}
                                        </p>
                                    </td>
                                    <td className="px-6 py-5">
                                        <div className="font-bold text-gray-900 dark:text-white">
                                            {fmtCurrency(s.amount)}
                                        </div>
                                    </td>
                                    <td className="px-6 py-5">
                                        <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                                            <Clock size={14} />
                                            {s.next_occurrence ? fmtDate(s.next_occurrence) : '—'}
                                        </div>
                                    </td>
                                    <td className="px-6 py-5 text-right">
                                        <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <button 
                                                onClick={(e) => { e.stopPropagation(); enrichMutation.mutate(s.id); }}
                                                disabled={enrichMutation.isPending}
                                                className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-500 hover:text-indigo-600 transition-all"
                                                title={t('subscriptions.enrichWithAI')}
                                            >
                                                <Sparkles size={18} className={enrichMutation.isPending ? "animate-spin" : ""} />
                                            </button>
                                            <button 
                                                onClick={(e) => handleDelete(e, s.id, s.merchant_name)}
                                                disabled={deleteMutation.isPending}
                                                className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-400 hover:text-red-600 transition-all"
                                                title={t('subscriptions.deleteTitle')}
                                            >
                                                <Trash2 size={18} />
                                            </button>
                                            <button 
                                                onClick={(e) => { e.stopPropagation(); navigate(`/subscriptions/${s.id}?cancel=true`); }}
                                                className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-500 hover:text-indigo-600 transition-all"
                                                title={t('subscriptions.cancel')}
                                            >
                                                <XCircle size={18} />
                                            </button>
                                            <button className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-500 transition-all">
                                                <ChevronRight size={18} />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            {activeGroup.length === 0 && !isLoadingSubs && (
                                <tr>
                                    <td colSpan={6} className="px-6 py-20 text-center">
                                        <div className="max-w-xs mx-auto space-y-4">
                                            <div className="h-16 w-16 bg-gray-50 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto text-gray-300 dark:text-gray-600">
                                                <RefreshCcw size={32} />
                                            </div>
                                            <div className="space-y-1">
                                                <p className="font-bold text-gray-900 dark:text-white">{t('subscriptions.emptyTitle')}</p>
                                                <p className="text-sm text-gray-500 dark:text-gray-400">{t('subscriptions.emptyDesc')}</p>
                                            </div>
                                        </div>
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* cancelled Subscriptions List */}
            {cancelledGroup.length > 0 && (
                <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl shadow-sm overflow-hidden opacity-75">
                    <div className="p-6 border-b border-gray-100 dark:border-gray-800 flex flex-col md:flex-row md:items-center justify-between gap-4">
                        <h3 className="text-xl font-bold text-gray-500">{t('subscriptions.cancelledSubscriptions')}</h3>
                    </div>

                    <div className="overflow-x-auto">
                        <table className="w-full text-left">
                            <thead>
                                <tr className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest border-b border-gray-100 dark:border-gray-800">
                                    <th className="px-6 py-4">{t('subscriptions.merchantService')}</th>
                                    <th className="px-6 py-4">{t('subscriptions.statusHeader')}</th>
                                    <th className="px-6 py-4">{t('subscriptions.billing')}</th>
                                    <th className="px-6 py-4">{t('subscriptions.amountHeader')}</th>
                                    <th className="px-6 py-4">{t('subscriptions.nextPayment')}</th>
                                    <th className="px-6 py-4 text-right">{t('common.actions')}</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                                {cancelledGroup.map((s) => (
                                    <tr 
                                        key={s.id} 
                                        onClick={() => navigate(`/subscriptions/${s.id}`)}
                                        className="group hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
                                    >
                                        <td className="px-6 py-5">
                                            <div className="flex items-center gap-4">
                                                <div className="h-12 w-12 rounded-2xl bg-gray-100 dark:bg-gray-800 flex items-center justify-center font-bold text-gray-400 shrink-0 transition-colors">
                                                    {s.merchant_name.charAt(0)}
                                                </div>
                                                <div className="space-y-0.5">
                                                    <div className="font-bold text-gray-500 dark:text-gray-400 flex items-center gap-2">
                                                        {s.merchant_name}
                                                    </div>
                                                    <p className="text-sm text-gray-400 dark:text-gray-500">
                                                        {categories.find((c: Category) => c.id === s.category_id)?.name || t('common.uncategorized')}
                                                    </p>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-6 py-5">
                                            <StatusBadge status={s.status} />
                                        </td>
                                        <td className="px-6 py-5">
                                            <p className="text-sm font-medium capitalize text-gray-400">
                                                {s.billing_cycle}
                                                {s.billing_interval > 1 && ` ${t('subscriptions.everyXMonths', { count: s.billing_interval })}`}
                                            </p>
                                        </td>
                                        <td className="px-6 py-5">
                                            <div className="font-bold text-gray-400 dark:text-gray-500">
                                                {fmtCurrency(s.amount)}
                                            </div>
                                        </td>
                                        <td className="px-6 py-5">
                                            <div className="flex items-center gap-2 text-sm text-gray-400 dark:text-gray-500">
                                                <Clock size={14} />
                                                {s.next_occurrence ? fmtDate(s.next_occurrence) : '—'}
                                            </div>
                                        </td>
                                        <td className="px-6 py-5 text-right">
                                            <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <button 
                                                    onClick={(e) => handleDelete(e, s.id, s.merchant_name)}
                                                    disabled={deleteMutation.isPending}
                                                    className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-400 hover:text-red-600 transition-all"
                                                    title={t('subscriptions.deleteTitle')}
                                                >
                                                    <Trash2 size={18} />
                                                </button>
                                                <button className="p-2 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-400 transition-all">
                                                    <ChevronRight size={18} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}

            {/* Ignored Patterns List */}
            {declined.length > 0 && (
                <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl p-6 shadow-sm">
                    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
                        <div className="space-y-1">
                            <h3 className="text-xl font-bold flex items-center gap-2">
                                <XCircle className="text-gray-400" /> {t('subscriptions.ignoredPatterns')}
                            </h3>
                            <p className="text-sm text-gray-500 dark:text-gray-400">
                                {t('subscriptions.ignoredPatternsDesc')}
                            </p>
                        </div>
                    </div>
                    
                    <div className="flex flex-wrap gap-3">
                        {declined.map((d: DiscoveryFeedback, idx: number) => (
                            <div key={idx} className="group relative flex items-center gap-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-2.5 hover:border-indigo-200 dark:hover:border-indigo-900 transition-all">
                                <div className={`p-1 rounded-md ${d.source === 'AI' ? 'bg-indigo-50 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400' : 'bg-gray-100 dark:bg-gray-700 text-gray-500'}`}>
                                    {d.source === 'AI' ? <Sparkles size={14} /> : <User size={14} />}
                                </div>
                                <span className="text-sm font-bold text-gray-900 dark:text-gray-100">{d.merchant_name}</span>
                                <button 
                                    onClick={() => undeclineMutation.mutate(d.merchant_name)}
                                    disabled={undeclineMutation.isPending}
                                    className="ml-2 p-1.5 rounded-lg hover:bg-white dark:hover:bg-gray-700 shadow-sm border border-transparent hover:border-gray-200 dark:hover:border-gray-600 text-gray-400 hover:text-indigo-600 transition-all"
                                    title={t('subscriptions.restoreAndWhitelist')}
                                >
                                    <RotateCcw size={16} />
                                </button>
                                
                                {d.source === 'AI' && (
                                    <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-3 py-1 bg-gray-900 text-white text-[10px] font-bold rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap">
                                        {t('subscriptions.aiFiltered')}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
}

function StatusBadge({ status }: { status: string }) {
    const { t } = useTranslation();
    switch (status) {
        case 'active':
            return (
                <span className="inline-flex items-center gap-1.5 bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-400 px-3 py-1 rounded-full text-xs font-bold border border-green-100 dark:border-green-800/50">
                    <CheckCircle2 size={12} /> {t('subscriptions.status.active')}
                </span>
            );
        case 'cancellation_pending':
            return (
                <span className="inline-flex items-center gap-1.5 bg-amber-50 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 px-3 py-1 rounded-full text-xs font-bold border border-amber-100 dark:border-amber-800/50">
                    <Clock size={12} /> {t('subscriptions.status.cancellation_pending')}
                </span>
            );
        case 'cancelled':
            return (
                <span className="inline-flex items-center gap-1.5 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 px-3 py-1 rounded-full text-xs font-bold border border-gray-100 dark:border-gray-700">
                    <XCircle size={12} /> {t('subscriptions.status.cancelled')}
                </span>
            );
        case 'paused':
            return (
                <span className="inline-flex items-center gap-1.5 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 px-3 py-1 rounded-full text-xs font-bold border border-blue-100 dark:border-blue-800/50">
                    <Clock size={12} /> {t('subscriptions.status.paused')}
                </span>
            );
        default:
            return (
                <span className="inline-flex items-center gap-1.5 bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 px-3 py-1 rounded-full text-xs font-bold border border-blue-100 dark:border-blue-800/50">
                    {status}
                </span>
            );
    }
}

function DiscoverySettings() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [isOpen, setIsOpen] = useState(false);

    const { data: settings = {}, isLoading } = useQuery({
        queryKey: ['settings'],
        queryFn: settingsService.fetchSettings
    });

    const updateMutation = useMutation({
        mutationFn: settingsService.updateSettings,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['settings'] });
            queryClient.invalidateQueries({ queryKey: ['suggestedSubscriptions'] });
            setIsOpen(false);
        }
    });

    const [form, setForm] = useState({
        subscription_lookback_years: '3',
        subscription_discovery_amount_tolerance: '0.10',
        subscription_discovery_min_transactions_generic: '3',
        subscription_discovery_date_tolerance: '3.0'
    });

    // Sync form when settings are loaded
    useEffect(() => {
        if (Object.keys(settings).length > 0) {
            setForm({
                subscription_lookback_years: settings.subscription_lookback_years || '3',
                subscription_discovery_amount_tolerance: settings.subscription_discovery_amount_tolerance || '0.10',
                subscription_discovery_min_transactions_generic: settings.subscription_discovery_min_transactions_generic || '3',
                subscription_discovery_date_tolerance: settings.subscription_discovery_date_tolerance || '3.0'
            });
        }
    }, [settings]);

    const handleSave = () => {
        updateMutation.mutate(form);
    };

    if (isLoading) return null;

    return (
        <div className="space-y-4">
            <button 
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 text-sm font-bold text-gray-500 hover:text-indigo-600 transition-colors"
            >
                <Settings2 size={16} />
                {t('subscriptions.discoverySettings')}
            </button>

            {isOpen && (
                <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl p-6 shadow-sm animate-in slide-in-from-top-4 duration-300">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        {/* Lookback Years */}
                        <div className="space-y-3">
                            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest flex items-center gap-2">
                                <History size={14} />
                                {t('subscriptions.settings.lookback')}
                            </label>
                            <div className="flex items-center gap-3">
                                <input 
                                    type="range" min="1" max="10" 
                                    value={form.subscription_lookback_years}
                                    onChange={(e) => setForm({...form, subscription_lookback_years: e.target.value})}
                                    className="flex-1 accent-indigo-600"
                                />
                                <span className="font-bold min-w-[3ch]">{form.subscription_lookback_years}y</span>
                            </div>
                        </div>

                        {/* Amount Tolerance */}
                        <div className="space-y-3">
                            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">
                                {t('subscriptions.settings.amountTolerance')}
                            </label>
                            <div className="flex items-center gap-3">
                                <input 
                                    type="range" min="0.01" max="0.5" step="0.01"
                                    value={form.subscription_discovery_amount_tolerance}
                                    onChange={(e) => setForm({...form, subscription_discovery_amount_tolerance: e.target.value})}
                                    className="flex-1 accent-indigo-600"
                                />
                                <span className="font-bold min-w-[3ch]">{Math.round(parseFloat(form.subscription_discovery_amount_tolerance) * 100)}%</span>
                            </div>
                        </div>

                        {/* Generic Min Transactions */}
                        <div className="space-y-3">
                            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">
                                {t('subscriptions.settings.minTxGeneric')}
                            </label>
                            <div className="flex items-center gap-3">
                                <input 
                                    type="range" min="2" max="6" 
                                    value={form.subscription_discovery_min_transactions_generic}
                                    onChange={(e) => setForm({...form, subscription_discovery_min_transactions_generic: e.target.value})}
                                    className="flex-1 accent-indigo-600"
                                />
                                <span className="font-bold min-w-[3ch]">{form.subscription_discovery_min_transactions_generic}</span>
                            </div>
                        </div>

                        {/* Date Tolerance */}
                        <div className="space-y-3">
                            <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">
                                {t('subscriptions.settings.dateTolerance')}
                            </label>
                            <div className="flex items-center gap-3">
                                <input 
                                    type="range" min="1" max="7" step="0.5"
                                    value={form.subscription_discovery_date_tolerance}
                                    onChange={(e) => setForm({...form, subscription_discovery_date_tolerance: e.target.value})}
                                    className="flex-1 accent-indigo-600"
                                />
                                <span className="font-bold min-w-[3ch]">{form.subscription_discovery_date_tolerance}d</span>
                            </div>
                        </div>
                    </div>

                    <div className="mt-6 flex justify-end">
                        <button 
                            onClick={handleSave}
                            disabled={updateMutation.isPending}
                            className="bg-indigo-600 text-white px-6 py-2 rounded-xl font-bold text-sm hover:bg-indigo-700 transition-colors flex items-center gap-2 disabled:opacity-50"
                        >
                            <Save size={16} />
                            {t('common.saveChanges')}
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
