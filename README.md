# ARFF-OS — Airport Rescue & Firefighting Operating System

**ARFF-OS** (Airport Rescue & Firefighting Operating System) untuk Bandar Udara Internasional Hang Nadim menggunakan arsitektur modern terdesentralisasi:

```
┌─────────────────────────────────┐
│     Frontend (Next.js 14)       │ ──> Host di Vercel (Edge Network)
└─────────────────────────────────┘
                 │
                 ▼ API Proxy / CORS
┌─────────────────────────────────┐
│      Backend (Golang API)       │ ──> Host di Railway (Containerized Service)
└─────────────────────────────────┘
                 │
                 ▼ PostgreSQL Connection Pooler
┌─────────────────────────────────┐
│      Database (Supabase)        │ ──> PostgreSQL Serverless Database
└─────────────────────────────────┘
```

---

## 📁 Struktur Monorepo

```
arff-os/
├── backend/                  # GOLANG BACKEND SERVICE (Railway target)
│   ├── cmd/
│   │   └── api/
│   │       └── main.go       # Entrypoint REST API & Graceful shutdown
│   ├── internal/
│   │   ├── config/           # Environment loader (PORT, DATABASE_URL)
│   │   ├── database/         # Supabase PostgreSQL connection pooler (lib/pq)
│   │   ├── handler/          # HTTP Handlers (Healthcheck, Dashboard stats)
│   │   └── model/            # Data models & entities
│   ├── Dockerfile            # Multi-stage lightweight Docker image (~15MB)
│   ├── go.mod                # Go module: github.com/Nurdiansyah0/arff-os-backend
│   └── railway.toml          # Konfigurasi deployment Railway
│
├── frontend/                 # NEXT.JS FRONTEND (Vercel target)
│   ├── src/
│   │   ├── app/              # Next.js App Router (layout.jsx, page.jsx, globals.css)
│   │   └── components/       # Dashboard components (Executive, Operations, Director, Squad Leader)
│   ├── public/               # Static assets & airport logo
│   ├── package.json          # Package: arff-os-frontend
│   ├── next.config.mjs       # Next.js config with API proxy rewrites to Railway
│   └── tailwind.config.js    # Tailwind styling tokens
│
├── .github/                  # CI Pipeline (Go Vet/Build & Next.js Build)
├── .gitignore                # Aturan ignore untuk Go, Next.js, dan Node
└── .env.example              # Template environment variable
```

---

## 🚀 Panduan Deployment

### 1. Database (Supabase)
1. Buat project baru di [supabase.com](https://supabase.com).
2. Salin **Connection String URI** dari menu **Project Settings > Database > Connection Pooling**:
   ```
   postgres://postgres.[PROJECT_REF]:[PASSWORD]@aws-0-[REGION].pooler.supabase.com:6543/postgres?sslmode=require
   ```

### 2. Backend (Railway)
1. Hubungkan repository GitHub ke project di [railway.com](https://railway.com).
2. Buat Service baru dari repo dengan **Root Directory**: `/backend`.
3. Tambahkan Environment Variables di Railway:
   - `PORT`: `8080` (Railway mengisinya secara otomatis)
   - `DATABASE_URL`: `[Connection String Supabase Anda]`
   - `APP_ENV`: `production`
4. Railway akan build dan deploy menggunakan `backend/Dockerfile` & `backend/railway.toml`.

### 3. Frontend (Vercel)
1. Import repository ini di dashboard [vercel.com](https://vercel.com).
2. Atur **Root Directory** ke `frontend`.
3. Framework Preset akan otomatis terdeteksi sebagai **Next.js**.
4. Tambahkan Environment Variable di Vercel:
   - `BACKEND_API_URL`: URL public service Railway Anda (contoh: `https://arff-os-backend.up.railway.app`).
5. Klik **Deploy**.
