import {useState, useRef, useMemo, useEffect} from 'react';
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import {
    Briefcase, Pencil, TrendingUp, Wallet, Filter, Columns, Check, FileUp, ArrowUpRight, ArrowDownRight, BrainCircuit, BarChart3
} from 'lucide-react';
import { payslipService } from '../api/services/payslipService';
import { settingsService } from '../api/services/settingsService';
import type { Payslip } from "../api/types/payslip";
import {fmtCurrency} from '../utils/formatters';

import {type ColKey, type SortDirection, formatYearMonth, getAdjustedNetto} from '../components/payslips/utils';
import {PayslipChart} from '../components/payslips/PayslipChart';
import {PayslipTable} from '../components/payslips/PayslipTable';
import {
    ViewPayslipModal,
    PreviewPayslipModal,
    ImportPayslipModal,
    EditPayslipModal,
    BatchResultsModal
} from '../components/payslips/PayslipModals';

export default function PayslipsPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const inputRef = useRef<HTMLInputElement>(null);

    // Main Dropzone State
    const [dragOver, setDragOver] = useState(false);
    const [useAI, setUseAI] = useState(false);

    // Modal States
    const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
    const [editingPayslip, setEditingPayslip] = useState<Payslip | null>(null);
    const [viewingPayslip, setViewingPayslip] = useState<Payslip | null>(null);
    const [previewingPayslip, setPreviewingPayslip] = useState<Payslip | null>(null);
    const [previewInfo, setPreviewInfo] = useState<{url: string, mimeType: string} | null>(null);
    const [isPreviewLoading, setIsPreviewLoading] = useState<string | null>(null);
    const [batchResults, setBatchResults] = useState<{ successful: Payslip[], failed: { filename: string, error: string }[] } | null>(null);

    // Filter States
    const [selectedStartPeriod, setSelectedStartPeriod] = useState<string>('All');
    const [selectedEndPeriod, setSelectedEndPeriod] = useState<string>('All');
    const [selectedEmployer, setSelectedEmployer] = useState<string>('All');
    const [selectedTaxClass, setSelectedTaxClass] = useState<string>('All');

    const [appliedStartPeriod, setAppliedStartPeriod] = useState<string>('All');
    const [appliedEndPeriod, setAppliedEndPeriod] = useState<string>('All');
    const [appliedEmployer, setAppliedEmployer] = useState<string>('All');
    const [appliedTaxClass, setAppliedTaxClass] = useState<string>('All');

    // Shared Chart/Table States
    const [excludedBonuses, setExcludedBonuses] = useState<Set<string>>(new Set());
    const [excludeLeasing] = useState(true);
    const [useProportionalMath, setUseProportionalMath] = useState(true);
    const [ignoredPayslipIds, setIgnoredPayslipIds] = useState<Set<string>>(new Set());
    const [initializedBonuses, setInitializedBonuses] = useState(false);

    // Table Settings
    const [showColMenu, setShowColMenu] = useState(false);
    const [visibleCols, setVisibleCols] = useState<Record<ColKey, boolean>>({
        period: true, employer: true, gross: true, net: true, adjNet: false, payout: true, leasing: false,
    });
    const [sortConfig, setSortConfig] = useState<{ key: ColKey; direction: SortDirection }>({
        key: 'period', direction: 'desc'
    });

    // ── Queries & Mutations ──
    const {data: payslips = [], isLoading} = useQuery<Payslip[], Error>({
        queryKey: ['payslips', appliedEmployer],
        queryFn: () => payslipService.fetchPayslips(appliedEmployer === 'All' ? undefined : appliedEmployer)
    });

    // We need all payslips to populate the filter dropdowns correctly
    const {data: allPayslips = []} = useQuery<Payslip[], Error>({
        queryKey: ['payslips', 'all'],
        queryFn: () => payslipService.fetchPayslips()
    });
    const {data: summary} = useQuery({
        queryKey: ['payslips', 'summary'],
        queryFn: () => payslipService.fetchSummary()
    });
    const {data: settings} = useQuery({queryKey: ['settings'], queryFn: () => settingsService.fetchSettings()});

    useEffect(() => {
        if (settings?.payslips_visible_cols) {
            try {
                setVisibleCols(JSON.parse(settings.payslips_visible_cols));
            } catch (e) {
                console.error("Failed to parse column settings", e);
            }
        }
    }, [settings?.payslips_visible_cols]);

    const uploadMutation = useMutation({
        mutationFn: (args: { file: File; overrides?: Partial<Payslip>; useAI?: boolean }) => payslipService.import(args),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['payslips']});
            setIsUploadModalOpen(false);
        },
    });

    const batchUploadMutation = useMutation({
        mutationFn: (files: File[]) => payslipService.importBatch(files),
        onSuccess: (data) => {
            queryClient.invalidateQueries({queryKey: ['payslips']});
            setBatchResults(data);
        },
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => payslipService.delete(id),
        onSuccess: () => queryClient.invalidateQueries({queryKey: ['payslips']}),
    });

    const updateMutation = useMutation({
        mutationFn: ({id, data}: {
            id: string;
            data: Partial<Payslip> | FormData
        }) => payslipService.update(id, data as Partial<Payslip> | FormData),
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ['payslips']});
            setEditingPayslip(null);
            setPreviewingPayslip(null);
        },
    });

    const updateSettingsMut = useMutation({
        mutationFn: (params: Record<string, string>) => settingsService.updateSettings(params),
        onSuccess: () => queryClient.invalidateQueries({queryKey: ['settings']})
    });

    // ── Handlers ──
    const handleFiles = (files: FileList | File[]) => {
        const fileList = Array.from(files);
        if (fileList.length === 0) return;

        if (fileList.length === 1) {
            uploadMutation.reset();
            uploadMutation.mutate({file: fileList[0], useAI});
        } else {
            batchUploadMutation.reset();
            batchUploadMutation.mutate(fileList);
        }

        if (inputRef.current) inputRef.current.value = '';
    };

    const onDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setDragOver(false);
        if (e.dataTransfer.files?.length > 0) handleFiles(e.dataTransfer.files);
    };

    const handleApplyFilters = () => {
        setAppliedStartPeriod(selectedStartPeriod);
        setAppliedEndPeriod(selectedEndPeriod);
        setAppliedEmployer(selectedEmployer);
        setAppliedTaxClass(selectedTaxClass);
    };

    const toggleColumn = (col: ColKey) => {
        setVisibleCols(prev => {
            const next = {...prev, [col]: !prev[col]};
            updateSettingsMut.mutate({payslips_visible_cols: JSON.stringify(next)});
            return next;
        });
    };

    const handleSort = (key: ColKey) => {
        setSortConfig(prev => ({key, direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'}));
    };

    const handlePreview = async (id: string) => {
        try {
            setIsPreviewLoading(id);
            const p = payslips.find(ps => ps.id === id);
            if (p) setPreviewingPayslip(p);
            const info = await payslipService.getPreviewUrl(id);
            setPreviewInfo(info);
        } catch {
            alert("Could not load the document preview.");
        } finally {
            setIsPreviewLoading(null);
        }
    };

    const closePreview = () => {
        if (previewInfo) {
            URL.revokeObjectURL(previewInfo.url);
        }
        setPreviewInfo(null);
        setPreviewingPayslip(null);
    };

    // ── Data Processing ──
    const uniqueBonuses = useMemo(() => {
        const bonuses = new Set<string>();
        ((allPayslips as Payslip[]) || []).forEach(p => p.bonuses?.forEach(sz => bonuses.add(sz.description)));
        return Array.from(bonuses).sort();
    }, [allPayslips]);

    useEffect(() => {
        if (allPayslips && (allPayslips as Payslip[]).length > 0 && !initializedBonuses && uniqueBonuses.length > 0) {
            setExcludedBonuses(new Set(uniqueBonuses));
            setInitializedBonuses(true);
        }
    }, [allPayslips, uniqueBonuses, initializedBonuses]);

    const {uniquePeriods, uniqueEmployers, uniqueTaxClasses} = useMemo(() => {
        if (!allPayslips || (allPayslips as Payslip[]).length === 0) return {uniquePeriods: [], uniqueEmployers: [], uniqueTaxClasses: []};
        return {
            uniquePeriods: Array.from(new Set((allPayslips as Payslip[]).map(p => formatYearMonth(p.period_year, p.period_month_num)))).sort((a, b) => b.localeCompare(a)),
            uniqueEmployers: Array.from(new Set((allPayslips as Payslip[]).map(p => p.employer_name).filter(Boolean))).sort(),
            uniqueTaxClasses: Array.from(new Set((allPayslips as Payslip[]).map(p => p.tax_class).filter(Boolean))).sort(),
        };
    }, [allPayslips]);

    const filteredPayslips = useMemo(() => {
        const filtered = ((payslips as Payslip[]) || []).filter(p => {
            const period = formatYearMonth(p.period_year, p.period_month_num);
            if (appliedStartPeriod !== 'All' && period < appliedStartPeriod) return false;
            if (appliedEndPeriod !== 'All' && period > appliedEndPeriod) return false;
            if (appliedEmployer !== 'All' && p.employer_name !== appliedEmployer) return false;
            if (appliedTaxClass !== 'All' && p.tax_class !== appliedTaxClass) return false;
            return true;
        });

        filtered.sort((a, b) => {
            let comparison = 0;
            switch (sortConfig.key) {
                case 'period':
                    comparison = (a.period_year - b.period_year) || ((a.period_month_num || 0) - (b.period_month_num || 0));
                    break;
                case 'employer':
                    comparison = (a.employer_name || '').localeCompare(b.employer_name || '');
                    break;
                case 'gross':
                    comparison = a.gross_pay - b.gross_pay;
                    break;
                case 'net':
                    comparison = a.net_pay - b.net_pay;
                    break;
                case 'adjNet':
                    comparison = getAdjustedNetto(a, excludedBonuses, excludeLeasing, useProportionalMath) - getAdjustedNetto(b, excludedBonuses, excludeLeasing, useProportionalMath);
                    break;
                case 'payout':
                    comparison = a.payout_amount - b.payout_amount;
                    break;
                case 'leasing':
                    comparison = a.custom_deductions - b.custom_deductions;
                    break;
            }
            return sortConfig.direction === 'asc' ? comparison : -comparison;
        });
        return filtered;
    }, [payslips, appliedStartPeriod, appliedEndPeriod, appliedEmployer, appliedTaxClass, sortConfig, excludedBonuses, excludeLeasing, useProportionalMath]);

    const hasFilter = appliedStartPeriod !== 'All' || appliedEndPeriod !== 'All' || appliedEmployer !== 'All' || appliedTaxClass !== 'All';

    const periodTotals = useMemo(() => {
        if (!hasFilter && summary) {
            return {
                gross: summary.total_gross,
                net: summary.total_net,
                payout: summary.total_payout,
                bonuses: summary.total_bonuses,
            };
        }
        return filteredPayslips.reduce((acc, p) => {
            acc.gross += p.gross_pay || 0;
            acc.net += p.net_pay || 0;
            acc.payout += p.payout_amount || 0;
            acc.bonuses += (p.bonuses || []).reduce((sum, b) => sum + (b.amount || 0), 0);
            return acc;
        }, {gross: 0, net: 0, payout: 0, bonuses: 0});
    }, [filteredPayslips, hasFilter, summary]);

    const latestPayslip = [...filteredPayslips].sort((a, b) => (b.period_year - a.period_year) || ((b.period_month_num || 0) - (a.period_month_num || 0)))[0];
    const previousPayslip = [...filteredPayslips].sort((a, b) => (b.period_year - a.period_year) || ((b.period_month_num || 0) - (a.period_month_num || 0)))[1];

    let totalPercentChange = 0;
    let adjPercentChange = 0;
    
    if (!hasFilter && summary) {
        totalPercentChange = summary.net_pay_trend;
    } else if (latestPayslip && previousPayslip && previousPayslip.net_pay > 0) {
        totalPercentChange = ((latestPayslip.net_pay - previousPayslip.net_pay) / previousPayslip.net_pay) * 100;
    }

    if (latestPayslip && previousPayslip) {
        const prevAdjNet = getAdjustedNetto(previousPayslip, excludedBonuses, excludeLeasing, useProportionalMath);
        if (prevAdjNet > 0) adjPercentChange = ((getAdjustedNetto(latestPayslip, excludedBonuses, excludeLeasing, useProportionalMath) - prevAdjNet) / prevAdjNet) * 100;
    }

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Briefcase className="text-indigo-600 dark:text-indigo-400"/> {t('payslips.title')}</h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {t('payslips.subtitle')}
                    </p>
                </div>
            </div>

            {/* Quick Upload Dropzone */}
            <div
                className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row gap-4 items-center">
                <div
                    onDragOver={(e) => {
                        e.preventDefault();
                        setDragOver(true);
                    }} onDragLeave={() => setDragOver(false)} onDrop={onDrop} onClick={() => inputRef.current?.click()}
                    className={`flex-1 w-full border-2 border-dashed rounded-xl p-4 flex items-center justify-center gap-4 cursor-pointer transition-all duration-200 ${dragOver ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20' : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50 dark:bg-gray-800/30'}`}
                >
                    <input type="file" className="hidden" ref={inputRef} accept=".pdf,.png,.jpg,.jpeg,.webp,.gif"
                           onChange={(e) => e.target.files && handleFiles(e.target.files)}/>
                    <div
                        className="p-2 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg text-indigo-600 dark:text-indigo-400">
                        <FileUp size={24}/></div>
                    <div>
                        {uploadMutation.isPending || batchUploadMutation.isPending ?
                            <p className="text-sm font-medium text-gray-900 dark:text-gray-100 animate-pulse">{t('payslips.uploadingParsing')}</p> : <><p
                                className="text-sm font-medium text-gray-900 dark:text-gray-100"><span
                                className="text-indigo-600 dark:text-indigo-400">{t('payslips.clickToParse')}</span> {t('bankStatements.import.orDrag')}</p><p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('payslips.pdfFormat')}</p></>}
                    </div>
                </div>

                {/* AI Toggle and Manual Override Button Container */}
                <div
                    className="flex flex-row md:flex-col gap-3 w-full md:w-auto md:min-w-[200px] justify-center items-center md:items-stretch">
                    <label
                        className="flex items-center gap-2 cursor-pointer group px-1 justify-center md:justify-start">
                        <div className="relative flex items-center">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={useAI}
                                onChange={(e) => setUseAI(e.target.checked)}
                            />
                            <div
                                className="w-9 h-5 bg-gray-200 dark:bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 dark:after:border-gray-600 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-indigo-600"></div>
                        </div>
                        <span
                            className="text-sm font-medium text-gray-600 dark:text-gray-400 flex items-center gap-1.5 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                            <BrainCircuit size={16}/>
                            {t('payslips.forceAi')}
                        </span>
                    </label>

                    <button onClick={() => setIsUploadModalOpen(true)} disabled={uploadMutation.isPending}
                            className="flex items-center justify-center gap-2 w-full md:w-auto px-4 py-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 text-gray-700 dark:text-gray-300 text-sm font-medium rounded-xl transition-colors disabled:opacity-70">
                        <Pencil size={16}/> {t('payslips.manualOverride')}
                    </button>
                </div>
            </div>
            {uploadMutation.isError && <div
                className="p-4 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 rounded-xl border border-red-200 dark:border-red-800/50 text-sm">{t('payslips.uploadFailed')}</div>}
            {batchUploadMutation.isError && <div
                className="p-4 bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 rounded-xl border border-red-200 dark:border-red-800/50 text-sm">{t('payslips.uploadFailed')}</div>}

            {/* Filters Bar */}
            <div
                className="bg-white dark:bg-gray-900 p-4 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <div className="flex flex-wrap items-center gap-4">
                    <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400"><Filter size={18}/><span
                        className="text-sm font-medium">{t('common.filters')}:</span></div>
                    <div className="flex items-center gap-2">
                        <select value={selectedStartPeriod} onChange={e => setSelectedStartPeriod(e.target.value)}
                                className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500">
                            <option value="All">{t('common.from')}</option>
                            {uniquePeriods.map(p => <option key={p} value={p}>{p}</option>)}</select>
                        <span className="text-gray-400 text-sm">-</span>
                        <select value={selectedEndPeriod} onChange={e => setSelectedEndPeriod(e.target.value)}
                                className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500">
                            <option value="All">{t('common.to')}</option>
                            {uniquePeriods.map(p => <option key={p} value={p}>{p}</option>)}</select>
                    </div>
                    <select value={selectedEmployer} onChange={e => setSelectedEmployer(e.target.value)}
                            className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500">
                        <option value="All">{t('payslips.allEmployers')}</option>
                        {uniqueEmployers.map(e => <option key={e} value={e}>{e}</option>)}</select>
                    <select value={selectedTaxClass} onChange={e => setSelectedTaxClass(e.target.value)}
                            className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500">
                        <option value="All">{t('payslips.allTaxClasses')}</option>
                        {uniqueTaxClasses.map(taxClass => (
                            <option key={taxClass} value={taxClass}>
                                {t('payslips.taxClassCount', { count: taxClass })}
                            </option>
                        ))}</select>
                    <button onClick={handleApplyFilters}
                            className="px-4 py-1.5 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-sm font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors">{t('payslips.applyFilters')}
                    </button>
                </div>
                <div className="relative">
                    <button onClick={() => setShowColMenu(!showColMenu)}
                            className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
                        <Columns size={16}/> {t('payslips.columns')}
                    </button>
                    {showColMenu && (
                        <div
                            className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg z-10 overflow-hidden p-2 space-y-1">
                            {[{key: 'period', label: t('payslips.modals.period')}, {key: 'employer', label: t('payslips.modals.employer')}, {
                                key: 'gross',
                                label: t('payslips.modals.gross')
                            }, {key: 'net', label: t('payslips.modals.net')}, {
                                key: 'adjNet',
                                label: t('payslips.adjustedNet')
                            }, {key: 'payout', label: t('payslips.modals.payout')}, {key: 'leasing', label: t('payslips.modals.leasing')}].map(({
                                                                                                                   key,
                                                                                                                   label
                                                                                                               }) => (
                                <button key={key} onClick={() => toggleColumn(key as ColKey)}
                                        className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors">
                                    {label} {visibleCols[key as ColKey] &&
                                    <Check size={16} className="text-indigo-600 dark:text-indigo-400"/>}
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            {/* KPI Cards */}
            {latestPayslip && (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    <div
                        className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
                        <div
                            className="w-12 h-12 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 flex items-center justify-center shrink-0">
                            <Wallet className="text-emerald-600 dark:text-emerald-400" size={24}/></div>
                        <div className="flex-1">
                            <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">
                                {t('payslips.latestTotalNet')}
                            </p>
                            <div className="flex items-end gap-2">
                                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 leading-none">{fmtCurrency(latestPayslip.net_pay, 'EUR')}</p>
                                {totalPercentChange !== 0 && <span
                                    className={`flex items-center text-xs font-medium mb-0.5 ${totalPercentChange > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'}`}>{totalPercentChange > 0 ?
                                    <ArrowUpRight size={14}/> :
                                    <ArrowDownRight size={14}/>}{Math.abs(totalPercentChange).toFixed(1)}%</span>}
                            </div>
                            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{formatYearMonth(latestPayslip.period_year, latestPayslip.period_month_num)}</p>
                        </div>
                    </div>
                    <div
                        className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4">
                        <div
                            className="w-12 h-12 rounded-xl bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center shrink-0">
                            <TrendingUp className="text-blue-600 dark:text-blue-400" size={24}/></div>
                        <div className="flex-1">
                            <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">
                                {t('payslips.adjustedNet')}
                            </p>
                            <div className="flex items-end gap-2">
                                <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 leading-none">{fmtCurrency(getAdjustedNetto(latestPayslip, excludedBonuses, excludeLeasing, useProportionalMath), 'EUR')}</p>
                                {adjPercentChange !== 0 && <span
                                    className={`flex items-center text-xs font-medium mb-0.5 ${adjPercentChange > 0 ? 'text-blue-600 dark:text-blue-400' : 'text-red-600 dark:text-red-400'}`}>{adjPercentChange > 0 ?
                                    <ArrowUpRight size={14}/> :
                                    <ArrowDownRight size={14}/>}{Math.abs(adjPercentChange).toFixed(1)}%</span>}
                            </div>
                            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">{t('payslips.momGrowth')}</p>
                        </div>
                    </div>
                    <div
                        className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex items-center gap-4 hidden lg:flex">
                        <div
                            className="w-12 h-12 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 flex items-center justify-center shrink-0">
                            <Briefcase className="text-indigo-600 dark:text-indigo-400" size={24}/></div>
                        <div>
                            <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">
                                {t('payslips.latestGrossIncome')}
                            </p>
                            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-0.5 leading-none">{fmtCurrency(latestPayslip.gross_pay, 'EUR')}</p>
                            <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                                {t('payslips.taxClass', { taxClass: latestPayslip.tax_class })}
                            </p>
                        </div>
                    </div>
                </div>
            )}

            {/* Period Summary */}
            <div className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm animate-in fade-in slide-in-from-top-2 duration-300">
                <div className="flex items-center gap-2 mb-4">
                    <BarChart3 className="text-indigo-500" size={20} />
                    <h3 className="font-semibold text-gray-900 dark:text-gray-100">{t('payslips.periodSummary.title')}</h3>
                </div>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                    <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.gross')}</p>
                        <p className="text-xl font-bold text-gray-900 dark:text-gray-100 font-mono">{fmtCurrency(periodTotals.gross, 'EUR')}</p>
                    </div>
                    <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.net')}</p>
                        <p className="text-xl font-bold text-gray-900 dark:text-gray-100 font-mono">{fmtCurrency(periodTotals.net, 'EUR')}</p>
                    </div>
                    <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.payout')}</p>
                        <p className="text-xl font-bold text-emerald-600 dark:text-emerald-400 font-mono">{fmtCurrency(periodTotals.payout, 'EUR')}</p>
                    </div>
                    <div>
                        <p className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1">{t('payslips.modals.bonuses')}</p>
                        <p className="text-xl font-bold text-indigo-600 dark:text-indigo-400 font-mono">{fmtCurrency(periodTotals.bonuses, 'EUR')}</p>
                    </div>
                </div>
            </div>

            {/* Charts Section */}
            <PayslipChart
                filteredPayslips={filteredPayslips}
                ignoredPayslipIds={ignoredPayslipIds}
                uniqueBonuses={uniqueBonuses}
                excludedBonuses={excludedBonuses}
                setExcludedBonuses={setExcludedBonuses}
                useProportionalMath={useProportionalMath}
                setUseProportionalMath={setUseProportionalMath}
            />

            {/* Injected Table Component */}
            <PayslipTable
                isLoading={isLoading}
                filteredPayslips={filteredPayslips}
                visibleCols={visibleCols}
                sortConfig={sortConfig}
                onSort={handleSort}
                ignoredPayslipIds={ignoredPayslipIds}
                onToggleIgnore={(id) => setIgnoredPayslipIds(prev => {
                    const next = new Set(prev);
                    if (next.has(id)) {
                        next.delete(id);
                    } else {
                        next.add(id);
                    }
                    return next;
                })}
                onPreview={handlePreview}
                isPreviewLoading={isPreviewLoading}
                onView={(p) => setViewingPayslip(p)}
                onEdit={(p) => setEditingPayslip(p)}
                onDownload={payslipService.downloadFile}
                onDelete={(id) => deleteMutation.mutate(id)}
                excludedBonuses={excludedBonuses}
                excludeLeasing={excludeLeasing}
                useProportionalMath={useProportionalMath}
            />

            {/* Modals */}
            {viewingPayslip && <ViewPayslipModal payslip={viewingPayslip} onClose={() => setViewingPayslip(null)}/>}
            {previewInfo && previewingPayslip && (
                <PreviewPayslipModal 
                    previewUrl={previewInfo.url} 
                    mimeType={previewInfo.mimeType}
                    payslip={previewingPayslip} 
                    onClose={closePreview}
                    onUpdate={(id, data) => updateMutation.mutate({id, data})}
                    isPending={updateMutation.isPending}
                />
            )}
            {isUploadModalOpen &&
                <ImportPayslipModal onClose={() => setIsUploadModalOpen(false)}
                                    onImport={(file, overrides) => uploadMutation.mutate({file, overrides, useAI})}
                                    isPending={uploadMutation.isPending}
                                    useAI={useAI} />}
            {editingPayslip && <EditPayslipModal payslip={editingPayslip} onClose={() => setEditingPayslip(null)}
                                                 onUpdate={(id, payload) => updateMutation.mutate({id, data: payload})}
                                                 isPending={updateMutation.isPending}/>}
            {batchResults && <BatchResultsModal {...batchResults} onClose={() => setBatchResults(null)} />}

        </div>
    );
}