import { api } from '../client';
import type { User } from "../types/system";
import type { AxiosResponse } from 'axios';

export const userService = {
    fetchUsers: (search?: string): Promise<User[]> =>
        api.get<User[]>('users/', { params: { q: search } }).then((r: AxiosResponse<User[]>) => r.data ?? []),

    fetchUser: (id: string): Promise<User> =>
        api.get<User>(`users/${id}/`).then((r: AxiosResponse<User>) => r.data),

    createUser: (data: Partial<User> & { password?: string }): Promise<User> =>
        api.post<User>('users/', data).then((r: AxiosResponse<User>) => r.data),

    updateUser: (id: string, data: Partial<User>): Promise<User> =>
        api.put<User>(`users/${id}/`, data).then((r: AxiosResponse<User>) => r.data),

    deleteUser: (id: string): Promise<void> =>
        api.delete(`users/${id}/`).then(() => undefined)
};