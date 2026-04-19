import type { Payslip } from "../../api/types/payslip";
export type ColKey = 'period' | 'employer' | 'gross' | 'net' | 'adjNet' | 'payout' | 'leasing';
export type SortDirection = 'asc' | 'desc';

export const formatYearMonth = (year: number, monthNum: number) => {
    return `${year}-${String(monthNum || 0).padStart(2, '0')}`;
};

export const getAdjustedNetto = (payslip: Payslip, excludedBonuses: Set<string>, excludeLeasing: boolean, useProportional: boolean) => {
    let adjusted = payslip.net_pay;

    const excludedBonusTotal = payslip.bonuses?.reduce((sum, bonus) => {
        return excludedBonuses.has(bonus.description) ? sum + Number(bonus.amount) : sum;
    }, 0) || 0;

    if (excludedBonusTotal > 0) {
        if (useProportional && payslip.gross_pay > 0) {
            const baseGross = payslip.gross_pay - excludedBonusTotal;
            const ratio = baseGross / payslip.gross_pay;
            adjusted = payslip.net_pay * ratio;
        } else {
            adjusted -= excludedBonusTotal;
        }
    }

    if (excludeLeasing && payslip.custom_deductions) {
        adjusted += Math.abs(payslip.custom_deductions);
    }

    return adjusted;
};