import { api } from '../client';
import type { Category } from "../types/category";
export const categoryService = {
    fetchCategories: (): Promise<Category[]> =>
        api.get<Category[]>('categories/').then(r => r.data ?? []),

    fetchAverage: (id: string, strategy: string): Promise<{ average: number }> =>
        api.get<{ average: number }>(`categories/${id}/average/`, { params: { strategy } }).then(r => r.data),

    create: (name: string, color: string, isVariableSpending: boolean = false, forecastStrategy: string = '3y'): Promise<Category> =>
        api.post<Category>('categories/', { name, color, is_variable_spending: isVariableSpending, forecast_strategy: forecastStrategy }).then(r => r.data),

    update: (id: string, name: string, color: string, isVariableSpending: boolean = false, forecastStrategy?: string): Promise<Category> =>
        api.put<Category>(`categories/${id}/`, { name, color, is_variable_spending: isVariableSpending, forecast_strategy: forecastStrategy }).then(r => r.data),

    delete: (id: string): Promise<void> =>
        api.delete(`categories/${id}/`).then(() => undefined),

    restore: (id: string): Promise<Category> =>
        api.post<Category>(`categories/${id}/restore/`).then(r => r.data),

    share: (id: string, userId: string, permission: 'view' | 'edit' = 'view'): Promise<void> =>
        api.post(`categories/${id}/share/`, { user_id: userId, permission }).then(() => undefined),

    revokeShare: (id: string, userId: string): Promise<void> =>
        api.delete(`categories/${id}/share/${userId}/`).then(() => undefined),

    fetchShares: (id: string): Promise<string[]> =>
        api.get<string[]>(`categories/${id}/shares/`).then(r => r.data ?? [])
};