import { useTranslation } from 'react-i18next';
import { Loader2, Receipt, Eye, Pencil, Download, Trash2, LineChart as LineChartIcon, ArrowUpDown, ArrowUp, ArrowDown, FileText } from 'lucide-react';
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
}

export function PayslipTable({
                                 isLoading, filteredPayslips, visibleCols, sortConfig, onSort, ignoredPayslipIds,
                                 onToggleIgnore, onPreview, isPreviewLoading, onView, onEdit, onDownload, onDelete,
                                 excludedBonuses, excludeLeasing, useProportionalMath
                             }: PayslipTableProps) {

    const { t } = useTranslation();

    const renderSortIcon = (colKey: ColKey) => {
        if (sortConfig.key !== colKey) return <ArrowUpDown size={14} className="text-gray-400 opacity-50 group-hover/th:opacity-100 transition-opacity" />;
        return sortConfig.direction === 'asc'
            ? <ArrowUp size={14} className="text-indigo-600 dark:text-indigo-400" />
            : <ArrowDown size={14} className="text-indigo-600 dark:text-indigo-400" />;
    };

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col w-full overflow-hidden">
            <div className="w-full overflow-x-auto pb-2">
                <table className="min-w-[1000px] w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm">
                    <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide select-none">
                    <tr>
                        {visibleCols.period && <th onClick={() => onSort('period')} className="px-5 py-3.5 text-left font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center gap-1.5">{t('payslips.table.period')} {renderSortIcon('period')}</div></th>}
                        {visibleCols.employer && <th onClick={() => onSort('employer')} className="px-5 py-3.5 text-left font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center gap-1.5">{t('payslips.modals.employer')} {renderSortIcon('employer')}</div></th>}
                        {visibleCols.gross && <th onClick={() => onSort('gross')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('gross')} {t('payslips.table.gross')}</div></th>}
                        {visibleCols.net && <th onClick={() => onSort('net')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('net')} {t('payslips.table.net')}</div></th>}
                        {visibleCols.adjNet && <th onClick={() => onSort('adjNet')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('adjNet')} {t('payslips.table.adjNet')}</div></th>}
                        {visibleCols.payout && <th onClick={() => onSort('payout')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('payout')} {t('payslips.table.payout')}</div></th>}
                        {visibleCols.leasing && <th onClick={() => onSort('leasing')} className="px-5 py-3.5 text-right font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group/th"><div className="flex items-center justify-end gap-1.5">{renderSortIcon('leasing')} {t('payslips.table.leasing')}</div></th>}
                        <th className="px-5 py-3.5 text-right font-medium">{t('payslips.table.actions')}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                    {isLoading ? (
                        <tr><td colSpan={8} className="px-6 py-12 text-center text-gray-500"><Loader2 size={24} className="animate-spin mx-auto text-indigo-500" /></td></tr>
                    ) : filteredPayslips.length === 0 ? (
                        <tr>
                            <td colSpan={8}>
                                <div className="flex flex-col items-center justify-center py-16 text-gray-400 dark:text-gray-600">
                                    <Receipt size={48} className="mb-4 opacity-20" />
                                    <p className="text-lg font-medium text-gray-700 dark:text-gray-300">{t('payslips.table.noPayslips')}</p>
                                </div>
                            </td>
                        </tr>
                    ) : (
                        filteredPayslips.map((p) => {
                            const isIgnored = ignoredPayslipIds.has(p.id);
                            return (
                                <tr key={p.id} className={`transition-colors group ${isIgnored ? 'bg-gray-50/50 dark:bg-gray-800/20 opacity-60 hover:opacity-100' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'}`}>
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
                        })
                    )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}