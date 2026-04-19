import { useState, useEffect } from 'react';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { 
    ArrowLeft, 
    Calendar, 
    CreditCard, 
    Mail, 
    Globe, 
    Phone, 
    ExternalLink, 
    History, 
    XCircle, 
    Clock, 
    Send,
    AlertTriangle,
    CheckCircle2,
    RefreshCcw,
    Shield,
    Trash2,
    Edit3,
    Save,
    X,
    Sparkles
} from 'lucide-react';
import { subscriptionService } from '../api/services/subscriptionService';
import { transactionService } from '../api/services/transactionService';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import type { 
    Subscription, 
    CancellationLetterResult, 
    SubscriptionEvent,
    SubscriptionStatus
} from '../api/types/subscription';

export default function SubscriptionDetailPage() {
    const { id } = useParams<{ id: string }>();
    const navigate = useNavigate();
    const { t, i18n } = useTranslation();
    const queryClient = useQueryClient();
    const [showCancelModal, setShowCancelModal] = useState(false);
    const [cancellationDraft, setCancellationDraft] = useState<CancellationLetterResult | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [editForm, setEditForm] = useState({
        merchant_name: '',
        amount: 0,
        billing_cycle: 'monthly' as 'monthly' | 'yearly',
        status: 'active' as SubscriptionStatus,
        customer_number: '',
        contact_email: '',
        contact_phone: '',
        contact_website: '',
        support_url: '',
        cancellation_url: '',
        is_trial: false,
        notes: ''
    });

    const { data: subscription, isLoading: isLoadingSub } = useQuery<Subscription>({
        queryKey: ['subscription', id],
        queryFn: () => subscriptionService.getSubscription(id!),
        enabled: !!id
    });

    const { data: transactions = [] } = useQuery({
        queryKey: ['subscriptionTransactions', id],
        queryFn: () => transactionService.fetchTransactions(undefined, false, undefined, undefined, undefined, false, false, id),
        enabled: !!id
    });

    const { data: events = [] } = useQuery<SubscriptionEvent[]>({
        queryKey: ['subscriptionEvents', id],
        queryFn: () => subscriptionService.fetchSubscriptionEvents(id!),
        enabled: !!id
    });

    const totalSpent = transactions.reduce((acc, tx) => acc + Math.abs(tx.amount), 0);

    const previewMutation = useMutation({
        mutationFn: () => subscriptionService.previewCancellation(id!, i18n.language.toUpperCase()),
        onSuccess: (data) => setCancellationDraft(data)
    });

    const cancelMutation = useMutation({
        mutationFn: (data: { subject: string, body: string }) => subscriptionService.cancelSubscription(id!, data.subject, data.body),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscription', id] });
            queryClient.invalidateQueries({ queryKey: ['subscriptionEvents', id] });
            setShowCancelModal(false);
            setCancellationDraft(null);
        }
    });

    const deleteMutation = useMutation({
        mutationFn: () => subscriptionService.deleteSubscription(id!),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
            navigate('/subscriptions');
        }
    });

    const updateMutation = useMutation({
        mutationFn: (data: typeof editForm) => subscriptionService.updateSubscription(id!, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscription', id] });
            queryClient.invalidateQueries({ queryKey: ['subscriptions'] });
            setIsEditing(false);
        }
    });

    const [searchParams, setSearchParams] = useSearchParams();
    const [showEnrichedBadge, setShowEnrichedBadge] = useState(false);

    useEffect(() => {
        if (searchParams.get('enriched') === 'true') {
            setShowEnrichedBadge(true);
            const timer = setTimeout(() => {
                setShowEnrichedBadge(false);
                // Clear the param without a full navigation
                searchParams.delete('enriched');
                setSearchParams(searchParams, { replace: true });
            }, 10000);
            return () => clearTimeout(timer);
        }

        if (searchParams.get('cancel') === 'true') {
            setShowCancelModal(true);
            previewMutation.mutate();
            // Clear the param
            searchParams.delete('cancel');
            setSearchParams(searchParams, { replace: true });
        }
    }, [searchParams, setSearchParams, previewMutation]);

    const enrichMutation = useMutation({
        mutationFn: () => subscriptionService.enrichSubscription(id!),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['subscription', id] });
            setShowEnrichedBadge(true);
            setTimeout(() => setShowEnrichedBadge(false), 10000);
        }
    });

    const startEditing = () => {
        if (!subscription) return;
        setEditForm({
            merchant_name: subscription.merchant_name,
            amount: subscription.amount,
            billing_cycle: subscription.billing_cycle as 'monthly' | 'yearly',
            status: subscription.status,
            customer_number: subscription.customer_number || '',
            contact_email: subscription.contact_email || '',
            contact_phone: subscription.contact_phone || '',
            contact_website: subscription.contact_website || '',
            support_url: subscription.support_url || '',
            cancellation_url: subscription.cancellation_url || '',
            is_trial: subscription.is_trial,
            notes: subscription.notes || ''
        });
        setIsEditing(true);
    };

    const handleSave = (e: React.FormEvent) => {
        e.preventDefault();
        updateMutation.mutate(editForm);
    };

    const handleDelete = () => {
        if (window.confirm(t('subscriptions.deleteConfirm', { name: subscription?.merchant_name }))) {
            deleteMutation.mutate();
        }
    };

    if (isLoadingSub || !subscription) return null;

    return (
        <div className="max-w-5xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                <div className="flex items-center gap-4 flex-1">
                    <button 
                        onClick={() => navigate('/subscriptions')}
                        className="p-2.5 rounded-xl bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 text-gray-500 hover:text-indigo-600 transition-all shadow-sm shrink-0"
                    >
                        <ArrowLeft size={20} />
                    </button>
                    <div className="space-y-1 flex-1">
                        <div className="flex items-center gap-3">
                            {isEditing ? (
                                <input
                                    type="text"
                                    value={editForm.merchant_name}
                                    onChange={(e) => setEditForm({ ...editForm, merchant_name: e.target.value })}
                                    className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white bg-transparent border-b-2 border-indigo-500 focus:outline-none w-full max-w-lg"
                                    autoFocus
                                />
                            ) : (
                                <h1 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white">
                                    {subscription.merchant_name}
                                </h1>
                            )}
                            {!isEditing && <StatusBadge status={subscription.status} />}
                            {showEnrichedBadge && (
                                <span className="inline-flex items-center gap-1.5 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 px-3 py-1 rounded-full text-xs font-bold border border-indigo-100 dark:border-indigo-800/50 animate-bounce">
                                    <Sparkles size={12} /> {t('subscriptions.justEnriched')}
                                </span>
                            )}
                        </div>
                        {showEnrichedBadge && (
                            <p className="text-xs font-bold text-indigo-600 dark:text-indigo-400 animate-in fade-in slide-in-from-left-2">
                                {t('subscriptions.enrichSuccess')}
                            </p>
                        )}
                        {!isEditing && !showEnrichedBadge && (
                            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-gray-500 dark:text-gray-400">
                                <div className="flex items-center gap-2">
                                    <CreditCard size={14} />
                                    {fmtCurrency(subscription.amount)} / {subscription.billing_cycle}
                                </div>
                                <div className="h-1 w-1 rounded-full bg-gray-300 dark:bg-gray-700 hidden md:block" />
                                <div className="flex items-center gap-2">
                                    <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400">{t('subscriptions.totalSpent')}:</span>
                                    <span className="font-bold text-gray-700 dark:text-gray-300">{fmtCurrency(totalSpent)}</span>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
                
                <div className="flex items-center gap-3 shrink-0">
                    {isEditing ? (
                        <>
                            <button 
                                onClick={handleSave}
                                disabled={updateMutation.isPending}
                                className="px-6 py-2.5 bg-indigo-600 text-white rounded-xl font-bold flex items-center gap-2 hover:bg-indigo-700 transition-all shadow-lg shadow-indigo-200 dark:shadow-none"
                            >
                                {updateMutation.isPending ? <RefreshCcw size={18} className="animate-spin" /> : <Save size={18} />}
                                {t('common.save')}
                            </button>
                            <button 
                                onClick={() => setIsEditing(false)}
                                className="px-4 py-2.5 bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400 rounded-xl font-bold flex items-center gap-2 hover:bg-gray-200 dark:hover:bg-gray-700 transition-all"
                            >
                                <X size={18} />
                                {t('common.cancel')}
                            </button>
                        </>
                    ) : (
                        <button 
                            onClick={startEditing}
                            className="px-6 py-2.5 bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700 rounded-xl font-bold flex items-center gap-2 hover:border-indigo-500 transition-all shadow-sm"
                        >
                            <Edit3 size={18} />
                            {t('common.edit')}
                        </button>
                    )}
                    {!isEditing && (
                        <button 
                            onClick={() => enrichMutation.mutate()}
                            disabled={enrichMutation.isPending}
                            className="px-6 py-2.5 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 border border-indigo-100 dark:border-indigo-800/50 rounded-xl font-bold flex items-center gap-2 hover:bg-indigo-100 dark:hover:bg-indigo-900/30 transition-all shadow-sm disabled:opacity-50"
                        >
                            <Sparkles size={18} className={enrichMutation.isPending ? "animate-spin" : ""} />
                            {enrichMutation.isPending ? t('subscriptions.enriching') : t('subscriptions.enrichWithAI')}
                        </button>
                    )}
                </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Left: Metadata & Contact */}
                <div className="lg:col-span-2 space-y-8">
                    {/* Information Grid */}
                    <div className={`bg-white dark:bg-gray-900 border rounded-3xl p-6 md:p-8 shadow-sm grid grid-cols-1 md:grid-cols-2 gap-8 transition-all duration-1000 ${showEnrichedBadge ? 'border-indigo-500 shadow-indigo-100 dark:shadow-none scale-[1.01]' : 'border-gray-200 dark:border-gray-800'}`}>
                        <div className="space-y-6">
                            {isEditing ? (
                                <div className="space-y-4">
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.amountLabel')}</label>
                                        <input
                                            type="number"
                                            step="0.01"
                                            value={editForm.amount}
                                            onChange={(e) => setEditForm({ ...editForm, amount: parseFloat(e.target.value) })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.billingCycleLabel')}</label>
                                        <select
                                            value={editForm.billing_cycle}
                                            onChange={(e) => setEditForm({ ...editForm, billing_cycle: e.target.value as 'monthly' | 'yearly' })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold appearance-none"
                                        >
                                            <option value="monthly">{t('forecasting.interval.monthly')}</option>
                                            <option value="yearly">{t('forecasting.interval.yearly')}</option>
                                        </select>
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.statusLabel')}</label>
                                        <select
                                            value={editForm.status}
                                            onChange={(e) => setEditForm({ ...editForm, status: e.target.value as SubscriptionStatus })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold appearance-none"
                                        >
                                            <option value="active">{t('subscriptions.status.active')}</option>
                                            <option value="canceled">{t('subscriptions.status.canceledPast')}</option>
                                            <option value="paused">{t('subscriptions.status.paused')}</option>
                                        </select>
                                    </div>
                                </div>
                            ) : (
                                <>
                                    <InfoItem 
                                        icon={<Calendar size={18} />} 
                                        label={t('subscriptions.billingCycle')} 
                                        value={`${subscription.billing_cycle}${subscription.billing_interval > 1 ? ` ${t('subscriptions.everyXMonths', { count: subscription.billing_interval })}` : ''}`} 
                                    />
                                    <InfoItem 
                                        icon={<Shield size={18} />} 
                                        label={t('subscriptions.noticePeriod')} 
                                        value={`${subscription.notice_period_days} ${t('common.days')}`} 
                                    />

                                </>
                            )}
                        </div>
                        <div className="space-y-6">
                            {isEditing ? (
                                <div className="space-y-4">
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.customerNumber')}</label>
                                        <input
                                            type="text"
                                            value={editForm.customer_number}
                                            onChange={(e) => setEditForm({ ...editForm, customer_number: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.website')}</label>
                                        <input
                                            type="text"
                                            value={editForm.contact_website}
                                            onChange={(e) => setEditForm({ ...editForm, contact_website: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.contactEmail')}</label>
                                        <input
                                            type="email"
                                            value={editForm.contact_email}
                                            onChange={(e) => setEditForm({ ...editForm, contact_email: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.phone')}</label>
                                        <input
                                            type="text"
                                            value={editForm.contact_phone}
                                            onChange={(e) => setEditForm({ ...editForm, contact_phone: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                </div>
                            ) : (
                                <>
                                    <InfoItem 
                                        icon={<Shield size={18} />} 
                                        label={t('subscriptions.customerNumber')} 
                                        value={subscription.customer_number} 
                                    />
                                    <InfoItem 
                                        icon={<Globe size={18} />} 
                                        label={t('subscriptions.website')} 
                                        value={subscription.contact_website} 
                                        isLink 
                                    />
                                    <InfoItem 
                                        icon={<Mail size={18} />} 
                                        label={t('subscriptions.contactEmail')} 
                                        value={subscription.contact_email} 
                                    />
                                    <InfoItem 
                                        icon={<Phone size={18} />} 
                                        label={t('subscriptions.phone')} 
                                        value={subscription.contact_phone} 
                                    />
                                </>
                            )}
                        </div>
                    </div>

                    {/* Extended Metadata */}
                    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl p-6 md:p-8 shadow-sm">
                        {isEditing ? (
                            <div className="space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.supportUrl')}</label>
                                        <input
                                            type="text"
                                            value={editForm.support_url}
                                            onChange={(e) => setEditForm({ ...editForm, support_url: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                    <div className="space-y-1">
                                        <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.cancellationUrl')}</label>
                                        <input
                                            type="text"
                                            value={editForm.cancellation_url}
                                            onChange={(e) => setEditForm({ ...editForm, cancellation_url: e.target.value })}
                                            className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                        />
                                    </div>
                                </div>
                                <div className="flex items-center gap-3">
                                    <input
                                        type="checkbox"
                                        id="is_trial"
                                        checked={editForm.is_trial}
                                        onChange={(e) => setEditForm({ ...editForm, is_trial: e.target.checked })}
                                        className="h-5 w-5 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
                                    />
                                    <label htmlFor="is_trial" className="text-sm font-bold text-gray-700 dark:text-gray-300 cursor-pointer">
                                        {t('subscriptions.trial')}
                                    </label>
                                </div>
                                <div className="space-y-1">
                                    <label className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.notes')}</label>
                                    <textarea
                                        rows={4}
                                        value={editForm.notes}
                                        onChange={(e) => setEditForm({ ...editForm, notes: e.target.value })}
                                        className="w-full p-3 bg-gray-50 dark:bg-gray-800 rounded-xl border-none focus:ring-2 focus:ring-indigo-500 font-medium text-sm"
                                    />
                                </div>
                            </div>
                        ) : (
                            <div className="space-y-6">
                                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                                    <InfoItem 
                                        icon={<Shield size={18} />} 
                                        label={t('subscriptions.supportUrl')} 
                                        value={subscription.support_url} 
                                        isLink
                                    />
                                    <InfoItem 
                                        icon={<XCircle size={18} />} 
                                        label={t('subscriptions.cancellationUrl')} 
                                        value={subscription.cancellation_url} 
                                        isLink
                                    />
                                </div>
                                {subscription.is_trial && (
                                    <div className="flex items-center gap-2 bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800/50 rounded-xl px-4 py-2 w-fit">
                                        <Shield className="text-amber-600 dark:text-amber-400" size={16} />
                                        <span className="text-sm font-bold text-amber-700 dark:text-amber-400 uppercase tracking-wider">{t('subscriptions.trial')}</span>
                                    </div>
                                )}
                                {subscription.notes && (
                                    <div className="space-y-1.5">
                                        <p className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{t('subscriptions.notes')}</p>
                                        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-2xl text-sm text-gray-700 dark:text-gray-300 leading-relaxed whitespace-pre-wrap">
                                            {subscription.notes}
                                        </div>
                                    </div>
                                )}
                                {!subscription.support_url && !subscription.cancellation_url && !subscription.is_trial && !subscription.notes && (
                                    <p className="text-sm text-gray-400 italic text-center py-4">{t('common.noData')}</p>
                                )}
                            </div>
                        )}
                    </div>

                    {/* History: Transactions */}
                    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl shadow-sm overflow-hidden">
                        <div className="p-6 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
                            <h3 className="text-lg font-bold flex items-center gap-2">
                                <History size={18} className="text-gray-400" />
                                {t('subscriptions.paymentHistory')}
                            </h3>
                            <span className="text-xs font-bold text-gray-400 bg-gray-50 dark:bg-gray-800 px-2.5 py-1 rounded-full uppercase tracking-widest">
                                {transactions.length} {t('common.all')}
                            </span>
                        </div>
                        <div className="divide-y divide-gray-100 dark:divide-gray-800">
                            {transactions.map(tx => (
                                <div key={tx.id} className="p-4 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                    <div className="space-y-0.5">
                                        <p className="text-sm font-bold text-gray-900 dark:text-white">{fmtDate(tx.booking_date)}</p>
                                        <p className="text-xs text-gray-500 dark:text-gray-400 truncate max-w-[200px] md:max-w-md">{tx.description}</p>
                                    </div>
                                    <p className="font-mono font-bold text-gray-900 dark:text-white">
                                        {fmtCurrency(tx.amount)}
                                    </p>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {/* Right: Actions & Audit Trail */}
                <div className="space-y-8">
                    {/* Primary Actions */}
                    <div className="bg-indigo-600 rounded-3xl p-6 text-white shadow-xl shadow-indigo-500/20 space-y-4">
                        <h3 className="font-bold flex items-center gap-2">
                            <Shield size={18} />
                            {t('subscriptions.management')}
                        </h3>
                        <p className="text-sm text-indigo-100 leading-relaxed">
                            {t('subscriptions.managementDesc')}
                        </p>
                        <button 
                            onClick={() => { setShowCancelModal(true); previewMutation.mutate(); }}
                            disabled={subscription.status === 'canceled' || subscription.status === 'cancellation_pending'}
                            className="w-full bg-white text-indigo-600 py-3 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-indigo-50 transition-colors disabled:opacity-50"
                        >
                            <XCircle size={18} />
                            {t('subscriptions.cancel')}
                        </button>
                        <button 
                            onClick={handleDelete}
                            disabled={deleteMutation.isPending}
                            className="w-full border border-white/20 py-3 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-white/10 transition-colors disabled:opacity-50"
                        >
                            <Trash2 size={18} />
                            {t('subscriptions.deleteLocal')}
                        </button>
                        {subscription.cancellation_url && (
                            <a 
                                href={subscription.cancellation_url} 
                                target="_blank" 
                                rel="noopener noreferrer"
                                className="w-full border border-white/20 py-3 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-white/10 transition-colors"
                            >
                                <ExternalLink size={18} />
                                {t('subscriptions.cancellationUrl')}
                            </a>
                        )}
                    </div>

                    {/* Audit Trail */}
                    <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-3xl p-6 shadow-sm">
                        <h3 className="text-sm font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-6">{t('subscriptions.activityLog')}</h3>
                        <div className="space-y-6">
                            {events.length === 0 ? (
                                <p className="text-sm text-gray-400 text-center py-4 italic">{t('subscriptions.emptyEvents')}</p>
                            ) : (
                                events.map(e => (
                                    <div key={e.id} className="relative pl-6 border-l-2 border-gray-100 dark:border-gray-800 space-y-1">
                                        <div className="absolute -left-[9px] top-0 h-4 w-4 rounded-full bg-white dark:bg-gray-900 border-2 border-indigo-500" />
                                        <p className="text-xs text-gray-400 dark:text-gray-500 font-bold uppercase tracking-tighter">{fmtDate(e.created_at)}</p>
                                        <p className="text-sm font-bold text-gray-900 dark:text-white leading-tight">{e.title}</p>
                                        <p className="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{e.content}</p>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                </div>
            </div>

            {/* Cancellation Modal */}
            {showCancelModal && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 md:p-8">
                    <div className="absolute inset-0 bg-gray-900/60 backdrop-blur-md" onClick={() => setShowCancelModal(false)} />
                    <div className="relative bg-white dark:bg-gray-900 rounded-[2rem] w-full max-w-2xl max-h-[90dvh] flex flex-col shadow-2xl animate-in zoom-in-95 duration-200 border border-gray-200 dark:border-gray-800">
                        <div className="p-8 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
                            <div className="space-y-1">
                                <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{t('subscriptions.cancelModalTitle')}</h2>
                                <p className="text-sm text-gray-500 dark:text-gray-400">{t('subscriptions.cancelModalSubtitle')}</p>
                            </div>
                            <button onClick={() => setShowCancelModal(false)} className="p-2 text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors">
                                <XCircle size={24} />
                            </button>
                        </div>

                        <div className="flex-1 overflow-y-auto p-8 space-y-6">
                            {previewMutation.isPending ? (
                                <div className="flex flex-col items-center justify-center py-20 gap-4 text-gray-400">
                                    <RefreshCcw size={40} className="animate-spin text-indigo-500" />
                                    <p className="font-bold">{t('subscriptions.cancelDrafting')}</p>
                                </div>
                            ) : cancellationDraft ? (
                                <div className="space-y-6 animate-in fade-in duration-300">
                                    <div className="space-y-2">
                                        <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">{t('subscriptions.letterRecipient')}</label>
                                        <div className="p-4 bg-gray-50 dark:bg-gray-800 rounded-2xl flex items-center gap-3 border border-gray-100 dark:border-gray-700">
                                            <Mail size={18} className="text-indigo-500" />
                                            <span className="font-bold">{subscription.contact_email}</span>
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">{t('subscriptions.letterSubject')}</label>
                                        <input 
                                            type="text" 
                                            value={cancellationDraft.subject}
                                            onChange={(e) => setCancellationDraft({ ...cancellationDraft, subject: e.target.value })}
                                            className="w-full p-4 bg-gray-50 dark:bg-gray-800 rounded-2xl font-bold border-none focus:ring-2 focus:ring-indigo-500 transition-all"
                                        />
                                    </div>
                                    <div className="space-y-2">
                                        <label className="text-xs font-bold text-gray-400 uppercase tracking-widest">{t('subscriptions.letterBody')}</label>
                                        <textarea 
                                            rows={12}
                                            value={cancellationDraft.body}
                                            onChange={(e) => setCancellationDraft({ ...cancellationDraft, body: e.target.value })}
                                            className="w-full p-4 bg-gray-50 dark:bg-gray-800 rounded-2xl text-sm leading-relaxed border-none focus:ring-2 focus:ring-indigo-500 transition-all resize-none"
                                        />
                                    </div>
                                    <div className="p-4 bg-amber-50 dark:bg-amber-900/20 border border-amber-100 dark:border-amber-800 rounded-2xl flex gap-3 text-amber-800 dark:text-amber-400 text-xs">
                                        <AlertTriangle size={18} className="shrink-0" />
                                        <p className="leading-normal">
                                            {t('subscriptions.cancelConfirm')}
                                        </p>
                                    </div>
                                </div>
                            ) : (
                                <div className="text-center py-10">
                                    <p className="text-red-500 font-bold">{t('subscriptions.cancelDraftError')}</p>
                                </div>
                            )}
                        </div>

                        <div className="p-8 border-t border-gray-100 dark:border-gray-800 flex gap-4">
                            <button 
                                onClick={() => setShowCancelModal(false)}
                                className="flex-1 px-6 py-4 rounded-2xl font-bold text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                            >
                                {t('common.cancel')}
                            </button>
                            <button 
                                onClick={() => cancellationDraft && cancelMutation.mutate(cancellationDraft)}
                                disabled={cancelMutation.isPending || !cancellationDraft}
                                className="flex-[2] bg-indigo-600 text-white px-6 py-4 rounded-2xl font-bold flex items-center justify-center gap-2 hover:bg-indigo-700 transition-all shadow-lg shadow-indigo-500/25 disabled:opacity-50"
                            >
                                {cancelMutation.isPending ? <RefreshCcw size={20} className="animate-spin" /> : <Send size={20} />}
                                {t('subscriptions.cancelSend')}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

function InfoItem({ icon, label, value, isLink }: { icon: React.ReactNode, label: string, value?: string, isLink?: boolean }) {
    if (!value) return null;
    return (
        <div className="flex gap-4">
            <div className="h-10 w-10 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center text-gray-400 shrink-0">
                {icon}
            </div>
            <div className="space-y-0.5">
                <p className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">{label}</p>
                {isLink ? (
                    <a href={value.startsWith('http') ? value : `https://${value}`} target="_blank" rel="noopener noreferrer" className="text-sm font-bold text-indigo-600 dark:text-indigo-400 flex items-center gap-1 hover:underline">
                        {value.replace(/^https?:\/\//, '')} <ExternalLink size={12} />
                    </a>
                ) : (
                    <p className="text-sm font-bold text-gray-900 dark:text-white truncate">{value}</p>
                )}
            </div>
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
        case 'canceled':
            return (
                <span className="inline-flex items-center gap-1.5 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 px-3 py-1 rounded-full text-xs font-bold border border-gray-100 dark:border-gray-700">
                    <XCircle size={12} /> {t('subscriptions.status.cancelled')}
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
