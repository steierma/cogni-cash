import { getClientId } from './clientId';

export const mergeEffectiveSettings = (rawSettings: Record<string, string>): Record<string, string> => {
    const clientId = getClientId();
    const prefix = `client.${clientId}.`;
    const merged: Record<string, string> = { ...rawSettings };

    Object.keys(rawSettings).forEach((key) => {
        if (key.startsWith(prefix)) {
            const originalKey = key.replace(prefix, '');
            merged[originalKey] = rawSettings[key];
        }
    });

    return merged;
};

export const getNamespacedKey = (key: string, isDeviceSpecific: boolean): string => {
    if (!isDeviceSpecific) return key;
    const clientId = getClientId();
    return `client.${clientId}.${key}`;
};
