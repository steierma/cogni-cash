import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { bankService } from '../api/services/bankService';
import { settingsService } from '../api/services/settingsService';
import { authService } from '../api/services/authService';
import {
    Plus,
    Trash2,
    RefreshCw,
    ExternalLink,
    Search,
    X,
    Loader2,
    AlertCircle,
    Info,
    Landmark,
    Calendar,
    ChevronRight,
    History
} from 'lucide-react';
import type { BankConnection, BankInstitution } from "../api/types/bank";
import { fmtCurrency } from '../utils/formatters';

export default function BankConnectionsPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [searchParams, setSearchParams] = useSearchParams();

    // ── State ──
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedCountry, setSelectedCountry] = useState('DE');
    const [historyDays, setHistoryDays] = useState('90');

    // NEU: Zieht den Wert nun aus der Umgebungsvariablen. Default ist false.
    const [isSandbox, setIsSandbox] = useState(import.meta.env.VITE_ENABLE_SANDBOX === 'true');
    const [isSyncing, setIsSyncing] = useState(false);

    // ── Queries ──
    const { data: connections, isLoading: isConnsLoading } = useQuery<BankConnection[]>({
        queryKey: ['bank-connections'],
        queryFn: bankService.fetchConnections,
    });

    const { data: user } = useQuery<any>({
        queryKey: ['me'],
        queryFn: authService.fetchMe,
    });

    const { data: systemInfo } = useQuery({
        queryKey: ['system-info'],
        queryFn: settingsService.fetchSystemInfo,
        enabled: user?.role === 'admin',
    });

    const { data: settings } = useQuery({
        queryKey: ['settings'],
        queryFn: settingsService.fetchSettings,
    });

    const { data: institutions, isLoading: isInstsLoading, error: instsError } = useQuery<BankInstitution[]>({
        queryKey: ['bank-institutions', selectedCountry, isSandbox],
        queryFn: () => bankService.fetchInstitutions(selectedCountry, isSandbox),
        enabled: isAddModalOpen,
        retry: false,
    });

    // ── Mutations ──
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

    const updateAccTypeMut = useMutation({
        mutationFn: ({ id, type }: { id: string, type: string }) => bankService.updateAccountType(id, type),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-connections'] });
        },
    });

    const syncMut = useMutation({
        mutationFn: bankService.syncAccounts,
        onSuccess: () => {
            setIsSyncing(false);
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
            queryClient.invalidateQueries({ queryKey: ['bank-statements'] });
        },
        onError: () => {
            setIsSyncing(false);
        }
    });

    const updateSettingsMut = useMutation({
        mutationFn: settingsService.updateSettings,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['settings'] });
        }
    });

    // ── Effects ──
    useEffect(() => {
        const gcReqId = searchParams.get('requisition_id');
        const ebState = searchParams.get('state');
        const ebCode = searchParams.get('code');

        if (gcReqId) {
            finishConnMut.mutate({ reqId: gcReqId });
        } else if (ebState && ebCode) {
            finishConnMut.mutate({ reqId: ebState, code: ebCode });
        }
    }, [searchParams]);

    useEffect(() => {
        if (settings && settings['bank_sync_history_days']) {
            setHistoryDays(settings['bank_sync_history_days']);
        }
    }, [settings]);

    // ── Handlers ──
    const handleSyncAll = () => {
        setIsSyncing(true);
        syncMut.mutate();
    };

    const handleDelete = (id: string) => {
        if (confirm(t('bankConnections.deleteConfirm'))) {
            deleteConnMut.mutate(id);
        }
    };

    const handleHistoryDaysChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setHistoryDays(e.target.value);
    };

    const handleHistoryDaysBlur = () => {
        if (settings?.['bank_sync_history_days'] !== historyDays) {
            updateSettingsMut.mutate({ bank_sync_history_days: historyDays });
        }
    };

    const filteredInstitutions = institutions?.filter(inst =>
        inst.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        inst.bic.toLowerCase().includes(searchTerm.toLowerCase())
    ) || [];

    if (isConnsLoading) {
        return (
            <div className="flex flex-col items-center justify-center h-64 text-gray-500">
                <Loader2 size={32} className="animate-spin text-indigo-500 mb-4" />
                <p>{t('common.loading')}</p>
            </div>
        );
    }

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <div className="flex items-center gap-3">
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                            <Landmark className="text-indigo-600 dark:text-indigo-400" /> {t('bankConnections.title')}
                        </h1>
                        <span className="px-2.5 py-0.5 rounded-full bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 text-xs font-semibold border border-indigo-100 dark:border-indigo-800/50 uppercase tracking-wider">
                            {systemInfo?.bank_provider || '...'}
                        </span>
                    </div>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('bankConnections.subtitle')}</p>
                </div>
                <div className="flex items-center gap-3">
                    {/* Settings Control for History Days */}
                    <div className="flex items-center gap-2 px-3 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-sm">
                        <History size={16} className="text-gray-400" />
                        <span className="text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">{t('bankConnections.historyDays', 'History (days):')}</span>
                        <input
                            type="number"
                            min="1"
                            max="730"
                            value={historyDays}
                            onChange={handleHistoryDaysChange}
                            onBlur={handleHistoryDaysBlur}
                            className="w-14 bg-transparent border-none p-0 text-sm font-medium text-gray-900 dark:text-gray-100 focus:ring-0 text-right outline-none"
                        />
                    </div>

                    <button onClick={handleSyncAll} disabled={isSyncing || connections?.length === 0} className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-all shadow-sm disabled:opacity-50">
                        <RefreshCw size={18} className={isSyncing ? 'animate-spin' : ''} />
                        {isSyncing ? t('bankConnections.syncing') : t('bankConnections.syncAll')}
                    </button>
                    <button onClick={() => setIsAddModalOpen(true)} className="flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-xl hover:bg-indigo-700 transition-all shadow-md shadow-indigo-500/20">
                        <Plus size={18} />
                        {t('bankConnections.addConnection')}
                    </button>
                </div>
            </div>

            {connections?.length === 0 ? (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-12 text-center shadow-sm">
                    <div className="w-16 h-16 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-2xl flex items-center justify-center mx-auto mb-4">
                        <Landmark size={32} />
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">{t('bankConnections.noConnections')}</h3>
                    <button onClick={() => setIsAddModalOpen(true)} className="text-indigo-600 dark:text-indigo-400 font-medium hover:underline">
                        {t('bankConnections.addConnection')}
                    </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {connections?.map(conn => (
                        <div key={conn.id} className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-6 shadow-sm hover:shadow-md transition-shadow relative group">
                            <div className="flex items-start justify-between mb-4">
                                <div className="flex items-center gap-3">
                                    <div className="w-10 h-10 bg-gray-100 dark:bg-gray-800 rounded-xl flex items-center justify-center text-indigo-600 dark:text-indigo-400">
                                        <Landmark size={20} />
                                    </div>
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <h4 className="font-semibold text-gray-900 dark:text-gray-100">{conn.institution_name}</h4>
                                            <span className="text-[10px] px-1.5 py-0.5 rounded-md bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 font-mono">
                                                {conn.provider === 'enablebanking' ? 'EB' : 'GC'}
                                            </span>
                                        </div>
                                        <div className="flex items-center gap-1.5 mt-0.5">
                                            <span className={`w-2 h-2 rounded-full ${
                                                conn.status === 'linked' ? 'bg-green-500' :
                                                    conn.status === 'initialized' ? 'bg-yellow-500' : 'bg-red-500'
                                            }`} />
                                            <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                                                {t(`bankConnections.${conn.status}`)}
                                            </span>
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
                                                    <div className="text-right ml-2">
                                                        <p className="text-sm font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(acc.balance, acc.currency)}</p>
                                                        {acc.last_synced_at && (
                                                            <p className="text-[10px] text-gray-400 dark:text-gray-500">
                                                                {new Date(acc.last_synced_at).toLocaleDateString()}
                                                            </p>
                                                        )}
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
                                        {t('bankConnections.connectBank')}
                                    </a>
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            )}

            {/* ── Add Connection Modal ── */}
            {isAddModalOpen && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm animate-in fade-in duration-200">
                    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200">
                        <div className="flex items-center justify-between p-6 border-b border-gray-100 dark:border-gray-800">
                            <div>
                                <h3 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('bankConnections.addConnection')}</h3>
                                <p className="text-sm text-gray-500 mt-1">{t('bankConnections.selectInstitution')}</p>
                            </div>
                            <button onClick={() => setIsAddModalOpen(false)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-2 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
                                <X size={20} />
                            </button>
                        </div>

                        <div className="p-6 space-y-4 flex-1 overflow-y-auto">
                            <div className="flex flex-wrap gap-3">
                                <div className="relative flex-1 min-w-[200px]">
                                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                                    <input
                                        type="text"
                                        placeholder={t('bankConnections.searchInstitution')}
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="w-full pl-10 pr-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl focus:ring-2 focus:ring-indigo-500 outline-none transition-shadow"
                                    />
                                </div>

                                <select
                                    value={selectedCountry}
                                    onChange={(e) => setSelectedCountry(e.target.value)}
                                    className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-4 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-shadow"
                                >
                                    <option value="DE">DE</option>
                                    <option value="FI">FI</option>
                                    <option value="AT">AT</option>
                                    <option value="CH">CH</option>
                                    <option value="FR">FR</option>
                                    <option value="ES">ES</option>
                                    <option value="GB">GB</option>
                                </select>

                                {/* Sandbox Toggle Button - Only visible if enabled via environment */}
                                {import.meta.env.VITE_ENABLE_SANDBOX === 'true' && (
                                    <label className="flex items-center gap-2 px-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                                        <input
                                            type="checkbox"
                                            checked={isSandbox}
                                            onChange={(e) => setIsSandbox(e.target.checked)}
                                            className="w-4 h-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500 dark:bg-gray-900 dark:border-gray-600"
                                        />
                                        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Sandbox</span>
                                    </label>
                                )}
                            </div>

                            {isInstsLoading ? (
                                <div className="py-12 flex flex-col items-center justify-center text-gray-500">
                                    <Loader2 size={32} className="animate-spin text-indigo-500 mb-4" />
                                    <p>{t('common.loading')}</p>
                                </div>
                            ) : instsError ? (
                                <div className="py-12 flex flex-col items-center justify-center text-center p-6 bg-red-50 dark:bg-red-900/10 rounded-2xl border border-red-100 dark:border-red-900/20">
                                    <AlertCircle size={32} className="text-red-500 mb-4" />
                                    <h4 className="text-lg font-semibold text-red-900 dark:text-red-400 mb-1">{t('bankConnections.error')}</h4>
                                    <p className="text-sm text-red-700 dark:text-red-300">
                                        {(instsError as any)?.response?.data?.error || (instsError as Error).message}
                                    </p>
                                </div>
                            ) : (
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                                    {filteredInstitutions.map(inst => (
                                        <button
                                            key={inst.id}
                                            onClick={() => createConnMut.mutate({ id: inst.id, name: inst.name, country: inst.country })}
                                            disabled={createConnMut.isPending}
                                            className="flex items-center gap-3 p-4 border border-gray-100 dark:border-gray-800 rounded-2xl hover:border-indigo-200 dark:hover:border-indigo-800 hover:bg-indigo-50/50 dark:hover:bg-indigo-900/20 text-left transition-all group disabled:opacity-50"
                                        >
                                            <div className="w-10 h-10 bg-white dark:bg-gray-800 border border-gray-100 dark:border-gray-700 rounded-xl flex items-center justify-center overflow-hidden shrink-0">
                                                {inst.logo ? (
                                                    <img src={inst.logo} alt={inst.name} className="w-full h-full object-contain p-1" />
                                                ) : (
                                                    <Landmark size={20} className="text-gray-400" />
                                                )}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <p className="font-medium text-gray-900 dark:text-gray-100 truncate">{inst.name}</p>
                                                <p className="text-xs text-gray-500 truncate">{inst.bic}</p>
                                            </div>
                                            <ChevronRight size={18} className="text-gray-300 group-hover:text-indigo-400 transition-colors shrink-0" />
                                        </button>
                                    ))}
                                </div>
                            )}
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
        </div>
    );
}