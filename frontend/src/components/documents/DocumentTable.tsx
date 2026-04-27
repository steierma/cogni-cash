import * as React from 'react';
import { useTranslation } from 'react-i18next';
import { useVirtualizer } from '@tanstack/react-virtual';
import { 
    Archive, FileText, Briefcase, ShieldCheck, File, Pencil, Download, Trash2, Loader2,
    type LucideIcon 
} from 'lucide-react';
import { fmtDate } from '../../utils/formatters';
import type { Document } from "../../api/types/document";
import { documentService } from '../../api/services/documentService';

const TYPE_ICONS: Record<string, LucideIcon> = {
    tax_certificate: ShieldCheck,
    receipt: FileText,
    contract: Briefcase,
    other: File
};

interface DocumentTableProps {
    documents: Document[];
    isLoading: boolean;
    onPreview: (doc: Document) => void;
    onEdit: (doc: Document) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
}

const DocumentRow = React.memo(({
    doc,
    onPreview,
    onEdit,
    onDelete,
    isPreviewLoading,
    t
}: {
    doc: Document;
    onPreview: (doc: Document) => void;
    onEdit: (doc: Document) => void;
    onDelete: (id: string) => void;
    isPreviewLoading: string | null;
    t: any;
}) => {
    const Icon = TYPE_ICONS[doc.type] || File;
    const docDate = String(doc.metadata?.date || doc.created_at);

    return (
        <tr className="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors group border-b border-gray-100 dark:border-gray-800/50">
            <td className="px-6 py-4">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                        <Icon className="text-indigo-600 dark:text-indigo-400" size={20} />
                    </div>
                    <div>
                        <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate max-w-[200px] sm:max-w-xs" title={doc.file_name}>{doc.file_name}</p>
                        <p className="text-xs text-gray-500 dark:text-gray-400 line-clamp-1">{t(`documents.types.${doc.type}`)}</p>
                    </div>
                </div>
            </td>
            <td className="px-6 py-4">
                <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200">
                    {t(`documents.types.${doc.type}`)}
                </span>
            </td>
            <td className="px-6 py-4">
                <p className="text-sm text-gray-600 dark:text-gray-300 font-medium whitespace-nowrap">{fmtDate(docDate)}</p>
                <p className="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-tight whitespace-nowrap">{t('documents.uploadDate')}: {fmtDate(doc.created_at)}</p>
            </td>
            <td className="px-6 py-4 text-right">
                <div className="flex items-center justify-end gap-2">
                    <button
                        onClick={() => onPreview(doc)}
                        disabled={isPreviewLoading === doc.id}
                        className="p-2 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-all disabled:opacity-50"
                        title={t('documents.actions.preview')}
                    >
                        {isPreviewLoading === doc.id ? (
                            <Loader2 className="animate-spin" size={18} />
                        ) : (
                            <FileText size={18} />
                        )}
                    </button>
                    <button
                        onClick={() => onEdit(doc)}
                        className="p-2 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-all"
                        title={t('common.edit')}
                    >
                        <Pencil size={18} />
                    </button>
                    <a
                        href={documentService.getDownloadUrl(doc.id)}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-2 text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-all"
                        title={t('documents.actions.download')}
                    >
                        <Download size={18} />
                    </a>
                    <button
                        onClick={() => onDelete(doc.id)}
                        className="p-2 text-gray-400 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-all"
                        title={t('documents.actions.delete')}
                    >
                        <Trash2 size={18} />
                    </button>
                </div>
            </td>
        </tr>
    );
});

export default function DocumentTable({
    documents,
    isLoading,
    onPreview,
    onEdit,
    onDelete,
    isPreviewLoading
}: DocumentTableProps) {
    const { t } = useTranslation();
    const parentRef = React.useRef<HTMLDivElement>(null);

    const rowVirtualizer = useVirtualizer({
        count: documents.length,
        getScrollElement: () => parentRef.current,
        estimateSize: () => 72,
        overscan: 10,
    });

    const virtualRows = rowVirtualizer.getVirtualItems();
    const totalSize = rowVirtualizer.getTotalSize();

    const paddingTop = virtualRows.length > 0 ? virtualRows[0].start : 0;
    const paddingBottom = virtualRows.length > 0 ? totalSize - virtualRows[virtualRows.length - 1].end : 0;

    if (isLoading) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-8 flex flex-col gap-4">
                {[1, 2, 3, 4, 5].map(i => (
                    <div key={i} className="h-16 bg-gray-50 dark:bg-gray-800/50 rounded-xl animate-pulse" />
                ))}
            </div>
        );
    }

    if (documents.length === 0) {
        return (
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-20 flex flex-col items-center justify-center text-center">
                <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-full mb-4">
                    <Archive size={48} className="text-gray-300 dark:text-gray-700 opacity-40" />
                </div>
                <p className="text-gray-500 dark:text-gray-400">{t('documents.noDocuments')}</p>
            </div>
        );
    }

    return (
        <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden flex flex-col">
            <div 
                ref={parentRef}
                className="overflow-x-auto overflow-y-auto max-h-[calc(100vh-280px)] sm:max-h-[calc(100vh-320px)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-gray-700"
            >
                <table className="w-full text-left border-separate border-spacing-0">
                    <thead className="bg-gray-50 dark:bg-gray-800/90 backdrop-blur-sm sticky top-0 z-10 shadow-sm">
                    <tr>
                        <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-inherit border-b border-gray-100 dark:border-gray-800">{t('common.name')}</th>
                        <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-inherit border-b border-gray-100 dark:border-gray-800">{t('documents.type')}</th>
                        <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 bg-inherit border-b border-gray-100 dark:border-gray-800">{t('documents.documentDate')}</th>
                        <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 text-right bg-inherit border-b border-gray-100 dark:border-gray-800">{t('common.actions')}</th>
                    </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                    {paddingTop > 0 && (
                        <tr>
                            <td style={{ height: `${paddingTop}px` }} colSpan={4} />
                        </tr>
                    )}
                    {virtualRows.map((virtualRow) => {
                        const doc = documents[virtualRow.index];
                        return (
                            <DocumentRow
                                key={doc.id}
                                doc={doc}
                                onPreview={onPreview}
                                onEdit={onEdit}
                                onDelete={onDelete}
                                isPreviewLoading={isPreviewLoading}
                                t={t}
                            />
                        );
                    })}
                    {paddingBottom > 0 && (
                        <tr>
                            <td style={{ height: `${paddingBottom}px` }} colSpan={4} />
                        </tr>
                    )}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
