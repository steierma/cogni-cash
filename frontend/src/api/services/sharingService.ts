import { api } from '../client';
import type { SharingDashboard } from "../types/category";
export const sharingService = {
    fetchSharingDashboard: (): Promise<SharingDashboard> =>
        api.get<SharingDashboard>('sharing/dashboard/').then(r => r.data)
};
