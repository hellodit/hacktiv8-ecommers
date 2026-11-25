# ecommers

REST API e-commerce sederhana dengan Go, dipakai sebagai studi kasus sesi **"Bikin REST API Golang, dari Ngoding sampai Load Testing"**. Isinya katalog produk publik, login JWT, dan checkout yang wajib login, lengkap dengan Swagger dan load test k6.

> **Status:** endpoint checkout masih mengembalikan `501 Not Implemented`. Fungsi `reserveStock` di `internal/handler/checkout.go` sengaja dibiarkan TODO sebagai latihan.

## Stack

| Komponen | Versi |
|---|---|
| Go | 1.27 |
| Echo | v4.16 |
| GORM + driver Postgres (pgx v5) | v1.31 / v1.6 |
| PostgreSQL | jalan di host lokal (tidak ada di Docker Compose) |
| JWT | golang-jwt v5, HS256 |
| Swagger | swaggo (echo-swagger) |
| Load test | k6 |

## Persiapan

Yang dibutuhkan: Go 1.27, PostgreSQL yang jalan di mesin lokal, dan (opsional) Docker, k6, serta `swag`.

```bash
# 1. Buat database
createdb ecommers

# 2. Siapkan env, lalu isi JWT_SECRET
cp .env.example .env
openssl rand -hex 32   # tempel hasilnya ke JWT_SECRET di .env

# 3. Jalankan API
make run
```

Saat start, aplikasi otomatis menjalankan **AutoMigrate** dan **seeder** (30 user, 100 produk). Seeder dilewati kalau tabel sudah berisi. Tidak ada perintah migrate atau seed terpisah.

## Environment variables

| Nama | Default | Keterangan |
|---|---|---|
| `JWT_SECRET` | – | **Wajib.** Aplikasi berhenti kalau kosong. |
| `JWT_TTL` | `24h` | Masa berlaku token (format durasi Go) |
| `PORT` | `8080` | Port HTTP |
| `DB_HOST` | `localhost` | |
| `DB_PORT` | `5432` | |
| `DB_USER` | `postgres` | |
| `DB_PASSWORD` | `postgres` | |
| `DB_NAME` | `ecommers` | |
| `DB_SSLMODE` | `disable` | |
| `ACCESS_LOG` | aktif | Isi `false` untuk mematikan request log |

## Perintah

| Perintah | Fungsi |
|---|---|
| `make run` | Jalankan API (`go run ./cmd/api`) |
| `make build` | Build binary ke `bin/api` |
| `make swag` | Generate ulang dokumentasi Swagger ke `docs/` |
| `make up` / `make down` / `make logs` | Jalankan API di Docker Compose |
| `make k6` | Load test (`BASE_URL` default `http://localhost:8080`) |
| `go vet ./...` | Pemeriksaan statis |

`make swag` butuh CLI swag: `go install github.com/swaggo/swag/cmd/swag@latest`.

Belum ada test otomatis (`go test ./...` belum punya file test) dan belum ada linter.

### Docker

Compose hanya menjalankan service `api`. API konek ke Postgres di host lewat `host.docker.internal`, jadi Postgres di host harus menerima koneksi dari jaringan Docker (atur `listen_addresses` dan `pg_hba.conf`). Image dibangun multi-stage, `golang:1.27-alpine` ke `distroless/static-debian12:nonroot`.

## API

Base URL: `http://localhost:8080/api/v1` · Swagger UI: `http://localhost:8080/swagger/index.html`

| Method | Path | Auth | Fungsi | Status |
|---|---|---|---|---|
| POST | `/auth/login` | – | Login, dapat JWT | 200, 400, 401 |
| GET | `/products?page=1&limit=20` | – | List produk (limit maks 100, urut id) | 200 |
| GET | `/products/:id` | – | Detail produk | 200, 400, 404 |
| POST | `/checkout` | Bearer JWT | Beli satu produk, `quantity` 1–100 | 201, 400, 401, 404, 409 (saat ini 501) |

Format error: `{"message": "..."}`.

### Coba cepat

User seed: `user1@example.com` sampai `user30@example.com`, password `password123`.

```bash
# Login
curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user1@example.com","password":"password123"}'
# → {"token":"eyJ...","expires_at":...}

# List produk
curl 'localhost:8080/api/v1/products?page=1&limit=20'

# Checkout (saat ini masih 501)
curl -X POST localhost:8080/api/v1/checkout \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"product_id":1,"quantity":2}'
```

## Data model

- **users**: `id`, `name`, `email` (unique), `password` (hash bcrypt, tidak pernah ikut di JSON), timestamps
- **products**: `id`, `name`, `description`, `image_url`, `price` (int64, rupiah), `stock` (`CHECK stock >= 0`), timestamps
- **orders**: `id`, `user_id` (FK ke users, `ON DELETE RESTRICT`), `product_snapshot` (JSONB), timestamps

`orders` sengaja tidak punya FK ke `products`. Saat checkout, data produk (nama, harga, qty, total) dibekukan ke `product_snapshot`, jadi riwayat order tidak ikut berubah kalau produknya berubah.

### Alur checkout (target implementasi)

JWT middleware → ambil `user_id` dari claim `sub` → `BEGIN` → `SELECT ... FOR UPDATE` baris produk → stok kurang? `409` → kurangi stok → insert order beserta snapshot → `COMMIT` → `201`.

## Load test

`k6/load-test.js` login 30 user sekali di `setup()` lalu memakai ulang tokennya, supaya yang diukur API-nya, bukan bcrypt. Ada dua skenario:

- **browse**: naik 0 ke 50 VU, buka list lalu detail
- **checkout**: 20 VU konstan selama 1 menit 45 detik

Threshold: error < 1%, p95 browse < 300 ms, p95 checkout < 500 ms.

## Struktur

```
cmd/api/main.go       entry point: env → config → DB → migrate → seed → routes → graceful shutdown
internal/config/      baca env (DSN, port, JWT)
internal/database/    koneksi + connection pool + migrate, seeder
internal/model/       User, Product, Order (+ ProductSnapshot JSONB)
internal/dto/         struct request/response
internal/handler/     auth, product, checkout
internal/middleware/  JWT (Bearer → user_id di context)
docs/                 hasil generate swag, jangan diedit manual
k6/                   script load test
```
