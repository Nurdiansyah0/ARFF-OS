'use client';

import React from 'react';
import dynamic from 'next/dynamic';
const Chart = dynamic(() => import('react-apexcharts'), { ssr: false });

import { 
  Activity, CheckCircle2, AlertTriangle, XCircle, 
  Clock, ShieldAlert, PlaneTakeoff, PlaneLanding, Users, 
  Settings, Battery, Target, TrendingUp, AlertOctagon, Info
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
    <div className="flex-1 p-5 overflow-hidden flex flex-col">
      {children}
    </div>
  </div>
);

// Fleet Readiness Radial Chart
const FleetReadiness = ({ data }) => {
  const readyVal = data?.total > 0 ? Math.round((data.ready / data.total) * 100) : 0;
  const vorVal = data?.total > 0 ? Math.round((data.vor / data.total) * 100) : 0;
  const unknownVal = data?.total > 0 ? Math.round((data.unknown / data.total) * 100) : 0;

  const options = {
    chart: { type: 'radialBar', background: 'transparent' },
    colors: ['#16A34A', '#DC2626', '#EAB308'], // Semantic Success, Danger, Warning
    plotOptions: {
      radialBar: {
        hollow: { size: '40%', background: 'transparent' },
        track: { background: '#F1F5F9', strokeWidth: '100%', margin: 5 },
        dataLabels: {
          name: { fontSize: '12px', color: '#64748B', offsetY: -10, fontWeight: 600 },
          value: { fontSize: '28px', color: '#0F172A', fontWeight: 800, offsetY: 10, formatter: (val) => val + "%" },
          total: { show: true, label: 'READY', color: '#16A34A', fontSize: '13px', fontWeight: 700, formatter: () => readyVal + "%" }
        }
      }
    },
    labels: ['Ready', 'Unserviceable', 'Unchecked'],
    stroke: { lineCap: 'round' },
    theme: { mode: 'light' }
  };

  return (
    <Card title="Fleet Readiness" icon={Battery}>
      <div className="h-full flex items-center justify-center min-h-[260px]">
        <Chart options={options} series={[readyVal, vorVal, unknownVal]} type="radialBar" height="100%" width="100%" />
      </div>
    </Card>
  );
};

const ActiveDuty = ({ data }) => (
  <Card title="Active Duty Status" icon={Users}>
    <div className="grid grid-cols-2 gap-3 mb-4">
      <div className="bg-slate-50 rounded-lg p-3 border border-slate-200">
        <p className="text-[11px] text-slate-500 uppercase tracking-widest font-bold">Shift On Duty</p>
        <p className="text-lg font-bold text-slate-900 mt-1">{data?.shift} <span className="text-xs text-slate-500 font-medium">({data?.period})</span></p>
      </div>
      <div className="bg-slate-50 rounded-lg p-3 border border-slate-200">
        <p className="text-[11px] text-slate-500 uppercase tracking-widest font-bold">Team Leader</p>
        <p className="text-sm font-bold text-slate-900 mt-1 truncate">{data?.team_leader}</p>
      </div>
      <div className="bg-slate-50 rounded-lg p-3 border border-slate-200">
        <p className="text-[11px] text-slate-500 uppercase tracking-widest font-bold">Personnel</p>
        <p className="text-lg font-bold text-semantic-success mt-1">{data?.personnel_count} <span className="text-xs text-slate-500">Active</span></p>
      </div>
      <div className="bg-slate-50 rounded-lg p-3 border border-slate-200">
        <p className="text-[11px] text-slate-500 uppercase tracking-widest font-bold">On Leave</p>
        <p className={`text-lg font-bold mt-1 ${data?.leave_count > 0 ? 'text-semantic-warning' : 'text-slate-700'}`}>{data?.leave_count}</p>
      </div>
    </div>
    <div className="flex-1 flex flex-col min-h-0">
      <p className="text-[11px] text-slate-500 uppercase tracking-widest mb-2 font-bold">Assigned Drivers</p>
      <div className="space-y-2 overflow-y-auto pr-2 custom-scrollbar">
        {data?.drivers?.length > 0 ? data.drivers.map((driver, i) => (
          <div key={i} className="flex justify-between items-center text-sm py-1.5 px-3 rounded-md bg-white border border-slate-200 shadow-sm">
            <span className={`font-semibold ${driver.is_unserviceable ? 'text-slate-400 line-through' : 'text-slate-700'}`}>{driver.vehicle}</span>
            {driver.is_unserviceable ? (
              <span className="text-semantic-danger text-xs font-bold uppercase tracking-wider">Unserviceable</span>
            ) : (
              <span className="text-semantic-info font-bold text-xs">{driver.name}</span>
            )}
          </div>
        )) : (
          <div className="text-sm text-slate-500 italic py-2">No assigned drivers</div>
        )}
      </div>
    </div>
  </Card>
);

const TLPerformance = ({ data }) => (
  <Card title="Team Leader Ops (7 Days)" icon={Target}>
    <div className="flex flex-col gap-3 h-full overflow-y-auto pr-2 custom-scrollbar">
      {data?.map((tl, i) => (
        <div key={i} className="bg-white border border-slate-200 rounded-lg p-4 flex items-center justify-between hover:bg-slate-50 transition-colors shadow-sm">
          <div className="flex items-center gap-4">
            <img src={tl.photo_url || `https://ui-avatars.com/api/?name=${tl.tl_name}&background=E2E8F0&color=0F172A`} alt={tl.tl_name} className="w-10 h-10 rounded-full border border-slate-300" />
            <div>
              <h5 className="text-sm font-bold text-slate-900">{tl.tl_name} <span className="text-xs text-slate-500 font-medium">({tl.shift})</span></h5>
              <div className="flex items-center gap-3 mt-1.5">
                <span className="flex items-center gap-1.5 text-xs font-semibold text-slate-600"><div className="w-1.5 h-1.5 rounded-full bg-semantic-success"></div>{tl.approved}</span>
                <span className="flex items-center gap-1.5 text-xs font-semibold text-slate-600"><div className="w-1.5 h-1.5 rounded-full bg-semantic-warning"></div>{tl.pending}</span>
                <span className="flex items-center gap-1.5 text-xs font-semibold text-slate-600"><div className="w-1.5 h-1.5 rounded-full bg-semantic-danger"></div>{tl.missing}</span>
              </div>
            </div>
          </div>
          <div className="text-right">
            <span className={`text-xl font-black ${tl.score >= 90 ? 'text-semantic-success' : tl.score >= 75 ? 'text-semantic-warning' : 'text-semantic-danger'}`}>{tl.score}%</span>
            <div className="text-[10px] text-slate-500 mt-1 font-bold uppercase tracking-wider">CHK {tl.checklist_score}% • FIT {tl.samapta_score}%</div>
          </div>
        </div>
      ))}
    </div>
  </Card>
);

const TrendChart = ({ title, icon, data, type, color, target = null, yTitle = '' }) => {
  const options = {
    chart: { type: type, toolbar: { show: false }, background: 'transparent' },
    colors: [color],
    fill: type === 'area' ? { type: 'gradient', gradient: { shadeIntensity: 1, opacityFrom: 0.2, opacityTo: 0.01, stops: [0, 100] } } : {},
    dataLabels: { enabled: false },
    stroke: { curve: 'smooth', width: 3 },
    xaxis: { 
      categories: data?.categories || data?.map(d => `Month ${d.month}`) || [], 
      labels: { style: { colors: '#64748B', fontSize: '11px', fontWeight: 500 } },
      axisBorder: { show: false }, axisTicks: { show: false }
    },
    yaxis: { labels: { style: { colors: '#64748B', fontSize: '11px', fontWeight: 500 } } },
    grid: { borderColor: '#E2E8F0', strokeDashArray: 4, position: 'back' },
    theme: { mode: 'light' },
    annotations: target ? {
      yaxis: [{ y: target, borderColor: '#DC2626', label: { style: { color: '#fff', background: '#DC2626', fontWeight: 600, padding: { left: 5, right: 5, top: 2, bottom: 2 } }, text: `Target (${target})` } }]
    } : {}
  };

  const series = [{ name: yTitle, data: data?.data || data?.map(d => d.score) || [] }];

  return (
    <Card title={title} icon={icon}>
      <div className="h-full min-h-[250px]">
        <Chart options={options} series={series} type={type} height="100%" width="100%" />
      </div>
    </Card>
  );
}

const DataTable = ({ columns, data, renderRow, emptyMessage }) => (
  <div className="overflow-x-auto">
    <table className="w-full text-left border-collapse">
      <thead>
        <tr className="bg-slate-50 border-b border-slate-200">
          {columns.map((col, i) => (
            <th key={i} className={`py-3 px-4 text-[11px] font-bold text-slate-500 uppercase tracking-widest ${col.className || ''}`}>{col.label}</th>
          ))}
        </tr>
      </thead>
      <tbody className="text-sm">
        {data?.length > 0 ? data.map((row, i) => renderRow(row, i)) : (
          <tr>
            <td colSpan={columns.length} className="py-8 text-center text-slate-500 text-sm font-medium">{emptyMessage}</td>
          </tr>
        )}
      </tbody>
    </table>
  </div>
);

export default function ExecutiveOverview({ 
  fleet_status, current_duty, tl_performance, 
  recent_flights, recent_emergencies, recent_incidents,
  defective_parts_trend, response_time_trend, reliability_trend, activeIncident 
}) {
  return (
    <div className="bg-slate-950 -mt-6 -mx-4 sm:-mx-6 lg:-mx-8 px-4 sm:px-6 lg:px-8 py-8 min-h-screen">
      <div className="flex flex-col gap-6 max-w-[1400px] mx-auto w-full h-full">
      
      {/* OSC Active Incident Alert */}
      {activeIncident && (
        <div className="bg-red-50 border-l-4 border-semantic-danger rounded-r-xl p-5 flex items-center gap-4 shadow-sm">
          <AlertOctagon className="w-8 h-8 text-semantic-danger animate-pulse" />
          <div>
            <h2 className="text-sm font-black text-semantic-danger uppercase tracking-widest">Active OSC Command Required</h2>
            <p className="text-sm text-red-800 mt-1 font-medium">Incident: <strong>{activeIncident.title}</strong>. Immediate attention required.</p>
          </div>
        </div>
      )}

      {/* Top Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 h-auto lg:h-[400px]">
        <FleetReadiness data={fleet_status} />
        <ActiveDuty data={current_duty} />
        <TLPerformance data={tl_performance} />
      </div>

      {/* Middle Row: Trend Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {defective_parts_trend && <TrendChart title="Defective Parts Trend" icon={Settings} data={defective_parts_trend} type="bar" color="#E3A008" yTitle="Reports" />}
        {response_time_trend && <TrendChart title="Response Time (Minutes)" icon={Clock} data={response_time_trend} type="area" color="#0F172A" target={3} yTitle="Minutes" />}
      </div>
      
      {/* Reliability Full Width */}
      {reliability_trend && (
        <div className="w-full">
          <TrendChart title="Fleet Uptime (Monthly Reliability)" icon={Activity} data={reliability_trend} type="area" color="#2563EB" yTitle="Score %" />
        </div>
      )}

      {/* Tables Section */}
      <div className="grid grid-cols-1 gap-6">
        
        <Card title="Watchroom Movement Logs (Live)" icon={PlaneTakeoff}>
          <DataTable 
            columns={[
              { label: 'Time' }, { label: 'Flight' }, { label: 'Type' }, 
              { label: 'Route' }, { label: 'Rwy' }, { label: 'Status', className: 'text-right' }
            ]}
            data={recent_flights}
            emptyMessage="No recent flights recorded"
            renderRow={(flight, i) => (
              <tr key={i} className="border-b border-slate-100 hover:bg-blue-50 transition-colors">
                <td className="py-3 px-4 font-bold text-slate-700">{new Date(flight.movement_time).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</td>
                <td className="py-3 px-4">
                  <div className="font-bold text-slate-900">{flight.flight_number}</div>
                  <div className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">{flight.operator}</div>
                </td>
                <td className="py-3 px-4 text-slate-600 font-medium">{flight.aircraft_type}</td>
                <td className="py-3 px-4 text-slate-700 font-bold">{flight.origin_iata} → {flight.destination_iata}</td>
                <td className="py-3 px-4 text-slate-500 font-semibold">{flight.runway_used}</td>
                <td className="py-3 px-4 text-right">
                  <span className={`px-2 py-1 rounded-md text-[10px] font-bold uppercase tracking-widest ${flight.movement_type === 'ARR' ? 'bg-green-100 text-green-700' : 'bg-yellow-100 text-yellow-700'}`}>
                    {flight.movement_type === 'ARR' ? 'Arrival' : 'Departure'}
                  </span>
                </td>
              </tr>
            )}
          />
        </Card>

        <Card title="Emergency & Standbys (Recent 12H)" icon={ShieldAlert}>
          <DataTable 
            columns={[
              { label: 'Received' }, { label: 'Operator / AC' }, { label: 'Trouble' }, 
              { label: 'ARFF Action' }, { label: 'Status', className: 'text-right' }
            ]}
            data={recent_emergencies}
            emptyMessage="No emergencies in the last 12 hours"
            renderRow={(em, i) => (
              <tr key={i} className="border-b border-slate-100 hover:bg-blue-50 transition-colors">
                <td className="py-3 px-4 font-bold text-slate-700">{new Date(em.received_at_lt).toLocaleString([], {day: '2-digit', month: 'short', hour: '2-digit', minute:'2-digit'})}</td>
                <td className="py-3 px-4">
                  <div className="font-bold text-slate-900">{em.operator}</div>
                  <div className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">{em.aircraft_type} (POB: {em.pob || '-'})</div>
                </td>
                <td className="py-3 px-4 text-semantic-danger font-semibold">{em.trouble_type?.substring(0, 40)}</td>
                <td className="py-3 px-4 text-primary-gold text-[10px] font-black uppercase tracking-widest">{em.arff_action?.replace(/_/g, ' ')}</td>
                <td className="py-3 px-4 text-right">
                  <span className={`px-2 py-1 rounded-md text-[10px] font-bold uppercase tracking-widest ${em.status === 'active' ? 'bg-red-100 text-red-700 animate-pulse border border-red-200' : 'bg-slate-100 text-slate-600'}`}>
                    {em.status}
                  </span>
                </td>
              </tr>
            )}
          />
        </Card>

      </div>
    </div>
    </div>
  );
}
