import React, { useState } from 'react';
import { Plus, Trash2, Edit2, Check, Eye, EyeOff, Bot, Power } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export interface LLMProfile {
    id: string;
    name: string;
    type: 'gemini' | 'ollama' | 'openai';
    url: string;
    token: string;
    model: string;
    is_active: boolean;
}

interface LLMProfileManagerProps {
    profiles: LLMProfile[];
    onChange: (profiles: LLMProfile[]) => void;
    title?: string;
}

const MASKED_TOKEN = '********';

export const LLMProfileManager: React.FC<LLMProfileManagerProps> = ({ profiles, onChange, title }) => {
    const { t } = useTranslation();
    const [isAdding, setIsAdding] = useState(false);
    const [editingId, setEditingId] = useState<string | null>(null);
    const [showToken, setShowToken] = useState<Record<string, boolean>>({});

    const [editForm, setEditForm] = useState<LLMProfile>({
        id: '',
        name: '',
        type: 'openai',
        url: '',
        token: '',
        model: '',
        is_active: true
    });

    const handleAdd = () => {
        setEditForm({
            id: crypto.randomUUID(),
            name: '',
            type: 'openai',
            url: '',
            token: '',
            model: '',
            is_active: profiles.length === 0 // Default to active if it's the first one
        });
        setIsAdding(true);
        setEditingId(null);
    };

    const handleEdit = (profile: LLMProfile) => {
        setEditForm({ ...profile, token: MASKED_TOKEN });
        setEditingId(profile.id);
        setIsAdding(false);
    };

    const handleSave = () => {
        if (!editForm.name || !editForm.model) return;

        let updatedProfiles: LLMProfile[];
        
        const finalForm = { ...editForm };
        
        if (isAdding) {
            updatedProfiles = [...profiles, finalForm];
        } else {
            updatedProfiles = profiles.map(p => {
                if (p.id === editingId) {
                    const tokenToSave = finalForm.token === MASKED_TOKEN ? p.token : finalForm.token;
                    return { ...finalForm, token: tokenToSave };
                }
                return p;
            });
        }

        // If the saved profile is active, deactivate others
        if (finalForm.is_active) {
            updatedProfiles = updatedProfiles.map(p => ({
                ...p,
                is_active: p.id === (isAdding ? finalForm.id : editingId)
            }));
        } else if (updatedProfiles.every(p => !p.is_active) && updatedProfiles.length > 0) {
            // If none are active after save, make the saved one (or the first one) active
            // Actually, if we just saved it as inactive, maybe we should respect that,
            // but the user's issue was none being active.
            // Let's force at least one active if possible.
            updatedProfiles[0].is_active = true;
        }

        onChange(updatedProfiles);
        setIsAdding(false);
        setEditingId(null);
    };

    const handleDelete = (id: string) => {
        if (window.confirm(t('settings.llm.confirmDelete'))) {
            const newProfiles = profiles.filter(p => p.id !== id);
            // Ensure at least one is active if list is not empty
            if (newProfiles.length > 0 && newProfiles.every(p => !p.is_active)) {
                newProfiles[0].is_active = true;
            }
            onChange(newProfiles);
        }
    };

    const toggleActive = (id: string) => {
        const updatedProfiles = profiles.map(p => ({
            ...p,
            is_active: p.id === id ? true : false // Clicking one makes it active, others inactive
        }));
        onChange(updatedProfiles);
    };

    const toggleTokenVisibility = (id: string) => {
        setShowToken(prev => ({ ...prev, [id]: !prev[id] }));
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                    <Bot size={18} className="text-indigo-500" />
                    {title || t('settings.llm.profiles')}
                </h3>
                {!isAdding && !editingId && (
                    <button
                        type="button"
                        onClick={handleAdd}
                        className="flex items-center gap-1 text-xs font-bold text-indigo-600 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 px-2 py-1 rounded-lg transition-colors"
                    >
                        <Plus size={14} />
                        {t('settings.llm.addProfile')}
                    </button>
                )}
            </div>

            {(isAdding || editingId) && (
                <div className="p-4 bg-gray-50 dark:bg-gray-800/50 border border-indigo-100 dark:border-indigo-900/50 rounded-xl space-y-4 animate-in fade-in slide-in-from-top-2 duration-200">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div className="space-y-1">
                            <label className="text-[11px] font-bold text-gray-500 uppercase">{t('settings.llm.profileName')}</label>
                            <input
                                type="text"
                                value={editForm.name}
                                onChange={e => setEditForm({ ...editForm, name: e.target.value })}
                                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900"
                                placeholder="e.g. My Gemini Pro"
                            />
                        </div>
                        <div className="space-y-1">
                            <label className="text-[11px] font-bold text-gray-500 uppercase">{t('settings.llm.type')}</label>
                            <select
                                value={editForm.type}
                                onChange={e => setEditForm({ ...editForm, type: e.target.value as any })}
                                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900"
                            >
                                <option value="gemini">Google Gemini</option>
                                <option value="ollama">Ollama (Local)</option>
                                <option value="openai">Universal (OpenAI-compatible)</option>
                            </select>
                        </div>
                        <div className="space-y-1 md:col-span-2">
                            <label className="text-[11px] font-bold text-gray-500 uppercase">{t('settings.llm.apiUrl')}</label>
                            <input
                                type="text"
                                value={editForm.url}
                                onChange={e => setEditForm({ ...editForm, url: e.target.value })}
                                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 font-mono"
                                placeholder={editForm.type === 'ollama' ? 'http://localhost:11434' : 'https://api.openai.com/v1'}
                            />
                        </div>
                        <div className="space-y-1">
                            <label className="text-[11px] font-bold text-gray-500 uppercase">{t('settings.llm.apiToken')}</label>
                            <div className="relative">
                                <input
                                    type={showToken['edit'] ? 'text' : 'password'}
                                    value={editForm.token}
                                    onChange={e => setEditForm({ ...editForm, token: e.target.value })}
                                    className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => toggleTokenVisibility('edit')}
                                    className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                                >
                                    {showToken['edit'] ? <EyeOff size={16} /> : <Eye size={16} />}
                                </button>
                            </div>
                        </div>
                        <div className="space-y-1">
                            <label className="text-[11px] font-bold text-gray-500 uppercase">{t('settings.llm.model')}</label>
                            <input
                                type="text"
                                value={editForm.model}
                                onChange={e => setEditForm({ ...editForm, model: e.target.value })}
                                className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 font-mono"
                                placeholder="e.g. gpt-4o, gemini-1.5-pro, llama3"
                            />
                        </div>
                    </div>

                    <div className="flex items-center justify-between pt-2 border-t border-gray-100 dark:border-gray-800">
                        <label className="flex items-center gap-2 cursor-pointer">
                            <input
                                type="checkbox"
                                checked={editForm.is_active}
                                onChange={e => setEditForm({ ...editForm, is_active: e.target.checked })}
                                className="rounded text-indigo-600 focus:ring-indigo-500"
                            />
                            <span className="text-xs font-medium text-gray-700 dark:text-gray-300">{t('settings.llm.activeProfile')}</span>
                        </label>
                        <div className="flex gap-2">
                            <button
                                type="button"
                                onClick={() => { setIsAdding(false); setEditingId(null); }}
                                className="px-3 py-1.5 text-xs font-bold text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                            >
                                {t('common.cancel')}
                            </button>
                            <button
                                type="button"
                                onClick={handleSave}
                                className="px-3 py-1.5 text-xs font-bold bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center gap-1"
                            >
                                <Check size={14} />
                                {t('common.save')}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            <div className="grid grid-cols-1 gap-3">
                {profiles.length === 0 && !isAdding && (
                    <div className="text-center py-6 bg-gray-50 dark:bg-gray-800/30 rounded-xl border border-dashed border-gray-200 dark:border-gray-800 text-gray-400 text-sm italic">
                        {t('settings.llm.noProfiles')}
                    </div>
                )}
                {profiles.map(profile => (
                    <div
                        key={profile.id}
                        className={`group p-3 rounded-xl border transition-all ${
                            profile.is_active
                                ? 'bg-indigo-50/50 dark:bg-indigo-900/10 border-indigo-200 dark:border-indigo-800/50'
                                : 'bg-white dark:bg-gray-900 border-gray-100 dark:border-gray-800 hover:border-gray-200 dark:hover:border-gray-700'
                        }`}
                    >
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3 min-w-0 flex-1">
                                <div className={`w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center ${
                                    profile.is_active ? 'bg-indigo-600 text-white shadow-md shadow-indigo-200 dark:shadow-none' : 'bg-gray-100 dark:bg-gray-800 text-gray-400'
                                }`}>
                                    <Bot size={20} />
                                </div>
                                <div className="min-w-0 flex-1">
                                    <div className="flex flex-wrap items-center gap-2">
                                        <p className="text-sm font-bold text-gray-900 dark:text-gray-100 truncate">{profile.name}</p>
                                        <span className="px-1.5 py-0.5 text-[10px] font-black uppercase tracking-wider bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded">
                                            {profile.type}
                                        </span>
                                        {profile.is_active && (
                                            <span className="flex items-center gap-0.5 text-[10px] font-bold text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900/20 px-1.5 py-0.5 rounded-full">
                                                <Power size={10} /> {t('settings.llm.active')}
                                            </span>
                                        )}
                                    </div>
                                    <p className="text-[11px] text-gray-500 dark:text-gray-400 font-mono mt-0.5 truncate">
                                        {profile.model} • {profile.url || 'Default URL'}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center gap-1 md:opacity-0 md:group-hover:opacity-100 transition-opacity flex-shrink-0 ml-2">
                                <button
                                    type="button"
                                    onClick={() => toggleActive(profile.id)}
                                    className={`p-1.5 rounded-lg transition-colors ${profile.is_active ? 'text-indigo-600 bg-indigo-50 dark:bg-indigo-900/30' : 'text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}`}
                                    title={profile.is_active ? t('settings.llm.deactivate') : t('settings.llm.activate')}
                                >
                                    <Power size={16} />
                                </button>
                                <button
                                    type="button"
                                    onClick={() => handleEdit(profile)}
                                    className="p-1.5 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-colors"
                                    title={t('common.edit')}
                                >
                                    <Edit2 size={16} />
                                </button>
                                <button
                                    type="button"
                                    onClick={() => handleDelete(profile.id)}
                                    className="p-1.5 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                                    title={t('common.delete')}
                                >
                                    <Trash2 size={16} />
                                </button>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
