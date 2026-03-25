import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { AlertCircle, Lock, LogIn } from 'lucide-react';
import { login } from '../api/client';

export default function LoginPage() {
    const { t } = useTranslation();
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const navigate = useNavigate();

    const loginMut = useMutation({
        mutationFn: () => login(username, password),
        onSuccess: (data) => {
            localStorage.setItem('auth_token', data.token);
            navigate('/');
        },
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (username.trim() && password.trim()) {
            loginMut.mutate();
        }
    };

    return (
        <div className="min-h-[80vh] flex flex-col items-center justify-center animate-in fade-in duration-300">
            <div className="w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-8">
                <div className="flex flex-col items-center mb-8">
                    <div className="bg-indigo-50 dark:bg-indigo-900/20 p-3 rounded-full mb-4">
                        <Lock size={28} className="text-indigo-600 dark:text-indigo-400" />
                    </div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('login.title')}</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('login.subtitle')}</p>
                </div>

                {loginMut.isError && (
                    <div className="mb-6 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl flex items-center gap-2 text-red-700 dark:text-red-400 text-sm">
                        <AlertCircle size={16} />
                        {t('login.invalidCredentials')}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-5">
                    <div className="space-y-1.5">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('login.username')}</label>
                        <input
                            type="text"
                            required
                            autoFocus
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 transition-shadow"
                            placeholder="admin"
                        />
                    </div>

                    <div className="space-y-1.5">
                        <label className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('login.password')}</label>
                        <input
                            type="password"
                            required
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 transition-shadow"
                            placeholder="••••••••"
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={loginMut.isPending || !username || !password}
                        className="w-full flex items-center justify-center gap-2 py-2.5 px-4 mt-2 bg-indigo-600 text-white font-medium rounded-xl hover:bg-indigo-700 focus:outline-none focus:ring-4 focus:ring-indigo-100 dark:focus:ring-indigo-900 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
                    >
                        {loginMut.isPending ? t('login.authenticating') : (
                            <>
                                <LogIn size={18} />
                                {t('login.signIn')}
                            </>
                        )}
                    </button>
                </form>
            </div>
        </div>
    );
}