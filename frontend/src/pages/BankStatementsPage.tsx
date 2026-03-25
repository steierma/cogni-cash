import { useMemo, useState, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
    deleteBankStatement, downloadBankStatement, fetchBankStatements, fetchBankStatementBlob, importBankStatement,
    fetchSettings, updateSettings
} from '../api/client';
import {
    AlertTriangle, CheckCircle2, ChevronDown, ChevronUp, Download, FileText,
    Trash2, FileUp, BrainCircuit, AlertCircle, XCircle, X, Filter, Eye, List, Columns, Check, Loader2, FileSpreadsheet
} from 'lucide-react';
import type { BankStatementSummary, ImportBatchResponse, ImportResult } from '../api/types';
import { fmtCurrency } from '../utils/formatters';

// ── Types & Interfaces ──
type SortField = 'statement_info' | 'statement_type' | 'period' | 'transaction_count' | 'new_balance';
type SortDirection = 'asc' | 'desc';
type ColKey = 'statementInfo' | 'type' | 'period' | 'transactions' | 'newBalance';

export interface PreviewData {
    url: string | null;
    text: string | null;
    type: string | null;
    statementId: string;
}

export default function BankStatementsPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const navigate = useNavigate();
    const inputRef = useRef<HTMLInputElement>(null);

    // ── Import States ──
    const [dragOver, setDragOver] = useState(false);
    const [useAI, setUseAI] = useState(false);
    const [statementType, setStatementType] = useState('auto');

    // ── Modal & Action States ──
    const [deletingId, setDeletingId] = useState<string | null>(null);
    const [viewingStatement, setViewingStatement] = useState<BankStatementSummary | null>(null);
    const [previewData, setPreviewData] = useState<PreviewData | null>(null);
    const [isPreviewLoading, setIsPreviewLoading] = useState<string | null>(null);

    // ── Filter States ──
    const [selectedYear, setSelectedYear] = useState<string>('All');
    const [selectedType, setSelectedType] = useState<string>('All');
    const [selectedAccount, setSelectedAccount] = useState<string>('All');
    const [appliedYear, setAppliedYear] = useState<string>('All');
    const [appliedType, setAppliedType] = useState<string>('All');
    const [appliedAccount, setAppliedAccount] = useState<string>('All');

    // ── Sorting State ──
    const [sortField, setSortField] = useState<SortField>('period');
    const [sortDirection, setSortDirection] = useState<SortDirection>('desc');

    // ── Column Visibility State ──
    const [showColMenu, setShowColMenu] = useState(false);
    const [visibleCols, setVisibleCols] = useState<Record<ColKey, boolean>>({
        statementInfo: true,
        type: true,
        period: true,
        transactions: true,
        newBalance: true,
    });

    // ── Queries ──
    const { data: statements, isLoading } = useQuery<BankStatementSummary[]>({
        queryKey: ['bank-statements'],
        queryFn: fetchBankStatements,
    });

    const { data: settings } = useQuery({
        queryKey: ['settings'],
        queryFn: fetchSettings,
    });

    // Apply saved column settings on load
    useEffect(() => {
        if (settings?.bank_statements_visible_cols) {
            try {
                setVisibleCols(JSON.parse(settings.bank_statements_visible_cols));
            } catch (e) {
                console.error("Failed to parse column settings", e);
            }
        }
    }, [settings?.bank_statements_visible_cols]);

    // Cleanup object URLs on unmount to prevent memory leaks
    useEffect(() => {
        return () => {
            if (previewData?.url) {
                URL.revokeObjectURL(previewData.url);
            }
        };
    }, [previewData?.url]);

    // ── Mutations ──
    const importMut = useMutation({
        mutationFn: (data: { files: File[], useAI: boolean, statementType: string }) =>
            importBankStatement(data.files, data.useAI, data.statementType),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['categories'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
            queryClient.invalidateQueries({ queryKey: ['bank-statements'] });
        }
    });

    const deleteMut = useMutation({
        mutationFn: deleteBankStatement,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['bank-statements'] });
            queryClient.invalidateQueries({ queryKey: ['transactions'] });
            setDeletingId(null);
        },
        onError: () => {
            setDeletingId(null);
        }
    });

    const updateSettingsMut = useMutation({
        mutationFn: updateSettings,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['settings'] });
        }
    });

    // ── Handlers ──
    const handleFiles = (files: FileList | File[]) => {
        const fileArray = Array.from(files);
        if (fileArray.length === 0) return;
        importMut.reset();
        importMut.mutate({ files: fileArray, useAI, statementType });
    };

    const handleDelete = (id: string) => {
        if (confirm(t('bankStatements.deleteConfirm', 'Are you sure you want to delete this statement?'))) {
            setDeletingId(id);
            deleteMut.mutate(id);
        }
    };

    const handleApplyFilters = () => {
        setAppliedYear(selectedYear);
        setAppliedType(selectedType);
        setAppliedAccount(selectedAccount);
    };

    const handleSort = (field: SortField) => {
        if (sortField === field) {
            setSortDirection(prev => prev === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
    };

    const toggleColumn = (col: ColKey) => {
        setVisibleCols(prev => {
            const next = { ...prev, [col]: !prev[col] };
            updateSettingsMut.mutate({ bank_statements_visible_cols: JSON.stringify(next) });
            return next;
        });
    };

    const handlePreview = async (id: string) => {
        setIsPreviewLoading(id);

        if (previewData?.url) {
            URL.revokeObjectURL(previewData.url);
        }

        try {
            const blob = await fetchBankStatementBlob(id);
            const type = blob.type.toLowerCase();

            if (type.includes('pdf')) {
                const url = URL.createObjectURL(blob);
                setPreviewData({ url, text: null, type, statementId: id });
            } else if (type.includes('csv')) {
                const text = await blob.text();
                setPreviewData({ url: null, text, type, statementId: id });
            } else {
                setPreviewData({ url: null, text: null, type, statementId: id });
            }
        } catch (error) {
            console.error("Failed to load preview:", error);
            alert(t('bankStatements.errors.previewFailed', 'Failed to load document preview.'));
        } finally {
            setIsPreviewLoading(null);
        }
    };

    // ── Data Processing & Filtering ──
    const { uniqueYears, uniqueTypes, uniqueAccounts } = useMemo(() => {
        if (!statements) return { uniqueYears: [], uniqueTypes: [], uniqueAccounts: [] };
        return {
            uniqueYears: Array.from(new Set(statements.map(s => new Date(s.end_date).getFullYear().toString()))).sort((a, b) => b.localeCompare(a)),
            uniqueTypes: Array.from(new Set(statements.map(s => s.statement_type))).sort(),
            uniqueAccounts: Array.from(new Set(statements.map(s => s.iban))).sort(),
        };
    }, [statements]);

    const filteredStatements = useMemo(() => {
        return (statements || []).filter(s => {
            if (appliedYear !== 'All' && new Date(s.end_date).getFullYear().toString() !== appliedYear) return false;
            if (appliedType !== 'All' && s.statement_type !== appliedType) return false;
            if (appliedAccount !== 'All' && s.iban !== appliedAccount) return false;
            return true;
        });
    }, [statements, appliedYear, appliedType, appliedAccount]);

    const sortedStatements = useMemo(() => {
        return [...filteredStatements].sort((a, b) => {
            let valA: any, valB: any;
            switch (sortField) {
                case 'statement_info': valA = a.iban; valB = b.iban; break;
                case 'statement_type': valA = a.statement_type; valB = b.statement_type; break;
                case 'period': valA = new Date(a.end_date).getTime(); valB = new Date(b.end_date).getTime(); break;
                case 'transaction_count': valA = a.transaction_count; valB = b.transaction_count; break;
                case 'new_balance': valA = a.new_balance; valB = b.new_balance; break;
                default: valA = 0; valB = 0;
            }
            if (valA < valB) return sortDirection === 'asc' ? -1 : 1;
            if (valA > valB) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        });
    }, [filteredStatements, sortField, sortDirection]);

    const importResponse: ImportBatchResponse | undefined = importMut.data;
    const hasReconcilableImport = importResponse?.results.some((r: ImportResult) => {
        const fname = r.filename.toLowerCase();
        return r.status === 'imported' && (fname.includes('.xls') || fname.includes('extra') || fname.includes('tagesgeld'));
    });

    const SortableHeader = ({ field, label, align = 'left' }: { field: SortField, label: string, align?: 'left' | 'right' | 'center' }) => (
        <th className={`px-6 py-4 font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800/80 transition-colors select-none ${align === 'right' ? 'text-right' : align === 'center' ? 'text-center' : 'text-left'}`} onClick={() => handleSort(field)}>
            <div className={`flex items-center gap-1.5 inline-flex ${align === 'right' ? 'flex-row-reverse' : ''}`}>
                {label}
                <div className="flex flex-col opacity-50 ml-1">
                    {sortField === field ? (
                        sortDirection === 'asc' ? <ChevronUp size={14} className="text-indigo-600 dark:text-indigo-400 opacity-100" /> : <ChevronDown size={14} className="text-indigo-600 dark:text-indigo-400 opacity-100" />
                    ) : <ChevronDown size={14} className="opacity-0 group-hover:opacity-50" />}
                </div>
            </div>
        </th>
    );

    const visibleColCount = Object.values(visibleCols).filter(Boolean).length + 1;

    if (isLoading) return <div className="p-8 text-center text-gray-500"><Loader2 size={24} className="animate-spin mx-auto text-indigo-500 mb-2" />{t('bankStatements.table.loading')}</div>;

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <div className="flex justify-between items-center">
                <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100">{t('bankStatements.title')}</h1>
            </div>

            {/* ── Drag & Drop Import Zone ── */}
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row gap-4 items-center">
                <div
                    onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                    onDragLeave={() => setDragOver(false)}
                    onDrop={(e) => { e.preventDefault(); setDragOver(false); if (e.dataTransfer.files?.length > 0) handleFiles(e.dataTransfer.files); }}
                    onClick={() => inputRef.current?.click()}
                    className={`flex-1 w-full border-2 border-dashed rounded-xl p-4 flex items-center justify-center gap-4 cursor-pointer transition-all duration-200 ${
                        dragOver ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20 scale-[1.01]' : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50 dark:bg-gray-800/30 hover:bg-gray-100 dark:hover:bg-gray-800/50'
                    }`}
                >
                    <input
                        type="file"
                        multiple
                        className="hidden"
                        ref={inputRef}
                        accept=".pdf,.csv,.xls,.xlsx"
                        onChange={(e) => e.target.files && handleFiles(e.target.files)}
                    />
                    <div className={`p-2 rounded-lg transition-colors ${dragOver ? 'bg-indigo-200 dark:bg-indigo-800 text-indigo-700 dark:text-indigo-300' : 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400'}`}>
                        <FileUp size={24} />
                    </div>
                    <div>
                        {importMut.isPending ? (
                            <p className="text-sm font-medium text-gray-900 dark:text-gray-100 animate-pulse">{t('bankStatements.import.uploading')}</p>
                        ) : (
                            <>
                                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                                    <span className="text-indigo-600 dark:text-indigo-400">{t('bankStatements.import.clickToUpload')}</span> {t('bankStatements.import.orDrag')}
                                </p>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('bankStatements.import.acceptedFiles')}</p>
                            </>
                        )}
                    </div>
                </div>

                <div className="flex flex-row md:flex-col gap-3 w-full md:w-auto md:min-w-[200px] justify-center">
                    <select
                        value={statementType}
                        onChange={(e) => setStatementType(e.target.value)}
                        className="text-sm bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 rounded-lg px-3 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-shadow w-full"
                    >
                        <option value="auto">{t('bankStatements.import.typeAuto')}</option>
                        <option value="giro">{t('bankStatements.import.typeGiro')}</option>
                        <option value="credit_card">{t('bankStatements.import.typeCc')}</option>
                        <option value="extra_account">{t('bankStatements.import.typeExtra')}</option>
                    </select>

                    <label className="flex items-center gap-2 cursor-pointer group px-1">
                        <div className="relative flex items-center">
                            <input
                                type="checkbox"
                                className="sr-only peer"
                                checked={useAI}
                                onChange={(e) => setUseAI(e.target.checked)}
                            />
                            <div className="w-9 h-5 bg-gray-200 dark:bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 dark:after:border-gray-600 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-indigo-600"></div>
                        </div>
                        <span className="text-sm font-medium text-gray-600 dark:text-gray-400 flex items-center gap-1.5 group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">
                            <BrainCircuit size={16} />
                            {t('bankStatements.import.forceAi')}
                        </span>
                    </label>
                </div>
            </div>

            {/* ── Import Results Banner ── */}
            {importResponse && (
                <div className="animate-in slide-in-from-top-4 space-y-4">
                    <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                        <div className="px-4 py-3 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30 flex justify-between items-center">
                            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100">{t('bankStatements.results.title')}</h3>
                            <div className="flex items-center gap-4">
                                <span className="text-sm font-medium text-gray-500 dark:text-gray-400">
                                    {t('bankStatements.results.importedOf', { imported: importResponse.summary.imported, total: importResponse.summary.total })}
                                </span>
                                <button onClick={() => importMut.reset()} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                                    <X size={16} />
                                </button>
                            </div>
                        </div>
                        <ul className="divide-y divide-gray-100 dark:divide-gray-800/50">
                            {importResponse.results.map((res: ImportResult, idx: number) => (
                                <li key={idx} className="px-4 py-3 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                                    <div className="flex items-center gap-3 truncate">
                                        {res.status === 'imported' && <CheckCircle2 size={16} className="text-green-500 dark:text-green-400 shrink-0" />}
                                        {res.status === 'duplicate' && <AlertCircle size={16} className="text-yellow-500 dark:text-yellow-400 shrink-0" />}
                                        {res.status === 'error' && <XCircle size={16} className="text-red-500 dark:text-red-400 shrink-0" />}
                                        <div className="truncate">
                                            <p className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">{res.filename}</p>
                                            {res.error && <p className="text-xs text-red-600 dark:text-red-400 mt-0.5 truncate">{res.error}</p>}
                                        </div>
                                    </div>
                                    <div className="shrink-0 ml-4 text-xs">
                                        {res.status === 'imported' && <span className="inline-flex items-center px-2 py-1 rounded-md bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 font-medium">{t('bankStatements.results.statusImported')}</span>}
                                        {res.status === 'duplicate' && <span className="inline-flex items-center px-2 py-1 rounded-md bg-yellow-50 dark:bg-yellow-900/20 text-yellow-700 dark:text-yellow-400 font-medium">{t('bankStatements.results.statusSkipped')}</span>}
                                        {res.status === 'error' && <span className="inline-flex items-center px-2 py-1 rounded-md bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 font-medium">{t('bankStatements.results.statusFailed')}</span>}
                                    </div>
                                </li>
                            ))}
                        </ul>
                    </div>

                    {hasReconcilableImport && (
                        <div className="rounded-xl bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-200 dark:border-indigo-800/50 shadow-sm p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                            <div>
                                <h3 className="font-medium text-sm text-indigo-900 dark:text-indigo-300">{t('bankStatements.results.reconcilableTitle')}</h3>
                                <p className="text-xs text-indigo-700 dark:text-indigo-400 mt-0.5">{t('bankStatements.results.reconcilableDesc')}</p>
                            </div>
                            <Link to="/reconcile" className="whitespace-nowrap px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition-colors">
                                {t('bankStatements.results.goToReconcile')}
                            </Link>
                        </div>
                    )}
                </div>
            )}

            {/* ── Filters & Column Toggles ── */}
            <div className="bg-white dark:bg-gray-900 p-4 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-4">
                <div className="flex flex-wrap items-center gap-4">
                    <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400">
                        <Filter size={18} />
                        <span className="text-sm font-medium">{t('bankStatements.filters.title')}</span>
                    </div>

                    <select value={selectedYear} onChange={e => setSelectedYear(e.target.value)} className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500 transition-shadow">
                        <option value="All">{t('bankStatements.filters.allYears')}</option>
                        {uniqueYears.map(y => <option key={y} value={y}>{y}</option>)}
                    </select>

                    <select value={selectedType} onChange={e => setSelectedType(e.target.value)} className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500 transition-shadow">
                        <option value="All">{t('bankStatements.filters.allTypes')}</option>
                        {uniqueTypes.map(type => (
                            <option key={type} value={type}>
                                {type === 'credit_card' ? t('bankStatements.filters.typeCc') : type === 'extra_account' ? t('bankStatements.filters.typeExtra') : t('bankStatements.filters.typeGiro')}
                            </option>
                        ))}
                    </select>

                    <select value={selectedAccount} onChange={e => setSelectedAccount(e.target.value)} className="bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-lg px-3 py-1.5 outline-none focus:ring-2 focus:ring-indigo-500 transition-shadow">
                        <option value="All">{t('bankStatements.filters.allAccounts')}</option>
                        {uniqueAccounts.map(a => <option key={a} value={a}>{a}</option>)}
                    </select>

                    <button
                        onClick={handleApplyFilters}
                        className="px-4 py-1.5 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-sm font-medium rounded-lg hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors"
                    >
                        {t('bankStatements.filters.apply')}
                    </button>
                </div>

                <div className="relative">
                    <button
                        onClick={() => setShowColMenu(!showColMenu)}
                        className="flex items-center gap-2 px-3 py-1.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                    >
                        <Columns size={16} />
                        {t('bankStatements.filters.columns')}
                    </button>

                    {showColMenu && (
                        <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-lg z-10 overflow-hidden">
                            <div className="p-2 space-y-1">
                                {[
                                    { key: 'statementInfo', label: t('bankStatements.columns.statementInfo') },
                                    { key: 'type', label: t('bankStatements.columns.type') },
                                    { key: 'period', label: t('bankStatements.columns.period') },
                                    { key: 'transactions', label: t('bankStatements.columns.transactions') },
                                    { key: 'newBalance', label: t('bankStatements.columns.newBalance') },
                                ].map(({ key, label }) => (
                                    <button
                                        key={key}
                                        onClick={() => toggleColumn(key as ColKey)}
                                        className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
                                    >
                                        {label}
                                        {visibleCols[key as ColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
                                    </button>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* ── Data Table ── */}
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
                <div className="overflow-x-auto [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700">
                    <table className="w-full text-left text-sm whitespace-nowrap">
                        <thead className="bg-gray-50 dark:bg-gray-800/50 text-gray-600 dark:text-gray-400 border-b border-gray-200 dark:border-gray-800">
                        <tr>
                            {visibleCols.statementInfo && <SortableHeader field="statement_info" label={t('bankStatements.columns.statementInfo')} />}
                            {visibleCols.type && <SortableHeader field="statement_type" label={t('bankStatements.columns.type')} />}
                            {visibleCols.period && <SortableHeader field="period" label={t('bankStatements.columns.period')} />}
                            {visibleCols.transactions && <SortableHeader field="transaction_count" label={t('bankStatements.columns.transactions')} align="right" />}
                            {visibleCols.newBalance && <SortableHeader field="new_balance" label={t('bankStatements.columns.newBalance')} align="right" />}
                            <th className="px-6 py-4 font-medium text-right">{t('bankStatements.columns.actions')}</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
                        {sortedStatements.length === 0 ? (
                            <tr>
                                <td colSpan={visibleColCount} className="px-6 py-12 text-center text-gray-500">
                                    <div className="flex flex-col items-center justify-center">
                                        <FileText size={40} className="text-gray-300 dark:text-gray-700 mb-3" />
                                        {t('bankStatements.table.noData', 'No statements found')}
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            sortedStatements.map((stmt) => (
                                <tr key={stmt.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors group">
                                    {visibleCols.statementInfo && (
                                        <td className="px-6 py-4 flex items-center gap-3">
                                            <div className="p-2 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-lg group-hover:bg-indigo-100 dark:group-hover:bg-indigo-900/50 transition-colors">
                                                <FileText size={18} />
                                            </div>
                                            <div>
                                                <div className="font-medium text-gray-900 dark:text-gray-100">{stmt.iban}</div>
                                                <div className="text-xs text-gray-500">{t('bankStatements.table.statementNo', { no: stmt.statement_no })}</div>
                                            </div>
                                        </td>
                                    )}
                                    {visibleCols.type && (
                                        <td className="px-6 py-4">
                                            <span className={`px-2.5 py-1 text-xs font-medium rounded-full ${
                                                stmt.statement_type === 'credit_card' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
                                                    : stmt.statement_type === 'extra_account' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                                                        : 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                                            }`}>
                                                {stmt.statement_type === 'credit_card' ? t('bankStatements.filters.typeCc') : stmt.statement_type === 'extra_account' ? t('bankStatements.filters.typeExtra') : t('bankStatements.filters.typeGiro')}
                                            </span>
                                        </td>
                                    )}
                                    {visibleCols.period && (
                                        <td className="px-6 py-4 text-gray-600 dark:text-gray-400">
                                            {new Date(stmt.start_date).toLocaleDateString()} - {new Date(stmt.end_date).toLocaleDateString()}
                                        </td>
                                    )}
                                    {visibleCols.transactions && (
                                        <td className="px-6 py-4 text-right font-medium text-gray-900 dark:text-gray-100">{stmt.transaction_count}</td>
                                    )}
                                    {visibleCols.newBalance && (
                                        <td className="px-6 py-4 text-right font-medium text-gray-900 dark:text-gray-100">{fmtCurrency(stmt.new_balance, stmt.currency)}</td>
                                    )}
                                    <td className="px-6 py-4 text-right">
                                        <div className="flex items-center justify-end gap-2">
                                            {/* NEW SHORTCUT BUTTON */}
                                            <button
                                                onClick={() => navigate(`/transactions?statement=${stmt.id}`)}
                                                className="p-2 text-gray-400 hover:text-cyan-600 hover:bg-cyan-50 dark:hover:bg-cyan-900/30 rounded-lg transition-colors"
                                                title={t('bankStatements.actions.viewTransactions', 'View Transactions')}
                                            >
                                                <List size={18} />
                                            </button>
                                            <button
                                                onClick={() => handlePreview(stmt.id)}
                                                disabled={isPreviewLoading === stmt.id}
                                                className="p-2 text-gray-400 hover:text-fuchsia-600 hover:bg-fuchsia-50 dark:hover:bg-fuchsia-900/30 rounded-lg transition-colors disabled:opacity-50"
                                                title={t('bankStatements.actions.preview', 'Preview')}
                                            >
                                                {isPreviewLoading === stmt.id ? <Loader2 size={18} className="animate-spin" /> : <FileText size={18} />}
                                            </button>
                                            <button
                                                onClick={() => setViewingStatement(stmt)}
                                                className="p-2 text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 rounded-lg transition-colors"
                                                title={t('bankStatements.actions.view')}
                                            >
                                                <Eye size={18} />
                                            </button>
                                            <button
                                                onClick={() => downloadBankStatement(stmt.id)}
                                                className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-colors"
                                                title={t('bankStatements.actions.download')}
                                            >
                                                <Download size={18} />
                                            </button>
                                            <button
                                                onClick={() => handleDelete(stmt.id)}
                                                disabled={deletingId === stmt.id}
                                                className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors disabled:opacity-50"
                                                title={t('bankStatements.actions.delete')}
                                            >
                                                {deletingId === stmt.id ? <AlertTriangle size={18} className="animate-pulse" /> : <Trash2 size={18} />}
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))
                        )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* ── Read-Only Details Modal ── */}
            {viewingStatement && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
                    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden max-h-[90vh] flex flex-col">
                        <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                                <Eye className="text-emerald-500" size={20} />
                                {t('bankStatements.modal.title')}
                            </h3>
                            <button onClick={() => setViewingStatement(null)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
                                <X size={20} />
                            </button>
                        </div>
                        <div className="p-6 space-y-6 overflow-y-auto [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700">
                            <div className="grid grid-cols-2 gap-y-4 gap-x-6 text-sm">
                                <div>
                                    <p className="text-gray-500 dark:text-gray-400 mb-1">{t('bankStatements.modal.account')}</p>
                                    <p className="font-medium text-gray-900 dark:text-gray-100">{viewingStatement.iban}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500 dark:text-gray-400 mb-1">{t('bankStatements.modal.statementNo')}</p>
                                    <p className="font-medium text-gray-900 dark:text-gray-100">#{viewingStatement.statement_no}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500 dark:text-gray-400 mb-1">{t('bankStatements.modal.startDate')}</p>
                                    <p className="font-medium text-gray-900 dark:text-gray-100">{new Date(viewingStatement.start_date).toLocaleDateString()}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500 dark:text-gray-400 mb-1">{t('bankStatements.modal.endDate')}</p>
                                    <p className="font-medium text-gray-900 dark:text-gray-100">{new Date(viewingStatement.end_date).toLocaleDateString()}</p>
                                </div>
                                <div>
                                    <p className="text-gray-500 dark:text-gray-400 mb-1">{t('bankStatements.modal.type')}</p>
                                    <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${
                                        viewingStatement.statement_type === 'credit_card' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400'
                                            : viewingStatement.statement_type === 'extra_account' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                                                : 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                                    }`}>
                                        {viewingStatement.statement_type === 'credit_card' ? t('bankStatements.filters.typeCc') : viewingStatement.statement_type === 'extra_account' ? t('bankStatements.filters.typeExtra') : t('bankStatements.filters.typeGiro')}
                                    </span>
                                </div>
                            </div>

                            <hr className="border-gray-100 dark:border-gray-800" />

                            <div className="space-y-3 text-sm">
                                <div className="flex justify-between items-center">
                                    <span className="text-gray-500 dark:text-gray-400">{t('bankStatements.modal.totalTxns')}</span>
                                    <span className="font-medium text-gray-900 dark:text-gray-100">{viewingStatement.transaction_count}</span>
                                </div>
                                <div className="flex justify-between items-center pt-2 border-t border-gray-100 dark:border-gray-800">
                                    <span className="font-medium text-gray-700 dark:text-gray-300">{t('bankStatements.modal.endingBalance')}</span>
                                    <span className="font-mono font-bold text-lg text-emerald-600 dark:text-emerald-400">{fmtCurrency(viewingStatement.new_balance, viewingStatement.currency)}</span>
                                </div>
                            </div>

                            <div className="pt-4 flex justify-end gap-2 border-t border-gray-100 dark:border-gray-800">
                                <button type="button" onClick={() => setViewingStatement(null)} className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">
                                    {t('bankStatements.modal.close')}
                                </button>
                                <button
                                    onClick={() => navigate(`/transactions?statement=${viewingStatement.id}`)}
                                    className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 flex items-center gap-2"
                                >
                                    <List size={16} />
                                    {t('bankStatements.modal.viewTxns')}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* ── Preview Document Modal ── */}
            {previewData && (
                <PreviewStatementModal
                    data={previewData}
                    onClose={() => {
                        if (previewData.url) URL.revokeObjectURL(previewData.url);
                        setPreviewData(null);
                    }}
                    onDownload={(id) => downloadBankStatement(id)}
                />
            )}
        </div>
    );
}

// ── Shared Preview Modal Component ──
function PreviewStatementModal({
                                   data,
                                   onClose,
                                   onDownload
                               }: {
    data: PreviewData,
    onClose: () => void,
    onDownload: (id: string) => void
}) {
    const { t } = useTranslation();

    return (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-5xl h-[90vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        {data.type?.includes('pdf') ? <FileText className="text-fuchsia-500" size={20} /> : <FileSpreadsheet className="text-emerald-500" size={20} />}
                        {t('bankStatements.modal.previewTitle', 'Preview Statement')}
                    </h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                        <X size={20} />
                    </button>
                </div>
                <div className="flex-1 w-full bg-gray-200 dark:bg-gray-950 p-2 sm:p-4 overflow-hidden flex flex-col">
                    {data.url ? (
                        // PDF View
                        <iframe src={`${data.url}#toolbar=0`} className="w-full h-full rounded-xl border border-gray-300 dark:border-gray-800 shadow-inner bg-white" title="PDF Preview" />
                    ) : data.text !== null ? (
                        // CSV View
                        <div className="w-full h-full rounded-xl border border-gray-300 dark:border-gray-700 shadow-inner bg-gray-50 dark:bg-gray-900 overflow-auto p-4 md:p-6">
                            <pre className="text-[13px] leading-relaxed font-mono text-gray-800 dark:text-gray-300 whitespace-pre-wrap select-text">
                                {data.text}
                            </pre>
                        </div>
                    ) : (
                        // Excel Fallback View
                        <div className="flex-1 flex flex-col items-center justify-center text-center p-8 bg-white dark:bg-gray-900 rounded-xl border border-gray-300 dark:border-gray-800 shadow-inner">
                            <FileSpreadsheet size={48} className="text-emerald-500 mb-4 opacity-80" />
                            <h4 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">
                                {t('bankStatements.modal.previewNotSupportedTitle', 'Spreadsheet Preview Not Supported')}
                            </h4>
                            <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm mb-6">
                                {t('bankStatements.modal.previewNotSupportedDesc', 'Your browser cannot natively preview Excel files inline. Please download the file to view its contents.')}
                            </p>
                            <button
                                onClick={() => onDownload(data.statementId)}
                                className="px-5 py-2.5 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors flex items-center gap-2 shadow-sm"
                            >
                                <Download size={18} />
                                {t('bankStatements.actions.download', 'Download File')}
                            </button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}