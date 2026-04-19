import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import { 
    Users, Shield, PieChart,
    ChevronRight, ArrowRight, User, Wallet, Loader2 
} from 'lucide-react';
import { sharingService } from '../api/services/sharingService';
import { fmtCurrency } from '../utils/formatters';
import CategoryBadge from '../components/CategoryBadge';
import type { CategoryBalance, UserSpending, SharedCategorySummary } from "../api/types/category";

export default function SharingDashboard() {
    const { t } = useTranslation();
    const navigate = useNavigate();

    const { data: dashboard, isLoading, error } = useQuery({
        queryKey: ['sharing-dashboard'],
        queryFn: sharingService.fetchSharingDashboard,
        refetchInterval: 30000, // Refresh every 30s
    });

    if (isLoading) {
        return (
            <div className="flex flex-col items-center justify-center py-32 animate-in fade-in duration-500">
                <Loader2 size={48} className="text-indigo-500 animate-spin mb-4" />
                <p className="text-gray-500 font-medium">{t('common.loading')}</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="max-w-4xl mx-auto py-20 text-center space-y-4">
                <div className="bg-red-50 dark:bg-red-900/20 p-4 rounded-2xl inline-block">
                    <Shield className="text-red-500" size={32} />
                </div>
                <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('common.error')}</h2>
                <p className="text-gray-500 dark:text-gray-400">{(error as Error).message}</p>
            </div>
        );
    }

    const { shared_categories = [], shared_invoices = [], balances = [] } = dashboard || {};

    return (
        <div className="max-w-7xl mx-auto space-y-8 pb-20 animate-in fade-in duration-300">
            {/* Header */}
            <div className="space-y-1">
                <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                    <Users className="text-indigo-600 dark:text-indigo-400" /> 
                    {t('sharing.dashboardTitle')}
                </h1>
                <p className="text-sm text-gray-500 dark:text-gray-400">
                    {t('sharing.dashboardSubtitle')}
                </p>
            </div>

            {/* Balances: Who Paid What */}
            <section className="space-y-4">
                <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                    <Wallet className="text-emerald-500" size={20} />
                    {t('sharing.collaborativeBalances')}
                </h2>
                
                {balances.length === 0 ? (
                    <div className="bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-200 dark:border-gray-800 p-12 text-center space-y-3">
                        <PieChart className="mx-auto text-gray-300 dark:text-gray-700" size={48} />
                        <p className="text-gray-500 dark:text-gray-400 font-medium">{t('sharing.noBalancesYet')}</p>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {balances.map((balance: CategoryBalance) => (
                            <div 
                                key={balance.category_id}
                                className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col"
                            >
                                <div className="p-5 border-b border-gray-50 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30">
                                    <div className="flex justify-between items-start mb-1">
                                        <h3 className="font-bold text-gray-900 dark:text-gray-100 truncate pr-2">
                                            {balance.category_name}
                                        </h3>
                                        <div className="text-xs font-mono font-bold text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded-full">
                                            {fmtCurrency(balance.total_spent)}
                                        </div>
                                    </div>
                                    <p className="text-[10px] text-gray-400 uppercase tracking-widest font-bold">
                                        {t('sharing.totalCategorySpend')}
                                    </p>
                                </div>
                                <div className="p-5 flex-1 space-y-4">
                                    {balance.user_breakdown.map((user: UserSpending) => (
                                        <div key={user.user_id} className="space-y-1.5">
                                            <div className="flex justify-between text-xs font-medium">
                                                <span className="text-gray-600 dark:text-gray-400 flex items-center gap-1.5">
                                                    <User size={12} className="text-gray-400" />
                                                    {user.username}
                                                </span>
                                                <span className="text-gray-900 dark:text-gray-100 font-bold">{fmtCurrency(user.amount)}</span>
                                            </div>
                                            <div className="w-full h-1.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                                                <div 
                                                    className="h-full bg-indigo-500 transition-all duration-700 ease-out"
                                                    style={{ width: `${(user.amount / balance.total_spent) * 100}%` }}
                                                />
                                            </div>
                                        </div>
                                    ))}
                                </div>
                                <div className="px-5 py-3 bg-gray-50/30 dark:bg-gray-800/20 border-t border-gray-50 dark:border-gray-800">
                                    <button 
                                        onClick={() => navigate(`/transactions/?category=${balance.category_id}&include_shared=true`)}
                                        className="text-[11px] font-bold text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 flex items-center gap-1 group"
                                    >
                                        {t('sharing.viewAllTransactions')}
                                        <ArrowRight size={12} className="transition-transform group-hover:translate-x-0.5" />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </section>

            <div className="grid grid-cols-1 gap-8">
                {/* Shared Categories */}
                <section className="space-y-4">
                    <h2 className="text-lg font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Shield className="text-indigo-500" size={20} />
                        {t('sharing.sharedCategories')}
                    </h2>
                    <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                        {shared_categories.length === 0 ? (
                            <div className="p-12 text-center text-gray-400 dark:text-gray-600">
                                {t('sharing.noSharedCategories')}
                            </div>
                        ) : (
                            <div className="divide-y divide-gray-50 dark:divide-gray-800">
                                {shared_categories.map((cat: SharedCategorySummary) => (
                                    <div 
                                        key={cat.id} 
                                        onClick={() => navigate(`/transactions/?category=${cat.id}&include_shared=true`)}
                                        className="p-4 hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors flex items-center justify-between group cursor-pointer"
                                    >
                                        <div className="flex items-center gap-3">
                                            <CategoryBadge category={cat.id} />
                                            <div className="flex flex-col">
                                                <span className="text-[10px] font-bold uppercase tracking-wider text-gray-400">
                                                    {cat.permissions === 'owner' ? t('sharing.permOwner') : t('sharing.permSharedWithMe')}
                                                </span>
                                            </div>
                                        </div>
                                        <div className="flex items-center gap-4">
                                            {cat.shared_with && cat.shared_with.length > 0 && (
                                                <div className="flex -space-x-2">
                                                    {cat.shared_with.slice(0, 3).map((id: string) => (
                                                        <div key={id} className="w-6 h-6 rounded-full bg-gray-200 dark:bg-gray-700 border-2 border-white dark:border-gray-900 flex items-center justify-center text-[8px] font-bold text-gray-500">
                                                            {id.substring(0, 2).toUpperCase()}
                                                        </div>
                                                    ))}
                                                    {cat.shared_with.length > 3 && (
                                                        <div className="w-6 h-6 rounded-full bg-gray-100 dark:bg-gray-800 border-2 border-white dark:border-gray-900 flex items-center justify-center text-[8px] font-bold text-gray-400">
                                                            +{cat.shared_with.length - 3}
                                                        </div>
                                                    )}
                                                </div>
                                            )}
                                            <ChevronRight size={18} className="text-gray-300 group-hover:text-gray-500 transition-colors" />
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </section>
            </div>

            {/* Empty State Banner if everything is empty */}
            {shared_categories.length === 0 && shared_invoices.length === 0 && balances.length === 0 && (
                <div className="bg-gradient-to-br from-indigo-500 to-purple-600 rounded-3xl p-12 text-center text-white space-y-4 shadow-xl shadow-indigo-500/20">
                    <div className="bg-white/20 w-20 h-20 rounded-full flex items-center justify-center mx-auto backdrop-blur-md">
                        <Users size={40} />
                    </div>
                    <div className="max-w-md mx-auto space-y-2">
                        <h3 className="text-2xl font-bold">{t('sharing.welcomeTitle')}</h3>
                        <p className="text-indigo-100 text-sm leading-relaxed">
                            {t('sharing.welcomeSubtitle')}
                        </p>
                    </div>
                    <button className="bg-white text-indigo-600 px-8 py-3 rounded-2xl font-bold text-sm hover:bg-indigo-50 transition-colors shadow-lg shadow-black/10">
                        {t('sharing.startCollaborating')}
                    </button>
                </div>
            )}
        </div>
    );
}
