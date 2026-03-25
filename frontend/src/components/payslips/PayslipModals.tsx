import { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { X, Eye, FileText, Plus, FileUp, Pencil, Loader2 } from 'lucide-react';
import type { Payslip, User } from '../../api/types';
import { fmtCurrency } from '../../utils/formatters';
import { formatYearMonth } from './utils';

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
                        <div><p className="text-gray-500 dark:text-gray-400 mb-1">{t('payslips.modals.employee')}</p><p className="font-medium text-gray-900 dark:text-gray-100">{payslip.employee_name}</p></div>
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

export function PreviewPayslipModal({ previewUrl, onClose }: { previewUrl: string, onClose: () => void }) {
    const { t } = useTranslation();
    return (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm animate-in fade-in duration-200">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-5xl h-[90vh] flex flex-col overflow-hidden animate-in zoom-in-95 duration-200">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><FileText className="text-fuchsia-500" size={20} /> {t('payslips.modals.previewTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-900 dark:hover:text-gray-100 p-1.5 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"><X size={20} /></button>
                </div>
                <div className="flex-1 w-full bg-gray-200 dark:bg-gray-950 p-2 sm:p-4">
                    <iframe src={`${previewUrl}#toolbar=0`} className="w-full h-full rounded-xl border border-gray-300 dark:border-gray-800 shadow-inner bg-white" title="PDF Preview" />
                </div>
            </div>
        </div>
    );
}

export function ImportPayslipModal({ isOpen, onClose, currentUser, onImport, isPending }: { isOpen: boolean, onClose: () => void, currentUser?: User, onImport: (file: File, overrides: Partial<Payslip>) => void, isPending: boolean }) {
    const { t } = useTranslation();
    const [modalFile, setModalFile] = useState<File | null>(null);
    const [modalDragOver, setModalDragOver] = useState(false);
    const [importBonuses, setImportBonuses] = useState<{ description: string; amount: string }[]>([]);
    const modalInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (isOpen) { setModalFile(null); setImportBonuses([]); setModalDragOver(false); }
    }, [isOpen]);

    if (!isOpen) return null;

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const file = modalFile || (formData.get('file') as File);
        if (!file || file.size === 0) return;

        const overrides: Partial<Payslip> = {};
        const monthNum = formData.get('period_month_num') as string; if (monthNum) overrides.period_month_num = Number(monthNum);
        const year = formData.get('period_year') as string; if (year) overrides.period_year = Number(year);
        const empName = formData.get('employee_name') as string; if (empName) overrides.employee_name = empName;
        const taxClass = formData.get('tax_class') as string; if (taxClass) overrides.tax_class = taxClass;
        const taxId = formData.get('tax_id') as string; if (taxId) overrides.tax_id = taxId;
        const gross = formData.get('gross_pay') as string; if (gross) overrides.gross_pay = Number(gross);
        const net = formData.get('net_pay') as string; if (net) overrides.net_pay = Number(net);
        const payout = formData.get('payout_amount') as string; if (payout) overrides.payout_amount = Number(payout);
        const leasing = formData.get('custom_deductions') as string; if (leasing) overrides.custom_deductions = Number(leasing);
        if (importBonuses && importBonuses.length > 0) {
            overrides.bonuses = importBonuses.filter(b => b.description.trim() && b.amount.trim()).map(b => ({ description: b.description, amount: Number(b.amount) }));
        }
        onImport(file, overrides);
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden max-h-[90vh] flex flex-col">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><Plus className="text-indigo-500" size={20} /> {t('payslips.modals.importTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={20} /></button>
                </div>
                <form onSubmit={handleSubmit} className="p-4 space-y-4 overflow-y-auto">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.pdfFile')}</label>
                        <div
                            onDragOver={(e) => { e.preventDefault(); setModalDragOver(true); }}
                            onDragLeave={() => setModalDragOver(false)}
                            onDrop={(e) => { e.preventDefault(); setModalDragOver(false); if (e.dataTransfer.files?.length > 0) setModalFile(e.dataTransfer.files[0]); }}
                            onClick={() => modalInputRef.current?.click()}
                            className={`border-2 border-dashed rounded-xl p-4 flex flex-col items-center justify-center cursor-pointer transition-colors ${modalDragOver ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20' : 'border-gray-300 dark:border-gray-700 hover:border-indigo-400 bg-gray-50 dark:bg-gray-800/30'}`}
                        >
                            <input type="file" name="file" className="hidden" ref={modalInputRef} accept=".pdf" onChange={(e) => { if (e.target.files && e.target.files.length > 0) setModalFile(e.target.files[0]); }} required={!modalFile} />
                            <FileUp size={24} className="text-indigo-500 mb-2" />
                            {modalFile ? <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{modalFile.name}</span> : <><span className="text-sm font-medium text-gray-900 dark:text-gray-100">{t('payslips.modals.clickToSelect')}</span><span className="text-xs text-gray-500 dark:text-gray-400 mt-1">{t('payslips.modals.pdfRequired')}</span></>}
                        </div>
                    </div>
                    <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-100 dark:border-blue-900/50 p-3 rounded-xl text-sm text-blue-800 dark:text-blue-300"><strong>{t('payslips.modals.optionalOverrides')}</strong> {t('payslips.modals.optionalOverridesDesc')}</div>
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2"><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.empName')}</label><input name="employee_name" defaultValue={currentUser?.full_name || ''} placeholder="e.g. John Doe" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.month')}</label><input name="period_month_num" type="number" min="1" max="12" placeholder="e.g. 3" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.year')}</label><input name="period_year" type="number" placeholder="2024" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.taxClass')}</label><input name="tax_class" placeholder="e.g. 1" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div className="col-span-2 sm:col-span-1"><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.taxId')}</label><input name="tax_id" placeholder="e.g. 12 345 678 901" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                    </div>
                    <hr className="border-gray-100 dark:border-gray-800 my-4" />
                    <div className="grid grid-cols-2 gap-4">
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.gross')}</label><input name="gross_pay" type="number" step="0.01" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.net')}</label><input name="net_pay" type="number" step="0.01" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.leasing')}</label><input name="custom_deductions" type="number" step="0.01" placeholder="-0.00" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.payoutInput')}</label><input name="payout_amount" type="number" step="0.01" className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('payslips.modals.bonuses')}</label>
                        {importBonuses.map((b, idx) => (
                            <div key={idx} className="flex gap-2 mb-2">
                                <input type="text" placeholder={t('payslips.modals.descPlaceholder')} value={b.description} onChange={e => { const next = [...importBonuses]; next[idx].description = e.target.value; setImportBonuses(next); }} className="flex-1 px-2 py-1 border rounded-lg dark:bg-gray-800 dark:border-gray-700" />
                                <input type="number" placeholder={t('payslips.modals.amountPlaceholder')} value={b.amount} onChange={e => { const next = [...importBonuses]; next[idx].amount = e.target.value; setImportBonuses(next); }} className="w-28 px-2 py-1 border rounded-lg dark:bg-gray-800 dark:border-gray-700" />
                                <button type="button" onClick={() => setImportBonuses(importBonuses.filter((_, i) => i !== idx))} className="text-red-500 px-2">{t('payslips.modals.remove')}</button>
                            </div>
                        ))}
                        <button type="button" onClick={() => setImportBonuses([...importBonuses, { description: '', amount: '' }])} className="text-indigo-600 dark:text-indigo-400 text-sm mt-2">{t('payslips.modals.addBonus')}</button>
                    </div>
                    <div className="pt-4 flex justify-end gap-2 border-t border-gray-100 dark:border-gray-800">
                        <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">{t('payslips.modals.cancel')}</button>
                        <button type="submit" disabled={isPending} className="px-4 py-2 bg-indigo-600 text-white text-sm rounded-lg hover:bg-indigo-700 flex items-center gap-2">
                            {isPending && <Loader2 size={16} className="animate-spin" />} {t('payslips.modals.uploadDoc')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}

export function EditPayslipModal({ payslip, onClose, onUpdate, isPending }: { payslip: Payslip | null, onClose: () => void, onUpdate: (id: string, data: Partial<Payslip> | FormData) => void, isPending: boolean }) {
    const { t } = useTranslation();
    const [editFile, setEditFile] = useState<File | null>(null);
    const [editFileDragOver, setEditFileDragOver] = useState(false);
    const [editBonuses, setEditBonuses] = useState<{ description: string; amount: string }[]>([]);
    const editFileInputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (payslip) {
            setEditFile(null);
            setEditBonuses((payslip.bonuses ?? []).map(sz => ({ description: sz.description, amount: sz.amount.toString() })));
        }
    }, [payslip]);

    if (!payslip) return null;

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const data: any = {
            period_month_num: Number(formData.get('period_month_num')),
            period_year: Number(formData.get('period_year')),
            employee_name: formData.get('employee_name') as string,
            tax_class: formData.get('tax_class') as string,
            tax_id: formData.get('tax_id') as string,
            gross_pay: Number(formData.get('gross_pay')),
            net_pay: Number(formData.get('net_pay')),
            payout_amount: Number(formData.get('payout_amount')),
            custom_deductions: Number(formData.get('custom_deductions')),
        };
        data.bonuses = editBonuses.filter(b => b.description.trim() && b.amount.trim()).map(b => ({ description: b.description, amount: Number(b.amount) }));

        if (editFile) {
            const payload = new FormData();
            payload.append('file', editFile);
            payload.append('data', JSON.stringify(data));
            onUpdate(payslip.id, payload);
        } else {
            onUpdate(payslip.id, data);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white dark:bg-gray-900 rounded-2xl shadow-xl w-full max-w-lg overflow-hidden max-h-[90vh] flex flex-col">
                <div className="flex items-center justify-between p-4 border-b border-gray-100 dark:border-gray-800">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2"><Pencil className="text-blue-500" size={20} /> {t('payslips.modals.editTitle')}</h3>
                    <button onClick={onClose} className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"><X size={20} /></button>
                </div>
                <form onSubmit={handleSubmit} className="p-4 space-y-4 overflow-y-auto">
                    {!payslip.original_file_mime && (
                        <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.attachMissing')}</label>
                            <div
                                onDragOver={(e) => { e.preventDefault(); setEditFileDragOver(true); }}
                                onDragLeave={() => setEditFileDragOver(false)}
                                onDrop={(e) => { e.preventDefault(); setEditFileDragOver(false); if (e.dataTransfer.files?.length > 0) setEditFile(e.dataTransfer.files[0]); }}
                                onClick={() => editFileInputRef.current?.click()}
                                className={`border-2 border-dashed rounded-xl p-3 flex flex-col items-center justify-center cursor-pointer transition-colors ${editFileDragOver ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-gray-300 dark:border-gray-700 hover:border-blue-400 bg-gray-50 dark:bg-gray-800/30'}`}
                            >
                                <input type="file" className="hidden" ref={editFileInputRef} accept=".pdf" onChange={(e) => { if (e.target.files && e.target.files.length > 0) setEditFile(e.target.files[0]); }} />
                                <FileUp size={20} className="text-blue-500 mb-1" />
                                {editFile ? <span className="text-sm font-medium text-gray-900 dark:text-gray-100">{editFile.name}</span> : <span className="text-xs text-gray-500 dark:text-gray-400">{t('payslips.modals.clickOrDragMissing')}</span>}
                            </div>
                        </div>
                    )}
                    <div className="grid grid-cols-2 gap-4">
                        <div className="col-span-2"><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.empName')}</label><input name="employee_name" defaultValue={payslip.employee_name} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.month')}</label><input name="period_month_num" type="number" min="1" max="12" defaultValue={payslip.period_month_num} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.year')}</label><input name="period_year" type="number" defaultValue={payslip.period_year} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.taxClass')}</label><input name="tax_class" defaultValue={payslip.tax_class} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.taxId')}</label><input name="tax_id" defaultValue={payslip.tax_id} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                    </div>
                    <hr className="border-gray-100 dark:border-gray-800 my-4" />
                    <div className="grid grid-cols-2 gap-4">
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.gross')}</label><input name="gross_pay" type="number" step="0.01" defaultValue={payslip.gross_pay} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.net')}</label><input name="net_pay" type="number" step="0.01" defaultValue={payslip.net_pay} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.leasing')}</label><input name="custom_deductions" type="number" step="0.01" defaultValue={payslip.custom_deductions} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" /></div>
                        <div><label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('payslips.modals.payoutInput')}</label><input name="payout_amount" type="number" step="0.01" defaultValue={payslip.payout_amount} className="w-full px-3 py-2 border rounded-lg dark:bg-gray-800 dark:border-gray-700" required /></div>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{t('payslips.modals.bonuses')}</label>
                        {editBonuses.map((b, idx) => (
                            <div key={idx} className="flex gap-2 mb-2">
                                <input type="text" placeholder={t('payslips.modals.descPlaceholder')} value={b.description} onChange={e => { const next = [...editBonuses]; next[idx].description = e.target.value; setEditBonuses(next); }} className="flex-1 px-2 py-1 border rounded-lg dark:bg-gray-800 dark:border-gray-700" />
                                <input type="number" placeholder={t('payslips.modals.amountPlaceholder')} value={b.amount} onChange={e => { const next = [...editBonuses]; next[idx].amount = e.target.value; setEditBonuses(next); }} className="w-28 px-2 py-1 border rounded-lg dark:bg-gray-800 dark:border-gray-700" />
                                <button type="button" onClick={() => setEditBonuses(editBonuses.filter((_, i) => i !== idx))} className="text-red-500 px-2">{t('payslips.modals.remove')}</button>
                            </div>
                        ))}
                        <button type="button" onClick={() => setEditBonuses([...editBonuses, { description: '', amount: '' }])} className="text-blue-600 dark:text-blue-400 text-sm mt-2">{t('payslips.modals.addBonus')}</button>
                    </div>
                    <div className="pt-4 flex justify-end gap-2 border-t border-gray-100 dark:border-gray-800">
                        <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">{t('payslips.modals.cancel')}</button>
                        <button type="submit" disabled={isPending} className="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 flex items-center gap-2">
                            {isPending && <Loader2 size={16} className="animate-spin" />} {t('payslips.modals.saveChanges')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}