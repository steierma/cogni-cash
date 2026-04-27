import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { 
    CheckSquare, Square, Users, Layers, Eye, Download, Share2, Edit2, Trash2, Loader2,
    ChevronUp, ChevronDown 
} from 'lucide-react';
import { fmtCurrency, fmtDate } from '../../utils/formatters';
import type { Invoice } from "../../api/types/invoice";
import type { Category } from "../../api/types/category";

type SortKey = 'issued_at' | 'vendor' | 'category' | 'description' | 'amount';
type SortDir = 'asc' | 'desc';

interface InvoiceTableProps {
    invoices: Invoice[];
    categories: Category[];
    currentUserId?: string;
    selectedIds: Set<string>;
    onToggleSelect: (id: string) => void;
    onToggleSelectAll: () => void;
    sortKey: SortKey;
    sortDir: SortDir;
    onSort: (key: SortKey) => void;
    onCategoryChange: (id: string, categoryId: string | null) => void;
    onPreview: (inv: Invoice) => void;
    onDownload: (id: string, vendorName?: string) => void;
    onShare: (inv: Invoice) => void;
    onEdit: (inv: Invoice) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
    isDeleting: boolean;
}

const InvoiceRow = React.memo(({
    inv,
    categories,
    currentUserId,
    isSelected,
    onToggleSelect,
    onCategoryChange,
    onPreview,
    onDownload,
    onShare,
    onEdit,
    onDelete,
    isPreviewLoading,
    isDeleting,
    t
}: {
    inv: Invoice;
    categories: Category[];
    currentUserId?: string;
    isSelected: boolean;
    onToggleSelect: (id: string) => void;
    onCategoryChange: (id: string, categoryId: string | null) => void;
    onPreview: (inv: Invoice) => void;
    onDownload: (id: string, vendorName?: string) => void;
    onShare: (inv: Invoice) => void;
    onEdit: (inv: Invoice) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
    isDeleting: boolean;
    t: any;
}) => {
    const currentCat = categories.find((c) => c.id === inv.category_id);
    const hasSplits = inv.splits && inv.splits.length > 0;

    return (
        <tr className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors border-b border-gray-100 dark:border-gray-800/50 ${isSelected ? 'bg-indigo-50/30 dark:bg-indigo-900/20' : ''}`}>
            <td className="px-4 py-3">
                <button
                    type="button"
                    onClick={() => onToggleSelect(inv.id)}
                    className={`transition-colors ${isSelected ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                >
                    {isSelected ? <CheckSquare size={16}/> : <Square size={16}/>}
                </button>
            </td>
            <td className="px-4 py-3 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                {fmtDate(inv.issued_at, 'short')}
            </td>
            <td className="px-4 py-3 text-gray-800 dark:text-gray-200 font-medium whitespace-nowrap">
                <div className="flex items-center gap-2">
                    {inv.user_id !== currentUserId && (
                        <span title={t('transactions.table.shared')} className="inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[10px] font-medium bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400 border border-green-200 dark:border-green-800/50 shrink-0 uppercase tracking-tighter">
                            <Users size={9} /> {t('transactions.table.shared')}
                        </span>
                    )}
                    {inv.vendor?.name || t('invoices.unknownVendor')}
                </div>
            </td>
            <td className="px-4 py-3">
                {hasSplits ? (
                    <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-lg bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 text-xs font-bold border border-indigo-100 dark:border-indigo-800/50 uppercase tracking-wider">
                        <Layers size={10} /> {t('invoices.split', 'Split')}
                    </span>
                ) : (
                    <select
                        value={currentCat?.id ?? ''}
                        onChange={(e) => onCategoryChange(inv.id, e.target.value || null)}
                        className="text-xs rounded-lg border border-gray-200 dark:border-gray-700 px-2 py-1 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-300 dark:focus:ring-indigo-500 max-w-[11rem] truncate transition-colors hover:border-indigo-300 dark:hover:border-indigo-500"
                        style={currentCat ? { color: currentCat.color, borderColor: currentCat.color + '55' } : undefined}
                    >
                        <option value="">{t('transactions.table.unset')}</option>
                        {categories.filter(c => !c.deleted_at || c.id === currentCat?.id).map((c) => (
                            <option key={c.id} value={c.id}>{c.name}</option>
                        ))}
                    </select>
                )}
            </td>
            <td className="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-[12rem] sm:max-w-xs truncate" title={inv.description}>
                {inv.description || t('invoices.emptyDescription')}
            </td>
            <td className="px-4 py-3 text-right font-mono font-medium text-gray-900 dark:text-gray-100 whitespace-nowrap">
                <div className="flex flex-col items-end">
                    <span>{fmtCurrency(inv.amount, inv.currency)}</span>
                    {inv.base_currency && inv.base_currency !== inv.currency && inv.base_amount !== 0 && (
                        <span className="text-[10px] text-gray-400 dark:text-gray-500 font-normal">
                            {fmtCurrency(inv.base_amount, inv.base_currency)}
                        </span>
                    )}
                </div>
            </td>
            <td className="px-4 py-3 text-right">
                <div className="flex justify-end space-x-1">
                    <button
                        onClick={() => onPreview(inv)}
                        disabled={isPreviewLoading === inv.id}
                        title={t('invoices.preview')}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-fuchsia-600 dark:hover:text-fuchsia-400 hover:bg-fuchsia-50 dark:hover:bg-fuchsia-900/20 rounded-lg transition-colors disabled:opacity-50"
                    >
                        {isPreviewLoading === inv.id ? <Loader2 size={16} className="animate-spin" /> : <Eye size={16} />}
                    </button>
                    <button
                        onClick={() => onDownload(inv.id, inv.vendor?.name)}
                        title={t('invoices.download')}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                    >
                        <Download size={16} />
                    </button>
                    <button
                        onClick={() => onShare(inv)}
                        title={t('common.share')}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                    >
                        <Share2 size={16} />
                    </button>
                    <button
                        onClick={() => onEdit(inv)}
                        title={t('invoices.edit')}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                    >
                        <Edit2 size={16} />
                    </button>
                    <button
                        onClick={() => onDelete(inv.id)}
                        disabled={isDeleting}
                        title={t('common.delete')}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors disabled:opacity-50"
                    >
                        {isDeleting ? <Loader2 size={16} className="animate-spin" /> : <Trash2 size={16} />}
                    </button>
                </div>
            </td>
        </tr>
    );
});

export default function InvoiceTable({
    invoices,
    categories,
    currentUserId,
    selectedIds,
    onToggleSelect,
    onToggleSelectAll,
    sortKey,
    sortDir,
    onSort,
    onCategoryChange,
    onPreview,
    onDownload,
    onShare,
    onEdit,
    onDelete,
    isPreviewLoading,
    isDeleting
}: InvoiceTableProps) {
    const { t } = useTranslation();
    const parentRef = React.useRef<HTMLDivElement>(null);

    const rowVirtualizer = useVirtualizer({
        count: invoices.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 64,
        overscan: 10,
    });

    const virtualRows = rowVirtualizer.getVirtualItems();
    const totalSize = rowVirtualizer.getTotalSize();

    const paddingTop = virtualRows.length > 0 ? virtualRows[0].start : 0;
    const paddingBottom = virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1].end : 0;

    const renderSortIcon = (k: SortKey) =>
        sortKey === k ? (sortDir === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />) : null;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col">
            <div 
                ref={parentRef}
                className="overflow-x-auto overflow-y-auto max-h-[calc(100vh-280px)] sm:max-h-[calc(100vh-320px)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700"
            >
                <table className="min-w-full text-sm border-separate border-spacing-0">
                    <thead className="bg-gray-50 dark:bg-gray-800/90 backdrop-blur-sm text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide sticky top-0 z-10 shadow-sm">
                    <tr>
                        <th className="px-4 py-3 text-left w-10 bg-inherit border-b border-gray-100 dark:border-gray-800">
                            <button
                                type="button"
                                onClick={onToggleSelectAll}
                                className={`transition-colors ${selectedIds.size === invoices.length && invoices.length > 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-300 dark:text-gray-600 hover:text-gray-400 dark:hover:text-gray-400'}`}
                            >
                                {selectedIds.size === invoices.length && invoices.length > 0 ? <CheckSquare size={16}/> : <Square size={16}/>}
                            </button>
                        </th>
                        <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap bg-inherit border-b border-gray-100 dark:border-gray-800" onClick={() => onSort('issued_at')}>
                            <span className="inline-flex items-center gap-1">{t('invoices.date')} {renderSortIcon('issued_at')}</span>
                        </th>
                        <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap bg-inherit border-b border-gray-100 dark:border-gray-800" onClick={() => onSort('vendor')}>
                            <span className="inline-flex items-center gap-1">{t('invoices.vendor')} {renderSortIcon('vendor')}</span>
                        </th>
                        <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap bg-inherit border-b border-gray-100 dark:border-gray-800" onClick={() => onSort('category')}>
                            <span className="inline-flex items-center gap-1">{t('invoices.category')} {renderSortIcon('category')}</span>
                        </th>
                        <th className="px-4 py-3 text-left cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none bg-inherit border-b border-gray-100 dark:border-gray-800" onClick={() => onSort('description')}>
                            <span className="inline-flex items-center gap-1">{t('invoices.description')} {renderSortIcon('description')}</span>
                        </th>
                        <th className="px-4 py-3 text-right cursor-pointer hover:text-gray-700 dark:hover:text-gray-300 select-none whitespace-nowrap bg-inherit border-b border-gray-100 dark:border-gray-800" onClick={() => onSort('amount')}>
                            <span className="inline-flex items-center gap-1 justify-end">{t('invoices.amount')} {renderSortIcon('amount')}</span>
                        </th>
                        <th className="px-4 py-3 text-right bg-inherit border-b border-gray-100 dark:border-gray-800">{t('common.actions')}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                    {paddingTop > 0 && (
                        <tr>
                            <td style={{ height: `${paddingTop}px` }} colSpan={7} />
                        </tr>
                    )}
                    {virtualRows.map((virtualRow) => {
                        const inv = invoices[virtualRow.index];
                        return (
                            <InvoiceRow
                                key={inv.id}
                                inv={inv}
                                categories={categories}
                                currentUserId={currentUserId}
                                isSelected={selectedIds.has(inv.id)}
                                onToggleSelect={onToggleSelect}
                                onCategoryChange={onCategoryChange}
                                onPreview={onPreview}
                                onDownload={onDownload}
                                onShare={onShare}
                                onEdit={onEdit}
                                onDelete={onDelete}
                                isPreviewLoading={isPreviewLoading}
                                isDeleting={isDeleting}
                                t={t}
                            />
                        );
                    })}
                    {paddingBottom > 0 && (
                        <tr>
                            <td style={{ height: `${paddingBottom}px` }} colSpan={7} />
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>
            <div className="px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 dark:text-gray-500 text-right">
                {t('transactions.table.showing', { count: invoices.length })}
            </div>
        </div>
    );
}
