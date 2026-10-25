-- ========================================================
-- ARFF-OS Multi-Tenant SaaS Schema (PostgreSQL / Supabase)
-- ========================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. AIRPORTS (Tenant / Organisasi Bandara)
CREATE TABLE IF NOT EXISTS airports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    iata_code VARCHAR(10),
    icao_code VARCHAR(10),
    category INT DEFAULT 9, -- ICAO ARFF Category (1 - 10)
    country VARCHAR(100) DEFAULT 'ID',
    timezone VARCHAR(50) DEFAULT 'Asia/Jakarta',
    plan VARCHAR(50) DEFAULT 'trial', -- 'trial', 'pro', 'enterprise'
    status VARCHAR(50) DEFAULT 'active', -- 'active', 'suspended'
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_airports_icao ON airports(icao_code);
CREATE INDEX IF NOT EXISTS idx_airports_iata ON airports(iata_code);

-- 2. USERS (Personel / Pengguna yang terikat ke Bandara)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    airport_id UUID NOT NULL REFERENCES airports(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    nip VARCHAR(50),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'airport_admin', -- 'airport_admin', 'watchroom_officer', 'squad_leader', 'firefighter'
    phone VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_airport ON users(airport_id);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
