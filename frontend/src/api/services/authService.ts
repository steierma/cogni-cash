import { api } from '../client';
import type { AuthResponse, User } from "../types/system";
import type { AxiosResponse } from 'axios';

export const authService = {
    login: (username: string, password: string): Promise<AuthResponse> =>
        api.post<AuthResponse>('login/', { username, password }).then((r: AxiosResponse<AuthResponse>) => {
            if (r.data.refresh_token) localStorage.setItem('refresh_token', r.data.refresh_token);
            return r.data;
        }),

    logout: (): Promise<void> => {
        const refreshToken = localStorage.getItem('refresh_token');
        localStorage.removeItem('refresh_token');
        return api.post('logout/', { refresh_token: refreshToken }).then(() => undefined);
    },

    changePassword: (oldPassword: string, newPassword: string): Promise<void> =>
        api.post('auth/change-password/', { old_password: oldPassword, new_password: newPassword }).then(() => undefined),

    requestPasswordReset: (email: string): Promise<{ message: string }> =>
        api.post('auth/forgot-password/', { email }).then(r => r.data),

    validateResetToken: (token: string): Promise<{ valid: boolean }> =>
        api.get('auth/reset-password/validate/', { params: { token } }).then(r => r.data),

    confirmPasswordReset: (token: string, newPassword: string): Promise<{ message: string }> =>
        api.post('auth/reset-password/confirm/', { token, new_password: newPassword }).then(r => r.data),

    fetchMe: (): Promise<User> =>
        api.get<User>('auth/me/').then((r: AxiosResponse<User>) => r.data)
};