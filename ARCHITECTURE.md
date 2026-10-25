# ARFF-OS — Arsitektur Sistem & Spesifikasi Teknis

## 1. Ikhtisar Arsitektur

**ARFF-OS** (Airport Rescue & Firefighting Operating System) Bandar Udara Internasional Hang Nadim dibangun menggunakan arsitektur decoupled modern:

- **Frontend (Vercel)**: Next.js 14 App Router, React 18, Tailwind CSS, Lucide icons, dan dynamic charting via Apexcharts.
- **Backend (Railway)**: REST API performa tinggi berbasis Golang dengan koneksi pooler ke PostgreSQL.
- **Database (Supabase)**: Managed PostgreSQL database dengan connection pooler mode.

---

## 2. Struktur Modul & Layanan

### Backend (Golang)
- `cmd/api`: Entrypoint server HTTP dengan CORS dan graceful shutdown.
- `internal/config`: Parser environment variable (Port, Supabase connection string).
- `internal/database`: PostgreSQL connection pooler (Supabase).
- `internal/handler`: HTTP Handlers untuk health check dan data dashboard.
- `internal/model`: Domain entity data model.

### Frontend (Next.js)
- `src/app`: Root layout dan page render.
- `src/components`: UI Dashboards (Executive, Operations, Director, Squad Leader).
- `next.config.mjs`: Rewrites `/api/*` langsung ke Railway backend URL.
