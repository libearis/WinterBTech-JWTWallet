# JWT Auth & Wallet (Go + Vue 3)

Aplikasi full-stack tangguh yang dilengkapi autentikasi JWT, fitur auto-logout 15 menit, serta transfer dompet digital dengan perlindungan idempotensi (aman dari transfer ganda saat jaringan tidak stabil).

---

## ⚡ Akun Demo Siap Pakai

Sudah tersedia akun siap uji dengan saldo awal **Rp 1.000,00** (100.000 minor unit):

| Email | Password | Keterangan |
|---|---|---|
| `jane@example.com` | `S3cure!pass` | Akun pengirim demo |
| `john@example.com` | `S3cure!pass` | Akun penerima demo |

> Kamu juga bisa klik **Sign up** pada halaman login untuk mendaftar akun baru kapan saja.

---

## 🚀 Cara Menjalankan

### Opsi A — Full Docker (Paling Mudah)
Jalankan Database, Backend, dan Frontend sekaligus dalam satu perintah:

```bash
# 1. Salin konfigurasi environment
cp .env.example .env

# 2. Build dan jalankan container
docker compose up --build
```

- **Frontend (Web)**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:3000/api](http://localhost:3000/api)
- **Health Check**: [http://localhost:3000/api/healthz](http://localhost:3000/api/healthz)

---

### Opsi B — Development Lokal (Host)

1. **Jalankan Database di Docker:**
   ```bash
   docker compose -f docker-compose.db.yml up -d
   ```

2. **Jalankan Backend (Go):**
   ```bash
   cd backend
   go run ./cmd/api
   ```
   *Berjalan di `http://localhost:8080`*

3. **Jalankan Frontend (Vite):**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   *Buka di `http://localhost:5173`*

---

## 🗄️ Konfigurasi Database (PostgreSQL)

Kredensial bawaan untuk container database:

- **Host**: `localhost` (Opsi B) / `db` (Opsi A Docker internal)
- **Port**: `5432`
- **Username**: `wallet_user`
- **Password**: `change-me`
- **Database**: `wallet_db`

Akses langsung melalui CLI:
```bash
docker exec -it wallet_db psql -U wallet_user -d wallet_db
```

---

## 🛡️ Fitur Unggulan

- **JWT Session & Auto Logout**: Sesi aktif otomatis keluar setelah 15 menit tidak ada aktivitas.
- **Pinia State Persistence**: Sesi login dan riwayat tersimpan aman di Pinia, tidak hilang saat browser di-refresh.
- **Idempotent Transfer**: Setiap transaksi menggunakan `Idempotency-Key` unik sehingga pengulangan transfer (retry) di jaringan buruk tidak akan memotong saldo dua kali.
- **Offline Resilient**: Notifikasi status offline dan data lokal tetap dapat dilihat meski tanpa koneksi internet.

---

## 🧪 Menjalankan Unit Test

Untuk menguji seluruh logika backend:

```bash
cd backend
go test ./... -v
```
