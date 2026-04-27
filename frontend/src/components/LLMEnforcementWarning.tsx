import React from 'react';
import { ShieldAlert, ArrowRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { authService } from '../api/services/authService';
import { useEffectiveSettings } from '../hooks/useEffectiveSettings';

export const LLMEnforcementWarning: React.FC = () => {
    const { t } = useTranslation();
    
    const { data: currentUser } = useQuery({
        queryKey: ['me'],
        queryFn: () => authService.fetchMe(),
    });

    const { data: settings } = useEffectiveSettings();

    if (!settings || !currentUser) return null;

    // Admins are exempt from enforcement as they use the global fallback
    if (currentUser.role === 'admin') return null;

    const enforcementEnabled = settings['llm_enforce_user_config'] === 'true';
    if (!enforcementEnabled) return null;

    let hasActiveProfile = false;
    try {
        const profiles = JSON.parse(settings['llm_profiles'] || '[]');
        hasActiveProfile = Array.isArray(profiles) && profiles.some((p: any) => p.is_active);
    } catch (e) {
        hasActiveProfile = false;
    }

    if (hasActiveProfile) return null;

    return (
        <div className="mb-6 animate-in fade-in slide-in-from-top-4 duration-300">
            <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800/50 rounded-2xl p-4 flex items-start gap-4">
                <div className="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center shrink-0">
                    <ShieldAlert className="text-amber-600 dark:text-amber-400" size={24} />
                </div>
                <div className="flex-1">
                    <h3 className="text-sm font-bold text-amber-900 dark:text-amber-100">
                        {t('settings.llm.enforcementTitle') || "AI Configuration Required"}
                    </h3>
                    <p className="text-xs text-amber-800/80 dark:text-amber-400/80 mt-1">
                        {t('settings.llm.enforcementMessage') || "The administrator requires users to configure their own AI profiles. Some features will be unavailable until you add an active profile in your settings."}
                    </p>
                    <Link
                        to="/settings"
                        className="inline-flex items-center gap-1 mt-3 text-xs font-bold text-amber-700 dark:text-amber-400 hover:underline"
                    >
                        {t('settings.llm.goToSettings') || "Go to Settings"}
                        <ArrowRight size={14} />
                    </Link>
                </div>
            </div>
        </div>
    );
};
