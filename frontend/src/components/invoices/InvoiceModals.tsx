import { useTranslation } from 'react-i18next';
import { X, Pencil, Split, FileText } from 'lucide-react';
import type { Invoice } from "../../api/types/invoice";
import type { Category } from "../../api/types/category";
import { InvoiceForm } from './InvoiceForm';
import { FilePreview } from '../FilePreview';

export function PreviewInvoiceModal({ 
    previewUrl, 
    mimeType, 
    invoice, 
    categories,
    onClose, 
    onUpdate, 
    isPending 
}: { 
    previewUrl: string, 
    mimeType: string,
    invoice: Invoice,
    categories: Category[],
    onClose: () => void,
    onUpdate: (id: string, data: any) => void,
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
                            {t('invoices.previewTitle')}
                        </h3>
                        <div className="hidden sm:flex items-center gap-2 px-3 py-1 bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300 text-xs font-medium rounded-full">
                            <FileText size={14} /> {invoice.vendor?.name || t('invoices.unknownVendor')}
                        </div>
                    </div>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                        <X size={20} />
                    </button>
                </div>
                
                <div className="flex-1 flex overflow-hidden">
                    {/* Left side: Document Preview */}
                    <div className="flex-1 bg-gray-200 dark:bg-gray-950 p-2 sm:p-4">
                        <FilePreview url={previewUrl} mimeType={mimeType} title={t('invoices.previewTitle')} />
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
                            <InvoiceForm 
                                initialData={invoice} 
                                categories={categories}
                                onSubmit={(data) => onUpdate(invoice.id, data)}
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

export function EditInvoiceModal({ 
    invoice, 
    categories,
    onUpdate, 
    onClose, 
    isPending 
}: { 
    invoice: Invoice, 
    categories: Category[],
    onUpdate: (id: string, data: any) => void, 
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
                        {t('invoices.editTitle')}
                    </h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors">
                        <X size={20} />
                    </button>
                </div>

                <div className="overflow-y-auto">
                    <InvoiceForm 
                        initialData={invoice} 
                        categories={categories}
                        onSubmit={(data) => onUpdate(invoice.id, data)}
                        isPending={isPending}
                        submitLabel={t('common.saveChanges')}
                    />
                </div>
            </div>
        </div>
    );
}
