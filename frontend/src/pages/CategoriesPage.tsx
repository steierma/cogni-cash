import { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Check, Pencil, Plus, Trash2, X, Tags, RotateCcw, Eye, EyeOff, ArrowRight, Users, Loader2 } from 'lucide-react';
import { categoryService } from '../api/services/categoryService';
import type { Category } from "../api/types/category";
import ShareCategoryModal from '../components/ShareCategoryModal';
import { fmtCurrency } from '../utils/formatters';

const PALETTE = [
    '#6366f1', '#8b5cf6', '#ec4899', '#f97316', '#f59e0b',
    '#22c55e', '#14b8a6', '#0ea5e9', '#3b82f6', '#6b7280',
    '#ef4444', '#84cc16',
];

function ColorPicker({ value, onChange }: { value: string; onChange: (c: string) => void }) {
    return (
        <div className="flex items-center gap-2 flex-wrap">
            {PALETTE.map((c) => (
                <button
                    key={c}
                    type="button"
                    onClick={() => onChange(c)}
                    className={`w-6 h-6 rounded-full border-2 transition-transform hover:scale-110 ${
                        value === c ? 'border-gray-900 dark:border-white scale-110' : 'border-transparent'
                    }`}
                    style={{ backgroundColor: c }}
                />
            ))}
            <input
                type="color"
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className="w-7 h-7 p-0 border-0 rounded cursor-pointer"
                title="Custom color"
            />
        </div>
    );
}

function StrategyPreview({ categoryId, strategy }: { categoryId?: string; strategy: string }) {
    const { t } = useTranslation();
    const [debouncedStrategy, setDebouncedStrategy] = useState(strategy);

    useEffect(() => {
        const timer = setTimeout(() => {
            if (/^(\d+[my]|all)$/.test(strategy)) {
                setDebouncedStrategy(strategy);
            }
        }, 500);
        return () => clearTimeout(timer);
    }, [strategy]);

    const { data, isLoading, isError } = useQuery({
        queryKey: ['category-average', categoryId, debouncedStrategy],
        queryFn: () => categoryService.fetchAverage(categoryId!, debouncedStrategy),
        enabled: !!categoryId && /^(\d+[my]|all)$/.test(debouncedStrategy),
        staleTime: 30000,
    });

    if (!/^(\d+[my]|all)$/.test(strategy)) return null;

    return (
        <div className="mt-1.5 flex items-center gap-2 min-h-[1.5rem]">
            <span className="text-[10px] font-bold text-indigo-600 dark:text-indigo-400 uppercase tracking-tight">
                {t('categories.predictionPreview')}:
            </span>
            {isLoading ? (
                <Loader2 size={12} className="animate-spin text-gray-400" />
            ) : isError ? (
                <span className="text-[10px] text-red-500">—</span>
            ) : (
                <span className="text-xs font-semibold text-gray-900 dark:text-gray-100">
                    {fmtCurrency(data?.average ?? 0)} / {t('forecasting.average').toLowerCase()}
                </span>
            )}
        </div>
    );
}

function CategoryRow({
                         cat,
                         onSaved,
                         onDelete,
                         onRestore,
                         onShare,
                     }: {
    cat: Category;
    onSaved: (id: string, name: string, color: string, isVariable: boolean, strategy: string) => void;
    onDelete: (id: string) => void;
    onRestore: (id: string) => void;
    onShare: (cat: Category) => void;
}) {
    const { t } = useTranslation();
    const [isEditing, setIsEditing] = useState(false);
    const [editName, setEditName] = useState(cat.name);
    const [editColor, setEditColor] = useState(cat.color);
    const [editIsVariable, setEditIsVariable] = useState(cat.is_variable_spending);
    const [editStrategy, setEditStrategy] = useState(cat.forecast_strategy || '3y');
    const isDeleted = !!cat.deleted_at;

    if (isEditing) {
        return (
            <tr className="bg-gray-50 dark:bg-gray-800/50">
                <td colSpan={3} className="px-4 py-4">
                    <div className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.name')}</label>
                                <input
                                    autoFocus
                                    type="text"
                                    value={editName}
                                    onChange={(e) => setEditName(e.target.value)}
                                    className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                                />
                            </div>
                            <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('categories.isVariableSpending')}</label>
                                <div className="flex items-center gap-3 py-2">
                                    <input
                                        type="checkbox"
                                        checked={editIsVariable}
                                        onChange={(e) => setEditIsVariable(e.target.checked)}
                                        className="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
                                    />
                                    <span className="text-sm text-gray-600 dark:text-gray-400">{t('categories.variableSpendingLabel')}</span>
                                </div>
                            </div>
                            {editIsVariable && (
                                <div>
                                    <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('categories.forecastStrategy')}</label>
                                    <div className="flex flex-col gap-1">
                                        <input
                                            type="text"
                                            value={editStrategy}
                                            onChange={(e) => setEditStrategy(e.target.value)}
                                            placeholder="e.g. 6m, 1y, all"
                                            className={`w-full px-3 py-2 text-sm rounded-lg border bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 ${
                                                !/^(\d+[my]|all)$/.test(editStrategy) ? 'border-red-500' : 'border-gray-200 dark:border-gray-700'
                                            }`}
                                        />
                                        <div className="flex flex-col">
                                            <p className="text-[10px] text-gray-400 leading-tight">
                                                {t('categories.forecastStrategyHelp')}
                                            </p>
                                            <StrategyPreview categoryId={cat.id} strategy={editStrategy} />
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">{t('common.color')}</label>
                            <ColorPicker value={editColor} onChange={setEditColor} />
                        </div>
                        <div className="flex items-center gap-2 pt-2">
                            <button
                                onClick={() => {
                                    onSaved(cat.id, editName, editColor, editIsVariable, editStrategy);
                                    setIsEditing(false);
                                }}
                                className="flex items-center gap-1.5 px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg transition-colors"
                            >
                                <Check size={16} /> {t('common.save')}
                            </button>
                            <button
                                onClick={() => {
                                    setEditName(cat.name);
                                    setEditColor(cat.color);
                                    setEditIsVariable(cat.is_variable_spending);
                                    setEditStrategy(cat.forecast_strategy || '3y');
                                    setIsEditing(false);
                                }}
                                className="flex items-center gap-1.5 px-3 py-1.5 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 text-sm font-medium rounded-lg transition-colors"
                            >
                                <X size={16} /> {t('common.cancel')}
                            </button>
                        </div>
                    </div>
                </td>
            </tr>
        );
    }

    return (
        <tr className={`hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors group ${isDeleted ? 'opacity-50' : ''}`}>
            <td className="px-4 py-4 whitespace-nowrap">
                <div className="flex items-center gap-3">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: cat.color }} />
                    <div className="flex flex-col">
                        <span className={`font-medium ${isDeleted ? 'text-gray-400 dark:text-gray-600 line-through' : 'text-gray-900 dark:text-gray-100'}`}>
                            {cat.name}
                        </span>
                        {cat.is_variable_spending && (
                            <span className="text-[10px] text-indigo-500 font-bold uppercase tracking-tighter">
                                {t('categories.isVariableSpending')}
                            </span>
                        )}
                        {cat.is_shared && (
                            <span className="text-[10px] text-green-500 font-bold uppercase tracking-tighter flex items-center gap-0.5">
                                <Users size={10} /> {t('categories.shared')}
                            </span>
                        )}
                        {isDeleted && (
                            <span className="text-[10px] text-red-500 font-bold uppercase tracking-tighter">
                                {t('categories.deleted')}
                            </span>
                        )}
                    </div>
                </div>
            </td>
            <td className="px-4 py-4 whitespace-nowrap">
                <span
                    className={`px-2.5 py-1 rounded-md text-xs font-semibold tracking-wide ${isDeleted ? 'grayscale' : ''}`}
                    style={{ backgroundColor: `${cat.color}20`, color: cat.color }}
                >
                    {cat.name}
                </span>
            </td>
            <td className="px-4 py-4 text-right whitespace-nowrap">
                {!isDeleted ? (
                    <>
                        <Link
                            to={`/transactions?category=${cat.id}`}
                            className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-colors mr-2 inline-flex items-center"
                            title={t('common.viewTransactions')}
                        >
                            <ArrowRight size={16} />
                        </Link>
                        <button
                            onClick={() => onShare(cat)}
                            className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-green-600 dark:hover:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/30 rounded-lg transition-colors mr-2"
                            title={t('categories.share')}
                        >
                            <Users size={16} />
                        </button>
                        <button
                            onClick={() => setIsEditing(true)}
                            className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-indigo-600 dark:hover:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg transition-colors mr-2"
                            title={t('common.edit')}
                        >
                            <Pencil size={16} />
                        </button>
                        <button
                            onClick={() => onDelete(cat.id)}
                            className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors"
                            title={t('common.delete')}
                        >
                            <Trash2 size={16} />
                        </button>
                    </>
                ) : (
                    <button
                        onClick={() => onRestore(cat.id)}
                        className="p-1.5 text-gray-400 dark:text-gray-500 hover:text-green-600 dark:hover:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/30 rounded-lg transition-colors"
                        title={t('categories.restore')}
                    >
                        <RotateCcw size={16} />
                    </button>
                )}
            </td>
        </tr>
    );
}

export default function CategoriesPage() {
    const { t } = useTranslation();
    const qc = useQueryClient();
    const [isCreating, setIsCreating] = useState(false);
    const [newName, setNewName] = useState('');
    const [newColor, setNewColor] = useState(PALETTE[0]);
    const [newIsVariable, setNewIsVariable] = useState(false);
    const [newStrategy, setNewStrategy] = useState('3y');
    const [showDeleted, setShowDeleted] = useState(false);
    const [shareTarget, setShareTarget] = useState<Category | null>(null);

    const { data: categories = [], isLoading } = useQuery({
        queryKey: ['categories'],
        queryFn: () => categoryService.fetchCategories(),
    });

    const createMut = useMutation({
        mutationFn: (c: { name: string; color: string; isVariable: boolean; strategy: string }) => categoryService.create(c.name, c.color, c.isVariable, c.strategy),
        onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    });

    const updateMut = useMutation({
        mutationFn: (c: { id: string; name: string; color: string; isVariable: boolean; strategy: string }) => categoryService.update(c.id, c.name, c.color, c.isVariable, c.strategy),
        onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    });

    const deleteMut = useMutation({
        mutationFn: (id: string) => categoryService.delete(id),
        onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    });

    const restoreMut = useMutation({
        mutationFn: (id: string) => categoryService.restore(id),
        onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
    });

    const handleCreate = () => {
        if (!newName.trim()) return;
        createMut.mutate({ name: newName.trim(), color: newColor, isVariable: newIsVariable, strategy: newStrategy });
        setIsCreating(false);
        setNewName('');
        setNewColor(PALETTE[0]);
        setNewIsVariable(false);
        setNewStrategy('3y');
    };

    const sorted = categories
        .filter(c => showDeleted || !c.deleted_at)
        .sort((a, b) => {
            // Deleted items at the bottom if showing both
            if (!!a.deleted_at !== !!b.deleted_at) {
                return a.deleted_at ? 1 : -1;
            }
            return a.name.localeCompare(b.name);
        });

    return (
        <div className="max-w-7xl mx-auto space-y-6 pb-20 animate-in fade-in duration-300">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-gray-900 dark:text-gray-100 flex items-center gap-2">
                        <Tags className="text-indigo-600 dark:text-indigo-400" /> {t('categories.title')}
                    </h1>
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t('categories.subtitle')}</p>
                </div>
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setShowDeleted(!showDeleted)}
                        className={`flex items-center justify-center gap-2 px-3 py-2 text-sm font-medium rounded-xl transition-all border ${
                            showDeleted
                                ? 'bg-indigo-50 border-indigo-200 text-indigo-700 dark:bg-indigo-900/30 dark:border-indigo-800 dark:text-indigo-300'
                                : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50 dark:bg-gray-900 dark:border-gray-800 dark:text-gray-400 dark:hover:bg-gray-800'
                        }`}
                    >
                        {showDeleted ? <Eye size={16} /> : <EyeOff size={16} />}
                        <span className="hidden sm:inline">{t('categories.showDeleted')}</span>
                    </button>

                    {!isCreating && (
                        <button
                            onClick={() => setIsCreating(true)}
                            className="flex items-center justify-center gap-2 px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-xl hover:bg-indigo-700 transition-colors shadow-sm"
                        >
                            <Plus size={16} /> {t('categories.newCategory')}
                        </button>
                    )}
                </div>
            </div>

            {isCreating && (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-5 shadow-sm">
                    <h3 className="font-semibold text-gray-900 dark:text-gray-100 mb-4">{t('categories.createCategory')}</h3>
                    <div className="space-y-4 max-w-lg">
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('common.name')}</label>
                                <input
                                    autoFocus
                                    type="text"
                                    value={newName}
                                    onChange={(e) => setNewName(e.target.value)}
                                    placeholder={t('categories.namePlaceholder')}
                                    className="w-full px-3 py-2 text-sm rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300"
                                />
                            </div>
                            <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('categories.isVariableSpending')}</label>
                                <div className="flex items-center gap-3 py-2">
                                    <input
                                        type="checkbox"
                                        checked={newIsVariable}
                                        onChange={(e) => setNewIsVariable(e.target.checked)}
                                        className="w-4 h-4 text-indigo-600 border-gray-300 rounded focus:ring-indigo-500"
                                    />
                                    <span className="text-sm text-gray-600 dark:text-gray-400">{t('categories.variableSpendingLabel')}</span>
                                </div>
                            </div>
                            {newIsVariable && (
                                <div>
                                    <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">{t('categories.forecastStrategy')}</label>
                                    <div className="flex flex-col gap-1">
                                        <input
                                            type="text"
                                            value={newStrategy}
                                            onChange={(e) => setNewStrategy(e.target.value)}
                                            placeholder="e.g. 6m, 1y, all"
                                            className={`w-full px-3 py-2 text-sm rounded-lg border bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-300 ${
                                                !/^(\d+[my]|all)$/.test(newStrategy) ? 'border-red-500' : 'border-gray-200 dark:border-gray-700'
                                            }`}
                                        />
                                        <div className="flex flex-col">
                                            <p className="text-[10px] text-gray-400 leading-tight">
                                                {t('categories.forecastStrategyHelp')}
                                            </p>
                                            <StrategyPreview strategy={newStrategy} />
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                        <p className="text-[10px] text-gray-400 leading-tight">
                            {t('categories.variableSpendingHelp')}
                        </p>
                        <div>
                            <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">{t('common.color')}</label>
                            <ColorPicker value={newColor} onChange={setNewColor} />
                        </div>
                        <div className="flex items-center gap-2 pt-2">
                            <button
                                onClick={handleCreate}
                                disabled={!newName.trim() || createMut.isPending}
                                className="flex items-center gap-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors disabled:opacity-50"
                            >
                                <Check size={16} /> {t('common.create')}
                            </button>
                            <button
                                onClick={() => setIsCreating(false)}
                                className="flex items-center gap-1.5 px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 text-sm font-medium rounded-lg transition-colors"
                            >
                                <X size={16} /> {t('common.cancel')}
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {isLoading ? (
                <div className="h-32 bg-gray-100 dark:bg-gray-800 rounded-2xl animate-pulse" />
            ) : (
                <div className="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm overflow-hidden">
                    <table className="min-w-full divide-y divide-gray-100 dark:divide-gray-800/50 text-sm">
                        <thead className="bg-gray-50 dark:bg-gray-800/50 text-xs uppercase text-gray-400 dark:text-gray-500 tracking-wide">
                        <tr>
                            <th className="px-4 py-3 text-left">{t('common.name')}</th>
                            <th className="px-4 py-3 text-left">{t('categories.badgePreview')}</th>
                            <th className="px-4 py-3 text-right">{t('common.actions')}</th>
                        </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-50 dark:divide-gray-800/50">
                        {sorted.map((cat) => (
                            <CategoryRow
                                key={cat.id}
                                cat={cat}
                                onSaved={(id, name, color, isVariable, strategy) => updateMut.mutate({ id, name, color, isVariable, strategy })}
                                onDelete={(id) => {
                                    const targetCat = categories.find(c => c.id === id);
                                    if (targetCat && confirm(t('categories.deleteConfirm', { name: targetCat.name }))) {
                                        deleteMut.mutate(id);
                                    }
                                }}
                                onRestore={(id) => {
                                    const targetCat = categories.find(c => c.id === id);
                                    if (targetCat && confirm(t('categories.restoreConfirm', { name: targetCat.name }))) {
                                        restoreMut.mutate(id);
                                    }
                                }}
                                onShare={(c) => setShareTarget(c)}
                            />
                        ))}
                        </tbody>
                    </table>
                    <div className="px-4 py-3 border-t border-gray-100 dark:border-gray-800 text-xs text-gray-400 dark:text-gray-500 text-right">
                        {t('categories.count', { count: sorted.length })}
                    </div>
                </div>
            )}

            {(updateMut.isError || deleteMut.isError) && (
                <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800/50 rounded-xl text-red-700 dark:text-red-400 text-sm mt-4">
                    {t('common.errorUpdateDb')}
                </div>
            )}

            {shareTarget && (
                <ShareCategoryModal
                    category={shareTarget}
                    onClose={() => setShareTarget(null)}
                />
            )}
        </div>
    );
}