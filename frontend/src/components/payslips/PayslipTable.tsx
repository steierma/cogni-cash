import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Loader2, Receipt, Eye, Pencil, Download, Trash2, LineChart as LineChartIcon, ArrowUpDown, ArrowUp, ArrowDown, FileText, Columns, Check } from 'lucide-react';
import { fmtCurrency } from '../../utils/formatters';
import { formatYearMonth, getAdjustedNetto, type ColKey, type SortDirection } from './utils';
import type { Payslip } from "../../api/types/payslip";

interface PayslipTableProps {
    isLoading: boolean;
    filteredPayslips: Payslip[];
    visibleCols: Record<ColKey, boolean>;
    sortConfig: { key: ColKey; direction: SortDirection };
    onSort: (key: ColKey) => void;
    ignoredPayslipIds: Set<string>;
    onToggleIgnore: (id: string) => void;
    onPreview: (id: string) => void;
    isPreviewLoading: string | null;
    onView: (p: Payslip) => void;
    onEdit: (p: Payslip) => void;
    onDownload: (id: string, name: string) => void;
    onDelete: (id: string) => void;
    excludedBonuses: Set<string>;
    excludeLeasing: boolean;
    useProportionalMath: boolean;
    showColMenu: boolean;
    onToggleColMenu: () => void;
    onToggleColumn: (key: ColKey) => void;
}

const PayslipRow = React.memo(({
    p,
    visibleCols,
    isIgnored,
    onToggleIgnore,
    onPreview,
    isPreviewLoading,
    onView,
    onEdit,
    onDownload,
    onDelete,
    excludedBonuses,
    excludeLeasing,
    useProportionalMath,
    t
}: {
    p: Payslip;
    visibleCols: Record<ColKey, boolean>;
    isIgnored: boolean;
    onToggleIgnore: (id: string) => void;
    onPreview: (id: string) => void;
    isPreviewLoading: string | null;
    onView: (p: Payslip) => void;
    onEdit: (p: Payslip) => void;
    onDownload: (id: string, name: string) => void;
    onDelete: (id: string) => void;
    excludedBonuses: Set<string>;
    excludeLeasing: boolean;
    useProportionalMath: boolean;
    t: any;
}) => {
    return (
        <tr className={`transition-colors group border-b border-gray-100 dark:border-gray-800/50 ${isIgnored ? 'bg-gray-50/50 dark:bg-gray-800/20 opacity-60 hover:opacity-100' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}>
            {visibleCols.period && <td className="px-5 py-3 font-medium text-gray-900 dark:text-gray-100 whitespace-nowrap">{formatYearMonth(p.period_year, p.period_month_num)}</td>}
            {visibleCols.employer && <td className="px-5 py-3 text-gray-600 dark:text-gray-400 whitespace-nowrap">{p.employer_name}</td>}
            {visibleCols.gross && <td className="px-5 py-3 text-gray-500 dark:text-gray-400 text-right whitespace-nowrap font-mono">{fmtCurrency(p.gross_pay, p.currency)}</td>}
            {visibleCols.net && <td className="px-5 py-3 font-medium text-gray-900 dark:text-gray-100 text-right whitespace-nowrap font-mono">{fmtCurrency(p.net_pay, p.currency)}</td>}
            {visibleCols.adjNet && <td className="px-5 py-3 font-medium text-blue-700 dark:text-blue-400 text-right whitespace-nowrap font-mono">{fmtCurrency(getAdjustedNetto(p, excludedBonuses, excludeLeasing, useProportionalMath), 'EUR')}</td>}
            {visibleCols.payout && <td className="px-5 py-3 text-gray-500 dark:text-gray-400 text-right whitespace-nowrap font-mono">{fmtCurrency(p.payout_amount, p.currency)}</td>}
            {visibleCols.leasing && <td className="px-5 py-3 text-gray-500 dark:text-gray-400 text-right whitespace-nowrap font-mono">{fmtCurrency(p.custom_deductions, p.currency)}</td>}
            <td className="px-5 py-3 whitespace-nowrap text-right">
                <div className="flex items-center justify-end gap-2 opacity-100 lg:opacity-40 lg:group-hover:opacity-100 transition-opacity">
                    <button onClick={() => onToggleIgnore(p.id)} className={`p-1.5 rounded-lg transition-colors ${isIgnored ? 'text-gray-400 hover:text-indigo-500 hover:bg-indigo-50 dark:hover:bg-indigo-900/30' : 'text-indigo-600 bg-indigo-50 dark:bg-indigo-900/30 hover:bg-indigo-100 dark:hover:bg-indigo-900/50'}`} title={isIgnored ? t('payslips.table.includeInChart') : t('payslips.table.excludeFromChart')}>
                        <LineChartIcon size={18} className={isIgnored ? 'opacity-50' : ''} />
                    </button>
                    <button onClick={() => onPreview(p.id)} disabled={isPreviewLoading === p.id || !p.original_file_name} className="p-1.5 text-gray-400 hover:text-fuchsia-600 hover:bg-fuchsia-50 dark:hover:bg-fuchsia-900/30 rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed" title={p.original_file_name ? t('payslips.table.previewPdf') : t('payslips.table.noFile')}>
                        {isPreviewLoading === p.id ? <Loader2 size={18} className="animate-spin" /> : <FileText size={18} />}
                    </button>
                    <button onClick={() => onView(p)} className="p-1.5 text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg"><Eye size={18} /></button>
                    <button onClick={() => onEdit(p)} className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg"><Pencil size={18} /></button>
                    <button onClick={() => onDownload(p.id, p.original_file_name)} disabled={!p.original_file_name} className="p-1.5 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg disabled:opacity-30 disabled:cursor-not-allowed" title={p.original_file_name ? t('payslips.table.downloadPdf') : t('payslips.table.noFile')}><Download size={18} /></button>
                    <button onClick={() => { if (confirm(t('payslips.table.deleteConfirm'))) onDelete(p.id); }} className="p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg"><Trash2 size={18} /></button>
                </div>
            </td>
        </tr>
    );
});

export function PayslipTable({
                                 isLoading, filteredPayslips, visibleCols, sortConfig, onSort, ignoredPayslipIds,
                                 onToggleIgnore, onPreview, isPreviewLoading, onView, onEdit, onDownload, onDelete,
                                 excludedBonuses, excludeLeasing, useProportionalMath,
                                 showColMenu, onToggleColMenu, onToggleColumn
                             }: PayslipTableProps) {

    const { t } = useTranslation();
    const parentRef = React.useRef<HTMLDivElement>(null);

    const rowVirtualizer = useVirtualizer({
        count: filteredPayslips.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 52,
        overscan: 10,
    });

    const virtualRows = rowVirtualizer.getVirtualItems();
    const totalSize = rowVirtualizer.getTotalSize();

    const paddingTop = virtualRows.length > 0 ? virtualRows[0].start : 0;
    const paddingBottom = virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1].end : 0;

    const renderSortIcon = (colKey: ColKey) => {
        if (sortConfig.key !== colKey) return <ArrowUpDown size={14} className="text-gray-400 opacity-50 group-hover/th:opacity-100 transition-opacity" />;
        return sortConfig.direction === 'asc'
            ? <ArrowUp size={14} className="text-indigo-600 dark:text-indigo-400" />
            : <ArrowDown size={14} className="text-indigo-600 dark:text-indigo-400" />;
    };

    if (isLoading) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-12 flex items-center justify-center">
                <Loader2 size={32} className="animate-spin text-indigo-500" />
            </div>
        );
    }

    if (filteredPayslips.length === 0) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-16 flex flex-col items-center justify-center text-gray-400 dark:text-gray-600">
                <Receipt size={48} className="mb-4 opacity-20" />
                <p className="text-lg font-medium text-gray-700 dark:text-gray-300">{t('payslips.table.noPayslips')}</p>
            </div>
        );
    }

    const colCount = Object.values(visibleCols).filter(Boolean).length + 1;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col w-full overflow-hidden">
            <div 
                ref={parentRef}
                className="w-full overflow-x-auto overflow-y-auto max-h-[calc(100vh-280px)] sm:max-h-[calc(100vh-320px)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700"
            >
                <table className="min-w-[1000px] w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm border-separate border-spacing-0">
                    <thead className="bg-gray-50 dark:bg-gray-800/90 backdrop-blur-sm text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide select-none sticky top-0 z-10 shadow-sm">
                    <tr>
                        {visibleCols.period && <th onClick={() => onSort('period')} className="px-5 py-3.5 text-left font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center gap-1.5">{t('payslips.table.period')} {renderSortIcon('period')}</div></th>}
                        {visibleCols.employer && <th onClick={() => onSort('employer')} className="px-5 py-3.5 text-left font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center gap-1.5">{t('payslips.modals.employer')} {renderSortIcon('employer')}</div></th>}
                        {visibleCols.gross && <th onClick={() => onSort('gross')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('gross')} {t('payslips.table.gross')}</div></th>}
                        {visibleCols.net && <th onClick={() => onSort('net')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('net')} {t('payslips.table.net')}</div></th>}
                        {visibleCols.adjNet && <th onClick={() => onSort('adjNet')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('adjNet')} {t('payslips.table.adjNet')}</div></th>}
                        {visibleCols.payout && <th onClick={() => onSort('payout')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('payout')} {t('payslips.table.payout')}</div></th>}
                        {visibleCols.leasing && <th onClick={() => onSort('leasing')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th bg-inherit border-b border-gray-100 dark:border-gray-800"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('leasing')} {t('payslips.table.leasing')}</div></th>}
                        <th className="px-5 py-3.5 text-right font-medium bg-inherit border-b border-gray-100 dark:border-gray-800">
                            <div className="flex items-center justify-end gap-3">
                                <span>{t('payslips.table.actions')}</span>
                                <div className="relative normal-case tracking-normal font-medium">
                                    <button
                                        onClick={onToggleColMenu}
                                        className={`p-1 rounded-lg border transition-all ${showColMenu
                                            ? 'bg-indigo-50 dark:bg-indigo-900/30 border-indigo-200 dark:border-indigo-800 text-indigo-600 dark:text-indigo-400'
                                            : 'bg-white dark:bg-gray-900 border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}
                                        title={t('transactions.columns', 'Columns')}
                                    >
                                        <Columns size={12} />
                                    </button>

                                    {showColMenu && (
                                        <>
                                            <div className="fixed inset-0 z-10" onClick={onToggleColMenu} />
                                            <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-800 rounded-xl shadow-xl border border-gray-200 dark:border-gray-700 p-2 z-20 animate-in fade-in zoom-in duration-150 origin-top-right">
                                                {[
                                                    { key: 'period', label: t('payslips.modals.period') },
                                                    { key: 'employer', label: t('payslips.modals.employer') },
                                                    { key: 'gross', label: t('payslips.modals.gross') },
                                                    { key: 'net', label: t('payslips.modals.net') },
                                                    { key: 'adjNet', label: t('payslips.adjustedNet') },
                                                    { key: 'payout', label: t('payslips.modals.payout') },
                                                    { key: 'leasing', label: t('payslips.modals.leasing') }
                                                ].map(({ key, label }) => (
                                                    <button
                                                        key={key}
                                                        onClick={() => onToggleColumn(key as ColKey)}
                                                        className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors text-left"
                                                    >
                                                        {label} {visibleCols[key as ColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
                                                    </button>
                                                ))}
                                            </div>
                                        </>
                                    )}
                                </div>
                            </div>
                        </th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                    {paddingTop > 0 && (
                        <tr>
                            <td style={{ height: `${paddingTop}px` }} colSpan={colCount} />
                        </tr>
                    )}
                    {virtualRows.map((virtualRow) => {
                        const p = filteredPayslips[virtualRow.index];
                        return (
                            <PayslipRow
                                key={p.id}
                                p={p}
                                visibleCols={visibleCols}
                                isIgnored={ignoredPayslipIds.has(p.id)}
                                onToggleIgnore={onToggleIgnore}
                                onPreview={onPreview}
                                isPreviewLoading={isPreviewLoading}
                                onView={onView}
                                onEdit={onEdit}
                                onDownload={onDownload}
                                onDelete={onDelete}
                                excludedBonuses={excludedBonuses}
                                excludeLeasing={excludeLeasing}
                                useProportionalMath={useProportionalMath}
                                t={t}
                            />
                        );
                    })}
                    {paddingBottom > 0 && (
                        <tr>
                            <td style={{ height: `${paddingBottom}px` }} colSpan={colCount} />
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}