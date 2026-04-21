import { api } from '../client';
import type { SystemInfo, CreateBridgeTokenResponse } from "../types/system";
export const settingsService = {
    fetchSystemInfo: (): Promise<SystemInfo> =>
        api.get<SystemInfo>('system/info/').then(r => r.data),

    fetchSettings: (): Promise<Record<string, string>> =>
        api.get<Record<string, string>>('settings/').then(r => r.data ?? {}),

    updateSettings: (settings: Record<string, string>): Promise<void> =>
        api.patch('settings/', settings).then(() => undefined),

    sendTestEmail: (to: string): Promise<void> =>
        api.post('settings/test-email/', { to }).then(() => undefined),

    // Log Level
    fetchLogLevel: (): Promise<{ level: string }> =>
        api.get<{ level: string }>('system/log-level/').then(r => r.data),

    updateLogLevel: (level: string): Promise<{ level: string }> =>
        api.put<{ level: string }>('system/log-level/', { level }).then(r => r.data),

    // Bridge Tokens
    fetchBridgeTokens: (): Promise<any[]> =>
        api.get<any[]>('bridge-tokens/').then(r => r.data ?? []),

    createBridgeToken: (name: string): Promise<CreateBridgeTokenResponse> =>
        api.post<CreateBridgeTokenResponse>('bridge-tokens/', { name }).then(r => r.data),

    revokeBridgeToken: (id: string): Promise<void> =>
        api.delete(`bridge-tokens/${id}/`).then(() => undefined)
};