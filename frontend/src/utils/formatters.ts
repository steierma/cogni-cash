export function fmtCurrency(amount: number, currency = 'EUR') {
    // Normalize and validate currency code. Intl.NumberFormat will throw
    // a RangeError for invalid/empty currency codes (which is the cause of
    // the runtime error seen in the browser). Accept a 3-letter code and
    // fallback to 'EUR' otherwise.
    const safeCurrency = (currency || 'EUR').toString().toUpperCase();
    const currencyCode = /^[A-Z]{3}$/.test(safeCurrency) ? safeCurrency : 'EUR';

    const safeAmount = Number(amount ?? 0);

    try {
        return new Intl.NumberFormat('de-DE', { style: 'currency', currency: currencyCode }).format(safeAmount);
    } catch (e) {
        // As a last resort, fall back to a plain numeric format plus currency code.
        return safeAmount.toLocaleString('de-DE', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) + ' ' + currencyCode;
    }
}

export function fmtDate(iso: string, format: 'short' | 'long' = 'long') {
    if (!iso) return '—';
    const options: Intl.DateTimeFormatOptions = format === 'short'
        ? { day: '2-digit', month: 'short', year: 'numeric' }
        : { day: '2-digit', month: '2-digit', year: 'numeric' };

    return new Date(iso).toLocaleDateString('de-DE', options);
}