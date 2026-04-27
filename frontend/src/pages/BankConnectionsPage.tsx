import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { bankService } from '../api/services/bankService';
import { authService } from '../api/services/authService';
import ShareBankAccountModal from '../components/ShareBankAccountModal';
import {
    Plus,
    Trash2,
    RefreshCw,
    ExternalLink,
    AlertCircle,
    Info,
    Landmark,
    Calendar,
    ChevronRight,
    History,
    Users,
    X,
    Loader2,
    Search
} from 'lucide-react';
import type { BankAccount } from "../api/types/bank";
import type { User } from "../api/types/system";
import { fmtCurrency } from '../utils/formatters';
import { LLMEnforcementWarning } from '../components/LLMEnforcementWarning';

export default function BankConnectionsPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [searchParams, setSearchParams] = useSearchParams();

    // ── State ──
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [isVirtualModalOpen, setIsVirtualModalOpen] = useState(false);
    const [selectedAccountToShare, setSelectedAccountToShare] = useState<BankAccount | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedCountry, setSelectedCountry] = useState('DE');

    // NEU: Zieht den Wert nun aus der Umgebungsvariablen. Default ist false.
    const isSandbox = import.meta.env.VITE_BANK_SANDBOX === 'true';

    const { data: user } = useQuery<User>({
        queryKey: ['user'],
        queryFn: authService.fetchMe,
    });

    const { data: connections = [], isLoading, isRefetching } = useQuery({
        queryKey: ['bank-connections'],
        queryFn: bankService.fetchConnections,
    });

    const { data: institutions = [] } = useQuery({
        queryKey: ['bank-institutions', selectedCountry],
        queryFn: () => bankService.fetchInstitutions(selectedCountry, isSandbox),
        enabled: isAddModalOpen,
    });

    const createConnMut = useMutation({
        mutationFn: ({ id, name, country }: { id: string, name: string, country: string }) => {
            const redirectUrl = window.location.origin + '/bank-connections';
            return bankService.createConnection(id, name, country, redirectUrl, isSandbox);
        },
        onSuccess: (data) => {
            if (data.auth_link) {
                window.location.href = data.auth_link;
            } else {
                queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
                setIsAddModalOpen(false);
            }
        },
    });

    const createVirtualMut = useMutation({
        mutationFn: ({ name, iban, currency, type, balance }: { name: string, iban: string, currency: string, type: string, balance: number }) =>
            bankService.createVirtualAccount(name, iban, currency, type, balance),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
            setIsVirtualModalOpen(false);
        },
    });

    const finishConnMut = useMutation({
        mutationFn: ({ reqId, code }: { reqId: string, code?: string }) => bankService.finishConnection(reqId, code),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
            setSearchParams({});
        },
    });

    const deleteConnMut = useMutation({
        mutationFn: (id: string) => bankService.deleteConnection(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
        },
    });

    const reauthConnMut = useMutation({
        mutationFn: (id: string) => {
            const redirectUrl = window.location.origin + '/bank-connections';
            return bankService.reauthenticateConnection(id, redirectUrl, isSandbox);
        },
        onSuccess: (data) => {
            if (data.auth_link) {
                window.location.href = data.auth_link;
            }
        },
    });

    const updateAccTypeMut = useMutation({
        mutationFn: ({ id, type }: { id: string, type: string }) => bankService.updateAccountType(id, type),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
        },
    });

    const syncAllMut = useMutation({
        mutationFn: () => bankService.syncAccounts(),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
            queryClient.invalidateQueries({ queryKey: ['bank-statements'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
        },
    });

    useEffect(() => {
        const reqId = searchParams.get('ref');
        const code = searchParams.get('code');
        if (reqId) {
            finishConnMut.mutate({ reqId, code: code || undefined });
        }
    }, [searchParams, finishConnMut]);

    const handleSyncAll = () => syncAllMut.mutate();

    const handleDelete = (id: string) => {
        if (window.confirm(t('bankConnections.deleteConfirm'))) {
            deleteConnMut.mutate(id);
        }
    };

    const filteredInstitutions = institutions.filter(inst =>
        inst.name.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const isSyncing = syncAllMut.isPending || isRefetching;

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            <LLMEnforcementWarning />
            {/* ── Header ── */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h2 className="text-3xl font-black text-gray-900 dark:text-white tracking-tight">
                        {t('bankConnections.title')}
                    </h2>
                    <p className="text-gray-500 dark:text-gray-400 mt-1 font-medium">
                        {t('bankConnections.subtitle')}
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    <button onClick={handleSyncAll} disabled={isSyncing || connections?.length === 0} className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-all shadow-sm disabled:opacity-50">
                        <RefreshCw size={18} className={isSyncing ? 'animate-spin' : ''} />
                        {isSyncing ? t('bankConnections.syncing') : t('bankConnections.syncAll')}
                    </button>
                    <button onClick={() => setIsVirtualModalOpen(true)} className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-all shadow-sm">
                        <Plus size={18} />
                        {t('bankConnections.addVirtual', 'Add Virtual Account')}
                    </button>
                    <button onClick={() => setIsAddModalOpen(true)} className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 transition-all shadow-md shadow-indigo-500/20">
                        <Plus size={18} />
                        {t('bankConnections.addConnection')}
                    </button>
                </div>
            </div>

            {/* ── Stats / Info ── */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700/50 shadow-sm">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2 bg-indigo-50 dark:bg-indigo-900/30 rounded-xl text-indigo-600 dark:text-indigo-400">
                            <Landmark size={20} />
                        </div>
                        <span className="text-sm font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">{t('bankConnections.stats.activeConnections')}</span>
                    </div>
                    <div className="text-3xl font-black text-gray-900 dark:text-white">
                        {connections.length}
                    </div>
                </div>

                <div className="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700/50 shadow-sm">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2 bg-emerald-50 dark:bg-emerald-900/30 rounded-xl text-emerald-600 dark:text-emerald-400">
                            <History size={20} />
                        </div>
                        <span className="text-sm font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">{t('bankConnections.stats.lastSync')}</span>
                    </div>
                    <div className="text-xl font-bold text-gray-900 dark:text-white">
                        {connections.some(c => c.accounts?.some(a => a.last_synced_at))
                            ? t('common.justNow')
                            : t('common.never')}
                    </div>
                </div>

                <div className="bg-white dark:bg-gray-800 p-6 rounded-3xl border border-gray-100 dark:border-gray-700/50 shadow-sm flex items-center gap-4">
                    <div className="p-3 bg-amber-50 dark:bg-amber-900/30 rounded-2xl text-amber-600 dark:text-amber-400">
                        <AlertCircle size={24} />
                    </div>
                    <div>
                        <p className="text-sm font-bold text-gray-900 dark:text-white">{t('bankConnections.psd2Info.title')}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('bankConnections.psd2Info.text')}</p>
                    </div>
                </div>
            </div>

            {/* ── Connections List ── */}
            {isLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {[1, 2, 3].map(i => (
                        <div key={i} className="h-64 bg-gray-100 dark:bg-gray-800 animate-pulse rounded-3xl border border-transparent" />
                    ))}
                </div>
            ) : connections.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-20 bg-white dark:bg-gray-800 rounded-3xl border-2 border-dashed border-gray-100 dark:border-gray-700">
                    <div className="p-4 bg-indigo-50 dark:bg-indigo-900/20 rounded-full text-indigo-600 dark:text-indigo-400 mb-4">
                        <Landmark size={48} strokeWidth={1.5} />
                    </div>
                    <h3 className="text-xl font-bold text-gray-900 dark:text-white">{t('bankConnections.noConnections')}</h3>
                    <p className="text-gray-500 dark:text-gray-400 mt-2 max-w-xs text-center font-medium">
                        {t('bankConnections.connectPrompt')}
                    </p>
                    <button
                        onClick={() => setIsAddModalOpen(true)}
                        className="mt-8 px-8 py-3 bg-indigo-600 text-white font-bold rounded-2xl hover:bg-indigo-700 transition-all shadow-xl shadow-indigo-500/20"
                    >
                        {t('bankConnections.addConnection')}
                    </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {connections.map((conn) => (
                        <div
                            key={conn.id}
                            className="group bg-white dark:bg-gray-800 rounded-3xl border border-gray-100 dark:border-gray-700/50 shadow-sm hover:shadow-xl hover:shadow-indigo-500/5 transition-all overflow-hidden flex flex-col"
                        >
                            <div className="p-6 flex-1">
                                <div className="flex items-start justify-between mb-6">
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 rounded-2xl bg-gray-50 dark:bg-gray-900 flex items-center justify-center border border-gray-100 dark:border-gray-700">
                                            <Landmark size={24} className="text-gray-400 group-hover:text-indigo-500 transition-colors" />
                                        </div>
                                        <div>
                                            <h4 className="font-bold text-gray-900 dark:text-white leading-tight">{conn.institution_name}</h4>
                                            <div className="flex items-center gap-2 mt-1">
                                                <span className={`w-2 h-2 rounded-full ${conn.status === 'linked' ? 'bg-emerald-500' : 'bg-amber-500'} shadow-sm`} />
                                                <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{conn.status}</span>
                                            </div>
                                        </div>
                                    </div>
                                    <button
                                        onClick={() => handleDelete(conn.id)}
                                        className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors opacity-0 group-hover:opacity-100"
                                    >
                                        <Trash2 size={18} />
                                    </button>
                                </div>

                                {conn.status !== 'initialized' && (
                                    <button
                                        onClick={() => reauthConnMut.mutate(conn.id)}
                                        disabled={reauthConnMut.isPending}
                                        className="mb-4 flex items-center justify-center gap-2 w-full py-2 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 text-sm font-medium rounded-xl hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors disabled:opacity-50"
                                    >
                                        {reauthConnMut.isPending ? <Loader2 size={16} className="animate-spin" /> : <RefreshCw size={16} />}
                                        {t('bankConnections.reauthenticate')}
                                    </button>
                                )}

                                <div className="space-y-3 pt-4 border-t border-gray-100 dark:border-gray-800">
                                    {conn.accounts && conn.accounts.length > 0 && (
                                        <div className="space-y-2 mb-4">
                                            {conn.accounts.map(acc => (
                                                <div key={acc.id} className="p-2.5 bg-gray-50/50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-700/50 space-y-2">
                                                    <div className="flex items-center justify-between">
                                                        <div className="min-w-0">
                                                            <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">{acc.name || 'Account'}</p>
                                                            <p className="text-[10px] text-gray-500 font-mono truncate">{acc.iban}</p>
                                                        </div>
                                                        <div className="text-right ml-2 flex flex-col items-end">
                                                            <p className="text-sm font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(acc.balance, acc.currency)}</p>
                                                            <div className="flex items-center gap-1.5 mt-1">
                                                                {acc.user_id === user?.id && (
                                                                    <button
                                                                        onClick={() => setSelectedAccountToShare(acc)}
                                                                        className="p-1 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-md transition-colors"
                                                                        title="Share account"
                                                                    >
                                                                        <Users size={14} />
                                                                    </button>
                                                                )}
                                                                {acc.is_shared && (
                                                                    <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 font-bold border border-indigo-100 dark:border-indigo-800/50">
                                                                        Shared
                                                                    </span>
                                                                )}
                                                            </div>
                                                        </div>
                                                    </div>
                                                    <div className="flex items-center gap-2 pt-1 border-t border-gray-100 dark:border-gray-700/50">
                                                        <span className="text-[10px] text-gray-400 dark:text-gray-500 uppercase font-bold tracking-wider">{t('bankStatements.columns.type')}:</span>
                                                        <select
                                                            value={acc.account_type || 'giro'}
                                                            onChange={(e) => updateAccTypeMut.mutate({ id: acc.id, type: e.target.value })}
                                                            disabled={updateAccTypeMut.isPending}
                                                            className="text-[10px] bg-transparent border-none p-0 focus:ring-0 font-medium text-indigo-600 dark:text-indigo-400 cursor-pointer outline-none"
                                                        >
                                                            <option value="giro" className="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">{t('reconcile.account.giro')}</option>
                                                            <option value="credit_card" className="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">{t('reconcile.account.credit_card')}</option>
                                                            <option value="extra_account" className="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">{t('reconcile.account.extra_account')}</option>
                                                        </select>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    )}

                                    <div className="flex items-center justify-between text-sm">
                                        <span className="text-gray-500 dark:text-gray-400 flex items-center gap-1.5">
                                            <Calendar size={14} />
                                            {t('bankConnections.expiresAt')}
                                        </span>
                                        <span className="font-medium text-gray-700 dark:text-gray-300">
                                            {conn.expires_at ? new Date(conn.expires_at).toLocaleDateString() : '—'}
                                        </span>
                                    </div>
                                </div>

                                {conn.status === 'initialized' && conn.auth_link && (
                                    <div className="mt-4">
                                        <a
                                            href={conn.auth_link}
                                            className="flex items-center justify-center gap-2 w-full py-2 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 text-sm font-medium rounded-xl hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors"
                                        >
                                            <ExternalLink size={16} />
                                            {t('bankConnections.completeAuth')}
                                        </a>
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {/* ── Add Connection Modal ── */}
            {isAddModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm animate-in fade-in duration-200">
                    <div className="bg-white dark:bg-gray-900 rounded-3xl shadow-2xl w-full max-w-2xl overflow-hidden animate-in zoom-in-95 duration-200 border border-gray-100 dark:border-gray-800">
                        <div className="flex items-center justify-between p-6 border-b border-gray-100 dark:border-gray-800">
                            <div>
                                <h3 className="text-xl font-bold text-gray-900 dark:text-white">{t('bankConnections.addConnection')}</h3>
                                <p className="text-sm text-gray-500 dark:text-gray-400 font-medium mt-0.5">{t('bankConnections.selectInstitution')}</p>
                            </div>
                            <button onClick={() => setIsAddModalOpen(false)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-2 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
                                <X size={24} />
                            </button>
                        </div>

                        <div className="p-6">
                            <div className="flex flex-col sm:flex-row gap-4 mb-6">
                                <div className="relative flex-1">
                                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" size={20} />
                                    <input
                                        type="text"
                                        placeholder={t('bankConnections.searchInstitutions')}
                                        className="w-full pl-12 pr-4 py-3 bg-gray-50 dark:bg-gray-800 border-none rounded-2xl outline-none focus:ring-2 focus:ring-indigo-500 font-medium"
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                    />
                                </div>
                                <select
                                    className="px-4 py-3 bg-gray-50 dark:bg-gray-800 border-none rounded-2xl outline-none focus:ring-2 focus:ring-indigo-500 font-bold"
                                    value={selectedCountry}
                                    onChange={(e) => setSelectedCountry(e.target.value)}
                                >
                                    <option value="DE">🇩🇪 Germany</option>
                                    <option value="AT">🇦🇹 Austria</option>
                                    <option value="CH">🇨🇭 Switzerland</option>
                                </select>
                            </div>

                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-h-[400px] overflow-y-auto pr-2 custom-scrollbar">
                                {filteredInstitutions.map((inst) => (
                                    <button
                                        key={inst.id}
                                        onClick={() => createConnMut.mutate({ id: inst.id, name: inst.name, country: selectedCountry })}
                                        disabled={createConnMut.isPending}
                                        className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800/50 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-2xl border border-transparent hover:border-indigo-100 dark:hover:border-indigo-800/50 transition-all group text-left disabled:opacity-50"
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className="w-10 h-10 rounded-xl bg-white dark:bg-gray-900 flex items-center justify-center border border-gray-100 dark:border-gray-800">
                                                <Landmark size={20} className="text-gray-400 group-hover:text-indigo-500" />
                                            </div>
                                            <div>
                                                <p className="font-bold text-gray-900 dark:text-white text-sm">{inst.name}</p>
                                                <p className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{inst.bic}</p>
                                            </div>
                                        </div>
                                        <ChevronRight size={18} className="text-gray-300 group-hover:text-indigo-400 transition-colors shrink-0" />
                                    </button>
                                ))}
                            </div>
                        </div>

                        <div className="p-6 bg-gray-50 dark:bg-gray-800/50 border-t border-gray-100 dark:border-gray-800">
                            <div className="flex items-start gap-3 text-xs text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-900 p-4 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm">
                                <Info size={16} className="text-indigo-500 shrink-0 mt-0.5" />
                                <p>{t('bankConnections.authRedirect')}</p>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* ── Virtual Account Modal ── */}
            {isVirtualModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm animate-in fade-in duration-200">
                    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in zoom-in-95 duration-200">
                        <div className="flex items-center justify-between p-6 border-b border-gray-100 dark:border-gray-800">
                            <h3 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('bankConnections.addVirtual', 'Add Virtual Account')}</h3>
                            <button onClick={() => setIsVirtualModalOpen(false)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-2 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
                                <X size={20} />
                            </button>
                        </div>

                        <form onSubmit={(e) => {
                            e.preventDefault();
                            const formData = new FormData(e.currentTarget);
                            createVirtualMut.mutate({
                                name: formData.get('name') as string,
                                iban: formData.get('iban') as string,
                                currency: formData.get('currency') as string,
                                type: formData.get('type') as string,
                                balance: parseFloat(formData.get('balance') as string || '0')
                            });
                        }} className="p-6 space-y-4">
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name</label>
                                <input name="name" required className="w-full px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500" />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">IBAN</label>
                                <input name="iban" required className="w-full px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500 font-mono" />
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Currency</label>
                                    <input name="currency" defaultValue="EUR" required className="w-full px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Type</label>
                                    <select name="type" className="w-full px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500">
                                        <option value="giro">Giro</option>
                                        <option value="credit_card">Credit Card</option>
                                        <option value="extra_account">Extra Account</option>
                                    </select>
                                </div>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Initial Balance</label>
                                <input name="balance" type="number" step="0.01" defaultValue="0.00" className="w-full px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl outline-none focus:ring-2 focus:ring-indigo-500" />
                            </div>

                            <button
                                type="submit"
                                disabled={createVirtualMut.isPending}
                                className="w-full py-3 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 transition-all shadow-md shadow-indigo-500/20 disabled:opacity-50 flex items-center justify-center gap-2"
                            >
                                {createVirtualMut.isPending ? <Loader2 size={20} className="animate-spin" /> : <Plus size={20} />}
                                Create Virtual Account
                            </button>
                        </form>
                    </div>
                </div>
            )}

            {/* ── Share Account Modal ── */}
            {selectedAccountToShare && (
                <ShareBankAccountModal
                    account={selectedAccountToShare!}
                    onClose={() => setSelectedAccountToShare(null)}
                />
            )}
        </div>
    );
}
