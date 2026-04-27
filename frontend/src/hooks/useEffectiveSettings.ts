import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { useMemo } from 'react';
import { settingsService } from '../api/services/settingsService';
import { mergeEffectiveSettings } from '../api/utils/settingsHelper';

export const useEffectiveSettings = (): UseQueryResult<Record<string, string>, Error> => {
    const query = useQuery<Record<string, string>, Error>({
        queryKey: ['settings'],
        queryFn: settingsService.fetchSettings,
    });

    const effectiveSettings = useMemo(() => {
        if (!query.data) return query.data;
        return mergeEffectiveSettings(query.data);
    }, [query.data]);

    return {
        ...query,
        data: effectiveSettings,
    } as UseQueryResult<Record<string, string>, Error>;
};
