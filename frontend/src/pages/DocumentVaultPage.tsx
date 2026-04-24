import { useState, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Archive, FileUp, Search, Filter, Trash2, Download,
    FileText, Briefcase, ShieldCheck, File, BrainCircuit, Pencil, Loader2,
    type LucideIcon
} from 'lucide-react';
import { documentService } from '../api/services/documentService';
import type { Document, DocumentType } from "../api/types/document";
import { fmtDate } from '../utils/formatters';
import { NavLink } from 'react-router-dom';
import { ImportDocumentModal, EditDocumentModal, PreviewDocumentModal } from '../components/documents/DocumentModals';

const TYPE_ICONS: Record<string, LucideIcon> = {
    tax_certificate: ShieldCheck,
    receipt: FileText,
    contract: Briefcase,
    other: File
};

export default function DocumentVaultPage() {
    const { t } = useTranslation();
    const queryClient = useQueryClient();
    const inputRef = useRef<HTMLInputElement>(null);

    const [searchTerm, setSearchTerm] = useState('');
    const [selectedType, setSelectedType] = useState<DocumentType | 'all'>('all');

    // UI States
    const [dragOver, setDragOver] = useState(false);
    const [useAI, setUseAI] = useState(true);

    // Modal States
    const [isImportModalOpen, setIsImportModalOpen] = useState(false);
    const [editingDoc, setEditingDoc] = useState<Document | null>(null);
    const [previewingDoc, setPreviewingDoc] = useState<Document | null>(null);
    const [previewInfo, setPreviewInfo] = useState<{ url: string, mimeType: string } | null>(null);
    const [isPreviewLoading, setIsPreviewLoading] = useState<string | null>(null);

    const { data: documents = [], isLoading } = useQuery({
        queryKey: ['documents', selectedType, searchTerm],
        queryFn: () => documentService.list(selectedType === 'all' ? undefined : selectedType, searchTerm),
    });

    const deleteMutation = useMutation({
        mutationFn: documentService.delete,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['documents'] })
    });

    const uploadMutation = useMutation({
        mutationFn: ({ file, overrides, useAI }: { file: File, overrides?: { file_name?: string, type?: DocumentType, document_date?: string }, useAI: boolean }) => {
            const skipAi = !useAI;
            return documentService.upload(file, overrides?.type, skipAi, overrides?.document_date, overrides?.file_name);
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['documents'] });
            setIsImportModalOpen(false);
        }
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string, data: { file_name?: string, type?: DocumentType, document_date?: string } }) =>
            documentService.update(id, {
                file_name: data.file_name,
                type: data.type,
                document_date: data.document_date || undefined
            }),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['documents'] });
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
            alert(t('documents.errors.previewFailed'));
        } finally {
            setIsPreviewLoading(null);
        }
    };

    const closePreview = () => {
        if (previewInfo) URL.revokeObjectURL(previewInfo.url);
        setPreviewInfo(null);
        setPreviewingDoc(null);
    };

    const handleFiles = (files: FileList | File[]) => {
        const fileList = Array.from(files);
        if (fileList.length === 0) return;

        // Currently we only handle one file at a time for documents
        const file = fileList[0];
        uploadMutation.mutate({ file, useAI });

        if (inputRef.current) inputRef.current.value = '';
    };

    const onDrop = (e: React.DragEvent) => {
        e.preventDefault();
        setDragOver(false);
        if (e.dataTransfer.files?.length > 0) handleFiles(e.dataTransfer.files);
    };

    const handleDelete = (id: string) => {
        if (window.confirm(t('documents.deleteConfirm'))) {
            deleteMutation.mutate(id);
        }
    };

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Header */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Archive className="text-indigo-600 dark:text-indigo-400" /> {t('documents.title')}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                        {t('documents.subtitle')}
                    </p>
                </div>
                <div className="flex items-center gap-3">
                    <NavLink
                        to={`/documents/tax/${new Date().getFullYear()}`}
                        className="flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300 text-sm font-medium rounded-xl hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                    >
                        <ShieldCheck size={18} className="text-indigo-500" />
                        {t('documents.taxSummary')}
                    </NavLink>
                </div>
            </div>

            {/* Quick Upload Dropzone (Aligned with Payslips) */}
            <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row gap-4 items-center">
                <div
                    onDragOver={(e) => { e.preventDefault(); setDragOver(true); }}
                    onDragLeave={() => setDragOver(false)}
                    onDrop={onDrop}
                    onClick={() => inputRef.current?.click()}
                    className={`flex-1 w-full border-2 border-dashed rounded-xl p-4 flex items-center justify-center gap-4 cursor-pointer transition-all duration-200 ${dragOver ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20' : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50 dark:bg-gray-800/30'}`}
                >
                    <input type="file" className="hidden" ref={inputRef} accept=".pdf,.png,.jpg,.jpeg,.webp,.gif"
                           onChange={(e) => e.target.files && handleFiles(e.target.files)}/>
                    <div className="p-2 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg text-indigo-600 dark:text-indigo-400">
                        <FileUp size={24}/>
                    </div>
                    <div>
                        {uploadMutation.isPending ? (
                            <p className="text-sm font-medium text-gray-900 dark:text-gray-100 animate-pulse">{t('documents.uploadingParsing')}</p>
                        ) : (
                            <>
                                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                                    <span className="text-indigo-600 dark:text-indigo-400">{t('documents.clickToParse')}</span> {t('bankStatements.import.orDrag')}
                                </p>
                                <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{t('payslips.pdfFormat')}</p>
                            </>
                        )}
                    </div>
                </div>

                <div className="flex flex-row md:flex-col gap-3 w-full md:w-auto md:min-w-[200px] justify-center items-center md:items-stretch">
                    <label className="flex items-center gap-2 cursor-pointer group px-1 justify-center md:justify-start">
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
                            <BrainCircuit size={16}/>
                            {t('documents.forceAi')}
                        </span>
                    </label>

                    <button
                        onClick={() => setIsImportModalOpen(true)}
                        disabled={uploadMutation.isPending}
                        className="flex items-center justify-center gap-2 w-full md:w-auto px-4 py-3 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 text-gray-700 dark:text-gray-300 text-sm font-medium rounded-xl transition-colors disabled:opacity-70"
                    >
                        <Pencil size={16}/> {t('documents.manualOverride')}
                    </button>
                </div>
            </div>

            {/* Filters Bar */}
            <div className="bg-white dark:bg-gray-900 p-4 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col md:flex-row items-center gap-4">
                <div className="relative flex-1 w-full">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" size={18} />
                    <input
                        type="text"
                        placeholder={t('documents.searchPlaceholder')}
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-sm focus:ring-2 focus:ring-indigo-500 outline-none transition-all"
                    />
                </div>
                <div className="flex items-center gap-3 w-full md:w-auto">
                    <div className="flex items-center gap-2 text-gray-500 dark:text-gray-400 whitespace-nowrap">
                        <Filter size={18} />
                        <span className="text-sm font-medium">{t('common.filters')}:</span>
                    </div>
                    <select
                        value={selectedType}
                        onChange={(e) => setSelectedType(e.target.value as DocumentType | 'all')}
                        className="flex-1 md:flex-none bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-xl px-4 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
                    >
                        <option value="all">{t('documents.allTypes')}</option>
                        <option value="tax_certificate">{t('documents.types.tax_certificate')}</option>
                        <option value="receipt">{t('documents.types.receipt')}</option>
                        <option value="contract">{t('documents.types.contract')}</option>
                        <option value="other">{t('documents.types.other')}</option>
                    </select>
                </div>
            </div>

            {/* Document List */}
            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
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
                        {isLoading ? (
                            Array.from({ length: 5 }).map((_, i) => (
                                <tr key={i} className="animate-pulse">
                                    <td colSpan={5} className="px-6 py-8"><div className="h-4 bg-gray-200 dark:bg-gray-800 rounded w-3/4 mx-auto"></div></td>
                                </tr>
                            ))
                        ) : documents.length === 0 ? (
                            <tr>
                                <td colSpan={5} className="px-6 py-20 text-center text-gray-500 dark:text-gray-400">
                                    <div className="flex flex-col items-center gap-3">
                                        <Archive size={48} className="opacity-20" />
                                        <p>{t('documents.noDocuments')}</p>
                                    </div>
                                </td>
                            </tr>
                        ) : (
                            documents.map((doc) => {
                                const Icon = TYPE_ICONS[doc.type] || File;
                                const docDate = String(doc.metadata?.date || doc.created_at);
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

            {/* Modals */}
            {isImportModalOpen && (
                <ImportDocumentModal
                    onClose={() => setIsImportModalOpen(false)}
                    onImport={(file, overrides, useAI) => uploadMutation.mutate({ file, overrides, useAI })}
                    isPending={uploadMutation.isPending}
                    useAI={useAI}
                />
            )}

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