# ARFF-OS Self-Hosted Mail Server (Stalwart)

Panduan deployment layanan mail server mandiri (*self-hosted*) menggunakan **[Stalwart Mail Server](https://github.com/stalwartlabs/mail-server)** untuk platform **ARFF-OS**.

Stalwart adalah all-in-one open-source mail server modern berbasis **Rust** yang hemat memori (~35MB RAM), aman, dan memiliki Web Admin Dashboard serta manajemen DKIM/SPF/DMARC bawaan.

---

## 🚀 Panduan Instalasi Cepat

### 1. Persiapan Environment
Salin template konfigurasi:
```bash
cd deploy/mail
cp .env.example .env
```
Edit file `.env` dan sesuaikan `MAIL_HOSTNAME` dengan domain/subdomain mail Anda (contoh: `mail.arff-os.com`).

### 2. Menjalankan Server
Jalankan container menggunakan Docker Compose:
```bash
docker compose up -d
```

### 3. Mendapatkan Password Administrator Awal
Saat pertama kali berjalan, Stalwart akan membuat kredensial administrator awal secara acak:
```bash
docker logs arff-os-mail-server | grep -i "admin"
```
Simpan username (`admin`) dan password yang tertera di log.

### 4. Mengakses Web Admin Console
Buka browser dan akses:
```
https://IP_SERVER_ANDA:8443
```
*(Terima peringatan self-signed SSL certificate jika baru pertama kali membuka).*

Login menggunakan akun admin di atas.

---

## ⚙️ Konfigurasi di Web Admin

1. **Tambahkan Domain**:
   - Masuk ke menu **Management > Directory > Domains**.
   - Tambahkan domain utama Anda (contoh: `arff-os.com`).
2. **Buat Akun Pengirim (Sender Account)**:
   - Masuk ke menu **Management > Directory > Accounts**.
   - Buat akun baru: `no-reply@arff-os.com`.
   - Buat password yang kuat untuk akun tersebut.
3. **Konfigurasi DNS (Anti-Spam)**:
   Di menu domain, Stalwart menyediakan kunci **DKIM**, **SPF**, dan **DMARC**. Pasang record tersebut pada DNS provider Anda (misal: Cloudflare / Namecheap) agar email OTP tidak masuk ke folder spam.

---

## 🔌 Menghubungkan ke Backend ARFF-OS

Buka file `.env` pada root atau `backend/.env` ARFF-OS Anda, lalu sesuaikan:

```env
# ========================================================
# Email SMTP Configuration (Stalwart Mail Server)
# ========================================================
SMTP_HOST=IP_SERVER_STALWART      # atau nama domain (misal: mail.arff-os.com)
SMTP_PORT=587                     # Port submission STARTTLS
SMTP_USER=no-reply@arff-os.com
SMTP_PASSWORD=password-dari-stalwart
SMTP_FROM_NAME="ARFF-OS Airport Security"
SMTP_FROM_EMAIL=no-reply@arff-os.com
```

Backend Golang ARFF-OS akan secara otomatis mengirimkan email verifikasi OTP 6-digit melalui server Stalwart Anda.

---

## 🔒 Catatan Keamanan & Privasi
- Direktori `./data` tempat penyimpanan konfigurasi, kunci enkripsi DKIM, dan pesan email telah secara ketat dimasukkan ke dalam `.gitignore` repositori sehingga **tidak akan pernah bocor ke GitHub**.
- Jangan pernah meng-commit file `.env` yang berisi password atau secret keys ke repository publik.
