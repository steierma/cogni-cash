import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import type { AuthResponse } from './types';

export const api = axios.create({
    baseURL: '/api/v1',
    withCredentials: true
});

api.interceptors.response.use(
    (response: AxiosResponse) => response,
    async (error: unknown) => {
        if (axios.isAxiosError(error) && error.response?.status === 401) {
            const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };
            if (originalRequest && !originalRequest._retry && window.location.pathname !== '/login') {
                originalRequest._retry = true;
                const refreshToken = localStorage.getItem('refresh_token');
                if (refreshToken) {
                    try {
                        const res = await axios.post<AuthResponse>('/api/v1/auth/refresh/', { refresh_token: refreshToken });
                        localStorage.setItem('refresh_token', res.data.refresh_token);
                        return api(originalRequest);
                    } catch {
                        localStorage.removeItem('refresh_token');
                        window.location.href = '/login';
                    }
                } else {
                    window.location.href = '/login';
                }
            }
        }
        return Promise.reject(error);
    }
);

export const fetchHealth = (): Promise<{ status: string }> =>
    axios.get<{ status: string }>('/health').then((r: AxiosResponse<{ status: string }>) => r.data);