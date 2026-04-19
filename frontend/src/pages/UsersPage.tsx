import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { userService } from '../api/services/userService';
import { authService } from '../api/services/authService';
import type { User } from "../api/types/system";
import {Plus, Edit2, ShieldAlert, Mail, MapPin, Search, Trash2, Users} from 'lucide-react';

export default function UsersPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const [searchQuery, setSearchQuery] = useState('');
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingUser, setEditingUser] = useState<User | null>(null);

    const [username, setUsername] = useState('');
    const [email, setEmail] = useState('');
    const [fullName, setFullName] = useState('');
    const [address, setAddress] = useState('');
    const [role, setRole] = useState('manager');
    const [password, setPassword] = useState('');
    const [errorMsg, setErrorMsg] = useState('');

    const { data: users, isLoading } = useQuery<User[]>({
        queryKey: ['users', searchQuery],
        queryFn: () => userService.fetchUsers(searchQuery),
    });

    const { data: currentUser } = useQuery<User>({
        queryKey: ['currentUser'],
        queryFn: () => authService.fetchMe(),
    });

    const createMutation = useMutation({
        mutationFn: (newUserData: Partial<User> & { password?: string }) => userService.createUser(newUserData),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] });
            closeModal();
        },
        onError: (err: any) => setErrorMsg(err.response?.data?.error || 'Failed to create user')
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<User> }) => userService.updateUser(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['users'] });
            closeModal();
        },
        onError: (err: any) => setErrorMsg(err.response?.data?.error || 'Failed to update user')
    });

    const delMutation = useMutation({
        mutationFn: (id: string) => userService.deleteUser(id),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] })
    });

    const openCreateModal = () => {
        setEditingUser(null);
        setUsername('');
        setEmail('');
        setFullName('');
        setAddress('');
        setRole('manager');
        setPassword('');
        setErrorMsg('');
        setIsModalOpen(true);
    };

    const openEditModal = (user: User) => {
        setEditingUser(user);
        setUsername(user.username);
        setEmail(user.email);
        setFullName(user.full_name);
        setAddress(user.address);
        setRole(user.role);
        setPassword('');
        setErrorMsg('');
        setIsModalOpen(true);
    };

    const handleDelete = (id: string) => {
        if (window.confirm(t('users.deleteConfirm'))) {
            delMutation.mutate(id);
        }
    };

    const closeModal = () => setIsModalOpen(false);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setErrorMsg('');

        if (editingUser) {
            updateMutation.mutate({
                id: editingUser.id,
                data: { username, email, full_name: fullName, address, role },
            });
        } else {
            if (!password) {
                setErrorMsg('Password is required for new users');
                return;
            }
            createMutation.mutate({
                username, email, full_name: fullName, address, role, password
            });
        }
    };

    return (
        <div className="max-w-6xl mx-auto space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <div className="flex justify-between items-center">
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                            <Users className="text-indigo-600 dark:text-indigo-400" /> {t('users.title')}
                        </h1>
                    </div>
                    <p className="text-gray-500 dark:text-gray-400 text-sm mt-1">{t('users.subtitle')}</p>
                </div>
                <button onClick={openCreateModal} className="flex items-center gap-2 bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700 transition-colors shadow-sm">
                    <Plus size={18} /><span>{t('users.addUser')}</span>
                </button>
            </div>

            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-sm overflow-hidden">
                <div className="p-4 border-b border-gray-200 dark:border-gray-800 flex items-center bg-gray-50/50 dark:bg-gray-800/50">
                    <div className="relative w-full max-w-md">
                        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input
                            type="text"
                            placeholder={t('users.search')}
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="w-full pl-9 pr-4 py-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:text-gray-100"
                        />
                    </div>
                </div>

                <div className="overflow-x-auto">
                    <table className="w-full text-left text-sm whitespace-nowrap">
                        <thead className="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                        <tr>
                            <th className="px-6 py-4 font-medium text-gray-500 dark:text-gray-400">{t('users.user')}</th>
                            <th className="px-6 py-4 font-medium text-gray-500 dark:text-gray-400">{t('users.contact')}</th>
                            <th className="px-6 py-4 font-medium text-gray-500 dark:text-gray-400">{t('users.role')}</th>
                            <th className="px-6 py-4 font-medium text-gray-500 dark:text-gray-400 text-right">{t('users.actions')}</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                        {isLoading ? (
                            <tr><td colSpan={4} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">{t('users.loading')}</td></tr>
                        ) : users?.length === 0 ? (
                            <tr><td colSpan={4} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">{t('users.noUsers')}</td></tr>
                        ) : (
                            users?.map((user) => (
                                <tr key={user.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                    <td className="px-6 py-4">
                                        <div className="font-medium text-gray-900 dark:text-gray-100 flex items-center gap-2">
                                            {user.full_name}
                                            {currentUser?.id === user.id && (
                                                <span className="text-[10px] uppercase bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-300 px-1.5 py-0.5 rounded-sm font-bold tracking-wider">{t('users.you')}</span>
                                            )}
                                        </div>
                                        <div className="text-gray-500 dark:text-gray-400 text-xs mt-0.5">@{user.username}</div>
                                    </td>
                                    <td className="px-6 py-4 text-gray-600 dark:text-gray-400">
                                        <div className="flex items-center gap-1.5"><Mail size={14} /> {user.email}</div>
                                        {user.address && <div className="flex items-center gap-1.5 text-xs mt-1"><MapPin size={14} /> {user.address}</div>}
                                    </td>
                                    <td className="px-6 py-4">
                                            <span className={`inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium border ${user.role === 'admin' ? 'bg-purple-50 text-purple-700 border-purple-200 dark:bg-purple-900/20 dark:text-purple-400 dark:border-purple-800/50' : 'bg-gray-50 text-gray-700 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700'}`}>
                                                {user.role === 'admin' && <ShieldAlert size={12} />}
                                                <span className="capitalize">{user.role}</span>
                                            </span>
                                    </td>
                                    <td className="px-6 py-4 text-right space-x-2">
                                        <button onClick={() => openEditModal(user)} title={t('users.editUser')} className="text-indigo-600 dark:text-indigo-400 hover:text-indigo-900 dark:hover:text-indigo-300 p-2 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/20 transition-colors inline-block">
                                            <Edit2 size={16} />
                                        </button>
                                        <button
                                            onClick={() => handleDelete(user.id)}
                                            disabled={currentUser?.id === user.id || delMutation.isPending}
                                            title={currentUser?.id === user.id ? t('users.cannotDeleteSelf') : t('users.deleteUser')}
                                            className="text-red-500 hover:text-red-700 p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors inline-block disabled:opacity-30 disabled:hover:bg-transparent"
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </td>
                                </tr>
                            ))
                        )}
                        </tbody>
                    </table>
                </div>
            </div>

            {isModalOpen && (
                <div className="fixed inset-0 bg-gray-900/50 dark:bg-gray-950/80 z-50 flex items-center justify-center p-4">
                    <div className="bg-white dark:bg-gray-900 w-full max-w-md rounded-xl shadow-xl border border-gray-200 dark:border-gray-800 flex flex-col max-h-[90vh]">
                        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between shrink-0">
                            <h2 className="text-lg font-semibold">{editingUser ? t('users.editTitle') : t('users.createTitle')}</h2>
                            <button onClick={closeModal} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">&times;</button>
                        </div>

                        <div className="p-6 overflow-y-auto">
                            {errorMsg && (
                                <div className="mb-4 p-3 bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400 rounded-lg text-sm border border-red-100 dark:border-red-900/50">
                                    {errorMsg}
                                </div>
                            )}

                            <form id="user-form" onSubmit={handleSubmit} className="space-y-4">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.username')}</label>
                                    <input required type="text" value={username} onChange={e => setUsername(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.email')}</label>
                                    <input required type="email" value={email} onChange={e => setEmail(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.fullName')}</label>
                                    <input required type="text" value={fullName} onChange={e => setFullName(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.address')}</label>
                                    <input required type="text" value={address} onChange={e => setAddress(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.role')}</label>
                                    <select value={role} disabled={editingUser?.id === currentUser?.id} onChange={e => setRole(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50">
                                        <option value="manager">{t('users.roleManager')}</option>
                                        <option value="admin">{t('users.roleAdmin')}</option>
                                    </select>
                                    {editingUser?.id === currentUser?.id && (
                                        <p className="text-xs text-gray-500 mt-1">{t('users.roleWarning')}</p>
                                    )}
                                </div>

                                {!editingUser && (
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('users.initialPwd')}</label>
                                        <input required type="password" value={password} onChange={e => setPassword(e.target.value)} className="w-full px-3 py-2 bg-white dark:bg-gray-900 border border-gray-300 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                                    </div>
                                )}
                            </form>
                        </div>

                        <div className="px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50 flex justify-end gap-3 shrink-0 rounded-b-xl">
                            <button type="button" onClick={closeModal} className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg font-medium transition-colors">
                                {t('users.cancel')}
                            </button>
                            <button type="submit" form="user-form" disabled={createMutation.isPending || updateMutation.isPending} className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 font-medium transition-colors disabled:opacity-50">
                                {editingUser ? t('users.saveChanges') : t('users.createUser')}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}