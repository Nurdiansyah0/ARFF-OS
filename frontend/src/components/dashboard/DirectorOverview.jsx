'use client';

import React from 'react';
import dynamic from 'next/dynamic';
const Chart = dynamic(() => import('react-apexcharts'), { ssr: false });

import { ShieldAlert, CheckCircle, Activity, LayoutDashboard, Target } from 'lucide-react';

const Card = ({ title, icon: Icon, children, className = '' }) => (
    <section className={`bg-white border border-slate-200 rounded-xl overflow-hidden flex flex-col shadow-sm transition-all hover:shadow-md ${className}`}>
        <header className="px-5 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
            <div className="flex items-center gap-3">
                {Icon && <Icon aria-hidden="true" className="w-4 h-4 text-slate-500" />}
                <h3 className="text-sm font-bold text-slate-900 tracking-wide uppercase">{title}</h3>
            </div>
        </header>
        <div className="flex-1 p-5 flex flex-col">
            {children}
        </div>
    </section>
);

const DirectorOverview = ({ arff_health_index, critical_incidents_ytd, overall_compliance, arff_health_matrix }) => {
    
    // Format numeric values
    const fmtIndex = Number(arff_health_index).toFixed(1);
    const fmtCompliance = Number(overall_compliance).toFixed(1);

    // Prepare chart data
    const chartLabels = arff_health_matrix.map(item => item.name.replace('LAPORAN ', ''));
    const chartSeries = arff_health_matrix.map(item => Number(item.weight));
    const colors = ['#C62828', '#f59e0b', '#10b981', '#3b82f6', '#8b5cf6'];

    const chartOptions = {
        chart: { type: 'donut', fontFamily: 'inherit' },
        labels: chartLabels,
        colors: colors,
        plotOptions: {
            pie: { donut: { size: '75%' } }
        },
        dataLabels: { enabled: false },
        stroke: { show: false },
        legend: {
            position: 'right',
            labels: { colors: '#64748b' },
            fontFamily: 'inherit',
            fontWeight: 700,
            markers: { shape: 'circle' }
        },
        tooltip: {
            y: { formatter: (val) => `${val}% Weight` }
        }
    };

    return (
        <main className="flex flex-col gap-6 pb-32 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 pt-6">
            <div className="grid grid-cols-1 xl:grid-cols-12 gap-6">
                
                {/* LEFT COLUMN */}
                <div className="xl:col-span-4 flex flex-col gap-6">
                    <Card title="Health Metrics" icon={Activity}>
                        <div className="text-center mb-8">
                            <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-2" id="health-index-label">ARFF Health Index</p>
                            <div className="flex items-baseline justify-center gap-2" aria-labelledby="health-index-label">
                                <span className="text-6xl font-black text-slate-900 tracking-tighter">{fmtIndex}</span>
                                <span className="text-xl font-bold text-slate-400">/ 100%</span>
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-4 border-t border-slate-100 pt-6">
                            <div className="text-center border-r border-slate-100">
                                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Critical Incidents</p>
                                <p className="text-3xl font-black text-[#C62828]">{critical_incidents_ytd}</p>
                                <p className="text-[8px] font-bold text-slate-400 mt-1 uppercase tracking-widest">Year to Date</p>
                            </div>
                            <div className="text-center">
                                <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-1">Compliance</p>
                                <p className="text-3xl font-black text-emerald-600">{fmtCompliance}%</p>
                                <p className="text-[8px] font-bold text-slate-400 mt-1 uppercase tracking-widest">ICAO Standards</p>
                            </div>
                        </div>
                    </Card>

                    <Card title="Readiness Breakdown" icon={Target}>
                        <div className="space-y-5">
                            {arff_health_matrix.map((item, idx) => (
                                <div key={idx}>
                                    <div className="flex justify-between text-xs mb-1.5 font-bold uppercase tracking-widest">
                                        <span className="text-slate-600">{item.name.replace('LAPORAN ', '')}</span>
                                        <span className="text-slate-900">{item.percentage}%</span>
                                    </div>
                                    <div className="w-full bg-slate-100 rounded-full h-1.5">
                                        <div className="bg-emerald-500 h-1.5 rounded-full" style={{ width: `${Math.min(100, item.percentage)}%` }}></div>
                                    </div>
                                    <p className="text-[9px] font-bold text-slate-400 mt-1.5 text-right uppercase tracking-wider">
                                        Weight: {item.weight}% &bull; Contribution: {item.total}%
                                    </p>
                                </div>
                            ))}
                        </div>
                    </Card>
                </div>

                {/* RIGHT COLUMN */}
                <div className="xl:col-span-8 flex flex-col gap-6">
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                        <Card title="Matrix Weights" icon={LayoutDashboard}>
                            <div className="relative h-[250px] flex items-center justify-center">
                                <Chart options={chartOptions} series={chartSeries} type="donut" width="100%" height="250" />
                            </div>
                        </Card>

                        <Card title="Management by Exception" icon={ShieldAlert}>
                            <div 
                                className="flex-1 flex flex-col items-center justify-center py-6 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50"
                                aria-live="polite"
                            >
                                {critical_incidents_ytd > 0 ? (
                                    <>
                                        <div className="w-16 h-16 bg-red-50 rounded-full flex items-center justify-center mb-4 border border-red-100 shadow-sm">
                                            <ShieldAlert aria-hidden="true" className="w-8 h-8 text-[#C62828]" />
                                        </div>
                                        <p className="text-sm font-black text-slate-900 uppercase tracking-widest">{critical_incidents_ytd} Critical Incidents</p>
                                        <p className="text-[10px] font-bold text-slate-500 mt-2 uppercase tracking-widest max-w-[200px]">Require immediate executive review.</p>
                                    </>
                                ) : (
                                    <>
                                        <div className="w-16 h-16 bg-emerald-50 rounded-full flex items-center justify-center mb-4 border border-emerald-100 shadow-sm">
                                            <CheckCircle aria-hidden="true" className="w-8 h-8 text-emerald-600" />
                                        </div>
                                        <p className="text-sm font-black text-slate-900 uppercase tracking-widest">All Systems Operational</p>
                                        <p className="text-[10px] font-bold text-slate-500 mt-2 uppercase tracking-widest max-w-[200px]">No critical issues or exceptions found.</p>
                                    </>
                                )}
                            </div>
                        </Card>
                    </div>

                    <Card title="Complete Matrix" icon={Activity} className="flex-1">
                        <div className="overflow-x-auto rounded-lg border border-slate-200 shadow-sm">
                            <table className="w-full text-left border-collapse">
                                <caption className="sr-only">Detailed breakdown of ARFF health metrics and contributions.</caption>
                                <thead className="bg-slate-50 border-b border-slate-200">
                                    <tr>
                                        <th scope="col" className="py-3 px-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest">Component</th>
                                        <th scope="col" className="py-3 px-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest text-center">Score</th>
                                        <th scope="col" className="py-3 px-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest text-center">Weight</th>
                                        <th scope="col" className="py-3 px-4 text-[10px] font-bold text-slate-500 uppercase tracking-widest text-right">Contribution</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-100 bg-white">
                                    {arff_health_matrix.map((item, idx) => (
                                        <tr key={idx} className="hover:bg-slate-50 transition-colors">
                                            <td className="py-4 px-4 text-xs font-bold text-slate-900 uppercase tracking-wider">{item.name}</td>
                                            <td className="py-4 px-4 text-xs font-bold text-slate-600 text-center">{Number(item.percentage).toFixed(2)}%</td>
                                            <td className="py-4 px-4 text-xs font-bold text-slate-500 text-center">{item.weight}%</td>
                                            <td className="py-4 px-4 text-sm font-black text-emerald-600 text-right">{item.total}%</td>
                                        </tr>
                                    ))}
                                </tbody>
                                <tfoot className="bg-slate-50 border-t border-slate-200">
                                    <tr>
                                        <th scope="row" colSpan="3" className="py-4 px-4 text-[10px] font-bold text-slate-500 text-right uppercase tracking-widest">
                                            Total ARFF Health Index
                                        </th>
                                        <td className="py-4 px-4 text-right">
                                            <span className="text-lg font-black text-slate-900">{fmtIndex}%</span>
                                        </td>
                                    </tr>
                                </tfoot>
                            </table>
                        </div>
                    </Card>
                </div>
            </div>
        </main>
    );
};

export default DirectorOverview;
