import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Save, Loader2, FileText, Calendar, Tag, FileUp } from 'lucide-react';
import type { DocumentType } from "../../api/types/document";

interface DocumentFormProps {
    initialData: {
        file_name?: string;
        type?: DocumentType;
        document_date?: string;
    };
    onSubmit: (data: {
        file_name?: string;
        type?: DocumentType;
        document_date?: string;
        file?: File;
    }) => void;
    isPending: boolean;
    submitLabel?: string;
    showFileUpload?: boolean;
}

export function DocumentForm({ initialData, onSubmit, isPending, submitLabel, showFileUpload }: DocumentFormProps) {
    const { t } = useTranslation();
    const [fileName, setFileName] = useState(initialData.file_name || '');
    const [type, setType] = useState<DocumentType>(initialData.type || 'other');
    const [date, setDate] = useState(initialData.document_date || '');
    const [file, setFile] = useState<File | null>(null);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit({
            file_name: fileName,
            type,
            document_date: date || undefined,
            file: file || undefined
        });
    };

    return (
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {showFileUpload && (
                <div className="space-y-1.5 pb-2 border-b border-gray-100 dark:border-gray-800">
                    <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center gap-2">
                        <FileUp size={14} /> {t('documents.selectFile')} ({t('common.optional')})
                    </label>
                    <input
                        type="file"
                        accept=".pdf,.png,.jpg,.jpeg,.webp,.gif"
                        onChange={(e) => setFile(e.target.files?.[0] || null)}
                        className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 dark:file:bg-indigo-900/30 file:text-indigo-700 dark:file:text-indigo-300 hover:file:bg-indigo-100 dark:hover:file:bg-indigo-900/50 transition-all cursor-pointer"
                    />
                    <p className="text-[10px] text-amber-600 dark:text-amber-500 font-medium mt-1">
                        {t('documents.reuploadWarning', 'Note: Uploading a new file will replace the current content and re-run AI extraction.')}
                    </p>
                </div>
            )}
            <div className="space-y-1.5">
                <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center gap-2">
                    <FileText size={14} /> {t('common.name')}
                </label>
                <input
                    type="text"
                    value={fileName}
                    onChange={(e) => setFileName(e.target.value)}
                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-xl px-4 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-all text-gray-900 dark:text-gray-100"
                    placeholder="e.g. Contract_2024.pdf"
                />
            </div>

            <div className="space-y-1.5">
                <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center gap-2">
                    <Tag size={14} /> {t('documents.type')}
                </label>
                <select
                    value={type}
                    onChange={(e) => setType(e.target.value as DocumentType)}
                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-xl px-4 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-all text-gray-900 dark:text-gray-100"
                >
                    <option value="tax_certificate">{t('documents.types.tax_certificate')}</option>
                    <option value="receipt">{t('documents.types.receipt')}</option>
                    <option value="contract">{t('documents.types.contract')}</option>
                    <option value="other">{t('documents.types.other')}</option>
                </select>
            </div>

            <div className="space-y-1.5">
                <label className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center gap-2">
                    <Calendar size={14} /> {t('documents.documentDate')}
                </label>
                <input
                    type="date"
                    value={date}
                    onChange={(e) => setDate(e.target.value)}
                    className="w-full bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-sm rounded-xl px-4 py-2 outline-none focus:ring-2 focus:ring-indigo-500 transition-all text-gray-900 dark:text-gray-100 [color-scheme:light] dark:[color-scheme:dark]"
                />
            </div>

            <div className="pt-4">
                <button
                    type="submit"
                    disabled={isPending}
                    className="w-full py-3 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-medium shadow-sm transition-all disabled:opacity-50 flex items-center justify-center gap-2"
                >
                    {isPending ? <Loader2 size={18} className="animate-spin" /> : <Save size={18} />}
                    {submitLabel || t('common.save')}
                </button>
            </div>
        </form>
    );
}
