import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { X, UserPlus, Trash2, Shield, Loader2 } from 'lucide-react';
import { userService } from '../api/services/userService';
import { bankService } from '../api/services/bankService';
import type { BankAccount } from "../api/types/bank";

interface ShareBankAccountModalProps {
    account: BankAccount;
    onClose: () => void;
}

export default function ShareBankAccountModal({ account, onClose }: ShareBankAccountModalProps) {
    const { t } = useTranslation();
    const qc = useQueryClient();
    const [selectedUserId, setSelectedUserId] = useState<string>('');
    const [permission, setPermission] = useState<'view' | 'edit'>('view');

    const { data: users = [] } = useQuery({
        queryKey: ['users'],
        queryFn: () => userService.fetchUsers(),
    });

    const { data: sharedUserIds = [], isLoading: isLoadingShares } = useQuery({
        queryKey: ['bank-account-shares', account.id],
        queryFn: () => bankService.listShares(account.id),
    });

    const shareMut = useMutation({
        mutationFn: ({ userId, perm }: { userId: string; perm: 'view' | 'edit' }) => 
            bankService.shareAccount(account.id, userId, perm),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['bank-account-shares', account.id] });
            qc.invalidateQueries({ queryKey: ['bank-connections'] });
            setSelectedUserId('');
        },
    });

    const revokeMut = useMutation({
        mutationFn: (userId: string) => bankService.revokeShare(account.id, userId),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['bank-account-shares', account.id] });
            qc.invalidateQueries({ queryKey: ['bank-connections'] });
        },
    });

    const availableUsers = users.filter(
        u => u.id !== account.owner_id && !sharedUserIds.includes(u.id)
    );

    const activeParticipants = users.filter(u => sharedUserIds.includes(u.id));

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl w-full max-w-md shadow-2xl border border-gray-200 dark:border-gray-800 overflow-hidden">
                <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Shield className="text-indigo-600 dark:text-indigo-400" size={20} />
                        Share Account: {account.name}
                    </h3>
                    <button
                        onClick={onClose}
                        className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-xl transition-colors"
                    >
                        <X size={20} />
                    </button>
                </div>

                <div className="p-6 space-y-6">
                    <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">
                        Sharing this bank account will also share all its linked transactions, statements, and recurring forecast items.
                    </p>

                    <div className="space-y-3">
                        <label className="block text-xs font-bold text-gray-400 uppercase tracking-wider">
                            {t('categories.inviteUser')}
                        </label>
                        <div className="flex flex-col gap-3">
                            <select
                                value={selectedUserId}
                                onChange={(e) => setSelectedUserId(e.target.value)}
                                className="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-sm focus:ring-2 focus:ring-indigo-500 outline-none transition-all"
                            >
                                <option value="">{t('categories.selectUser')}</option>
                                {availableUsers.map((u) => (
                                    <option key={u.id} value={u.id}>
                                        {u.full_name || u.username} ({u.email})
                                    </option>
                                ))}
                            </select>

                            <div className="flex items-center gap-2">
                                <select
                                    value={permission}
                                    onChange={(e) => setPermission(e.target.value as 'view' | 'edit')}
                                    className="flex-1 px-4 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-sm focus:ring-2 focus:ring-indigo-500 outline-none transition-all"
                                >
                                    <option value="view">{t('categories.permissionView')}</option>
                                    <option value="edit">{t('categories.permissionEdit')}</option>
                                </select>

                                <button
                                    onClick={() => shareMut.mutate({ userId: selectedUserId, perm: permission })}
                                    disabled={!selectedUserId || shareMut.isPending}
                                    className="flex items-center justify-center gap-2 px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-bold rounded-xl transition-all shadow-lg shadow-indigo-500/20"
                                >
                                    {shareMut.isPending ? <Loader2 size={18} className="animate-spin" /> : <UserPlus size={18} />}
                                    {t('common.add')}
                                </button>
                            </div>
                        </div>
                    </div>

                    <div className="space-y-4 pt-4">
                        <label className="block text-xs font-bold text-gray-400 uppercase tracking-wider">
                            {t('categories.activeParticipants')}
                        </label>
                        
                        {isLoadingShares ? (
                            <div className="flex justify-center py-4">
                                <Loader2 size={24} className="animate-spin text-indigo-500" />
                            </div>
                        ) : activeParticipants.length === 0 ? (
                            <div className="text-center py-6 border-2 border-dashed border-gray-100 dark:border-gray-800 rounded-2xl">
                                <p className="text-sm text-gray-400">{t('categories.noParticipants')}</p>
                            </div>
                        ) : (
                            <div className="space-y-2">
                                {activeParticipants.map((user) => (
                                    <div 
                                        key={user.id}
                                        className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-800 group"
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/50 flex items-center justify-center text-indigo-600 dark:text-indigo-400 font-bold text-xs">
                                                {user.username.substring(0, 2).toUpperCase()}
                                            </div>
                                            <div>
                                                <div className="text-sm font-semibold text-gray-900 dark:text-gray-100">
                                                    {user.full_name || user.username}
                                                </div>
                                                <div className="text-[10px] text-gray-500 dark:text-gray-400">
                                                    {user.email}
                                                </div>
                                            </div>
                                        </div>
                                        <button
                                            onClick={() => revokeMut.mutate(user.id)}
                                            disabled={revokeMut.isPending}
                                            className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-all opacity-0 group-hover:opacity-100"
                                            title={t('categories.revokeAccess')}
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>

                <div className="px-6 py-4 bg-gray-50 dark:bg-gray-800/50 border-t border-gray-100 dark:border-gray-800 flex justify-end">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
                    >
                        {t('common.close')}
                    </button>
                </div>
            </div>
        </div>
    );
}
