import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { FileText, List, Eye, Download, AlertTriangle, Trash2, Loader2, ChevronUp, ChevronDown, Columns, Check } from 'lucide-react';
import { fmtCurrency } from '../utils/formatters';
import type { BankStatementSummary } from "../api/types/bank";

type SortField = 'statement_info' | 'statement_type' | 'period' | 'transaction_count' | 'new_balance';
type SortDirection = 'asc' | 'desc';
type ColKey = 'statementInfo' | 'type' | 'period' | 'transactions' | 'newBalance';

interface BankStatementTableProps {
    statements: BankStatementSummary[];
    visibleCols: Record<ColKey, boolean>;
    sortField: SortField;
    sortDirection: SortDirection;
    onSort: (field: SortField) => void;
    onNavigateTransactions: (id: string) => void;
    onPreview: (id: string) => void;
    onView: (stmt: BankStatementSummary) => void;
    onDownload: (id: string) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
    deletingId: string | null;
    showColMenu: boolean;
    onToggleColMenu: () => void;
    onToggleColumn: (key: ColKey) => void;
}

const BankStatementRow = React.memo(({
    stmt,
    visibleCols,
    onNavigateTransactions,
    onPreview,
    onView,
    onDownload,
    onDelete,
    isPreviewLoading,
    deletingId,
    t
}: {
    stmt: BankStatementSummary;
    visibleCols: Record<ColKey, boolean>;
    onNavigateTransactions: (id: string) => void;
    onPreview: (id: string) => void;
    onView: (stmt: BankStatementSummary) => void;
    onDownload: (id: string) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
    deletingId: string | null;
    t: any;
}) => {
    return (
        <tr className="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors group border-b border-gray-100 dark:border-gray-800/50">
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
                    <button
                        onClick={() => onNavigateTransactions(stmt.id)}
                        className="p-2 text-gray-400 hover:text-cyan-600 hover:bg-cyan-50 dark:hover:bg-cyan-900/30 rounded-lg transition-colors"
                        title={t('bankStatements.actions.viewTransactions')}
                    >
                        <List size={18} />
                    </button>
                    <button
                        onClick={() => onPreview(stmt.id)}
                        disabled={isPreviewLoading === stmt.id || !stmt.has_original_file}
                        className="p-2 text-gray-400 hover:text-fuchsia-600 hover:bg-fuchsia-50 dark:hover:bg-fuchsia-900/30 rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                        title={stmt.has_original_file ? t('bankStatements.actions.preview') : t('bankStatements.actions.previewUnavailable')}
                    >
                        {isPreviewLoading === stmt.id ? <Loader2 size={18} className="animate-spin" /> : <FileText size={18} />}
                    </button>
                    <button
                        onClick={() => onView(stmt)}
                        className="p-2 text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 rounded-lg transition-colors"
                        title={t('bankStatements.actions.view')}
                    >
                        <Eye size={18} />
                    </button>
                    <button
                        onClick={() => onDownload(stmt.id)}
                        disabled={!stmt.has_original_file}
                        className="p-2 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
                        title={stmt.has_original_file ? t('bankStatements.actions.download') : t('bankStatements.actions.downloadUnavailable')}
                    >
                        <Download size={18} />
                    </button>
                    <button
                        onClick={() => onDelete(stmt.id)}
                        disabled={deletingId === stmt.id}
                        className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors disabled:opacity-50"
                        title={t('bankStatements.actions.delete')}
                    >
                        {deletingId === stmt.id ? <AlertTriangle size={18} className="animate-pulse" /> : <Trash2 size={18} />}
                    </button>
                </div>
            </td>
        </tr>
    );
});

export default function BankStatementTable({
    statements,
    visibleCols,
    sortField,
    sortDirection,
    onSort,
    onNavigateTransactions,
    onPreview,
    onView,
    onDownload,
    onDelete,
    isPreviewLoading,
    deletingId,
    showColMenu,
    onToggleColMenu,
    onToggleColumn
}: BankStatementTableProps) {
    const { t } = useTranslation();
    const parentRef = React.useRef<HTMLDivElement>(null);

    const rowVirtualizer = useVirtualizer({
        count: statements.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 72,
        overscan: 10,
    });

    const virtualRows = rowVirtualizer.getVirtualItems();
    const totalSize = rowVirtualizer.getTotalSize();

    const paddingTop = virtualRows.length > 0 ? virtualRows[0].start : 0;
    const paddingBottom = virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1].end : 0;

    const SortableHeader = ({ field, label, align = 'left' }: { field: SortField, label: string, align?: 'left' | 'right' | 'center' }) => (
        <th className={`px-6 py-4 font-medium cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 group transition-colors select-none bg-inherit border-b border-gray-100 dark:border-gray-800 ${align === 'right' ? 'text-right' : align === 'center' ? 'text-center' : 'text-left'}`} onClick={() => onSort(field)}>
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

    if (statements.length === 0) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-16 flex flex-col items-center justify-center text-gray-500 shadow-sm">
                <FileText size={40} className="text-gray-300 dark:text-gray-700 mb-3" />
                {t('bankStatements.table.noData')}
            </div>
        );
    }

    const colCount = Object.values(visibleCols).filter(Boolean).length + 1;

    return (
        <div className="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm flex flex-col">
            <div 
                ref={parentRef}
                className="overflow-x-auto overflow-y-auto max-h-[calc(100vh-280px)] sm:max-h-[calc(100vh-320px)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700"
            >
                <table className="w-full text-left text-sm whitespace-nowrap border-separate border-spacing-0">
                    <thead className="bg-gray-50 dark:bg-gray-800/90 backdrop-blur-sm text-gray-600 dark:text-gray-400 sticky top-0 z-10 shadow-sm">
                    <tr>
                        {visibleCols.statementInfo && <SortableHeader field="statement_info" label={t('bankStatements.columns.statementInfo')} />}
                        {visibleCols.type && <SortableHeader field="statement_type" label={t('bankStatements.columns.type')} />}
                        {visibleCols.period && <SortableHeader field="period" label={t('bankStatements.columns.period')} />}
                        {visibleCols.transactions && <SortableHeader field="transaction_count" label={t('bankStatements.columns.transactions')} align="right" />}
                        {visibleCols.newBalance && <SortableHeader field="new_balance" label={t('bankStatements.columns.newBalance')} align="right" />}
                        <th className="px-6 py-4 font-medium text-right bg-inherit border-b border-gray-100 dark:border-gray-800">
                            <div className="flex items-center justify-end gap-3">
                                <span>{t('bankStatements.columns.actions')}</span>
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
                                                    { key: 'statementInfo', label: t('bankStatements.columns.statementInfo') },
                                                    { key: 'type', label: t('bankStatements.columns.type') },
                                                    { key: 'period', label: t('bankStatements.columns.period') },
                                                    { key: 'transactions', label: t('bankStatements.columns.transactions') },
                                                    { key: 'newBalance', label: t('bankStatements.columns.newBalance') },
                                                ].map(({ key, label }) => (
                                                    <button
                                                        key={key}
                                                        onClick={() => onToggleColumn(key as ColKey)}
                                                        className="w-full flex items-center justify-between px-3 py-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors text-left"
                                                    >
                                                        {label}
                                                        {visibleCols[key as ColKey] && <Check size={16} className="text-indigo-600 dark:text-indigo-400" />}
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
                    <tbody className="divide-y divide-gray-100 dark:divide-gray-800 text-gray-700 dark:text-gray-300">
                    {paddingTop > 0 && (
                        <tr>
                            <td style={{ height: `${paddingTop}px` }} colSpan={colCount} />
                        </tr>
                    )}
                    {virtualRows.map((virtualRow) => {
                        const stmt = statements[virtualRow.index];
                        return (
                            <BankStatementRow
                                key={stmt.id}
                                stmt={stmt}
                                visibleCols={visibleCols}
                                onNavigateTransactions={onNavigateTransactions}
                                onPreview={onPreview}
                                onView={onView}
                                onDownload={onDownload}
                                onDelete={onDelete}
                                isPreviewLoading={isPreviewLoading}
                                deletingId={deletingId}
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
