-- ====================================================================
-- ARFF-OS Migration 002: Fleet Management, Extinguishing Agents,
-- ICAO Level of Protection (LoP) & Multi-Tier Approval Matrix
-- ====================================================================

-- 1. VEHICLES TABLE (Armada Pemadam Bandara terisolasi per airport_id)
CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    airport_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    call_sign VARCHAR(50) NOT NULL,              -- e.g. 'Foam Tender 01', 'RIV 01', 'Nurse Tender 01'
    registration_number VARCHAR(50),             -- Nomor Plat Dinas Bandara
    vehicle_type VARCHAR(50) NOT NULL,          -- 'foam_tender', 'riv', 'nurse_tender', 'ambulance', 'command_car', 'rescue_boat'
    brand_chassis VARCHAR(100),                  -- e.g. 'Rosenbauer Panther 6x6', 'Oshkosh Striker 3000'
    year_manufactured INT,
    
    -- Extinguishing Agent Capacities (Liter / kg)
    water_capacity INT NOT NULL DEFAULT 0,       -- Kapasitas Tangki Air (L)
    foam_capacity INT NOT NULL DEFAULT 0,        -- Kapasitas Tangki Busa AFFF (L)
    dcp_capacity INT NOT NULL DEFAULT 0,         -- Kapasitas Dry Chemical Powder (kg)
    co2_capacity INT NOT NULL DEFAULT 0,         -- Kapasitas CO2 (kg)
    
    -- Pumping & Discharge Rates (ICAO Annex 14)
    pump_discharge_rate INT NOT NULL DEFAULT 0,  -- Laju Pancaran Pompa Utama (L/min)
    roof_turret_rate INT NOT NULL DEFAULT 0,     -- Laju Pancaran Roof Turret (L/min)
    bumper_turret_rate INT NOT NULL DEFAULT 0,   -- Laju Pancaran Bumper Turret (L/min)
    under_truck_nozzles BOOLEAN DEFAULT TRUE,    -- Nozel Perlindungan Kolong Sasis
    
    -- Real-time Operational Status & Current Agent Quantities
    status VARCHAR(50) NOT NULL DEFAULT 'serviceable', -- 'serviceable', 'limited', 'vor' (Vehicle Off Road)
    current_water_level INT NOT NULL DEFAULT 0,
    current_foam_level INT NOT NULL DEFAULT 0,
    current_dcp_level INT NOT NULL DEFAULT 0,
    current_n2_pressure NUMERIC(6,2) DEFAULT 0.0,      -- Tekanan Nitrogen pendorong (bar)
    
    -- Telemetry & Usage
    odometer_km NUMERIC(10,2) DEFAULT 0.0,
    pump_hours NUMERIC(10,2) DEFAULT 0.0,
    notes TEXT,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vehicles_airport ON vehicles(airport_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status);
CREATE INDEX IF NOT EXISTS idx_vehicles_type ON vehicles(vehicle_type);

-- 2. EXTINGUISHING AGENT STOCKS (Gudang & Cadangan Logistik Bahan Pemadam)
CREATE TABLE IF NOT EXISTS extinguishing_agent_stocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    airport_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    agent_type VARCHAR(50) NOT NULL,            -- 'water', 'foam_afff_3', 'foam_afff_6', 'dcp_powder', 'co2', 'halotron'
    storage_location VARCHAR(100) NOT NULL,      -- e.g. 'Main Station Bulk Tank', 'Warehouse Drum B'
    current_quantity NUMERIC(12,2) NOT NULL DEFAULT 0.0,
    minimum_reserve_threshold NUMERIC(12,2) NOT NULL DEFAULT 0.0, -- Ambang batas cadangan 200% ICAO
    unit VARCHAR(20) NOT NULL DEFAULT 'liters', -- 'liters', 'kg'
    batch_number VARCHAR(100),
    expiry_date DATE,
    last_tested_at TIMESTAMPTZ,                  -- Uji laboratorium mutu busa berkala
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agent_stocks_airport ON extinguishing_agent_stocks(airport_id);

-- 3. VEHICLE DAILY INSPECTIONS (Digital Checklist & Test Fungsional)
CREATE TABLE IF NOT EXISTS vehicle_inspections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    airport_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    inspector_id UUID NOT NULL REFERENCES users(id),
    shift VARCHAR(20) NOT NULL,                  -- 'pagi', 'siang', 'malam'
    regu VARCHAR(10) NOT NULL,                   -- 'Regu A', 'Regu B', 'Regu C', 'Regu D'
    
    -- Checklist Item Statuses (PASS / FAIL / WARN)
    engine_status VARCHAR(20) DEFAULT 'pass',
    brake_status VARCHAR(20) DEFAULT 'pass',
    transmission_status VARCHAR(20) DEFAULT 'pass',
    lighting_siren_status VARCHAR(20) DEFAULT 'pass',
    pump_test_status VARCHAR(20) DEFAULT 'pass',
    turret_test_status VARCHAR(20) DEFAULT 'pass',
    under_truck_test_status VARCHAR(20) DEFAULT 'pass',
    scba_readiness_status VARCHAR(20) DEFAULT 'pass',
    extrication_tools_status VARCHAR(20) DEFAULT 'pass',
    
    -- Metric Observations
    water_level_liters INT NOT NULL,
    foam_level_liters INT NOT NULL,
    dcp_level_kg INT NOT NULL,
    fuel_level_percent INT NOT NULL,
    engine_oil_pressure NUMERIC(6,2),
    pump_pressure_bar NUMERIC(6,2),
    odometer_km NUMERIC(10,2),
    pump_hours NUMERIC(10,2),
    
    notes TEXT,
    overall_readiness VARCHAR(50) NOT NULL DEFAULT 'serviceable', -- 'serviceable', 'limited', 'vor'
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inspections_airport ON vehicle_inspections(airport_id);
CREATE INDEX IF NOT EXISTS idx_inspections_vehicle ON vehicle_inspections(vehicle_id);

-- 4. MULTI-TIER APPROVALS TABLE (Workflow Persetujuan Dokumen Operasional)
CREATE TABLE IF NOT EXISTS approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    airport_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,           -- 'vehicle_inspection', 'lop_degradation', 'incident_report', 'shift_handover'
    entity_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Workflow Stages:
    -- 'submitted' -> 'squad_leader_reviewed' -> 'supervisor_validated' -> 'chief_approved' / 'rejected'
    current_stage VARCHAR(50) NOT NULL DEFAULT 'submitted',
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    
    submitted_by UUID NOT NULL REFERENCES users(id),
    current_assignee_role VARCHAR(50) NOT NULL,   -- 'squad_leader', 'operations_supervisor', 'head_of_department'
    
    -- Cryptographic Integrity & Digital Signatures
    document_payload_hash VARCHAR(64) NOT NULL,   -- SHA-256 Hash of submitted data
    final_signature_hash VARCHAR(128),            -- HMAC SHA-256 seal signed by Chief ARFF
    signed_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_approvals_airport ON approvals(airport_id);
CREATE INDEX IF NOT EXISTS idx_approvals_status ON approvals(status);
CREATE INDEX IF NOT EXISTS idx_approvals_stage ON approvals(current_stage);

-- 5. APPROVAL ACTIONS (Audit Trail Riwayat Tanda Tangan & Komentar Tiap Tingkat)
CREATE TABLE IF NOT EXISTS approval_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    approval_id UUID NOT NULL REFERENCES approvals(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES users(id),
    actor_name VARCHAR(255) NOT NULL,
    actor_role VARCHAR(50) NOT NULL,
    
    action VARCHAR(50) NOT NULL,                  -- 'submit', 'review', 'validate', 'approve', 'reject'
    stage_from VARCHAR(50) NOT NULL,
    stage_to VARCHAR(50) NOT NULL,
    notes TEXT,
    signature_token VARCHAR(128) NOT NULL,        -- Cryptographic Signature of the Step
    
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_approval_actions ON approval_actions(approval_id);
