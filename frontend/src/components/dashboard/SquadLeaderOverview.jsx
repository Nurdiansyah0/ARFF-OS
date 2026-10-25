'use client';

import React, { useState } from 'react';
import { 
    CheckCircle2, AlertTriangle, XCircle, FileText, Activity, Users, ShieldAlert,
    Clock, Search
} from 'lucide-react';

const Card = ({ title, icon: Icon, children, className = "" }) => (
    <div className={`premium-card ${className}`}>
        {title && (
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
                <h3 className="font-bold text-slate-900 text-sm tracking-widest uppercase">{title}</h3>
                {Icon && <Icon className="w-4 h-4 text-slate-400" />}
            </div>
        )}
        <div className="p-6">{children}</div>
    </div>
);

const StatBox = ({ label, value, subtext, colorClass }) => (
    <div className="p-4 rounded-xl border border-slate-100 bg-slate-50">
        <span className="text-[10px] sm:text-xs font-black text-slate-400 uppercase tracking-widest block mb-1">{label}</span>
        <div className="flex items-end gap-2">
            <span className={`text-xl font-black ${colorClass} leading-none`}>{value}</span>
            {subtext && <span className="text-[10px] sm:text-xs font-bold text-slate-500 mb-0.5">{subtext}</span>}
        </div>
    </div>
);

export default function SquadLeaderOverview({ data }) {
    const { 
        user = { name: 'Squad Leader', vehicle: '' }, 
        shift_checklists = [], 
        vehicle_history = [], 
        routes = { bulkApprove: '', approveChecklist: '' }, 
        csrfToken = '' 
    } = data || {};
    const [selectedIds, setSelectedIds] = useState([]);

    const totalVehicles = 2; // standard for Squad Leader (Main + Support)
    const checklists = Object.values(shift_checklists || {});
    const totalCount = checklists.length;
    const pendingCount = checklists.filter(c => !c.squad_leader_approved_at).length;
    const servCount = checklists.filter(c => c.status === 'SERVICEABLE').length;
    const unservCount = checklists.filter(c => c.status === 'UNSERVICEABLE').length;

    const toggleAll = (e) => {
        if (e.target.checked) {
            setSelectedIds(checklists.map(c => c.id));
        } else {
            setSelectedIds([]);
        }
    };

    const toggleSelect = (id) => {
        if (selectedIds.includes(id)) {
            setSelectedIds(selectedIds.filter(i => i !== id));
        } else {
            setSelectedIds([...selectedIds, id]);
        }
    };

    return (
        <div className="flex flex-col gap-6 pb-32 max-w-7xl mx-auto w-full px-4 sm:px-6 lg:px-8 pt-6">
            
            {/* User Hero */}
            <div className="premium-card overflow-hidden flex flex-col relative border-l-4 border-l-primary-red">
                <div className="p-6 md:p-8 flex flex-col md:flex-row md:items-center justify-between gap-5 relative bg-white">
                    <div className="flex items-center gap-5">
                        <div className="w-16 h-16 md:w-20 md:h-20 rounded-2xl bg-slate-100 border border-slate-200 p-0.5 shadow-sm shrink-0 overflow-hidden relative">
                            {user?.profile_photo_url ? (
                                <img src={user.profile_photo_url} className="w-full h-full object-cover rounded-xl" />
                            ) : (
                                <div className="w-full h-full rounded-xl bg-slate-50 flex items-center justify-center text-slate-500 font-black text-3xl uppercase">
                                    {user?.name?.substring(0, 1) || '?'}
                                </div>
                            )}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-[10px] md:text-xs font-black text-primary-red uppercase tracking-widest mb-1">Squad Commander</p>
                            <h2 className="text-xl md:text-2xl font-black text-slate-900 leading-tight mb-1">{user?.name || 'Commander'}</h2>
                            <p className="text-xs md:text-sm font-bold text-slate-500 uppercase tracking-widest">{user?.vehicle || 'No Vehicle Assigned'}</p>
                        </div>
                    </div>
                </div>
            </div>

            {/* Stats Row */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <StatBox label="Total Inspections" value={totalCount} subtext={`/ ${totalVehicles}`} colorClass="text-slate-900" />
                <StatBox label="Pending Review" value={pendingCount} subtext={`/ ${totalVehicles}`} colorClass="text-primary-gold" />
                <StatBox label="Serviceable" value={servCount} subtext={`/ ${totalVehicles}`} colorClass="text-emerald-500" />
                <StatBox label="Unserviceable" value={unservCount} subtext={`/ ${totalVehicles}`} colorClass="text-primary-red" />
            </div>

            {/* Checklists Table */}
            <Card title="Daily Inspections Review (Today)" icon={FileText}>
                {checklists.length === 0 ? (
                    <div className="text-center py-10">
                        <div className="w-16 h-16 rounded-full bg-slate-50 flex items-center justify-center mx-auto mb-4">
                            <CheckCircle2 className="w-8 h-8 text-slate-300" />
                        </div>
                        <h4 className="text-lg font-bold text-slate-900 mb-1">No Inspections Pending</h4>
                        <p className="text-slate-500 text-sm">Waiting for personnel to submit daily vehicle checklists.</p>
                    </div>
                ) : (
                    <form action={routes.bulkApprove} method="POST">
                        <input type="hidden" name="_token" value={csrfToken} />
                        
                        {selectedIds.length > 0 && (
                            <div className="mb-4 p-3 bg-slate-50 border border-slate-200 rounded-lg flex items-center justify-between">
                                <span className="text-sm font-bold text-slate-700">{selectedIds.length} checklist(s) selected</span>
                                <button type="submit" className="button-pill bg-primary-red text-white py-1.5 px-4 text-xs">
                                    Approve Selected
                                </button>
                            </div>
                        )}

                        <div className="overflow-x-auto rounded-xl border border-slate-100">
                            <table className="w-full text-left border-collapse">
                                <thead>
                                    <tr className="bg-slate-50">
                                        <th className="py-3 px-4 w-12">
                                            <input type="checkbox" onChange={toggleAll} className="rounded border-slate-300 text-primary-red shadow-sm focus:ring-primary-red" />
                                        </th>
                                        <th className="py-3 px-4 text-[10px] font-black text-slate-500 uppercase tracking-widest">Time</th>
                                        <th className="py-3 px-4 text-[10px] font-black text-slate-500 uppercase tracking-widest">Vehicle</th>
                                        <th className="py-3 px-4 text-[10px] font-black text-slate-500 uppercase tracking-widest">Personnel</th>
                                        <th className="py-3 px-4 text-[10px] font-black text-slate-500 uppercase tracking-widest">Status</th>
                                        <th className="py-3 px-4 text-[10px] font-black text-slate-500 uppercase tracking-widest text-right">Action</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {checklists.map((c, i) => (
                                        <tr key={i} className="border-b border-slate-50 hover:bg-slate-50 transition-colors">
                                            <td className="py-4 px-4">
                                                {!c.squad_leader_approved_at && (
                                                    <input 
                                                        type="checkbox" 
                                                        name="checklist_ids[]" 
                                                        value={c.id} 
                                                        checked={selectedIds.includes(c.id)}
                                                        onChange={() => toggleSelect(c.id)}
                                                        className="rounded border-slate-300 text-primary-red shadow-sm focus:ring-primary-red" 
                                                    />
                                                )}
                                            </td>
                                            <td className="py-4 px-4 font-bold text-slate-700 text-sm">
                                                {new Date(c.check_date).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                                            </td>
                                            <td className="py-4 px-4">
                                                <div className="font-bold text-slate-900 text-sm">{c.vehicle_name}</div>
                                            </td>
                                            <td className="py-4 px-4">
                                                <div className="font-semibold text-slate-700 text-sm">{c.user?.name || 'Unknown'}</div>
                                            </td>
                                            <td className="py-4 px-4">
                                                {c.status === 'SERVICEABLE' ? (
                                                    <span className="px-2 py-1 bg-emerald-50 text-emerald-700 text-[10px] font-bold uppercase tracking-widest rounded-md">Serviceable</span>
                                                ) : (
                                                    <span className="px-2 py-1 bg-red-50 text-red-700 border border-red-200 animate-pulse text-[10px] font-bold uppercase tracking-widest rounded-md">U/S</span>
                                                )}
                                            </td>
                                            <td className="py-4 px-4 text-right">
                                                {c.squad_leader_approved_at ? (
                                                    <span className="text-[10px] font-bold text-emerald-600 uppercase tracking-widest flex items-center justify-end gap-1">
                                                        <CheckCircle2 className="w-3 h-3" /> Approved
                                                    </span>
                                                ) : (
                                                    <a href={routes.approveChecklist + '/' + c.id + '/approve'} className="text-primary-red hover:text-red-700 text-[10px] font-bold uppercase tracking-widest bg-red-50 hover:bg-red-100 px-3 py-1.5 rounded-lg transition-colors">
                                                        Review & Approve
                                                    </a>
                                                )}
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    </form>
                )}
            </Card>
        </div>
    );
}
