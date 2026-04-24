import i18n from '../i18n';

export function fmtCurrency(amount: number, currency = 'EUR') {
    // Normalize and validate currency code. Intl.NumberFormat will throw
    // a RangeError for invalid/empty currency codes.
    const safeCurrency = (currency || 'EUR').toString().toUpperCase();
    const currencyCode = /^[A-Z]{3}$/.test(safeCurrency) ? safeCurrency : 'EUR';

    const safeAmount = Number(amount ?? 0);
    const locale = i18n.language || 'en-US';

    try {
        return new Intl.NumberFormat(locale, { style: 'currency', currency: currencyCode }).format(safeAmount);
    } catch {
        // As a last resort, fall back to a plain numeric format plus currency code.
        try {
            return safeAmount.toLocaleString(locale, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' ' + currencyCode;
        } catch {
            return safeAmount.toFixed(2) + ' ' + currencyCode;
        }
    }
}

export function fmtDate(iso: string, format: 'short' | 'long' = 'long') {
    if (!iso) return '—';
    const options: Intl.DateTimeFormatOptions = format === 'short'
        ? { day: '2-digit', month: 'short', year: 'numeric' }
        : { day: '2-digit', month: '2-digit', year: 'numeric' };

    const locale = i18n.language || 'en-US';
    return new Date(iso).toLocaleDateString(locale, options);
}

/** Returns YYYY-MM-DD for a given date in LOCAL time */
export function getLocalISODate(date: Date = new Date()): string {
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
}
