import { useState, useMemo } from 'react';
import { useTranslation, Trans } from 'react-i18next';
import { LineChart as LineChartIcon, ChevronDown, BarChart3 } from 'lucide-react';
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ReferenceLine, BarChart, Bar, Legend } from 'recharts';
import { fmtCurrency } from '../../utils/formatters';
import { formatYearMonth, getAdjustedNetto } from './utils';
import type { Payslip } from '../../api/types';

const CustomChartTooltip = ({ active, payload, label, t }: any) => {
    if (active && payload && payload.length) {
        const data = payload[0].payload;
        const bonusesThisMonth = data.bonusesThisMonth || [];

        return (
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-3 rounded-xl shadow-lg text-sm z-50 min-w-[200px]">
                <p className="font-medium text-gray-900 dark:text-gray-100 mb-2 border-b border-gray-100 dark:border-gray-700 pb-1">{label}</p>
                {payload.map((entry: any, index: number) => {
                    let eurValue = 0;
                    if (entry.dataKey === 'adjGrowth') eurValue = entry.payload.adjEur;
                    else if (entry.dataKey === 'totalGrowth') eurValue = entry.payload.totalEur;
                    else if (entry.dataKey === 'payoutGrowth') eurValue = entry.payload.payoutEur;
                    else if (entry.dataKey === 'grossGrowth') eurValue = entry.payload.grossEur;

                    return (
                        <div key={index} className="flex items-center justify-between gap-4 mb-1">
                            <div className="flex items-center gap-2">
                                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: entry.color }} />
                                <span className="text-gray-600 dark:text-gray-400">{entry.name}:</span>
                            </div>
                            <div className="text-right flex items-center gap-2">
                                <span className="text-gray-400 text-xs font-mono">{fmtCurrency(eurValue, 'EUR')}</span>
                                <span className={`font-medium min-w-[3rem] ${entry.value > 0 ? 'text-emerald-500' : entry.value < 0 ? 'text-red-500' : 'text-gray-900 dark:text-gray-100'}`}>
                                    {entry.value > 0 ? '+' : ''}{entry.value}%
                                </span>
                            </div>
                        </div>
                    );
                })}
                {bonusesThisMonth.length > 0 && (
                    <div className="mt-3 pt-2 border-t border-gray-100 dark:border-gray-700">
                        <p className="text-[10px] font-bold text-gray-400 uppercase tracking-wider mb-1">{t('payslips.chart.bonusesIncluded')}</p>
                        <div className="space-y-1">
                            {bonusesThisMonth.map((b: any, idx: number) => (
                                <div key={idx} className="flex justify-between items-center text-xs">
                                    <span className="text-gray-500 dark:text-gray-400 truncate max-w-[160px]" title={b.description}>{b.description}</span>
                                    <span className="font-mono text-gray-700 dark:text-gray-300 ml-3">{fmtCurrency(b.amount, 'EUR')}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        );
    }
    return null;
};

const YearlyTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
        return (
            <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 p-3 rounded-xl shadow-lg text-sm z-50">
                <p className="font-bold text-gray-900 dark:text-gray-100 mb-2 border-b border-gray-100 dark:border-gray-700 pb-1">{label}</p>
                {payload.map((entry: any, index: number) => (
                    <div key={index} className="flex items-center justify-between gap-8 mb-1">
                        <div className="flex items-center gap-2">
                            <div className="w-3 h-3 rounded-full" style={{ backgroundColor: entry.color }} />
                            <span className="text-gray-600 dark:text-gray-400">{entry.name}:</span>
                        </div>
                        <span className="font-mono font-medium text-gray-900 dark:text-gray-100">{fmtCurrency(entry.value, 'EUR')}</span>
                    </div>
                ))}
            </div>
        );
    }
    return null;
};

interface PayslipChartProps {
    filteredPayslips: Payslip[];
    ignoredPayslipIds: Set<string>;
    uniqueBonuses: string[];
    excludedBonuses: Set<string>;
    setExcludedBonuses: React.Dispatch<React.SetStateAction<Set<string>>>;
    useProportionalMath: boolean;
    setUseProportionalMath: (val: boolean) => void;
}

export function PayslipChart({
                                 filteredPayslips, ignoredPayslipIds, uniqueBonuses, excludedBonuses, setExcludedBonuses,
                                 useProportionalMath, setUseProportionalMath
                             }: PayslipChartProps) {
    const { t } = useTranslation();
    const [showBonusDropdown, setShowBonusDropdown] = useState(false);
    const [viewMode, setViewMode] = useState<'growth' | 'yearly'>('growth');
    const [visibleChartLines, setVisibleChartLines] = useState({
        grossGrowth: false, adjGrowth: true, totalGrowth: true, payoutGrowth: false,
    });

    const chartData = useMemo(() => {
        const sortedForChart = [...filteredPayslips]
            .filter(p => !ignoredPayslipIds.has(p.id))
            .sort((a, b) => (a.period_year - b.period_year) || ((a.period_month_num || 0) - (b.period_month_num || 0)));

        if (sortedForChart.length === 0) return [];

        const baselinePayslip = sortedForChart[0];
        const baselineGross = baselinePayslip.gross_pay;
        const baselineAdjNet = getAdjustedNetto(baselinePayslip, excludedBonuses, true, useProportionalMath);
        const baselineTotalNet = baselinePayslip.net_pay;
        const baselinePayout = baselinePayslip.payout_amount;

        return sortedForChart.map((p) => {
            const grossNet = p.gross_pay;
            const adjNet = getAdjustedNetto(p, excludedBonuses, true, useProportionalMath);
            const totalNet = p.net_pay;
            const payout = p.payout_amount;

            let grossGrowth = 0, adjGrowth = 0, totalGrowth = 0, payoutGrowth = 0;
            if (baselineGross > 0) grossGrowth = ((grossNet - baselineGross) / baselineGross) * 100;
            if (baselineAdjNet > 0) adjGrowth = ((adjNet - baselineAdjNet) / baselineAdjNet) * 100;
            if (baselineTotalNet > 0) totalGrowth = ((totalNet - baselineTotalNet) / baselineTotalNet) * 100;
            if (baselinePayout > 0) payoutGrowth = ((payout - baselinePayout) / baselinePayout) * 100;

            return {
                period: formatYearMonth(p.period_year, p.period_month_num),
                grossGrowth: Number(grossGrowth.toFixed(1)),
                adjGrowth: Number(adjGrowth.toFixed(1)),
                totalGrowth: Number(totalGrowth.toFixed(1)),
                payoutGrowth: Number(payoutGrowth.toFixed(1)),
                grossEur: grossNet, adjEur: adjNet, totalEur: totalNet, payoutEur: payout,
                bonusesThisMonth: p.bonuses || []
            };
        });
    }, [filteredPayslips, excludedBonuses, useProportionalMath, ignoredPayslipIds]);

    const yearlyData = useMemo(() => {
        const sorted = [...filteredPayslips]
            .filter(p => !ignoredPayslipIds.has(p.id))
            .sort((a, b) => a.period_year - b.period_year);
        
        const years: Record<number, { year: string, gross: number, net: number, payout: number }> = {};
        
        sorted.forEach(p => {
            if (!years[p.period_year]) {
                years[p.period_year] = { year: p.period_year.toString(), gross: 0, net: 0, payout: 0 };
            }
            years[p.period_year].gross += p.gross_pay || 0;
            years[p.period_year].net += p.net_pay || 0;
            years[p.period_year].payout += p.payout_amount || 0;
        });
        
        return Object.values(years).sort((a, b) => Number(a.year) - Number(b.year));
    }, [filteredPayslips, ignoredPayslipIds]);

    if (!chartData || chartData.length <= 1) return null;

    return (
        <div className="bg-white dark:bg-gray-900 p-5 rounded-2xl border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col gap-4">
            <div className="flex flex-col xl:flex-row xl:items-start justify-between gap-4 mb-2">
                <div className="flex items-start gap-2">
                    {viewMode === 'growth' ? <LineChartIcon className="text-indigo-500 mt-0.5" size={20} /> : <BarChart3 className="text-indigo-500 mt-0.5" size={20} />}
                    <div>
                        <div className="flex items-center gap-3">
                            <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                                {viewMode === 'growth' ? t('payslips.chart.title') : t('payslips.chart.yearlyDevelopment')}
                            </h3>
                            <div className="flex bg-gray-100 dark:bg-gray-800 p-1 rounded-lg">
                                <button 
                                    onClick={() => setViewMode('growth')}
                                    className={`px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-md transition-all ${viewMode === 'growth' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}`}
                                >
                                    {t('payslips.chart.cumulative')}
                                </button>
                                <button 
                                    onClick={() => setViewMode('yearly')}
                                    className={`px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-md transition-all ${viewMode === 'yearly' ? 'bg-white dark:bg-gray-700 text-indigo-600 dark:text-indigo-400 shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'}`}
                                >
                                    {t('payslips.chart.yearly')}
                                </button>
                            </div>
                        </div>
                        <p className="text-xs text-gray-500 mt-1 max-w-2xl leading-relaxed">
                            {viewMode === 'growth' ? (
                                <Trans i18nKey="payslips.chart.desc">
                                    <strong>Total Net</strong> tracks your raw payslip amounts, including massive spikes from bonuses.
                                    <strong> Adjusted Net</strong> shows your base salary trend by neutralizing the specific bonuses and leasing rates you select below.
                                </Trans>
                            ) : (
                                t('payslips.chart.yearlyDesc')
                            )}
                        </p>
                    </div>
                </div>

                {viewMode === 'growth' && (
                    <div className="flex flex-col sm:flex-row sm:items-center gap-4 mt-2 xl:mt-0">
                        <div className="flex items-center gap-3 bg-gray-50 dark:bg-gray-800/50 px-3 py-1.5 rounded-lg border border-gray-200 dark:border-gray-700">
                            <div className="relative">
                                <button onClick={() => setShowBonusDropdown(!showBonusDropdown)} className={`flex items-center gap-1.5 text-xs font-medium transition-colors hover:text-gray-900 dark:hover:text-gray-100 ${excludedBonuses.size > 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-gray-600 dark:text-gray-400'}`}>
                                    {t('payslips.chart.filterBonuses', { count: excludedBonuses.size })} <ChevronDown size={14}/>
                                </button>
                                {showBonusDropdown && (
                                    <div className="absolute top-full left-0 mt-2 w-72 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl shadow-xl z-20 max-h-64 overflow-y-auto p-2 animate-in fade-in slide-in-from-top-2">
                                        <div className="flex justify-between items-center mb-2 px-2 pb-2 border-b border-gray-100 dark:border-gray-700">
                                            <span className="text-[10px] font-bold text-gray-400 uppercase tracking-wider">{t('payslips.chart.selectToExclude')}</span>
                                            <div className="flex items-center gap-2">
                                                <button onClick={(e) => { e.preventDefault(); setExcludedBonuses(new Set(uniqueBonuses)); }} className="text-[10px] font-medium text-indigo-600 dark:text-indigo-400 hover:underline">{t('payslips.chart.all')}</button>
                                                <span className="text-gray-300 dark:text-gray-600 text-[10px]">|</span>
                                                <button onClick={(e) => { e.preventDefault(); setExcludedBonuses(new Set()); }} className="text-[10px] font-medium text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:underline">{t('payslips.chart.none')}</button>
                                            </div>
                                        </div>
                                        {uniqueBonuses.length === 0 && <div className="px-2 text-xs text-gray-400">{t('payslips.chart.noBonuses')}</div>}
                                        {uniqueBonuses.map(b => (
                                            <label key={b} className="flex items-center gap-2 px-2 py-1.5 hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg cursor-pointer">
                                                <input
                                                    type="checkbox"
                                                    checked={excludedBonuses.has(b)}
                                                    onChange={(e) => setExcludedBonuses(prev => { const next = new Set(prev); e.target.checked ? next.add(b) : next.delete(b); return next; })}
                                                    className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 bg-white dark:bg-gray-900"
                                                />
                                                <span className="text-xs text-gray-700 dark:text-gray-300 truncate" title={b}>{b}</span>
                                            </label>
                                        ))}
                                    </div>
                                )}
                            </div>
                            <div className="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>
                            <label className="flex items-center gap-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 cursor-pointer hover:text-gray-900 dark:hover:text-gray-100" title={t('payslips.chart.propTaxMathTitle')}>
                                <input type="checkbox" checked={useProportionalMath} onChange={e => setUseProportionalMath(e.target.checked)} className="rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 bg-white dark:bg-gray-900" />
                                {t('payslips.chart.propTaxMath')}
                            </label>
                        </div>

                        <div className="flex flex-wrap items-center gap-2">
                            <button onClick={() => setVisibleChartLines(prev => ({ ...prev, grossGrowth: !prev.grossGrowth }))} className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${visibleChartLines.grossGrowth ? 'bg-amber-50 border-amber-200 text-amber-700 dark:bg-amber-900/30 dark:border-amber-800 dark:text-amber-400' : 'bg-gray-50 border-gray-200 text-gray-500 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700/50'}`}>
                                <div className={`w-2.5 h-2.5 rounded-full ${visibleChartLines.grossGrowth ? 'bg-amber-500' : 'bg-gray-400 dark:bg-gray-600'}`} /> {t('payslips.chart.grossGrowth')}
                            </button>
                            <button onClick={() => setVisibleChartLines(prev => ({ ...prev, adjGrowth: !prev.adjGrowth }))} className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${visibleChartLines.adjGrowth ? 'bg-blue-50 border-blue-200 text-blue-700 dark:bg-blue-900/30 dark:border-blue-800 dark:text-blue-400' : 'bg-gray-50 border-gray-200 text-gray-500 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700/50'}`}>
                                <div className={`w-2.5 h-2.5 rounded-full ${visibleChartLines.adjGrowth ? 'bg-blue-500' : 'bg-gray-400 dark:bg-gray-600'}`} /> {t('payslips.chart.adjGrowth')}
                            </button>
                            <button onClick={() => setVisibleChartLines(prev => ({ ...prev, totalGrowth: !prev.totalGrowth }))} className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${visibleChartLines.totalGrowth ? 'bg-emerald-50 border-emerald-200 text-emerald-700 dark:bg-emerald-900/30 dark:border-emerald-800 dark:text-emerald-400' : 'bg-gray-50 border-gray-200 text-gray-500 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700/50'}`}>
                                <div className={`w-2.5 h-2.5 rounded-full ${visibleChartLines.totalGrowth ? 'bg-emerald-500' : 'bg-gray-400 dark:bg-gray-600'}`} /> {t('payslips.chart.totalGrowth')}
                            </button>
                            <button onClick={() => setVisibleChartLines(prev => ({ ...prev, payoutGrowth: !prev.payoutGrowth }))} className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors border ${visibleChartLines.payoutGrowth ? 'bg-purple-50 border-purple-200 text-purple-700 dark:bg-purple-900/30 dark:border-purple-800 dark:text-purple-400' : 'bg-gray-50 border-gray-200 text-gray-500 dark:bg-gray-800 dark:border-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700/50'}`}>
                                <div className={`w-2.5 h-2.5 rounded-full ${visibleChartLines.payoutGrowth ? 'bg-purple-500' : 'bg-gray-400 dark:bg-gray-600'}`} /> {t('payslips.chart.payoutGrowth')}
                            </button>
                        </div>
                    </div>
                )}
            </div>

            <div className="h-72 w-full mt-4">
                <ResponsiveContainer width="100%" height="100%">
                    {viewMode === 'growth' ? (
                        <LineChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#374151" opacity={0.2} />
                            <XAxis dataKey="period" stroke="#6B7280" fontSize={12} tickLine={false} axisLine={false} dy={10} />
                            <YAxis stroke="#6B7280" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(val) => `${val}%`} />
                            <Tooltip content={<CustomChartTooltip t={t} />} cursor={{ stroke: '#6B7280', strokeWidth: 1, strokeDasharray: '3 3' }} />
                            <ReferenceLine y={0} stroke="#9CA3AF" strokeDasharray="3 3" />
                            <Line name={t('payslips.chart.grossGrowth')} hide={!visibleChartLines.grossGrowth} type="monotone" dataKey="grossGrowth" stroke="#f59e0b" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} />
                            <Line name={t('payslips.chart.adjGrowth')} hide={!visibleChartLines.adjGrowth} type="monotone" dataKey="adjGrowth" stroke="#3b82f6" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} />
                            <Line name={t('payslips.chart.totalGrowth')} hide={!visibleChartLines.totalGrowth} type="monotone" dataKey="totalGrowth" stroke="#10b981" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} />
                            <Line name={t('payslips.chart.payoutGrowth')} hide={!visibleChartLines.payoutGrowth} type="monotone" dataKey="payoutGrowth" stroke="#8b5cf6" strokeWidth={3} dot={{ r: 4, strokeWidth: 2 }} activeDot={{ r: 6 }} />
                        </LineChart>
                    ) : (
                        <BarChart data={yearlyData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                            <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#374151" opacity={0.2} />
                            <XAxis dataKey="year" stroke="#6B7280" fontSize={12} tickLine={false} axisLine={false} dy={10} />
                            <YAxis stroke="#6B7280" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(val) => `${(val / 1000).toFixed(0)}k`} />
                            <Tooltip content={<YearlyTooltip />} cursor={{ fill: 'rgba(107, 114, 128, 0.1)' }} />
                            <Legend verticalAlign="top" align="right" height={36} iconType="circle" />
                            <Bar name={t('payslips.modals.gross')} dataKey="gross" fill="#f59e0b" radius={[4, 4, 0, 0]} barSize={40} />
                            <Bar name={t('payslips.modals.net')} dataKey="net" fill="#10b981" radius={[4, 4, 0, 0]} barSize={40} />
                            <Bar name={t('payslips.modals.payout')} dataKey="payout" fill="#8b5cf6" radius={[4, 4, 0, 0]} barSize={40} />
                        </BarChart>
                    )}
                </ResponsiveContainer>
            </div>
        </div>
    );
}