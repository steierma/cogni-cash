import React, { useRef, useState, useMemo, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import {
    Trash2, Download, Edit2, Upload, X, Search, Database,
    ChevronUp, ChevronDown, Loader2, Filter, BarChart3,
    TrendingUp, Store, CheckSquare, Square, Layers, FileText
} from 'lucide-react';
import { FilePreview } from '../components/FilePreview';
import {
    deleteInvoice, fetchInvoices, importInvoice,
    downloadInvoiceFile, updateInvoice, fetchCategories, getInvoicePreviewUrl,
    type InvoiceUpdatePayload
} from '../api/client';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import type { Invoice, Category } from '../api/types';

type SortKey = 'issued_at' | 'vendor' | 'category' | 'description' | 'amount';
type SortDir = 'asc' | 'desc';

interface FilterState {
    search: string;
    category: string;
    from: string;
    to: string;
    amountMin: string;
    amountMax: string;
}

const initialFilters: FilterState = {
    search: '', category: 'all', from: '', to: '', amountMin: '', amountMax: ''
};

export default function InvoicesPage() {
    const { t, i18n } = useTranslation();
    const [searchParams, setSearchParams] = useSearchParams();
    const qc = useQueryClient();
    const fileInputRef = useRef<HTMLInputElement>(null);

    // --- State ---
    const [dragOver, setDragOver] = useState(false);

    const initialFiltersFromURL: FilterState = useMemo(() => {
        return {
            search: searchParams.get('search') || '',
            category: searchParams.get('category') || 'all',
            from: searchParams.get('from') || '',
            to: searchParams.get('to') || '',
            amountMin: searchParams.get('amountMin') || '',
            amountMax: searchParams.get('amountMax') || ''
        };
    }, [searchParams]);

    const [draftFilters, setDraftFilters] = useState<FilterState>(initialFiltersFromURL);
    const [appliedFilters, setAppliedFilters] = useState<FilterState>(initialFiltersFromURL);

    const [sortKey, setSortKey] = useState<SortKey>(() => {
        const sk = searchParams.get('sortKey') as SortKey;
        if (['issued_at', 'vendor', 'category', 'description', 'amount'].includes(sk)) return sk;
        return 'issued_at';
    });
    const [sortDir, setSortDir] = useState<SortDir>(() => {
        const sd = searchParams.get('sortDir') as SortDir;
        return sd === 'asc' ? 'asc' : 'desc';
    });
    const [showVisuals, setShowVisuals] = useState(searchParams.get('visuals') !== 'false');

    // Update URL when state changes
    useEffect(() => {
        const next = new URLSearchParams();
        if (appliedFilters.search) next.set('search', appliedFilters.search);
        if (appliedFilters.category !== 'all') next.set('category', appliedFilters.category);
        if (appliedFilters.from) next.set('from', appliedFilters.from);
        if (appliedFilters.to) next.set('to', appliedFilters.to);
        if (appliedFilters.amountMin) next.set('amountMin', appliedFilters.amountMin);
        if (appliedFilters.amountMax) next.set('amountMax', appliedFilters.amountMax);
        
        if (sortKey !== 'issued_at') next.set('sortKey', sortKey);
        if (sortDir !== 'desc') next.set('sortDir', sortDir);
        if (!showVisuals) next.set('visuals', 'false');

        const currentStr = searchParams.toString();
        const nextStr = next.toString();
        if (currentStr !== nextStr) {
            setSearchParams(next, { replace: true });
        }
    }, [appliedFilters, sortKey, sortDir, showVisuals, setSearchParams, searchParams]);

    // Batch Selection
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

    const [editingInvoice, setEditingInvoice] = useState<Invoice | null>(null);
    const [editDesc, setEditDesc] = useState('');
    const [editCatId, setEditCatId] = useState<string>('');
    const [editVendor, setEditVendor] = useState('');
    const [editAmount, setEditAmount] = useState('');
    const [editCurrency, setEditCurrency] = useState('EUR');
    const [editIssuedAt, setEditIssuedAt] = useState('');

    const [previewInfo, setPreviewInfo] = useState<{url: string, mimeType: string} | null>(null);
    const [isPreviewLoading, setIsPreviewLoading] = useState<string | null>(null);

    // --- Data Fetching ---
    const { data: invoices = [], isLoading } = useQuery({
        queryKey: ['invoices'],
        queryFn: fetchInvoices,
    });

    const { data: categories = [] } = useQuery({
        queryKey: ['categories'],
        queryFn: fetchCategories,
    });

    // --- Mutations ---
    const deleteMutation = useMutation({
        mutationFn: deleteInvoice,
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['invoices'] });
            setSelectedIds(new Set());
        },
    });

    const importMutation = useMutation({
        mutationFn: importInvoice,
        onSuccess: () => qc.invalidateQueries({ queryKey: ['invoices'] }),
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: InvoiceUpdatePayload }) => updateInvoice(id, data),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['invoices'] });
            setEditingInvoice(null);
        },
    });

    const batchCatMutation = useMutation({
        mutationFn: async ({ ids, categoryId }: { ids: string[]; categoryId: string }) => {
            await Promise.all(ids.map(id => updateInvoice(id, { category_id: categoryId || null })));
        },
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['invoices'] });
            setSelectedIds(new Set());
        },
    });

    // --- Handlers ---
    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (file) importMutation.mutate(file);
        if (fileInputRef.current) fileInputRef.current.value = '';
    };

    const handleFiles = (files: FileList | File[]) => {
        const file = files[0];
        if (file) {
            importMutation.mutate(file);
        }
        if (fileInputRef.current) fileInputRef.current.value = '';
    };

    const onDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setDragOver(false);
        if (e.dataTransfer.files?.length > 0) handleFiles(e.dataTransfer.files);
    };

    const openEditModal = (inv: Invoice) => {
        setEditingInvoice(inv);
        setEditDesc(inv.description || '');
        setEditCatId(inv.category_id || '');
        setEditVendor(inv.vendor?.name || '');
        setEditAmount(inv.amount != null ? String(inv.amount) : '');
        setEditCurrency(inv.currency || 'EUR');
        setEditIssuedAt(inv.issued_at ? inv.issued_at.slice(0, 10) : '');
    };

    const handleSaveEdit = () => {
        if (!editingInvoice) return;
        const parsedAmount = parseFloat(editAmount.replace(',', '.'));
        updateMutation.mutate({
            id: editingInvoice.id,
            data: {
                vendor: { id: editingInvoice.vendor?.id ?? '', name: editVendor },
                description: editDesc,
                category_id: editCatId || null,
                amount: isNaN(parsedAmount) ? editingInvoice.amount : parsedAmount,
                currency: editCurrency || editingInvoice.currency,
                issued_at: editIssuedAt || editingInvoice.issued_at,
            }
        });
    };

    const handlePreview = async (id: string) => {
        try {
            setIsPreviewLoading(id);
            const info = await getInvoicePreviewUrl(id);
            setPreviewInfo(info);
        } catch {
            alert(t('invoices.previewFailed'));
        } finally {
            setIsPreviewLoading(null);
        }
    };

    const closePreview = () => {
        if (previewInfo) URL.revokeObjectURL(previewInfo.url);
        setPreviewInfo(null);
    };

    const toggleSort = (key: SortKey) => {
        if (sortKey === key) setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
        else {
            setSortKey(key);
            setSortDir('desc');
        }
    };

    const handleApplyFilters = (e?: React.FormEvent) => {
        e?.preventDefault();
        setAppliedFilters(draftFilters);
        setSelectedIds(new Set());
    };

    const clearFilters = () => {
        setDraftFilters(initialFilters);
        setAppliedFilters(initialFilters);
        setSelectedIds(new Set());
    };

    // --- Derived Data & Filtering ---
    const [minDate, maxDate] = useMemo((): [string, string] => {
        const days = invoices.map((i) => i.issued_at ? i.issued_at.slice(0, 10) : '').filter(Boolean).sort();
        if (days.length === 0) return ['', ''];
        return [days[0], days[days.length - 1]];
    }, [invoices]);

    const filtered = useMemo(() => {
        let rows = invoices;

        if (appliedFilters.search.trim()) {
            const q = appliedFilters.search.toLowerCase();
            rows = rows.filter((inv) =>
                (inv.vendor?.name || '').toLowerCase().includes(q) ||
                (inv.description || '').toLowerCase().includes(q)
            );
        }

        if (appliedFilters.category !== 'all') {
            if (appliedFilters.category === 'uncategorized') rows = rows.filter((inv) => !inv.category_id);
            else rows = rows.filter((inv) => inv.category_id === appliedFilters.category);
        }

        const day = (d: string) => d?.length > 10 ? d.slice(0, 10) : d;
        if (appliedFilters.from) rows = rows.filter((inv) => day(inv.issued_at) >= appliedFilters.from);
        if (appliedFilters.to) rows = rows.filter((inv) => day(inv.issued_at) <= appliedFilters.to);

        const parseAmt = (val: string) => parseFloat(val.replace(',', '.'));
        if (appliedFilters.amountMin !== '') rows = rows.filter((inv) => inv.amount >= parseAmt(appliedFilters.amountMin));
        if (appliedFilters.amountMax !== '') rows = rows.filter((inv) => inv.amount <= parseAmt(appliedFilters.amountMax));

        rows = [...rows].sort((a, b) => {
            let cmp = 0;
            if (sortKey === 'issued_at') cmp = (a.issued_at || '').localeCompare(b.issued_at || '');
            else if (sortKey === 'vendor') cmp = (a.vendor?.name || '').localeCompare(b.vendor?.name || '');
            else if (sortKey === 'category') {
                const catA = categories.find(c => c.id === a.category_id)?.name || '';
                const catB = categories.find(c => c.id === b.category_id)?.name || '';
                cmp = catA.localeCompare(catB);
            }
            else if (sortKey === 'description') cmp = (a.description || '').localeCompare(b.description || '');
            else if (sortKey === 'amount') cmp = a.amount - b.amount;
            return sortDir === 'asc' ? cmp : -cmp;
        });

        return rows;
    }, [invoices, appliedFilters, sortKey, sortDir, categories]);

    const toggleSelect = (id: string) => {
        const next = new Set(selectedIds);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        setSelectedIds(next);
    };

    const toggleSelectAll = () => {
        if (selectedIds.size === filtered.length && filtered.length > 0) setSelectedIds(new Set());
        else setSelectedIds(new Set(filtered.map((i: Invoice) => i.id)));
    };

    const isDraftDirty = JSON.stringify(draftFilters) !== JSON.stringify(appliedFilters);
    const hasAppliedFilters = JSON.stringify(appliedFilters) !== JSON.stringify(initialFilters);
    const totalAmount = filtered.reduce((sum, inv) => sum + inv.amount, 0);

    // --- Statistics & Visuals ---
    const { topVendors, categoryBreakdown, maxVendorAmt, maxCatAmt } = useMemo(() => {
        const vMap: Record<string, number> = {};
        const cMap: Record<string, number> = {};

        filtered.forEach(inv => {
            const vName = inv.vendor?.name || 'Unknown';
            vMap[vName] = (vMap[vName] || 0) + inv.amount;

            const cName = categories.find(c => c.id === inv.category_id)?.name || 'Uncategorized';
            cMap[cName] = (cMap[cName] || 0) + inv.amount;
        });

        const vArr = Object.entries(vMap).sort((a, b) => b[1] - a[1]).slice(0, 5);
        const cArr = Object.entries(cMap).sort((a, b) => b[1] - a[1]).slice(0, 5);

        return {
            topVendors: vArr,
            categoryBreakdown: cArr,
            maxVendorAmt: vArr.length ? vArr[0][1] : 1,
            maxCatAmt: cArr.length ? cArr[0][1] : 1
        };
    }, [filtered, categories]);

    const renderSortIcon = (k: SortKey) =>
        sortKey === k ? (sortDir === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />) : null;

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-28 animate-in fade-in duration-300">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <div className="flex justify-between items-center">
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                            <FileText className="text-indigo-600 dark:text-indigo-400" /> {t('invoices.title')}
                        </h1>
                    </div>
                    <p className="text-sm text-gray-400 dark:text-gray-500 mt-0.5">
                        {isLoading ? t('common.loading') : `${filtered.length.toLocaleString(i18n.language)} / ${invoices.length.toLocaleString(i18n.language)} ${t('invoices.title').toLowerCase()}`}
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    {filtered.length > 0 && (
                        <button
                            onClick={() => setShowVisuals(!showVisuals)}
                            className={`flex items-center gap-1.5 text-sm px-3 py-2 rounded-xl border transition-colors ${showVisuals
                                ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-100 dark:border-indigo-800/50 text-indigo-600 dark:text-indigo-400'
                                : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'
                            }`}
                        >
                            <BarChart3 size={16} /> <span className="hidden sm:inline">{showVisuals ? t('transactions.hideCharts') : t('transactions.showCharts')}</span>
                        </button>
                    )}
                    <input type="file" ref={fileInputRef} className="hidden" accept=".pdf,.png,.jpg,.jpeg,.webp,.gif" onChange={handleFileChange} />
                    <button
                        onClick={() => fileInputRef.current?.click()}
                        disabled={importMutation.isPending}
                        className="flex items-center space-x-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl transition-all shadow-sm disabled:opacity-70 disabled:cursor-not-allowed font-medium text-sm"
                    >
                        {importMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <Upload size={16} />}
                        <span className="hidden sm:inline">{importMutation.isPending ? t('invoices.uploading') : t('invoices.import')}</span>
                    </button>
                </div>
            </div>

            {/* Quick Upload Dropzone (New, aligning with other pages) */}
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row gap-4 items-center animate-in fade-in duration-500">
                <div
                    onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                    onDragLeave={() => setDragOver(false)}
                    onDrop={onDrop}
                    onClick={() => fileInputRef.current?.click()}
                    className={`flex-1 w-full border-2 border-dashed rounded-xl p-6 flex items-center justify-center gap-4 cursor-pointer transition-all duration-200 ${
                        dragOver ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20 scale-[1.01]' : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50 dark:bg-gray-800/30 hover:bg-gray-100 dark:hover:bg-gray-800/50'
                    }`}
                >
                    <div className={`p-3 rounded-xl transition-colors ${dragOver ? 'bg-indigo-200 dark:bg-indigo-800 text-indigo-700 dark:text-indigo-300' : 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-600 dark:text-indigo-400'}`}>
                        <Upload size={24} />
                    </div>
                    <div>
                        {importMutation.isPending ? (
                            <p className="text-sm font-medium text-gray-900 dark:text-gray-100 animate-pulse">{t('invoices.uploading')}</p>
                        ) : (
                            <>
                                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                                    <span className="text-indigo-600 dark:text-indigo-400">{t('bankStatements.import.clickToUpload')}</span> {t('bankStatements.import.orDrag')}
                                </p>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('payslips.pdfFormat')}</p>
                            </>
                        )}
                    </div>
                </div>

                <div className="flex flex-col gap-3 w-full md:w-auto md:min-w-[180px] justify-center items-center md:items-stretch">
                    <div className="text-[10px] text-gray-400 dark:text-gray-500 text-center px-2 leading-relaxed">
                        {t('invoices.subtitle')}
                    </div>
                </div>
            </div>

            {/* Filter Panel */}
            <form onSubmit={handleApplyFilters} className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-4 space-y-3">
                <div className="flex items-center justify-between mb-1">
                    <div className="flex items-center gap-2 text-sm font-medium text-gray-500 dark:text-gray-400">
                        <Filter size={14} /> {t('transactions.filters.title')}
                    </div>
                    <div className="flex items-center gap-3">
                        {isDraftDirty && (
                            <span className="text-[10px] font-bold text-indigo-500 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 px-2 py-0.5 rounded-full uppercase tracking-wider animate-pulse">
                                {t('transactions.filters.unapplied')}
                            </span>
                        )}
                        {(hasAppliedFilters || isDraftDirty) && (
                            <button type="button" onClick={clearFilters} className="text-xs text-indigo-500 dark:text-indigo-400 hover:underline flex items-center gap-1">
                                <X size={12} /> {t('transactions.clearFilters')}
                            </button>
                        )}
                    </div>
                </div>

                {/* Row 1: Search & Category */}
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                    <div className="relative sm:col-span-2">
                        <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500" />
                        <input
                            value={draftFilters.search}
                            onChange={(e) => setDraftFilters(f => ({ ...f, search: e.target.value }))}
                            placeholder={t('invoices.searchPlaceholder')}
                            className="w-full pl-8 pr-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                        />
                    </div>
                    <div>
                        <select
                            value={draftFilters.category}
                            onChange={(e) => setDraftFilters(f => ({ ...f, category: e.target.value }))}
                            className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                        >
                            <option value="all">{t('transactions.filters.allCategories')}</option>
                            <option value="uncategorized">⚠️ {t('transactions.filters.uncategorizedOnly')}</option>
                            {categories.filter(c => !c.deleted_at || c.id === draftFilters.category).map((c: Category) => <option key={c.id} value={c.id}>{c.name}</option>)}
                        </select>
                    </div>
                </div>

                {/* Row 2: Dates, Amounts & Button */}
                <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
                    <div>
                        <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.from')}</label>
                        <input
                            type="date"
                            value={draftFilters.from ?? minDate}
                            onChange={(e) => setDraftFilters(f => ({ ...f, from: e.target.value }))}
                            min={minDate} max={maxDate}
                            className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 [color-scheme:light] dark:[color-scheme:dark]"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.to')}</label>
                        <input
                            type="date"
                            value={draftFilters.to ?? maxDate}
                            onChange={(e) => setDraftFilters(f => ({ ...f, to: e.target.value }))}
                            min={minDate} max={maxDate}
                            className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50 [color-scheme:light] dark:[color-scheme:dark]"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.minAmount')}</label>
                        <input
                            type="number"
                            step="0.01"
                            placeholder="-∞"
                            value={draftFilters.amountMin}
                            onChange={(e) => setDraftFilters(f => ({ ...f, amountMin: e.target.value }))}
                            className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                        />
                    </div>
                    <div>
                        <label className="text-xs text-gray-400 dark:text-gray-500 mb-1 block">{t('transactions.filters.maxAmount')}</label>
                        <input
                            type="number"
                            step="0.01"
                            placeholder="∞"
                            value={draftFilters.amountMax}
                            onChange={(e) => setDraftFilters(f => ({ ...f, amountMax: e.target.value }))}
                            className="w-full py-2 px-3 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500/50"
                        />
                    </div>
                    <div className="flex items-end">
                        <button
                            type="submit"
                            className={`w-full py-2 px-4 rounded-lg font-medium text-sm flex items-center justify-center gap-2 transition-all ${isDraftDirty
                                ? 'bg-indigo-600 dark:bg-indigo-500 text-white shadow-md hover:bg-indigo-700 dark:hover:bg-indigo-600'
                                : hasAppliedFilters
                                    ? 'bg-gray-100 dark:bg-gray-800 text-gray-400 dark:text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-700'
                                    : 'bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-100 dark:hover:bg-indigo-900/50'
                            }`}
                        >
                            <Search size={14} />
                            {t('transactions.filters.search')}
                        </button>
                    </div>
                </div>
            </form>

            {/* Statistics & Visuals */}
            {!isLoading && filtered.length > 0 && showVisuals && (
                <div className="space-y-4 animate-in fade-in duration-300">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 text-center shadow-sm flex flex-col justify-center">
                            <p className="text-xs text-gray-400 dark:text-gray-500 uppercase tracking-wide mb-1">{t('transactions.filteredRows')}</p>
                            <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{filtered.length.toLocaleString(i18n.language)}</p>
                        </div>
                        <div className="bg-indigo-50 dark:bg-indigo-900/20 rounded-xl border border-indigo-100 dark:border-indigo-800/50 p-4 text-center shadow-sm flex flex-col justify-center">
                            <p className="text-xs text-indigo-600 dark:text-indigo-400 uppercase tracking-wide mb-1">{t('invoices.totalAmount')}</p>
                            <p className="text-2xl font-bold text-indigo-700 dark:text-indigo-300">{fmtCurrency(totalAmount)}</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                        {/* Top Vendors */}
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                            <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                <Store size={16} className="text-indigo-500 dark:text-indigo-400" />
                                {t('invoices.topVendors')}
                            </h3>
                            <div className="space-y-4">
                                {topVendors.length > 0 ? topVendors.map(([name, amt]) => (
                                    <div key={name} className="space-y-1.5">
                                        <div className="flex justify-between text-xs font-medium text-gray-600 dark:text-gray-400">
                                            <span className="truncate pr-4">{name}</span>
                                            <span className="font-mono">{fmtCurrency(amt)}</span>
                                        </div>
                                        <div className="w-full bg-gray-100 dark:bg-gray-800 h-1.5 rounded-full overflow-hidden">
                                            <div
                                                className="h-full rounded-full transition-all duration-500 bg-indigo-400 dark:bg-indigo-500"
                                                style={{ width: `${(amt / maxVendorAmt) * 100}%` }}
                                            />
                                        </div>
                                    </div>
                                )) : (
                                    <div className="text-sm text-gray-400 text-center py-4">{t('invoices.noVendorData')}</div>
                                )}
                            </div>
                        </div>

                        {/* Category Breakdown */}
                        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                            <h3 className="text-sm font-semibold text-gray-800 dark:text-gray-200 mb-6 flex items-center gap-2">
                                <TrendingUp size={16} className="text-indigo-500 dark:text-indigo-400" />
                                {t('analytics.expensesTitle')}
                            </h3>
                            <div className="space-y-4">
                                {categoryBreakdown.length > 0 ? categoryBreakdown.map(([name, amt]) => {
                                    const catInfo = categories.find(c => c.name === name);
                                    return (
                                        <div key={name} className="space-y-1.5">
                                            <div className="flex justify-between text-xs font-medium text-gray-600 dark:text-gray-400">
                                                <span className="truncate pr-4">{name}</span>
                                                <span className="font-mono">{fmtCurrency(amt)}</span>
                                            </div>
                                            <div className="w-full bg-gray-100 dark:bg-gray-800 h-1.5 rounded-full overflow-hidden">
                                                <div
                                                    className="h-full rounded-full transition-all duration-500"
                                                    style={{ width: `${(amt / maxCatAmt) * 100}%`, backgroundColor: catInfo?.color || '#94a3b8' }}
                                                />
                                            </div>
                                        </div>
                                    );
                                }) : (
                                    <div className="text-sm text-gray-400 text-center py-4">{t('invoices.noCategoryData')}</div>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Empty State */}
            {!isLoading && filtered.length === 0 && (
                <div className="flex flex-col items-center justify-center py-32 bg-white dark:bg-gray-900 rounded-2xl border border-dashed border-gray-200 dark:border-gray-800 text-gray-400 dark:text-gray-500 mt-4">
                    <div className="bg-indigo-50 dark:bg-indigo-900/20 p-4 rounded-full mb-4">
                        {hasAppliedFilters ? <Search size={32} className="text-indigo-400 dark:text-indigo-500" /> : <Database size={32} className="text-indigo-400 dark:text-indigo-500" />}
                    </div>
                    <p className="text-base font-medium text-gray-600 dark:text-gray-300">
                        {hasAppliedFilters ? t('transactions.noMatches') : t('invoices.noInvoices')}
                    </p>
                    {hasAppliedFilters && (
                        <button onClick={clearFilters} className="mt-2 text-sm text-indigo-500 dark:text-indigo-400 hover:underline">
                            {t('transactions.clearFilters')}
                        </button>
                    )}
                </div>
            )}

            {/* Data Table */}
            {!isLoading && filtered.length > 0 && (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden relative">
                    <div className="overflow-x-auto [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700">
                        <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800 text-sm">
                            <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide">
                            <tr>
                                <th className="px-4 py-3 text-left w-10">
                                    <button
                                        type="button"
                                        onClick={toggleSelectAll}
                                        className={`transition-colors ${selectedIds.size === filtered.length && filtered.length > 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                                    >
                                        {selectedIds.size === filtered.length && filtered.length > 0 ? <CheckSquare size={16}/> : <Square size={16}/>}
                                    </button>
                                </th>
                                <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap" onClick={() => toggleSort('issued_at')}>
                                    <span className="inline-flex items-center gap-1">{t('invoices.date')} {renderSortIcon('issued_at')}</span>
                                </th>
                                <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap" onClick={() => toggleSort('vendor')}>
                                    <span className="inline-flex items-center gap-1">{t('invoices.vendor')} {renderSortIcon('vendor')}</span>
                                </th>
                                <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap" onClick={() => toggleSort('category')}>
                                    <span className="inline-flex items-center gap-1">{t('invoices.category')} {renderSortIcon('category')}</span>
                                </th>
                                <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none" onClick={() => toggleSort('description')}>
                                    <span className="inline-flex items-center gap-1">{t('invoices.description')} {renderSortIcon('description')}</span>
                                </th>
                                <th className="px-4 py-3 text-right cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap" onClick={() => toggleSort('amount')}>
                                    <span className="inline-flex items-center gap-1 justify-end">{t('invoices.amount')} {renderSortIcon('amount')}</span>
                                </th>
                                <th className="px-4 py-3 text-right">{t('common.actions')}</th>
                            </tr>
                            </thead>
                            <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                            {filtered.map((inv: Invoice) => {
                                const currentCat = categories.find((c) => c.id === inv.category_id);
                                const isSelected = selectedIds.has(inv.id);

                                return (
                                    <tr key={inv.id} className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors ${isSelected ? 'bg-indigo-50/30 dark:bg-indigo-900/20' : ''}`}>
                                        <td className="px-4 py-3">
                                            <button
                                                type="button"
                                                onClick={() => toggleSelect(inv.id)}
                                                className={`transition-colors ${isSelected ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                                            >
                                                {isSelected ? <CheckSquare size={16}/> : <Square size={16}/>}
                                            </button>
                                        </td>
                                        <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                                            {fmtDate(inv.issued_at, 'short')}
                                        </td>
                                        <td className="px-4 py-3 text-gray-800 dark:text-gray-200 font-medium whitespace-nowrap">
                                            {inv.vendor?.name || t('invoices.unknownVendor')}
                                        </td>
                                        <td className="px-4 py-3">
                                            <select
                                                value={currentCat?.id ?? ''}
                                                onChange={(e) => updateMutation.mutate({ id: inv.id, data: { category_id: e.target.value || null }})}
                                                className="text-xs rounded-lg border border-gray-200 dark:border-gray-700 px-2 py-1 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500 max-w-[11rem] truncate transition-colors hover:border-indigo-300 dark:hover:border-indigo-500"
                                                style={currentCat ? { color: currentCat.color, borderColor: currentCat.color + '55' } : undefined}
                                            >
                                                <option value="">{t('transactions.table.unset')}</option>
                                                {categories.filter(c => !c.deleted_at || c.id === currentCat?.id).map((c) => (
                                                    <option key={c.id} value={c.id}>{c.name}</option>
                                                ))}
                                            </select>
                                        </td>
                                        <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-[12rem] sm:max-w-xs truncate" title={inv.description}>
                                            {inv.description || t('invoices.emptyDescription')}
                                        </td>
                                        <td className="px-4 py-3 text-right font-mono font-medium text-gray-900 dark:text-gray-100 whitespace-nowrap">
                                            {fmtCurrency(inv.amount, inv.currency)}
                                        </td>
                                        <td className="px-4 py-3 text-right">
                                            <div className="flex justify-end space-x-1">
                                                <button
                                                    onClick={() => handlePreview(inv.id)}
                                                    disabled={isPreviewLoading === inv.id}
                                                    title={t('invoices.preview')}
                                                    className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-fuchsia-600 dark:hover:text-fuchsia-400 hover:bg-fuchsia-50 dark:hover:bg-fuchsia-900/20 rounded-lg transition-colors disabled:opacity-50"
                                                >
                                                    {isPreviewLoading === inv.id ? <Loader2 size={16} className="animate-spin" /> : <FileText size={16} />}
                                                </button>
                                                <button
                                                    onClick={() => downloadInvoiceFile(inv.id, inv.vendor?.name)}
                                                    title={t('invoices.download')}
                                                    className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                                                >
                                                    <Download size={16} />
                                                </button>
                                                <button
                                                    onClick={() => openEditModal(inv)}
                                                    title={t('invoices.edit')}
                                                    className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                                                >
                                                    <Edit2 size={16} />
                                                </button>
                                                <button
                                                    onClick={() => deleteMutation.mutate(inv.id)}
                                                    disabled={deleteMutation.isPending}
                                                    title={t('common.delete')}
                                                    className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors disabled:opacity-50"
                                                >
                                                    <Trash2 size={16} />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                );
                            })}
                            </tbody>
                        </table>
                    </div>
                    <div className="px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 dark:text-gray-500 text-right">
                        {t('transactions.table.showing', { count: filtered.length })}
                    </div>
                </div>
            )}

            {/* Batch Action Floating Bar */}
            {selectedIds.size > 0 && (
                <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4 duration-300">
                    <div className="bg-gray-900 dark:bg-gray-800 text-white rounded-2xl shadow-2xl px-6 py-4 flex items-center gap-6 border border-white/10 dark:border-gray-700">
                        <div className="flex items-center gap-2 pr-4 border-r border-gray-700 dark:border-gray-600">
                            <Layers size={18} className="text-indigo-400" />
                            <span className="font-bold text-sm">{t('invoices.selected', { count: selectedIds.size })}</span>
                        </div>

                        <div className="flex items-center gap-3">
                            <span className="text-xs text-gray-400 font-medium uppercase tracking-wider">{t('invoices.setCategory')}</span>
                            <select
                                className="bg-gray-800 dark:bg-gray-900 text-sm rounded-lg border border-gray-700 dark:border-gray-600 px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-indigo-500 min-w-[140px]"
                                defaultValue="placeholder"
                                onChange={(e) => {
                                    const val = e.target.value;
                                    if (val === 'placeholder') return;

                                    batchCatMutation.mutate({
                                        ids: Array.from(selectedIds),
                                        categoryId: val === 'unset' ? '' : val
                                    });
                                    e.target.value = 'placeholder';
                                }}
                                disabled={batchCatMutation.isPending}
                            >
                                <option value="placeholder" disabled hidden>{t('invoices.choose')}</option>
                                <option value="unset">{t('invoices.unset')}</option>
                                {categories.filter(c => !c.deleted_at).map((c: Category) => (
                                    <option key={c.id} value={c.id}>{c.name}</option>
                                ))}
                            </select>
                        </div>

                        <button onClick={() => setSelectedIds(new Set())} className="text-gray-400 hover:text-white transition-colors">
                            <X size={20} />
                        </button>
                    </div>
                </div>
            )}

            {/* Edit Modal */}
            {editingInvoice && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
                    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg border border-gray-100 dark:border-gray-800">
                        <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                                <Edit2 size={18} className="text-indigo-500" />
                                {t('invoices.editTitle')}
                            </h3>
                            <button onClick={() => setEditingInvoice(null)} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                                <X size={20} />
                            </button>
                        </div>
                        <div className="p-4 space-y-4">
                            {/* Vendor */}
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.vendor')}</label>
                                <input
                                    type="text"
                                    value={editVendor}
                                    onChange={(e) => setEditVendor(e.target.value)}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none"
                                />
                            </div>
                            {/* Date */}
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.date')}</label>
                                <input
                                    type="date"
                                    value={editIssuedAt}
                                    onChange={(e) => setEditIssuedAt(e.target.value)}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none [color-scheme:light] dark:[color-scheme:dark]"
                                />
                            </div>
                            {/* Amount + Currency */}
                            <div className="grid grid-cols-3 gap-3">
                                <div className="col-span-2">
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.amount')}</label>
                                    <input
                                        type="number"
                                        step="0.01"
                                        value={editAmount}
                                        onChange={(e) => setEditAmount(e.target.value)}
                                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none font-mono"
                                    />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.currency')}</label>
                                    <input
                                        type="text"
                                        maxLength={3}
                                        value={editCurrency}
                                        onChange={(e) => setEditCurrency(e.target.value.toUpperCase())}
                                        className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none font-mono uppercase"
                                    />
                                </div>
                            </div>
                            {/* Category */}
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.category')}</label>
                                <select
                                    value={editCatId}
                                    onChange={(e) => setEditCatId(e.target.value)}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none"
                                >
                                    <option value="">{t('common.none')}</option>
                                    {categories.filter(c => !c.deleted_at || c.id === editCatId).map((c: Category) => (
                                        <option key={c.id} value={c.id}>{c.name}</option>
                                    ))}
                                </select>
                            </div>
                            {/* Description */}
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('invoices.description')}</label>
                                <textarea
                                    value={editDesc}
                                    onChange={(e) => setEditDesc(e.target.value)}
                                    rows={3}
                                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-900 dark:text-gray-100 text-sm rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent block p-2.5 outline-none resize-none"
                                />
                            </div>
                        </div>
                        <div className="flex items-center justify-end space-x-3 p-4 border-t border-gray-100 dark:border-gray-800">
                            <button
                                onClick={() => setEditingInvoice(null)}
                                className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-xl transition-colors"
                            >
                                {t('common.cancel')}
                            </button>
                            <button
                                onClick={handleSaveEdit}
                                disabled={updateMutation.isPending}
                                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-xl transition-colors disabled:opacity-50 flex items-center gap-2"
                            >
                                {updateMutation.isPending && <Loader2 size={14} className="animate-spin" />}
                                {t('common.save')}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* PDF Preview Modal */}
            {previewInfo && (
                <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-in fade-in duration-200">
                    <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-5xl h-[90vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
                        <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                                <FileText className="text-fuchsia-500" size={20} />
                                {t('invoices.previewTitle')}
                            </h3>
                            <button onClick={closePreview} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                                <X size={20} />
                            </button>
                        </div>
                        <div className="flex-1 w-full bg-gray-200 dark:bg-gray-950 p-2 sm:p-4">
                            <FilePreview url={previewInfo.url} mimeType={previewInfo.mimeType} title={t('invoices.previewTitle')} />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}