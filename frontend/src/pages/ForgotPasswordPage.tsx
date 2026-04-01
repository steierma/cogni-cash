import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Mail, ArrowLeft, CheckCircle2, AlertCircle } from 'lucide-react';
import { requestPasswordReset } from '../api/client';

const ForgotPasswordPage = () => {
    const { t } = useTranslation();
    const [email, setEmail] = useState('');
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        try {
            await requestPasswordReset(email);
            setSuccess(true);
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to request password reset.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 p-6">
            <div className="max-w-md w-full">
                <div className="bg-white dark:bg-gray-900 rounded-3xl shadow-xl p-8 border border-gray-100 dark:border-gray-800">
                    <div className="text-center mb-8">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-2xl mb-4">
                            <Mail size={32} />
                        </div>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('login.requestReset.title')}</h1>
                        <p className="text-gray-500 dark:text-gray-400 mt-2 text-sm">
                            {t('login.requestReset.subtitle')}
                        </p>
                    </div>

                    {success ? (
                        <div className="text-center space-y-6">
                            <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800/50 rounded-2xl flex items-start gap-3 text-sm text-green-700 dark:text-green-300 text-left">
                                <CheckCircle2 className="shrink-0 mt-0.5" size={18} />
                                <p>{t('login.requestReset.success')}</p>
                            </div>
                            <a
                                href="/login"
                                className="inline-flex items-center gap-2 text-indigo-600 dark:text-indigo-400 font-semibold hover:underline"
                            >
                                <ArrowLeft size={16} /> {t('login.requestReset.back')}
                            </a>
                        </div>
                    ) : (
                        <form onSubmit={handleSubmit} className="space-y-6">
                            {error && (
                                <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800/50 rounded-2xl flex items-center gap-3 text-sm text-red-600 dark:text-red-400">
                                    <AlertCircle className="shrink-0" size={18} />
                                    <p>{error}</p>
                                </div>
                            )}

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300 ml-1">
                                    {t('login.requestReset.email')}
                                </label>
                                <input
                                    type="email"
                                    required
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    placeholder={t('login.requestReset.emailPlaceholder')}
                                    className="w-full px-5 py-3 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 transition-all outline-none"
                                />
                            </div>

                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-bold py-4 rounded-2xl shadow-lg shadow-indigo-200 dark:shadow-none transition-all flex items-center justify-center gap-2"
                            >
                                {loading ? t('common.loading') : t('login.requestReset.send')}
                            </button>

                            <div className="text-center">
                                <a
                                    href="/login"
                                    className="inline-flex items-center gap-2 text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors text-sm font-medium"
                                >
                                    <ArrowLeft size={16} /> {t('login.requestReset.back')}
                                </a>
                            </div>
                        </form>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ForgotPasswordPage;
