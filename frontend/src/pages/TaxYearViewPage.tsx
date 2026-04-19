import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
    ShieldCheck, ArrowLeft, FileText,
    Briefcase, TrendingUp, Calculator,
    File, Download, Trash2, Pencil, Loader2, Archive
} from 'lucide-react';
import { documentService } from '../api/services/documentService';
import { fmtCurrency, fmtDate } from '../utils/formatters';
import type { Document, DocumentType } from '../api/types';
import { EditDocumentModal, PreviewDocumentModal } from '../components/documents/DocumentModals';

const TYPE_ICONS: Record<string, any> = {
    tax_certificate: ShieldCheck,
    receipt: FileText,
    contract: Briefcase,
    other: File
};

export default function TaxYearViewPage() {
    const { year } = useParams<{ year: string }>();
    const navigate = useNavigate();
    const { t } = useTranslation();
    const queryClient = useQueryClient();

    const taxYear = parseInt(year || new Date().getFullYear().toString());

    // Modal States
    const [editingDoc, setEditingDoc] = useState<Document | null>(null);
    const [previewingDoc, setPreviewingDoc] = useState<Document | null>(null);
    const [previewInfo, setPreviewInfo] = useState<{ url: string, mimeType: string } | null>(null);
    const [isPreviewLoading, setIsPreviewLoading] = useState<string | null>(null);

    const { data: summary, isLoading } = useQuery({
        queryKey: ['tax-summary', taxYear],
        queryFn: () => documentService.getTaxSummary(taxYear),
    });

    const deleteMutation = useMutation({
        mutationFn: documentService.delete,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tax-summary', taxYear] })
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: { file_name?: string, type?: DocumentType, document_date?: string } }) =>
            documentService.update(id, {
                file_name: data.file_name,
                type: data.type,
                document_date: data.document_date || undefined
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['tax-summary', taxYear] });
            setEditingDoc(null);
            setPreviewingDoc(null);
            closePreview();
        }
    });

    const handlePreview = async (doc: Document) => {
        try {
            setIsPreviewLoading(doc.id);
            const info = await documentService.getPreview(doc.id);
            setPreviewInfo(info);
            setPreviewingDoc(doc);
        } catch (err) {
            console.error(err);
            alert(t('documents.errors.previewFailed') || 'Preview failed');
        } finally {
            setIsPreviewLoading(null);
        }
    };

    const closePreview = () => {
        if (previewInfo) URL.revokeObjectURL(previewInfo.url);
        setPreviewInfo(null);
        setPreviewingDoc(null);
    };

    const handleDelete = (id: string) => {
        if (window.confirm(t('documents.deleteConfirm'))) {
            deleteMutation.mutate(id);
        }
    };

    const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i);

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div className="flex items-center gap-4">
                    <button
                        onClick={() => navigate('/documents')}
                        className="p-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 transition-all"
                    >
                        <ArrowLeft size={20} />
                    </button>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                            <ShieldCheck className="text-indigo-600 dark:text-indigo-400" />
                            {t('tax.title', { year: taxYear })}
                        </h1>
                        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                            {t('tax.subtitle')}
                        </p>
                    </div>
                </div>
                <div className="flex items-center gap-3 bg-white dark:bg-gray-900 p-1 rounded-xl border border-gray-200 dark:border-gray-800">
                    {years.map(y => (
                        <button
                            key={y}
                            onClick={() => navigate(`/documents/tax/${y}`)}
                            className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${taxYear === y ? 'bg-indigo-600 text-white shadow-sm' : 'text-gray-500 hover:bg-gray-50 dark:hover:bg-gray-800'}`}
                        >
                            {y}
                        </button>
                    ))}
                </div>
            </div>

            {isLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                    {Array.from({ length: 4 }).map((_, i) => (
                        <div key={i} className="h-32 bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 animate-pulse"></div>
                    ))}
                </div>
            ) : !summary ? (
                <div className="bg-white dark:bg-gray-900 p-12 rounded-2xl border border-gray-200 dark:border-gray-800 text-center">
                    <ShieldCheck size={48} className="mx-auto text-gray-300 mb-4" />
                    <p className="text-gray-500 dark:text-gray-400">{t('tax.noData')}</p>
                </div>
            ) : (
                <>
                    {/* KPI Grid */}
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                        <div className="bg-white dark:bg-gray-900 p-6 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm relative overflow-hidden group">
                            <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:scale-110 transition-transform">
                                <TrendingUp size={64} className="text-indigo-600" />
                            </div>
                            <p className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">{t('tax.totalGrossIncome')}</p>
                            <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(summary.total_gross_income, 'EUR')}</p>
                            <p className="text-sm text-indigo-600 dark:text-indigo-400 font-medium mt-1">{t('tax.status.complete')}</p>
                        </div>

                        <div className="bg-white dark:bg-gray-900 p-6 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm relative overflow-hidden group">
                            <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:scale-110 transition-transform">
                                <Briefcase size={64} className="text-emerald-600" />
                            </div>
                            <p className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">{t('tax.totalNetIncome')}</p>
                            <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(summary.total_net_income, 'EUR')}</p>
                            <p className="text-sm text-emerald-600 dark:text-emerald-400 font-medium mt-1">{t('tax.totalIncomeTax')}: {fmtCurrency(summary.total_income_tax, 'EUR')}</p>
                        </div>

                        <div className="bg-white dark:bg-gray-900 p-6 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm relative overflow-hidden group">
                            <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:scale-110 transition-transform">
                                <FileText size={64} className="text-blue-600" />
                            </div>
                            <p className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">{t('tax.totalDeductible')}</p>
                            <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">{fmtCurrency(summary.total_deductible, 'EUR')}</p>
                            <p className="text-sm text-blue-600 dark:text-blue-400 font-medium mt-1">{summary.documents.filter(d => d.type === 'receipt').length} {t('documents.types.receipt')}</p>
                        </div>

                        <div className="bg-white dark:bg-gray-900 p-6 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm relative overflow-hidden group">
                            <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:scale-110 transition-transform">
                                <ShieldCheck size={64} className="text-orange-600" />
                            </div>
                            <p className="text-xs font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">{t('tax.documentCount')}</p>
                            <p className="text-3xl font-bold text-gray-900 dark:text-gray-100">{summary.documents.length}</p>
                            <p className="text-sm text-orange-600 dark:text-orange-400 font-medium mt-1">{t('tax.status.complete')}</p>
                        </div>
                    </div>

                    {/* Document List for the Year */}
                    <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden mt-8">
                        <div className="p-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between bg-gray-50/50 dark:bg-gray-800/30">
                            <h2 className="text-base font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-2">
                                <ShieldCheck size={18} className="text-indigo-500" />
                                {t('tax.relevantDocuments')}
                            </h2>
                            <span className="text-xs font-bold bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 px-2.5 py-1 rounded-full">
                                {summary.documents.length} {t('tax.documentCount')}
                            </span>
                        </div>

                        <div className="overflow-x-auto">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                <tr className="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-800">
                                    <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{t('common.name')}</th>
                                    <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{t('documents.type')}</th>
                                    <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{t('documents.documentDate')}</th>
                                    <th className="px-6 py-4 text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400 text-right">{t('common.actions')}</th>
                                </tr>
                                </thead>
                                <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                                {summary.documents.length === 0 ? (
                                    <tr>
                                        <td colSpan={5} className="px-6 py-20 text-center text-gray-500 dark:text-gray-400">
                                            <div className="flex flex-col items-center gap-3">
                                                <Archive size={48} className="opacity-20" />
                                                <p>{t('tax.noData')}</p>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    summary.documents.map((doc) => {
                                        const Icon = TYPE_ICONS[doc.type] || File;
                                        const docDate = doc.metadata?.date ? doc.metadata.date : doc.created_at;
                                        return (
                                            <tr key={doc.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors group">
                                                <td className="px-6 py-4">
                                                    <div className="flex items-center gap-3">
                                                        <div className="w-10 h-10 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 flex items-center justify-center shrink-0 group-hover:scale-110 transition-transform">
                                                            <Icon className="text-indigo-600 dark:text-indigo-400" size={20} />
                                                        </div>
                                                        <div>
                                                            <p className="text-sm font-semibold text-gray-900 dark:text-gray-100">{doc.file_name}</p>
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
                                                    <p className="text-sm text-gray-600 dark:text-gray-300 font-medium">{fmtDate(docDate)}</p>
                                                    <p className="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-tight">{t('documents.uploadDate')}: {fmtDate(doc.created_at)}</p>
                                                </td>
                                                <td className="px-6 py-4 text-right">
                                                    <div className="flex items-center justify-end gap-2">
                                                        <button
                                                            onClick={() => handlePreview(doc)}
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
                                                            onClick={() => setEditingDoc(doc)}
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
                                                            onClick={() => handleDelete(doc.id)}
                                                            className="p-2 text-gray-400 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-all"
                                                            title={t('documents.actions.delete')}
                                                        >
                                                            <Trash2 size={18} />
                                                        </button>
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

                    <div className="bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-800/50 p-6 rounded-2xl flex items-start gap-4">
                        <div className="p-2 bg-indigo-100 dark:bg-indigo-900/40 rounded-xl text-indigo-600 dark:text-indigo-400">
                            <Calculator size={24} />
                        </div>
                        <div>
                            <h3 className="font-bold text-indigo-900 dark:text-indigo-100">Tax Readiness Note</h3>
                            <p className="text-sm text-indigo-700 dark:text-indigo-300 mt-1">
                                This summary includes all documents marked as "Tax Relevant" for the year {taxYear}.
                                Ensure you have uploaded all relevant tax certificates and receipts to get an accurate representation for your tax return.
                            </p>
                        </div>
                    </div>
                </>
            )}

            {/* Modals */}
            {editingDoc && (
                <EditDocumentModal
                    document={editingDoc}
                    onClose={() => setEditingDoc(null)}
                    onUpdate={(id, data) => updateMutation.mutate({ id, data })}
                    isPending={updateMutation.isPending}
                />
            )}

            {previewingDoc && previewInfo && (
                <PreviewDocumentModal
                    document={previewingDoc}
                    previewUrl={previewInfo.url}
                    mimeType={previewInfo.mimeType}
                    onClose={closePreview}
                    onUpdate={(id, data) => updateMutation.mutate({ id, data })}
                    isPending={updateMutation.isPending}
                />
            )}
        </div>
    );
}