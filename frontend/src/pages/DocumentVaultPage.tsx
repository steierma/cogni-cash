import { useState, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Archive, FileUp, Search, Filter,
    BrainCircuit, Pencil, ShieldCheck
} from 'lucide-react';
import { documentService } from '../api/services/documentService';
import type { Document, DocumentType } from "../api/types/document";
import { NavLink } from 'react-router-dom';
import { ImportDocumentModal, EditDocumentModal, PreviewDocumentModal } from '../components/documents/DocumentModals';
import DocumentTable from '../components/documents/DocumentTable';
import { LLMEnforcementWarning } from '../components/LLMEnforcementWarning';

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
        mutationFn: ({ id, data }: { id: string, data: { file_name?: string, type?: DocumentType, document_date?: string, file?: File } }) =>
            documentService.update(id, {
                file_name: data.file_name,
                type: data.type,
                document_date: data.document_date || undefined,
                file: data.file
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
        <div className="space-y-6 pb-10 animate-in fade-in slide-in-from-bottom-4 duration-500">
            <LLMEnforcementWarning />
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
            <DocumentTable
                documents={documents}
                isLoading={isLoading}
                onPreview={handlePreview}
                onEdit={setEditingDoc}
                onDelete={handleDelete}
                isPreviewLoading={isPreviewLoading}
            />

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