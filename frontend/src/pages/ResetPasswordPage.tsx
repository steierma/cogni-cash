import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { KeyRound, CheckCircle2, AlertCircle, Eye, EyeOff } from 'lucide-react';
import { authService } from '../api/services/authService';

const ResetPasswordPage = () => {
    const { t } = useTranslation();
    const [searchParams] = useSearchParams();
    const token = searchParams.get('token');

    const [isValidating, setIsValidating] = useState(true);
    const [isTokenValid, setIsTokenValid] = useState(false);
    
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [showPassword, setShowPassword] = useState(false);
    
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        const checkToken = async () => {
            if (!token) {
                setIsTokenValid(false);
                setIsValidating(false);
                return;
            }
            try {
                const res = await authService.validateResetToken(token);
                setIsTokenValid(res.valid);
            } catch (err) {
                setIsTokenValid(false);
            } finally {
                setIsValidating(false);
            }
        };
        checkToken();
    }, [token]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (newPassword !== confirmPassword) {
            setError(t('login.resetConfirm.mismatch'));
            return;
        }

        setLoading(true);
        setError('');
        try {
            if (!token) throw new Error('Missing token');
            await authService.confirmPasswordReset(token, newPassword);
            setSuccess(true);
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to reset password.');
        } finally {
            setLoading(false);
        }
    };

    if (isValidating) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950 p-6">
            <div className="max-w-md w-full">
                <div className="bg-white dark:bg-gray-900 rounded-3xl shadow-xl p-8 border border-gray-100 dark:border-gray-800">
                    <div className="text-center mb-8">
                        <div className="inline-flex items-center justify-center w-16 h-16 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-2xl mb-4">
                            <KeyRound size={32} />
                        </div>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('login.resetConfirm.title')}</h1>
                        <p className="text-gray-500 dark:text-gray-400 mt-2 text-sm">
                            {t('login.resetConfirm.subtitle')}
                        </p>
                    </div>

                    {!isTokenValid ? (
                        <div className="text-center space-y-6">
                            <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800/50 rounded-2xl flex items-start gap-3 text-sm text-red-600 dark:text-red-400 text-left">
                                <AlertCircle className="shrink-0 mt-0.5" size={18} />
                                <p>{t('login.resetConfirm.invalidToken')}</p>
                            </div>
                            <a
                                href="/login"
                                className="inline-block bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-semibold px-6 py-3 rounded-2xl hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                            >
                                {t('login.requestReset.back')}
                            </a>
                        </div>
                    ) : success ? (
                        <div className="text-center space-y-6">
                            <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-100 dark:border-green-800/50 rounded-2xl flex items-start gap-3 text-sm text-green-700 dark:text-green-300 text-left">
                                <CheckCircle2 className="shrink-0 mt-0.5" size={18} />
                                <p>{t('login.resetConfirm.success')}</p>
                            </div>
                            <a
                                href="/login"
                                className="w-full inline-block bg-indigo-600 hover:bg-indigo-700 text-white font-bold py-4 rounded-2xl shadow-lg transition-all text-center"
                            >
                                {t('login.signIn')}
                            </a>
                        </div>
                    ) : (
                        <form onSubmit={handleSubmit} className="space-y-5">
                            {error && (
                                <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-100 dark:border-red-800/50 rounded-2xl flex items-center gap-3 text-sm text-red-600 dark:text-red-400">
                                    <AlertCircle className="shrink-0" size={18} />
                                    <p>{error}</p>
                                </div>
                            )}

                            <div className="space-y-2 relative">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300 ml-1">
                                    {t('login.resetConfirm.newPassword')}
                                </label>
                                <div className="relative">
                                    <input
                                        type={showPassword ? "text" : "password"}
                                        required
                                        value={newPassword}
                                        onChange={(e) => setNewPassword(e.target.value)}
                                        className="w-full px-5 py-3 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 transition-all outline-none pr-12"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowPassword(!showPassword)}
                                        className="absolute right-4 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                                    >
                                        {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                                    </button>
                                </div>
                            </div>

                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-700 dark:text-gray-300 ml-1">
                                    {t('login.resetConfirm.confirmPassword')}
                                </label>
                                <input
                                    type={showPassword ? "text" : "password"}
                                    required
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    className="w-full px-5 py-3 rounded-2xl border border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 transition-all outline-none"
                                />
                            </div>

                            <button
                                type="submit"
                                disabled={loading}
                                className="w-full bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white font-bold py-4 rounded-2xl shadow-lg shadow-indigo-200 dark:shadow-none transition-all flex items-center justify-center gap-2 mt-2"
                            >
                                {loading ? t('common.loading') : t('login.resetConfirm.submit')}
                            </button>
                        </form>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ResetPasswordPage;
