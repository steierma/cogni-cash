import {useEffect, useState} from 'react';
import {NavLink, useNavigate} from 'react-router-dom';
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
    ArrowLeftRight, BarChart3, Briefcase, ChevronLeft, ChevronRight, FileText, Landmark,
    LayoutDashboard, LogOut, Menu, Monitor, Moon, Settings, Sun, Tag, Users, X, List, Zap, Archive, RefreshCcw,
    type LucideIcon
} from 'lucide-react';
import {settingsService} from '../api/services/settingsService';
import {authService} from '../api/services/authService';
import { useEffectiveSettings } from '../hooks/useEffectiveSettings';
import { getNamespacedKey } from '../api/utils/settingsHelper';
import type { User, SystemInfo } from "../api/types/system";

interface NavItem {
    to: string;
    i18nKeyLabel: string;
    Icon: LucideIcon;
    adminOnly?: boolean;
}

interface NavGroup {
    i18nKeyTitle: string;
    adminOnly?: boolean;
    items: NavItem[];
}

const ALL_NAV_GROUPS: NavGroup[] = [
    {
        i18nKeyTitle: 'layout.overview',
        items: [
            {to: '/', i18nKeyLabel: 'layout.dashboard', Icon: LayoutDashboard},
            {to: '/analytics', i18nKeyLabel: 'layout.analytics', Icon: BarChart3},
            {to: '/forecasting', i18nKeyLabel: 'layout.forecasting', Icon: Zap},
        ]
    },
    {
        i18nKeyTitle: 'layout.bankCashflow',
        items: [
            {to: '/bank-connections', i18nKeyLabel: 'layout.bankConnections', Icon: Landmark},
            {to: '/bank-statements', i18nKeyLabel: 'layout.statements', Icon: List},
            {to: '/transactions', i18nKeyLabel: 'layout.transactions', Icon: ArrowLeftRight},
            {to: '/subscriptions', i18nKeyLabel: 'layout.subscriptions', Icon: RefreshCcw},
            {to: '/categories', i18nKeyLabel: 'layout.categories', Icon: Tag},
            {to: '/sharing', i18nKeyLabel: 'layout.sharing', Icon: Users},
        ]
    },
    {
        i18nKeyTitle: 'layout.incomeDocs',
        items: [
            {to: '/payslips', i18nKeyLabel: 'layout.payslips', Icon: Briefcase},
            {to: '/invoices', i18nKeyLabel: 'layout.invoices', Icon: FileText},
            {to: '/documents', i18nKeyLabel: 'layout.documents', Icon: Archive},
        ]
    },
    {
        i18nKeyTitle: 'layout.admin',
        adminOnly: true,
        items: [
            {to: '/users', i18nKeyLabel: 'layout.users', Icon: Users},
        ]
    }
];

export default function Layout({children}: { children?: React.ReactNode }) {
    const { t } = useTranslation();
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
    const navigate = useNavigate();
    const queryClient = useQueryClient();

    const {data: settings} = useEffectiveSettings();

    const {data: currentUser} = useQuery<User>({
        queryKey: ['currentUser'],
        queryFn: authService.fetchMe,
    });

    const {data: sysInfo} = useQuery<SystemInfo>({
        queryKey: ['systemInfo'],
        queryFn: settingsService.fetchSystemInfo,
        enabled: (currentUser as User)?.role === 'admin',
    });

    // Directly derive state from the query instead of syncing to a local useState via useEffect
    const theme = settings?.['theme'] || 'system';
    const layoutMode = settings?.['layout_mode'] || 'standard';
    const isDesktopExpanded = settings?.sidebar_expanded !== undefined ? settings.sidebar_expanded === 'true' : true;

    useEffect(() => {
        const root = window.document.documentElement;
        root.classList.remove('dark', 'light');
        if (theme === 'system') {
            const systemTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
            root.classList.add(systemTheme);
        } else {
            root.classList.add(theme);
        }
    }, [theme]);

    useEffect(() => {
        const root = window.document.documentElement;
        if (layoutMode === 'compact') {
            root.classList.add('layout-compact');
        } else {
            root.classList.remove('layout-compact');
        }
    }, [layoutMode]);

    const settingsMut = useMutation({
        mutationFn: (newSettings: Record<string, string>) => settingsService.updateSettings(newSettings),
        // Optimistic UI Update: Apply changes instantly to the UI before the server responds
        onMutate: async (partialSettings) => {
            await queryClient.cancelQueries({ queryKey: ['settings'] });
            const previousSettings = queryClient.getQueryData<Record<string, string>>(['settings']);
            queryClient.setQueryData(['settings'], {
                ...previousSettings,
                ...partialSettings
            });
            return { previousSettings };
        },
        // TS6133 behoben: _err und _newSettings mit Unterstrich markieren
        onError: (_err, _newSettings, context) => {
            // Roll back to the previous state if the API call fails
            if (context?.previousSettings) {
                queryClient.setQueryData(['settings'], context.previousSettings);
            }
        },
        onSettled: () => queryClient.invalidateQueries({queryKey: ['settings']})
    });

    const cycleTheme = () => {
        const nextTheme = theme === 'light' ? 'dark' : theme === 'dark' ? 'system' : 'light';
        settingsMut.mutate({ theme: nextTheme });
    };

    const toggleSidebar = () => {
        const nextState = !isDesktopExpanded;
        const key = getNamespacedKey('sidebar_expanded', true);
        settingsMut.mutate({ [key]: String(nextState) });
    };

    const handleLogout = async () => {
        try {
            await authService.logout();
        } catch (error) {
            console.error('Logout failed', error);
        }
        queryClient.clear();
        navigate('/login');
    };

    const ThemeIcon = theme === 'light' ? Sun : theme === 'dark' ? Monitor : Moon;
    const showLabels = isMobileMenuOpen || isDesktopExpanded;
    const sidebarWidth = showLabels ? 'w-64' : 'w-20';

    const visibleNavGroups = ALL_NAV_GROUPS.map(group => {
        if (group.adminOnly && currentUser?.role !== 'admin') return null;

        const visibleItems = group.items.filter(item => {
            if (item.adminOnly && currentUser?.role !== 'admin') return false;
            return true;
        });

        if (visibleItems.length === 0) return null;

        return {...group, items: visibleItems};
    }).filter(group => group !== null) as NavGroup[];

    return (
        <div className="min-h-[100dvh] flex bg-gray-50/50 dark:bg-gray-950 font-sans text-gray-900 dark:text-gray-100 transition-colors duration-200">
            {isMobileMenuOpen && (
                <div className="fixed inset-0 bg-gray-900/50 dark:bg-gray-950/80 z-40 md:hidden" onClick={() => setIsMobileMenuOpen(false)}/>
            )}

            <aside className={`fixed inset-y-0 left-0 md:sticky md:top-0 md:h-[100dvh] bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col z-50 transform transition-all duration-300 ease-in-out ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'} ${sidebarWidth}`}>

                <div className={`h-[68px] border-b border-gray-100 dark:border-gray-800/50 flex items-center shrink-0 transition-all duration-300 ${showLabels ? 'px-6 justify-between' : 'justify-center'}`}>
                    {showLabels ? (
                        <span className="text-lg font-bold text-indigo-600 dark:text-indigo-400 tracking-tight flex items-center gap-2 truncate"><img src="/logo.png" alt="Logo" className="h-6 w-6 object-contain"/>CogniCash</span>
                    ) : (
                        <span className="text-xl flex items-center justify-center">💰</span>
                    )}
                    <button className="md:hidden text-gray-500 dark:text-gray-400 p-2 -mr-2" onClick={() => setIsMobileMenuOpen(false)}><X size={20}/></button>
                </div>

                {/* Scrollable Navigation Area */}
                <div className="flex-1 overflow-y-auto [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]">
                    <nav className={`py-4 pb-8 space-y-6 transition-all duration-300 ${showLabels ? 'px-4' : 'px-3'}`}>
                        {visibleNavGroups.map((group, groupIdx) => (
                            <div key={groupIdx}>
                                {showLabels ? (
                                    <h3 className="px-3 mb-2 text-xs font-bold uppercase tracking-wider text-gray-400 dark:text-gray-500">{t(group.i18nKeyTitle)}</h3>
                                ) : (
                                    <div className="w-8 mx-auto mb-2 border-b border-gray-200 dark:border-gray-800"/>
                                )}
                                <div className="space-y-1">
                                    {group.items.map(({to, i18nKeyLabel, Icon}) => (
                                        <NavLink
                                            key={to}
                                            to={to}
                                            end={to === '/'}
                                            onClick={() => setIsMobileMenuOpen(false)}
                                            title={!showLabels ? t(i18nKeyLabel) : undefined}
                                            className={({isActive}) => `w-full flex items-center gap-3 py-2.5 rounded-xl text-sm font-medium transition-all ${isActive ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100'} ${showLabels ? 'px-3' : 'justify-center px-0'}`}
                                        >
                                            {({isActive}) => (
                                                <>
                                                    <Icon size={18} className={`shrink-0 ${isActive ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500'}`}/>
                                                    {showLabels && <span className="truncate">{t(i18nKeyLabel)}</span>}
                                                </>
                                            )}
                                        </NavLink>
                                    ))}
                                </div>
                            </div>
                        ))}
                    </nav>
                </div>

                {/* Pinned Bottom Actions */}
                <div className={`py-4 pb-[calc(1rem+env(safe-area-inset-bottom))] border-t border-gray-100 dark:border-gray-800/50 space-y-1 shrink-0 transition-all duration-300 ${showLabels ? 'px-4' : 'px-3'}`}>
                    <NavLink to="/settings" onClick={() => setIsMobileMenuOpen(false)} title={!showLabels ? t('layout.settings') : undefined} className={({isActive}) => `w-full flex items-center gap-3 py-2.5 rounded-xl text-sm font-medium transition-all ${isActive ? 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100'} ${showLabels ? 'px-3' : 'justify-center px-0'}`}>
                        {({isActive}) => (
                            <>
                                <Settings size={18} className={`shrink-0 ${isActive ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500'}`}/>
                                {showLabels && <span className="truncate">{t('layout.settings')}</span>}
                            </>
                        )}
                    </NavLink>

                    <button onClick={toggleSidebar} title={showLabels ? t('layout.collapse') : t('layout.expand')} className={`hidden md:flex w-full items-center gap-3 py-2.5 rounded-xl text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 transition-all group ${showLabels ? 'px-3' : 'justify-center px-0'}`}>
                        {showLabels ? (
                            <><ChevronLeft size={18} className="shrink-0 text-gray-400 dark:text-gray-500 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors"/><span className="truncate">{t('layout.collapseBtn')}</span></>
                        ) : (
                            <ChevronRight size={18} className="shrink-0 text-gray-400 dark:text-gray-500 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors"/>
                        )}
                    </button>

                    <button onClick={cycleTheme} title={!showLabels ? `${theme} ${t('layout.mode')}` : undefined} className={`w-full flex items-center gap-3 py-2.5 rounded-xl text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800/50 hover:text-gray-900 dark:hover:text-gray-100 transition-all group ${showLabels ? 'px-3' : 'justify-center px-0'}`}>
                        <ThemeIcon size={18} className="shrink-0 text-gray-400 dark:text-gray-500 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors"/>
                        {showLabels && <span className="capitalize truncate">{theme} {t('layout.mode')}</span>}
                    </button>

                    <button onClick={handleLogout} title={!showLabels ? t('layout.logout') : undefined} className={`w-full flex items-center gap-3 py-2.5 rounded-xl text-sm font-medium text-gray-600 dark:text-gray-400 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-600 dark:hover:text-red-400 transition-all group ${showLabels ? 'px-3' : 'justify-center px-0'}`}>
                        <LogOut size={18} className="shrink-0 text-gray-400 dark:text-gray-500 group-hover:text-red-600 dark:group-hover:text-red-400 transition-colors"/>
                        {showLabels && <span className="truncate">{t('layout.logout')}</span>}
                    </button>

                    {showLabels && sysInfo?.version && (
                        <div className="px-3 pt-2 text-[10px] font-medium text-gray-400 dark:text-gray-600 flex items-center gap-1 opacity-60">
                            <span className="w-1 h-1 rounded-full bg-gray-300 dark:bg-gray-700"></span>
                            {t('settings.version')}: {sysInfo.version}
                        </div>
                    )}
                </div>
            </aside>

            <main className="flex-1 flex flex-col min-w-0 h-[100dvh] overflow-hidden">
                <header className="md:hidden bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-4 py-3 flex items-center gap-3 shrink-0">
                    <button onClick={() => setIsMobileMenuOpen(true)} className="text-gray-600 dark:text-gray-400 p-1 -ml-1"><Menu size={24}/></button>
                    <span className="font-bold text-indigo-600 dark:text-indigo-400">CogniCash</span>
                </header>
                <div className="flex-1 overflow-auto p-4 md:p-8 pb-20 md:pb-8">
                    {children}
                </div>
            </main>
        </div>
    );
}