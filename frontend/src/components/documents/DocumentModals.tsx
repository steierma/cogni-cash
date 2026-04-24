import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { X, FileUp, Pencil, BrainCircuit, Split, FileText } from 'lucide-react';
import type { Document, DocumentType } from "../../api/types/document";
import { DocumentForm } from './DocumentForm';
import { FilePreview } from '../FilePreview';

export function PreviewDocumentModal({ 
    previewUrl, 
    mimeType, 
    document, 
    onClose, 
    onUpdate, 
    isPending 
}: { 
    previewUrl: string, 
    mimeType: string,
    document: Document,
    onClose: () => void,
    onUpdate: (id: string, data: { file_name?: string, type?: DocumentType, document_date?: string }) => void,
    isPending: boolean 
}) {
    const { t } = useTranslation();

    return (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-[95vw] h-[95vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                    <div className="flex items-center gap-4">
                        <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                            <Split className="text-indigo-500" size={20} /> 
                            {t('documents.actions.preview')}
                        </h3>
                        <div className="hidden sm:flex items-center gap-2 px-3 py-1 bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300 text-xs font-medium rounded-full">
                            <FileText size={14} /> {document.file_name}
                        </div>
                    </div>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                        <X size={20} />
                    </button>
                </div>
                
                <div className="flex-1 flex overflow-hidden">
                    {/* Left side: Document Preview */}
                    <div className="flex-1 bg-gray-200 dark:bg-gray-950 p-2 sm:p-4">
                        <FilePreview url={previewUrl} mimeType={mimeType} title={t('documents.actions.preview')} />
                    </div>

                    {/* Right side: Values form */}
                    <div className="w-full max-w-md border-l border-gray-100 dark:border-gray-800 flex flex-col bg-white dark:bg-gray-900 shadow-2xl z-10">
                        <div className="p-4 border-b border-gray-100 dark:border-gray-800">
                            <h4 className="font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                                <Pencil size={16} className="text-blue-500" />
                                {t('payslips.modals.optionalOverrides')}
                            </h4>
                            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                                {t('payslips.modals.optionalOverridesDesc')}
                            </p>
                        </div>

                        <div className="flex-1 overflow-y-auto">
                            <DocumentForm 
                                initialData={{
                                    file_name: document.file_name,
                                    type: document.type,
                                    document_date: String(document.metadata?.date || '')
                                }} 
                                onSubmit={(data) => onUpdate(document.id, data)}
                                isPending={isPending}
                                submitLabel={t('common.saveAndClose')}
                            />
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export function ImportDocumentModal({ 
    onImport, 
    onClose, 
    isPending, 
    useAI 
}: { 
    onImport: (file: File, overrides: { file_name?: string, type?: DocumentType, document_date?: string }, useAI: boolean) => void, 
    onClose: () => void,
    isPending: boolean,
    useAI: boolean
}) {
    const { t } = useTranslation();
    const [file, setFile] = useState<File | null>(null);

    const handleFormSubmit = (data: { file_name?: string, type?: DocumentType, document_date?: string }) => {
        if (!file) return;
        onImport(file, data, useAI);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <FileUp className="text-indigo-500" size={20} /> 
                        {t('documents.importTitle')}
                    </h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                        <X size={20} />
                    </button>
                </div>
                
                <div className="p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30">
                    <label className="block text-xs font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">{t('documents.selectFile')}</label>
                    <input 
                        type="file" 
                        accept=".pdf,.png,.jpg,.jpeg,.webp,.gif" 
                        onChange={(e) => {
                            const selectedFile = e.target.files?.[0] || null;
                            setFile(selectedFile);
                        }} 
                        className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 dark:file:bg-indigo-900/30 file:text-indigo-700 dark:file:text-indigo-300 hover:file:bg-indigo-100 dark:hover:file:bg-indigo-900/50 transition-all cursor-pointer" 
                        required 
                    />
                    <div className="mt-3 flex items-center gap-3">
                        <div className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${useAI ? 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300' : 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300'}`}>
                            <BrainCircuit size={12} className={useAI ? 'animate-pulse' : ''} />
                            {useAI ? t('documents.aiModeActive') : t('documents.staticModeActive')}
                        </div>
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto">
                    <div className="p-4 bg-indigo-50/30 dark:bg-indigo-900/10 border-b border-indigo-100 dark:border-indigo-900/20">
                        <p className="text-xs text-indigo-800 dark:text-indigo-400 font-semibold">{t('payslips.modals.optionalOverrides')}</p>
                        <p className="text-[10px] text-indigo-700/70 dark:text-indigo-400/60 mt-0.5">{t('payslips.modals.optionalOverridesDesc')}</p>
                    </div>
                    <DocumentForm 
                        initialData={file ? { file_name: file.name } : {}} 
                        onSubmit={handleFormSubmit}
                        isPending={isPending}
                        submitLabel={t('common.import')}
                    />
                </div>
            </div>
        </div>
    );
}

export function EditDocumentModal({ 
    document, 
    onUpdate, 
    onClose, 
    isPending 
}: { 
    document: Document, 
    onUpdate: (id: string, data: { file_name?: string, type?: DocumentType, document_date?: string }) => void, 
    onClose: () => void,
    isPending: boolean 
}) {
    const { t } = useTranslation();

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh] animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Pencil className="text-indigo-500" size={20} /> 
                        {t('common.edit')}
                    </h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                        <X size={20} />
                    </button>
                </div>

                <div className="overflow-y-auto">
                    <DocumentForm 
                        initialData={{
                            file_name: document.file_name,
                            type: document.type,
                            document_date: String(document.metadata?.date || '')
                        }} 
                        onSubmit={(data) => onUpdate(document.id, data)}
                        isPending={isPending}
                        submitLabel={t('common.saveChanges')}
                    />
                </div>
            </div>
        </div>
    );
}
