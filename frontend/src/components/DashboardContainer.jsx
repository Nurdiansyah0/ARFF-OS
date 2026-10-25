'use client';

import React, { useState, useEffect } from 'react';
import { Activity, Shield, Users, Compass, Server, CheckCircle2, AlertCircle } from 'lucide-react';
import ExecutiveOverview from './dashboard/ExecutiveOverview';
import OperationsOverview from './dashboard/OperationsOverview';
import DirectorOverview from './dashboard/DirectorOverview';
import SquadLeaderOverview from './dashboard/SquadLeaderOverview';

export default function DashboardContainer() {
  const [activeTab, setActiveTab] = useState('executive');
  const [backendStatus, setBackendStatus] = useState({ online: false, checking: true, message: '' });

  useEffect(() => {
    const checkBackend = async () => {
      try {
        const res = await fetch('/api/health');
        if (res.ok) {
          const data = await res.json();
          setBackendStatus({ online: true, checking: false, message: data.message || 'Go API Connected' });
        } else {
          setBackendStatus({ online: false, checking: false, message: 'Backend unreachable' });
        }
      } catch (err) {
        setBackendStatus({ online: false, checking: false, message: 'API Offline' });
      }
    };

    checkBackend();
    const interval = setInterval(checkBackend, 15000);
    return () => clearInterval(interval);
  }, []);

  const tabs = [
    { id: 'executive', name: 'Executive Overview', icon: Shield },
    { id: 'operations', name: 'Operations Overview', icon: Activity },
    { id: 'director', name: 'Director Overview', icon: Compass },
    { id: 'squad_leader', name: 'Squad Leader Overview', icon: Users },
  ];

  return (
    <div className="min-h-screen flex flex-col bg-slate-50">
      {/* Header */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-30 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center gap-3">
              <img src="/Logo_hangnadim.png" alt="Hang Nadim Logo" className="h-9 w-auto" />
              <div>
                <span className="font-extrabold text-slate-900 tracking-tight text-lg">ARFF-OS</span>
                <span className="hidden sm:inline-block ml-2 text-xs font-semibold px-2 py-0.5 rounded bg-primary-red/10 text-primary-red">
                  Hang Nadim Airport
                </span>
              </div>
            </div>

            {/* Backend Status indicator */}
            <div className="flex items-center gap-2 text-xs font-medium px-3 py-1.5 rounded-full border bg-slate-50">
              <Server className="w-3.5 h-3.5 text-slate-500" />
              <span className="text-slate-600 hidden md:inline">Go Backend (Railway):</span>
              {backendStatus.checking ? (
                <span className="text-slate-400">Connecting...</span>
              ) : backendStatus.online ? (
                <span className="flex items-center gap-1 text-emerald-600 font-semibold">
                  <CheckCircle2 className="w-3.5 h-3.5" /> Online
                </span>
              ) : (
                <span className="flex items-center gap-1 text-amber-600 font-semibold">
                  <AlertCircle className="w-3.5 h-3.5" /> Standby
                </span>
              )}
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav className="flex space-x-2 sm:space-x-4 overflow-x-auto pb-2 sm:pb-0" aria-label="Tabs">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              const isActive = activeTab === tab.id;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 py-3 px-3 border-b-2 font-medium text-xs sm:text-sm whitespace-nowrap transition-colors ${
                    isActive
                      ? 'border-primary-red text-primary-red'
                      : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  {tab.name}
                </button>
              );
            })}
          </nav>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {activeTab === 'executive' && <ExecutiveOverview />}
        {activeTab === 'operations' && <OperationsOverview />}
        {activeTab === 'director' && <DirectorOverview />}
        {activeTab === 'squad_leader' && <SquadLeaderOverview />}
      </main>

      {/* Footer */}
      <footer className="bg-white border-t border-slate-200 py-4 text-center text-xs text-slate-500">
        &copy; {new Date().getFullYear()} Hang Nadim International Airport — ARFF Operating System (ARFF-OS)
      </footer>
    </div>
  );
}
