'use client';

import React, { useState, useEffect } from 'react';
import { 
    AlertTriangle, Shield, Clock, Timer, CheckCircle, 
    XCircle, Activity, Info, Truck, Calendar, CheckSquare, Target, Flame
} from 'lucide-react';

// Reusable Enterprise Card Component
const Card = ({ title, icon: Icon, children, className = '', action = null }) => (
  <div className={`bg-white border border-slate-200 rounded-xl overflow-hidden flex flex-col shadow-sm transition-all hover:shadow-md ${className}`}>
    <div className="px-5 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
      <div className="flex items-center gap-3">
        {Icon && <Icon className="w-4 h-4 text-slate-500" />}
        <h3 className="text-sm font-bold text-slate-900 tracking-wide uppercase">{title}</h3>
      </div>
      {action && <div>{action}</div>}
    </div>
    <div className="flex-1 p-0 flex flex-col relative overflow-hidden">
      {children}
    </div>
  </div>
);

const OperationsOverview = (props) => {
    const {
        user, activeIncident, secondsRemaining: initialSeconds, isOnDuty, imageFilename,
        kpi_percentage, kpi_duty_days, kpi_submitted,
        fire_system_percentage, fire_system_quarter, fire_system_window, fire_system_inspected, fire_system_total,
        latest_samapta, todays_tasks, is_squad_leader, assetUrl, routes, csrfToken
    } = props;

    // Timer logic
    const [secondsRemaining, setSecondsRemaining] = useState(initialSeconds);
    const [formattedTime, setFormattedTime] = useState('00:00:00');

    useEffect(() => {
        let interval = null;
        if (isOnDuty && secondsRemaining > 0) {
            interval = setInterval(() => {
                setSecondsRemaining(prev => {
                    if (prev <= 1) {
                        clearInterval(interval);
                        setFormattedTime('SHIFT ENDED');
                        return 0;
                    }
                    const newSec = prev - 1;
                    const h = Math.floor(newSec / 3600);
                    const m = Math.floor((newSec % 3600) / 60);
                    const s = newSec % 60;
                    setFormattedTime([h, m, s].map(v => v.toString().padStart(2, '0')).join(':'));
                    return newSec;
                });
            }, 1000);
        } else if (!isOnDuty) {
            setFormattedTime('OFF DUTY');
        }
        return () => clearInterval(interval);
    }, [isOnDuty, secondsRemaining]);

    const dutyColorMap = {
        'Off Duty': 'bg-slate-100 text-slate-600 border-slate-200',
        'Morning Duty': 'bg-emerald-50 text-emerald-700 border-emerald-200',
        'Night Duty': 'bg-blue-50 text-blue-700 border-blue-200',
    };
    const dutyBadgeClass = dutyColorMap[user.duty_status] || dutyColorMap['Off Duty'];

    // Dispatch custom events for checklist
    const handleChecklistEdit = (task) => {
        if (task.is_fire_system) {
            window.dispatchEvent(new CustomEvent('open-fire-sys-inspect', { detail: { hash: task.hash } }));
            return;
        }

        const vName = task.vehicle_name?.toUpperCase() || '';
        let typeTask = 'standard';
        if (vName.includes('AMBULANCE')) typeTask = 'ambulance';
        else if (vName.startsWith('F')) typeTask = vName.includes('FCP') ? 'fcp' : 'foam';
        else if (vName.includes('COMMANDO')) typeTask = 'commando';
        else if (vName.startsWith('N')) typeTask = 'nurse';
        else if (vName.includes('UTILITY')) typeTask = 'utility';

        const eventStr = {
            'foam': 'open-foam-tender-checklist',
            'utility': 'open-utility-checklist',
            'nurse': 'open-nurse-tender-checklist',
            'fcp': 'open-fcp-checklist',
            'commando': 'open-commando-checklist',
            'ambulance': 'open-ambulance-checklist',
            'standard': 'open-ambulance-checklist'
        }[typeTask];

        window.dispatchEvent(new CustomEvent(eventStr, { detail: { vehicle: task.vehicle_name, checklistId: task.id } }));
    };

    return (
        <div className="flex flex-col gap-6 pb-32 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 pt-6">
            <div className="flex items-center justify-between mb-2">
                <h2 className="text-2xl font-black text-slate-900 tracking-tight">OFFICER COMMAND</h2>
            </div>

            {/* Emergency Alert Banner */}
            {activeIncident && (
                <div className="bg-red-50 border border-red-200 rounded-xl p-5 shadow-sm animate-pulse relative overflow-hidden flex flex-col sm:flex-row items-center gap-4">
                    <div className="absolute top-0 right-0 p-4 opacity-10">
                        <AlertTriangle className="w-24 h-24 text-red-600" />
                    </div>
                    <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center text-red-600 shrink-0 relative z-10 border border-red-200">
                        <AlertTriangle className="w-6 h-6 animate-bounce" />
                    </div>
                    <div className="flex-1 text-center sm:text-left relative z-10">
                        <span className="text-[10px] font-bold text-red-700 bg-red-100 px-2 py-0.5 rounded uppercase tracking-wider mb-1 inline-block border border-red-200">EMERGENCY STANDBY ACTIVE</span>
                        <h4 className="text-lg font-black text-red-900 uppercase">{activeIncident.title || 'KEJADIAN DARURAT AKTIF'}</h4>
                        <p className="text-sm text-red-700 font-medium uppercase mt-0.5">LOKASI: <span className="font-bold text-red-900">{activeIncident.location || 'AREA BANDARA'}</span></p>
                    </div>
                    {routes.watchroomShift && (
                        <a href={routes.watchroomShift} className="relative z-10 w-full sm:w-auto px-5 py-2.5 rounded-lg bg-red-600 hover:bg-red-700 text-white font-bold uppercase text-xs tracking-wide text-center transition-all shadow-sm">
                            View Shift Log
                        </a>
                    )}
                </div>
            )}

            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                
                {/* Identity & Status Card */}
                <div className="lg:col-span-6 flex flex-col gap-6 h-full">
                    <Card title="Duty Profile" icon={Shield} className="h-full">
                        <div className="p-6 flex flex-col h-full">
                            <div className="flex items-center gap-5 mb-6">
                                <div className="w-20 h-20 rounded-xl bg-slate-100 border border-slate-200 shadow-sm shrink-0 overflow-hidden relative">
                                    {user.profile_photo_url ? (
                                        <img src={user.profile_photo_url} className="w-full h-full object-cover" alt="Profile" />
                                    ) : (
                                        <div className="w-full h-full flex items-center justify-center text-slate-400 font-black text-3xl uppercase">
                                            {user.name.charAt(0)}
                                        </div>
                                    )}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <span className="text-[10px] font-bold text-slate-600 bg-slate-100 border border-slate-200 px-2 py-0.5 rounded uppercase tracking-wider block w-max mb-2">ARFF OFFICER</span>
                                    <h3 className="text-xl font-black text-slate-900 uppercase tracking-tight truncate">{user.name}</h3>
                                    <p className="text-xs font-semibold text-slate-500 uppercase tracking-wide mt-1">
                                        {user.position} <span className="mx-1 text-slate-300">|</span> <span className="text-[#C62828]">{user.employment_status || 'PNS'}</span>
                                    </p>
                                </div>
                            </div>

                            <div className="flex-1 flex flex-col justify-end">
                                <div className="flex flex-col md:flex-row md:items-end justify-between gap-4 border-t border-slate-100 pt-5 mt-auto">
                                    <div>
                                        <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest block mb-1">Shift & Status</span>
                                        <div className="flex items-center gap-3">
                                            <span className="text-2xl font-black text-slate-900 tracking-tight uppercase leading-none">{user.shift || 'N/A'}</span>
                                            <div className={`px-2.5 py-1 rounded-md border ${dutyBadgeClass}`}>
                                                <span className="text-[10px] font-bold uppercase tracking-wider">{user.duty_status}</span>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div className="text-left md:text-right">
                                        {isOnDuty && secondsRemaining > 0 ? (
                                            <>
                                                <div className="flex items-center gap-2 mb-1 justify-start md:justify-end">
                                                    <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Time Remaining</span>
                                                    <div className="w-2 h-2 rounded-full bg-[#C62828] animate-pulse"></div>
                                                </div>
                                                <span className="font-mono text-3xl font-black text-[#C62828] tracking-tight leading-none block">{formattedTime}</span>
                                            </>
                                        ) : (
                                            <>
                                                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest block mb-1">Tactical Time</span>
                                                <span className="text-xl font-black text-slate-400 uppercase tracking-wide block leading-none">{formattedTime}</span>
                                            </>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    </Card>
                </div>

                {/* Vehicle Inspection Hub */}
                <div className="lg:col-span-6 flex flex-col gap-6 h-full">
                    <Card title="Compliance Hub" icon={Activity} className="h-full">
                        {user.vehicle && user.vehicle !== 'Unassigned' ? (
                            <div className="flex flex-col h-full">
                                <div className="relative w-full h-32 sm:h-40 bg-slate-100 overflow-hidden group border-b border-slate-200">
                                    <div className="absolute inset-0 bg-gradient-to-t from-slate-900/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity z-10"></div>
                                    <img src={`${assetUrl}${imageFilename}`} alt={user.vehicle} className="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105" />
                                    <div className="absolute bottom-0 left-0 w-full p-4 bg-gradient-to-t from-slate-900 to-transparent z-20">
                                        <h4 className="text-xl font-black text-white tracking-wide uppercase leading-tight">{user.vehicle}</h4>
                                        <span className="text-[10px] font-bold text-slate-200 uppercase tracking-wider mt-0.5 block">
                                            Plot: <span className="text-white">{user.plotting || 'ARFF FLIGHTLINE'}</span>
                                        </span>
                                    </div>
                                </div>

                                <div className="grid grid-cols-2 divide-x divide-slate-100 bg-white flex-1">
                                    {/* Checklist KPI */}
                                    <div className="p-5 flex flex-col">
                                        <div className="flex items-center justify-between mb-3">
                                            <div className="flex items-center gap-1.5">
                                                <Truck className="w-4 h-4 text-slate-400" />
                                                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Vehicle (Mon)</span>
                                            </div>
                                            <span className={`text-xs font-black ${kpi_percentage >= 100 ? 'text-emerald-600' : kpi_percentage >= 80 ? 'text-blue-600' : 'text-amber-600'}`}>{kpi_percentage}%</span>
                                        </div>
                                        <div className="flex items-center bg-slate-50 rounded-lg border border-slate-200 p-2 mb-3">
                                            <div className="flex-1 text-center border-r border-slate-200">
                                                <span className="text-sm font-black text-slate-900">{kpi_duty_days}</span>
                                                <span className="text-[9px] font-semibold text-slate-500 uppercase block mt-0.5">Target</span>
                                            </div>
                                            <div className="flex-1 text-center">
                                                <span className="text-sm font-black text-slate-900">{kpi_submitted}</span>
                                                <span className="text-[9px] font-semibold text-slate-500 uppercase block mt-0.5">Submit</span>
                                            </div>
                                        </div>
                                        <div className="mt-auto w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                                            <div className={`h-full rounded-full ${kpi_percentage >= 100 ? 'bg-emerald-500' : kpi_percentage >= 80 ? 'bg-blue-500' : 'bg-amber-500'}`} style={{ width: `${Math.min(100, kpi_percentage)}%` }}></div>
                                        </div>
                                    </div>

                                    {/* Fire System KPI */}
                                    <div className="p-5 flex flex-col">
                                        <div className="flex items-center justify-between mb-3">
                                            <div className="flex items-center gap-1.5">
                                                <Flame className="w-4 h-4 text-slate-400" />
                                                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Fire Sys ({fire_system_quarter})</span>
                                            </div>
                                            <span className={`text-xs font-black ${fire_system_percentage >= 100 ? 'text-emerald-600' : fire_system_percentage >= 80 ? 'text-blue-600' : 'text-amber-600'}`}>{fire_system_percentage}%</span>
                                        </div>
                                        <div className="flex items-center bg-slate-50 rounded-lg border border-slate-200 p-2 mb-3">
                                            <div className="flex-1 text-center border-r border-slate-200">
                                                <span className="text-sm font-black text-[#C62828] truncate px-1">{fire_system_window}</span>
                                                <span className="text-[9px] font-semibold text-slate-500 uppercase block mt-0.5">Window</span>
                                            </div>
                                            <div className="flex-1 text-center">
                                                <span className="text-sm font-black text-slate-900">{fire_system_inspected} <span className="text-xs text-slate-400 font-semibold">/ {fire_system_total}</span></span>
                                                <span className="text-[9px] font-semibold text-slate-500 uppercase block mt-0.5">Checked</span>
                                            </div>
                                        </div>
                                        <div className="mt-auto w-full bg-slate-100 rounded-full h-1.5 overflow-hidden">
                                            <div className={`h-full rounded-full ${fire_system_percentage >= 100 ? 'bg-emerald-500' : fire_system_percentage >= 80 ? 'bg-blue-500' : 'bg-amber-500'}`} style={{ width: `${Math.min(100, fire_system_percentage)}%` }}></div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        ) : (
                            <div className="p-6 flex-1 flex flex-col items-center justify-center text-center bg-slate-50 m-4 rounded-xl border border-dashed border-slate-300">
                                <div className="w-16 h-16 rounded-full bg-slate-100 flex items-center justify-center text-slate-400 mb-4 shadow-sm border border-slate-200">
                                    <AlertTriangle className="w-8 h-8" />
                                </div>
                                <h4 className="text-base font-black text-slate-700 uppercase tracking-wider">No Vehicle Plotted</h4>
                                <p className="text-xs font-medium text-slate-500 mt-2 max-w-[250px]">
                                    KPI cannot be calculated. Contact your Team Leader to assign a vehicle for today's shift.
                                </p>
                            </div>
                        )}
                    </Card>
                </div>
            </div>

            {/* Battery Test (Samapta) - Full Width if available */}
            {latest_samapta && (
                <Card title="Battery Test Performance" icon={Activity}>
                    <div className="p-5">
                        {(() => {
                            const runScore = Math.min(100, (latest_samapta.run_12min / 2400) * 100);
                            const pushScore = Math.min(100, (latest_samapta.push_up / 40) * 100);
                            const sitScore = Math.min(100, (latest_samapta.sit_up / 40) * 100);
                            const pullScore = Math.min(100, (latest_samapta.pull_up / 15) * 100);
                            const shuttleScore = Math.min(100, (20 / Math.max(1, latest_samapta.shuttle_run)) * 100);
                            const overallScore = Math.round((runScore + pushScore + sitScore + pullScore + shuttleScore) / 5);
                            const theme = overallScore >= 80 ? 'text-emerald-600' : (overallScore >= 60 ? 'text-amber-500' : 'text-[#C62828]');
                            
                            return (
                                <div className="flex flex-col md:flex-row md:items-center justify-between gap-6">
                                    <div className="flex items-center justify-between md:justify-start md:w-1/4">
                                        <div>
                                            <h4 className="text-3xl font-black tracking-tight leading-none">
                                                <span className={theme}>{overallScore}</span>
                                                <span className="text-sm text-slate-400 font-bold ml-1">/ 100</span>
                                            </h4>
                                            <p className="text-xs font-bold text-slate-500 uppercase tracking-widest mt-1">Physical Readiness</p>
                                        </div>
                                    </div>
                                    
                                    <div className="flex-1 grid grid-cols-5 gap-3">
                                        {[
                                            { label: 'Run', val: latest_samapta.run_12min, unit: 'm' },
                                            { label: 'Push', val: latest_samapta.push_up, unit: 'x' },
                                            { label: 'Sit', val: latest_samapta.sit_up, unit: 'x' },
                                            { label: 'Pull', val: latest_samapta.pull_up, unit: 'x' },
                                            { label: 'Shuttle', val: latest_samapta.shuttle_run, unit: 's' }
                                        ].map((item, idx) => (
                                            <div key={idx} className="bg-slate-50 rounded-lg border border-slate-200 p-2.5 text-center flex flex-col justify-center">
                                                <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{item.label}</span>
                                                <span className="text-sm font-black text-slate-900 mt-1">{item.val}<span className="text-xs text-slate-400 font-medium ml-0.5">{item.unit}</span></span>
                                            </div>
                                        ))}
                                    </div>
                                    
                                    {routes.samaptaHistory && (
                                        <div className="md:w-auto w-full">
                                            <a href={routes.samaptaHistory} className="block w-full text-center px-4 py-2.5 bg-white border border-slate-300 rounded-lg text-xs font-bold text-slate-700 hover:bg-slate-50 hover:text-[#C62828] transition-colors">
                                                View History
                                            </a>
                                        </div>
                                    )}
                                </div>
                            );
                        })()}
                    </div>
                </Card>
            )}

            {/* Recent Completed Tasks */}
            <Card 
                title="Recent Inspections Log" 
                icon={CheckSquare}
                action={
                    <span className={`px-2.5 py-1 rounded-md border text-[10px] font-bold uppercase tracking-wider ${is_squad_leader ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-red-50 text-red-700 border-red-200'}`}>
                        {is_squad_leader ? 'Squad Leader' : 'Officer'}
                    </span>
                }
            >
                <div className="p-5">
                    {todays_tasks && todays_tasks.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            {todays_tasks.map((task, i) => {
                                const isFire = task.is_fire_system;
                                const isServiceable = task.status === 'SERVICEABLE';
                                const printUrl = isFire ? `${routes.fireSysPrint}/${task.id}` : `${routes.checklistPrint}/${task.id}/print`;
                                
                                return (
                                    <div key={i} className="flex flex-col bg-white border border-slate-200 rounded-xl overflow-hidden hover:border-slate-300 hover:shadow-sm transition-all">
                                        <a href={printUrl} className="p-4 flex-1 flex items-start justify-between cursor-pointer group hover:bg-slate-50/50">
                                            <div className="flex items-start gap-3">
                                                <div className={`mt-0.5 w-2.5 h-2.5 rounded-full shrink-0 ${isServiceable ? 'bg-emerald-500' : 'bg-[#C62828]'}`}></div>
                                                <div>
                                                    <p className="text-sm font-black text-slate-900 uppercase group-hover:text-[#C62828] transition-colors flex items-center gap-1.5 leading-tight">
                                                        {isFire && <span className="text-[9px] bg-slate-100 border border-slate-200 text-slate-700 px-1 py-0.5 rounded leading-none">APAR</span>}
                                                        {task.vehicle_name}
                                                    </p>
                                                    <p className="text-[10px] font-semibold text-slate-500 uppercase tracking-wide mt-1">
                                                        OP: {task.user?.name || 'Unknown'}
                                                    </p>
                                                </div>
                                            </div>
                                            <div className="flex flex-col items-end text-right ml-2">
                                                <span className={`text-[10px] font-black uppercase px-2 py-0.5 rounded bg-slate-50 border ${isServiceable ? 'text-emerald-700 border-emerald-200' : 'text-red-700 border-red-200'}`}>
                                                    {task.status}
                                                </span>
                                                <span className="text-[9px] font-bold text-slate-400 uppercase mt-1.5 whitespace-nowrap">
                                                    {new Date(task.created_at).toLocaleString('en-GB', { timeZone: 'Asia/Jakarta', day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' }).replace(',', '')}
                                                </span>
                                            </div>
                                        </a>
                                        
                                        <div className="border-t border-slate-100 p-2 flex gap-2 bg-slate-50">
                                            <button 
                                                onClick={(e) => { e.preventDefault(); handleChecklistEdit(task); }}
                                                className="flex-1 py-2 px-3 rounded-md bg-white border border-slate-200 hover:border-slate-300 text-slate-600 hover:text-slate-900 text-xs font-bold uppercase tracking-wide transition-colors shadow-sm"
                                            >
                                                Edit
                                            </button>
                                            {!task.approved_at && !isFire && routes.checklistDelete && (
                                                <form action={`${routes.checklistDelete}/${task.id}`} method="POST" className="flex-1 m-0" onSubmit={(e) => {
                                                    if (!window.confirm('Hapus laporan ini?')) e.preventDefault();
                                                }}>
                                                    <input type="hidden" name="_token" value={csrfToken} />
                                                    <input type="hidden" name="_method" value="DELETE" />
                                                    <button 
                                                        type="submit" 
                                                        className="w-full py-2 px-3 rounded-md bg-red-50 border border-red-200 text-red-600 hover:bg-red-100 hover:text-red-700 text-xs font-bold uppercase tracking-wide transition-colors"
                                                    >
                                                        Delete
                                                    </button>
                                                </form>
                                            )}
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    ) : (
                        <div className="py-12 px-6 text-center border-2 border-dashed border-slate-200 rounded-xl bg-slate-50">
                            <CheckSquare className="w-8 h-8 text-slate-300 mx-auto mb-3" />
                            <p className="text-xs font-bold text-slate-500 uppercase tracking-widest">No completed inspections today</p>
                        </div>
                    )}
                </div>
            </Card>

        </div>
    );
};

export default OperationsOverview;
