export interface Bonus {
    description: string;
    amount: number;
}

export interface Payslip {
    id: string;
    original_file_name: string;
    period_month_num: number;
    period_year: number;
    employer_name: string;
    tax_class: string;
    tax_id: string;
    gross_pay: number;
    net_pay: number;
    payout_amount: number;
    custom_deductions: number;
    bonuses: Bonus[];
    created_at: string;
}

export interface PayslipTrend {
    period: string;
    gross: number;
    net: number;
}

export interface PayslipSummary {
    total_gross: number;
    total_net: number;
    total_payout: number;
    total_bonuses: number;
    payslip_count: number;
    latest_net_pay: number;
    net_pay_trend: number;
    latest_period: string;
    trends: PayslipTrend[];
}