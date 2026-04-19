import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { X, Eye, FileText, FileUp, Pencil, Split } from 'lucide-react';
import { FilePreview } from '../FilePreview';
import type { Payslip } from "../../api/types/payslip";
import { fmtCurrency } from '../../utils/formatters';
import { formatYearMonth } from './utils';
import { PayslipForm } from './PayslipForm';

export function ViewPayslipModal({ payslip, onClose }: { payslip: Payslip, onClose: () => void }) {
    const { t } = useTranslation();
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden max-h-[90vh] flex flex-col">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><Eye className="text-emerald-500" size={20} /> {t('payslips.modals.viewTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={20} /></button>
                </div>
                <div className="p-6 space-y-6 overflow-y-auto">
                    <div className="grid grid-cols-2 gap-y-4 gap-x-6 text-sm">
                        <div><p className="text-gray-500 dark:text-gray-400 mb-1">{t('payslips.modals.period')}</p><p className="font-medium text-gray-900 dark:text-gray-100">{formatYearMonth(payslip.period_year, payslip.period_month_num)}</p></div>
                        <div><p className="text-gray-500 dark:text-gray-400 mb-1">{t('payslips.modals.employer')}</p><p className="font-medium text-gray-900 dark:text-gray-100">{payslip.employer_name}</p></div>
                        <div><p className="text-gray-500 dark:text-gray-400 mb-1">{t('payslips.modals.taxClass')}</p><p className="font-medium text-gray-900 dark:text-gray-100">{payslip.tax_class || '-'}</p></div>
                        <div><p className="text-gray-500 dark:text-gray-400 mb-1">{t('payslips.modals.taxId')}</p><p className="font-medium text-gray-900 dark:text-gray-100">{payslip.tax_id || '-'}</p></div>
                    </div>
                    <hr className="border-gray-100 dark:border-gray-800" />
                    <div className="space-y-3 text-sm">
                        <div className="flex justify-between items-center"><span className="text-gray-500 dark:text-gray-400">{t('payslips.modals.gross')}</span><span className="font-mono font-medium text-gray-900 dark:text-gray-100">{fmtCurrency(payslip.gross_pay, 'EUR')}</span></div>
                        <div className="flex justify-between items-center"><span className="text-gray-500 dark:text-gray-400">{t('payslips.modals.net')}</span><span className="font-mono font-medium text-gray-900 dark:text-gray-100">{fmtCurrency(payslip.net_pay, 'EUR')}</span></div>
                        <div className="flex justify-between items-center"><span className="text-gray-500 dark:text-gray-400">{t('payslips.modals.leasing')}</span><span className="font-mono text-red-600 dark:text-red-400">{fmtCurrency(payslip.custom_deductions, 'EUR')}</span></div>
                        <div className="flex justify-between items-center pt-2 border-t border-gray-100 dark:border-gray-800"><span className="font-medium text-gray-700 dark:text-gray-300">{t('payslips.modals.payout')}</span><span className="font-mono font-bold text-lg text-emerald-600 dark:text-emerald-400">{fmtCurrency(payslip.payout_amount, 'EUR')}</span></div>
                    </div>
                    {payslip.bonuses && payslip.bonuses.length > 0 && (
                        <><hr className="border-gray-100 dark:border-gray-800" /><div><h4 className="font-medium text-gray-900 dark:text-gray-100 mb-3 text-sm">{t('payslips.modals.bonuses')}</h4><div className="space-y-2 text-sm">{payslip.bonuses.map((sz, idx) => (<div key={idx} className="flex justify-between items-center bg-gray-50 dark:bg-gray-800/50 p-2 rounded-lg"><span className="text-gray-600 dark:text-gray-400">{sz.description}</span><span className="font-mono text-gray-900 dark:text-gray-100">{fmtCurrency(sz.amount, 'EUR')}</span></div>))}</div></div></>
                    )}
                </div>
            </div>
        </div>
    );
}

export function PreviewPayslipModal({ previewUrl, mimeType, payslip, onClose, onUpdate, isPending }: { 
    previewUrl: string, 
    mimeType: string,
    payslip: Payslip,
    onClose: () => void,
    onUpdate: (id: string, data: Partial<Payslip>) => void,
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
                            {t('payslips.modals.previewTitle')}
                        </h3>
                        <div className="hidden sm:flex items-center gap-2 px-3 py-1 bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-300 text-xs font-medium rounded-full">
                            <FileText size={14} /> {payslip.original_file_name}
                        </div>
                    </div>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors">
                        <X size={20} />
                    </button>
                </div>
                
                <div className="flex-1 flex overflow-hidden">
                    {/* Left side: Document Preview */}
                    <div className="flex-1 bg-gray-200 dark:bg-gray-950 p-2 sm:p-4">
                        <FilePreview url={previewUrl} mimeType={mimeType} title={t('payslips.modals.previewTitle')} />
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

                        <PayslipForm 
                            initialData={payslip} 
                            onSubmit={(data) => onUpdate(payslip.id, data)}
                            isPending={isPending}
                            submitLabel={t('common.saveAndClose')}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
}

export function ImportPayslipModal({ onImport, onClose, isPending, useAI }: { 
    onImport: (file: File, overrides: Partial<Payslip>, useAI: boolean) => void, 
    onClose: () => void,
    isPending: boolean,
    useAI: boolean
}) {
    const { t } = useTranslation();
    const [file, setFile] = useState<File | null>(null);

    const handleFormSubmit = (data: Partial<Payslip>) => {
        if (!file) return;
        onImport(file, data, useAI);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh]">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><FileUp className="text-indigo-500" size={20} /> {t('payslips.modals.importTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={20} /></button>
                </div>
                
                <div className="p-4 border-b border-gray-100 dark:border-gray-800">
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.selectFile')}</label>
                    <input type="file" accept=".pdf,.png,.jpg,.jpeg,.webp,.gif" onChange={(e) => setFile(e.target.files?.[0] || null)} className="w-full text-sm text-gray-500 file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100 dark:file:bg-gray-800 dark:file:text-gray-300" required />
                    <div className="mt-2 flex items-center gap-2">
                        <div className={`w-2 h-2 rounded-full ${useAI ? 'bg-amber-500 animate-pulse' : 'bg-blue-500'}`} />
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                            {useAI ? t('payslips.modals.aiModeActive') : t('payslips.modals.staticModeActive')}
                        </span>
                    </div>
                </div>

                <div className="flex-1 overflow-hidden">
                    <div className="p-4 bg-amber-50 dark:bg-amber-900/10 border-b border-amber-100 dark:border-amber-900/20">
                        <p className="text-xs text-amber-800 dark:text-amber-400 font-medium">{t('payslips.modals.optionalOverrides')}</p>
                        <p className="text-[10px] text-amber-700/70 dark:text-amber-400/60">{t('payslips.modals.optionalOverridesDesc')}</p>
                    </div>
                    <div className="overflow-y-auto max-h-[50vh]">
                        <PayslipForm 
                            initialData={{}} 
                            onSubmit={handleFormSubmit}
                            isPending={isPending}
                            submitLabel={t('common.import')}
                            showSubmitIcon={false}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
}

export function EditPayslipModal({ payslip, onUpdate, onClose, isPending }: { 
    payslip: Payslip, 
    onUpdate: (id: string, data: Partial<Payslip>) => void, 
    onClose: () => void,
    isPending: boolean 
}) {
    const { t } = useTranslation();

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden flex flex-col max-h-[90vh]">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><Pencil className="text-blue-500" size={20} /> {t('payslips.modals.editTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={20} /></button>
                </div>

                <div className="overflow-y-auto">
                    <PayslipForm 
                        initialData={payslip} 
                        onSubmit={(data) => onUpdate(payslip.id, data)}
                        isPending={isPending}
                    />
                </div>
            </div>
        </div>
    );
}

export function BatchResultsModal({ successful, failed, onClose }: { 
    successful: Payslip[], 
    failed: { filename: string, error: string }[],
    onClose: () => void 
}) {
    const { t } = useTranslation();
    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-2xl overflow-hidden flex flex-col max-h-[85vh] animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">{t('payslips.modals.batchResults')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"><X size={20} /></button>
                </div>
                <div className="p-6 overflow-y-auto space-y-6">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="bg-emerald-50 dark:bg-emerald-900/20 p-4 rounded-xl border border-emerald-100 dark:border-emerald-900/30">
                            <p className="text-sm text-emerald-600 dark:text-emerald-400 font-medium mb-1">{t('payslips.modals.successful')}</p>
                            <p className="text-2xl font-bold text-emerald-700 dark:text-emerald-300">{successful.length}</p>
                        </div>
                        <div className={`p-4 rounded-xl border ${failed.length > 0 ? 'bg-red-50 dark:bg-red-900/20 border-red-100 dark:border-red-900/30' : 'bg-gray-50 dark:bg-gray-800/50 border-gray-100 dark:border-gray-800'}`}>
                            <p className={`text-sm font-medium mb-1 ${failed.length > 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-500'}`}>{t('payslips.modals.failed')}</p>
                            <p className={`text-2xl font-bold ${failed.length > 0 ? 'text-red-700 dark:text-red-300' : 'text-gray-400'}`}>{failed.length}</p>
                        </div>
                    </div>

                    {failed.length > 0 && (
                        <div>
                            <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3 flex items-center gap-2">
                                <span className="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                                {t('payslips.modals.errorDetails')}
                            </h4>
                            <div className="space-y-2">
                                {failed.map((f, i) => (
                                    <div key={i} className="p-3 bg-red-50/50 dark:bg-red-900/10 rounded-lg border border-red-100/50 dark:border-red-900/20 text-sm">
                                        <p className="font-semibold text-red-800 dark:text-red-300 flex items-center gap-2">
                                            <FileText size={14} /> {f.filename}
                                        </p>
                                        <p className="text-red-600/80 dark:text-red-400/70 mt-1 pl-5">{f.error}</p>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {successful.length > 0 && (
                        <div>
                            <h4 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-3 flex items-center gap-2">
                                <span className="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                                {t('payslips.modals.importedFiles')}
                            </h4>
                            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                                {successful.map((s, i) => (
                                    <div key={i} className="p-2.5 bg-emerald-50/30 dark:bg-emerald-900/10 rounded-lg border border-emerald-100/30 dark:border-emerald-900/10 text-xs flex items-center justify-between">
                                        <span className="text-emerald-800 dark:text-emerald-300 truncate mr-2" title={s.original_file_name}>{s.original_file_name}</span>
                                        <span className="text-emerald-600 dark:text-emerald-500 font-medium shrink-0">{formatYearMonth(s.period_year, s.period_month_num)}</span>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
                <div className="p-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 flex justify-end">
                    <button onClick={onClose} className="px-6 py-2 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-xl font-medium hover:bg-gray-800 dark:hover:bg-gray-200 transition-colors">
                        {t('common.close')}
                    </button>
                </div>
            </div>
        </div>
    );
}
