import { useQuery } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { categoryService } from '../api/services/categoryService';
import type { Category } from "../api/types/category";
const FALLBACK: Record<string, string> = {
    'Haus und Hausrat':                           '#f59e0b',
    'Bildung, Gesundheit, Beauty und Wellness':   '#ec4899',
    'Restaurants':                                '#f97316',
    'Handel und Geschäfte':                       '#8b5cf6',
    'Sonstige Ausgaben':                          '#6b7280',
    'Einkommen':                                  '#22c55e',
    'Berufliche Ausgaben':                        '#3b82f6',
    'Freizeit und Reisen':                        '#0ea5e9',
    'Transport und Automobil':                    '#f59e0b',
    'Überweisungen, Bankkosten und Darlehen':     '#6366f1',
};

export default function CategoryBadge({ category }: { category?: string }) {
    const { t } = useTranslation();
    const { data: categories = [] } = useQuery<Category[]>({
        queryKey: ['categories'],
        queryFn: categoryService.fetchCategories,
        staleTime: 5 * 60 * 1000,
    });

    if (!category) {
        return (
            <span
                className="text-gray-300 text-xs"
                title={t('transactions.table.unset')}
            >
                —
            </span>
        );
    }

    // Allow passing either a category ID or a category Name. Prefer matching by ID,
    // fall back to matching by Name. If no match is found, treat the prop as the
    // display name directly.
    const match = categories.find((c) => c.id === category) ?? categories.find((c) => c.name === category);
    const displayName = match?.name ?? category;
    const color = match?.color ?? FALLBACK[displayName.trim()] ?? '#6366f1';

    return (
        <span
            className="inline-block px-2.5 py-0.5 rounded-full border text-xs font-medium max-w-[16rem] truncate"
            style={{ backgroundColor: color + '22', color, borderColor: color + '66' }}
            title={displayName}
        >
            {displayName}
        </span>
    );
}