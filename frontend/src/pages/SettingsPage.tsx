import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { KeyRound, CheckCircle2, AlertCircle, Settings, Server, Database, Save, Palette, Globe, ChevronDown, ChevronRight, MessageSquareCode } from 'lucide-react';
import { changePassword, fetchSystemInfo, fetchSettings, updateSettings } from '../api/client';

const DEFAULT_SINGLE_PROMPT = `Categorize the following invoice text. 
Use EXACTLY ONE category from: [{{CATEGORIES}}].
Return ONLY a valid JSON object. Do not include explanations.

JSON Schema:
{"category_name": "string", "vendor_name": "string", "amount": 12.34, "currency": "EUR", "description": "string"}

TEXT:
{{TEXT}}`;

const DEFAULT_BATCH_PROMPT = `Categorize these transactions using ONLY: [{{CATEGORIES}}].
Return ONLY a valid JSON array of objects.
Each object MUST have "hash" and "category".

DATA:
{{DATA}}`;

const DEFAULT_PAYSLIP_PROMPT = `Role: You are a precise financial data extraction system.
Task: Extract payroll information from the provided payslip and map it strictly to the JSON schema below.

Strict Extraction Rules:

No Hallucinations: Extract values exactly as they are represented. Do not calculate, guess, or infer numbers.

Missing Data: If a value is not explicitly found in the text, you must return null for that field.

Number Formatting: Convert localized number formats (e.g., 1.234,56 or 1,234.56) into standard float values (e.g., 1234.56) without thousands separators.

Date Mapping: Convert month names found in the text into their corresponding integer (e.g., "January" / "Januar" = 1, "May" / "Mai" = 5).

Output Constraint: Return ONLY raw, valid JSON. Do not wrap the output in markdown code blocks (do not use \`\`\`json). Do not include any conversational text, explanations, or formatting outside the JSON object.

JSON Schema Definition:
{
  "period_month_num": "integer (1-12)",
  "period_year": "integer (YYYY)",
  "employee_name": "string",
  "tax_class": "string",
  "tax_id": "string",
  "gross_pay": "float",
  "net_pay": "float",
  "payout_amount": "float",
  "custom_deductions": "float or null",
  "bonuses": [{"description": "string", "amount": "float"}]
}

Source Text:
{{TEXT}}`;

// Helper component for the expandable prompt accordions
const PromptAccordion = ({ title, settingKey, defaultPrompt, value, onChange, t }: any) => {
    const [isOpen, setIsOpen] = useState(false);

    return (
        <div className="border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden bg-white dark:bg-gray-800">
            <button
                type="button"
                onClick={() => setIsOpen(!isOpen)}
                className="w-full flex items-center justify-between px-4 py-3 bg-gray-50 dark:bg-gray-800/50 hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors text-sm font-medium text-gray-900 dark:text-gray-100"
            >
                <div className="flex items-center gap-2">
                    <MessageSquareCode size={16} className="text-gray-500 dark:text-gray-400" />
                    {title}
                </div>
                {isOpen ? <ChevronDown size={18} className="text-gray-500" /> : <ChevronRight size={18} className="text-gray-500" />}
            </button>
            {isOpen && (
                <div className="p-4 border-t border-gray-200 dark:border-gray-700 space-y-2 animate-in slide-in-from-top-2 duration-200">
                    <div className="flex items-center justify-between">
                        <span className="text-xs text-gray-500 dark:text-gray-400">{t('settings.editPromptTemplate') || "Edit the system prompt template below."}</span>
                        <button
                            type="button"
                            onClick={() => onChange(settingKey, defaultPrompt)}
                            className="text-xs text-indigo-600 dark:text-indigo-400 hover:underline"
                        >
                            {t('settings.insertDefault')}
                        </button>
                    </div>
                    <textarea
                        rows={8}
                        value={value || ''}
                        onChange={(e) => onChange(settingKey, e.target.value)}
                        placeholder={defaultPrompt}
                        className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono whitespace-pre-wrap"
                    />
                </div>
            )}
        </div>
    );
};

export default function SettingsPage() {
    const { t, i18n } = useTranslation();
    const queryClient = useQueryClient();

    const [settingsParams, setSettingsParams] = useState<Record<string, string>>({});
    const [settingsSuccess, setSettingsSuccess] = useState(false);
    const [settingsError, setSettingsError] = useState('');

    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [pwdErrorMsg, setPwdErrorMsg] = useState('');

    const { data: sysInfo } = useQuery({
        queryKey: ['systemInfo'],
        queryFn: fetchSystemInfo,
    });

    const { data: currentSettings, isSuccess: settingsLoaded } = useQuery({
        queryKey: ['settings'],
        queryFn: fetchSettings,
    });

    useEffect(() => {
        if (settingsLoaded && currentSettings) {
            setSettingsParams(currentSettings);
        }
    }, [currentSettings, settingsLoaded]);

    const settingsMut = useMutation({
        mutationFn: () => updateSettings(settingsParams),
        onSuccess: () => {
            setSettingsSuccess(true);
            setSettingsError('');
            queryClient.invalidateQueries({ queryKey: ['settings'] });
            setTimeout(() => setSettingsSuccess(false), 3000);
        },
        onError: (err: any) => {
            setSettingsError(err.response?.data?.error || 'Failed to save settings.');
            setSettingsSuccess(false);
        }
    });

    const passwordMut = useMutation({
        mutationFn: () => changePassword(oldPassword, newPassword),
        onSuccess: () => {
            setOldPassword('');
            setNewPassword('');
            setConfirmPassword('');
            setPwdErrorMsg('');
        },
        onError: (err: any) => {
            setPwdErrorMsg(err.response?.data?.error || 'Failed to change password.');
        }
    });

    const handleSettingChange = (key: string, value: string) => {
        setSettingsParams(prev => ({ ...prev, [key]: value }));
    };

    const handleSettingsSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        settingsMut.mutate();
    };

    const handlePasswordSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        setPwdErrorMsg('');

        if (newPassword !== confirmPassword) {
            setPwdErrorMsg(t('settings.pwdMismatch'));
            return;
        }
        passwordMut.mutate();
    };

    return (
        <div className="max-w-4xl mx-auto space-y-6 animate-in fade-in duration-300">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center gap-2">
                <Settings size={28} className="text-indigo-600 dark:text-indigo-400" />
                {t('settings.title')}
            </h1>

            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
                    <Server size={20} className="text-gray-500 dark:text-gray-400" />
                    {t('settings.systemStatus')}
                </h2>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-800">
                        <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1">{t('settings.storageMode')}</p>
                        <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 capitalize">{sysInfo?.storage_mode || 'Loading...'}</p>
                    </div>
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-800">
                        <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1">{t('settings.dbHost')}</p>
                        <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 font-mono">{sysInfo?.db_host || 'Loading...'}</p>
                    </div>
                    <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-800">
                        <p className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase mb-1 flex items-center gap-1">
                            <Database size={14} /> {t('settings.connState')}
                        </p>
                        <div className="flex items-center gap-2">
                            <div className={`w-2 h-2 rounded-full ${sysInfo?.db_state === 'connected' ? 'bg-green-500' : 'bg-red-500'}`}></div>
                            <p className="text-sm font-semibold text-gray-900 dark:text-gray-100 capitalize">{sysInfo?.db_state || 'Checking...'}</p>
                        </div>
                    </div>
                </div>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
                    <Settings size={20} className="text-gray-500 dark:text-gray-400" />
                    {t('settings.appConfig')}
                </h2>

                {settingsSuccess && (
                    <div className="mb-6 flex items-center gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/50 rounded-xl text-green-700 dark:text-green-400 text-sm">
                        <CheckCircle2 size={16} />
                        {t('settings.successSaved')}
                    </div>
                )}
                {settingsError && (
                    <div className="mb-6 flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm">
                        <AlertCircle size={16} />
                        {settingsError}
                    </div>
                )}

                <form onSubmit={handleSettingsSubmit} className="space-y-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div className="space-y-4">
                            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 border-b dark:border-gray-800 pb-2">{t('settings.llmConfig')}</h3>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.apiUrl')}</label>
                                <input
                                    type="text"
                                    value={settingsParams['llm_api_url'] || ''}
                                    onChange={(e) => handleSettingChange('llm_api_url', e.target.value)}
                                    placeholder="https://generativelanguage.googleapis.com or http://local-llm:11434"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.apiToken')}</label>
                                <input
                                    type="password"
                                    value={settingsParams['llm_api_token'] || ''}
                                    onChange={(e) => handleSettingChange('llm_api_token', e.target.value)}
                                    placeholder="Bearer token if required"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.modelName')}</label>
                                <input
                                    type="text"
                                    value={settingsParams['llm_model'] || ''}
                                    onChange={(e) => handleSettingChange('llm_model', e.target.value)}
                                    placeholder="llama3"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>

                            <div className="pt-4 space-y-3">
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('settings.promptEngineering') || "Prompt Engineering"}</label>

                                <PromptAccordion
                                    title={t('settings.singlePrompt') || "Single Invoice Prompt"}
                                    settingKey="llm_single_prompt"
                                    defaultPrompt={DEFAULT_SINGLE_PROMPT}
                                    value={settingsParams['llm_single_prompt']}
                                    onChange={handleSettingChange}
                                    t={t}
                                />

                                <PromptAccordion
                                    title={t('settings.batchPrompt') || "Batch Transaction Prompt"}
                                    settingKey="llm_batch_prompt"
                                    defaultPrompt={DEFAULT_BATCH_PROMPT}
                                    value={settingsParams['llm_batch_prompt']}
                                    onChange={handleSettingChange}
                                    t={t}
                                />

                                <PromptAccordion
                                    title={t('settings.payslipPrompt') || "Payslip Extraction Prompt"}
                                    settingKey="llm_payslip_prompt"
                                    defaultPrompt={DEFAULT_PAYSLIP_PROMPT}
                                    value={settingsParams['llm_payslip_prompt']}
                                    onChange={handleSettingChange}
                                    t={t}
                                />
                            </div>
                        </div>

                        <div className="space-y-4">
                            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 border-b dark:border-gray-800 pb-2">{t('settings.bgImport')}</h3>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.importDir')}</label>
                                <input
                                    type="text"
                                    value={settingsParams['import_dir'] || ''}
                                    onChange={(e) => handleSettingChange('import_dir', e.target.value)}
                                    placeholder="/path/to/watch"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.pollInterval')}</label>
                                <input
                                    type="text"
                                    value={settingsParams['import_interval'] || ''}
                                    onChange={(e) => handleSettingChange('import_interval', e.target.value)}
                                    placeholder="1h, 30m, 12h"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>

                            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 border-b dark:border-gray-800 pb-2 mt-6 pt-4">{t('settings.autoCat')}</h3>

                            <div className="flex items-center justify-between py-2">
                                <div>
                                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">{t('settings.enableAutoCat')}</label>
                                    <p className="text-xs text-gray-500 dark:text-gray-400">{t('settings.enableAutoCatDesc')}</p>
                                </div>
                                <label className="relative inline-flex items-center cursor-pointer">
                                    <input
                                        type="checkbox"
                                        className="sr-only peer"
                                        checked={settingsParams['auto_categorization_enabled'] === 'true'}
                                        onChange={(e) => handleSettingChange('auto_categorization_enabled', e.target.checked ? 'true' : 'false')}
                                    />
                                    <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-indigo-600"></div>
                                </label>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.pollInterval')}</label>
                                <input
                                    type="text"
                                    value={settingsParams['auto_categorization_interval'] || ''}
                                    onChange={(e) => handleSettingChange('auto_categorization_interval', e.target.value)}
                                    placeholder="5m, 1h"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.batchSize')}</label>
                                <input
                                    type="number"
                                    min="1"
                                    max="100"
                                    value={settingsParams['auto_categorization_batch_size'] || ''}
                                    onChange={(e) => handleSettingChange('auto_categorization_batch_size', e.target.value)}
                                    placeholder="10"
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 font-mono"
                                />
                            </div>

                            <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 border-b dark:border-gray-800 pb-2 mt-6 pt-4">{t('settings.uiPrefs')}</h3>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 flex items-center gap-2">
                                    <Globe size={16} className="text-gray-400"/> {t('settings.language')}
                                </label>
                                <select
                                    value={settingsParams['ui_language'] || i18n.language || 'en'}
                                    onChange={(e) => {
                                        handleSettingChange('ui_language', e.target.value);
                                        i18n.changeLanguage(e.target.value);
                                    }}
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                                >
                                    <option value="en">English</option>
                                    <option value="de">German</option>
                                    <option value="es">Spanish</option>
                                    <option value="fr">French</option>
                                </select>
                            </div>

                            <div className="pt-2">
                                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1 flex items-center gap-2">
                                    <Palette size={16} className="text-gray-400"/> {t('settings.theme')}
                                </label>
                                <select
                                    value={settingsParams['theme'] || 'system'}
                                    onChange={(e) => handleSettingChange('theme', e.target.value)}
                                    className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                                >
                                    <option value="system">{t('settings.themeSystem')}</option>
                                    <option value="light">{t('settings.themeLight')}</option>
                                    <option value="dark">{t('settings.themeDark')}</option>
                                </select>
                            </div>
                        </div>
                    </div>

                    <div className="pt-2 flex justify-end">
                        <button
                            type="submit"
                            disabled={settingsMut.isPending}
                            className="flex items-center gap-2 px-6 py-2.5 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-sm font-medium rounded-xl hover:bg-gray-800 dark:hover:bg-white disabled:opacity-50 transition-all"
                        >
                            <Save size={16} />
                            {settingsMut.isPending ? t('settings.saving') : t('settings.saveConfig')}
                        </button>
                    </div>
                </form>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm p-6">
                <h2 className="text-lg font-medium text-gray-900 dark:text-gray-100 mb-4 flex items-center gap-2">
                    <KeyRound size={20} className="text-gray-500 dark:text-gray-400" />
                    {t('settings.security')}
                </h2>

                {passwordMut.isSuccess && (
                    <div className="mb-6 flex items-center gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800/50 rounded-xl text-green-700 dark:text-green-400 text-sm">
                        <CheckCircle2 size={16} />
                        {t('settings.pwdSuccess')}
                    </div>
                )}
                {pwdErrorMsg && (
                    <div className="mb-6 flex items-center gap-2 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm">
                        <AlertCircle size={16} />
                        {pwdErrorMsg}
                    </div>
                )}

                <form onSubmit={handlePasswordSubmit} className="space-y-4 max-w-md">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.currentPwd')}</label>
                        <input
                            type="password"
                            required
                            value={oldPassword}
                            onChange={(e) => setOldPassword(e.target.value)}
                            className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                        />
                    </div>
                    <div className="pt-2">
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.newPwd')}</label>
                        <input
                            type="password"
                            required
                            value={newPassword}
                            onChange={(e) => setNewPassword(e.target.value)}
                            className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t('settings.confirmPwd')}</label>
                        <input
                            type="password"
                            required
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            className="w-full px-4 py-2.5 text-sm rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                        />
                    </div>

                    <div className="pt-4">
                        <button
                            type="submit"
                            disabled={passwordMut.isPending || !oldPassword || !newPassword || !confirmPassword}
                            className="px-6 py-2.5 bg-indigo-600 dark:bg-indigo-500 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 dark:hover:bg-indigo-600 disabled:opacity-50 transition-all"
                        >
                            {passwordMut.isPending ? t('settings.updating') : t('settings.updatePwd')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}