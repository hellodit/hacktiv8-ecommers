# AGENTS.md

Panduan untuk AI coding agent yang bekerja di repo ini. Baca sampai habis sebelum mengubah kode. Gambaran umum project ada di `README.md`.

## Konteks

REST API e-commerce (Go 1.27, Echo v4, GORM, PostgreSQL, JWT HS256) untuk demo sesi bootcamp. Kode sengaja dibuat **minimal dan mudah dibaca peserta**. Kesederhanaan lebih penting daripada abstraksi.

## Aturan wajib

1. **Jangan implementasikan atau ubah `reserveStock`** di `internal/handler/checkout.go`. Itu latihan untuk user. Selama masih TODO, `POST /api/v1/checkout` memang mengembalikan 501, dan itu bukan bug.
2. **Jangan commit `.env`** atau secret apa pun. Variabel baru ditambahkan ke `.env.example` dengan nilai placeholder.
3. **Jangan edit `docs/` secara manual.** Folder itu hasil generate `make swag`.
4. **Jangan tambah service Postgres ke `docker-compose.yml`.** Keputusan user: Postgres jalan di host lokal.
5. **Pertahankan struktur minimal.** Tidak ada layer service atau repository, handler langsung memakai GORM. Jangan tambah layer kecuali user memintanya.
6. Repo belum di-`git init`. Jangan membuat repo git, commit, atau push tanpa diminta.

## Perintah

```bash
make run            # go run ./cmd/api (butuh Postgres lokal + JWT_SECRET di .env)
make build          # go build -o bin/api ./cmd/api
make swag           # swag init -g cmd/api/main.go -o docs --parseInternal
go vet ./...        # wajib lolos sebelum selesai
make k6             # load test, BASE_URL default http://localhost:8080
make up / down / logs   # Docker Compose (hanya service api)
```

- Migrate dan seed jalan otomatis saat startup (AutoMigrate, lalu seeder yang idempotent). Tidak ada perintah terpisah.
- Belum ada test dan linter. Kalau menambah test, gunakan `go test ./...`.

## Struktur dan tanggung jawab

| Path | Isi |
|---|---|
| `cmd/api/main.go` | Wiring: env → config → DB → migrate → seed → route → graceful shutdown (10 detik) |
| `internal/config/` | Membaca env. `JWT_SECRET` wajib, sisanya punya default. |
| `internal/database/` | `database.go` (connect, pool 25/25/30m, migrate), `seeder.go` |
| `internal/model/` | Satu file per model: `user.go`, `product.go`, `order.go` |
| `internal/dto/` | Struct request/response, termasuk `ErrorResponse` |
| `internal/handler/` | `auth.go`, `product.go`, `checkout.go` |
| `internal/middleware/jwt.go` | Bearer token → `user_id` di context |
| `docs/` | Hasil generate swag |
| `k6/load-test.js` | Load test |

## Konvensi kode

- **Context:** setiap query GORM memakai `.WithContext(c.Request().Context())`.
- **Error:** pakai `echo.NewHTTPError(status, "pesan")`, hasilnya `{"message": "..."}`. Error domain dipetakan di handler: `ErrProductNotFound` → 404, `ErrInsufficientStock` → 409.
- **Validasi:** manual di handler, tanpa library validator.
  - Login: email dan password wajib, email di-trim dan di-lowercase.
  - Checkout: `product_id > 0`, `quantity` 1–100.
  - Pagination di-clamp: `page ≥ 1`, `limit` 1–100, default 20.
- **Model:** field `Password` memakai tag `json:"-"`. Harga bertipe `int64` dalam rupiah. Stok punya `CHECK (stock >= 0)`.
- **Order:** tidak punya FK ke products. Data produk disimpan di `product_snapshot` (JSONB, lewat `Value`/`Scan`). Jangan tambahkan FK ke products.
- **JWT:** algoritma dipin ke HS256, claim `exp` wajib, `sub` berisi user id. TTL dari `JWT_TTL` (default 24h).
- **Swagger:** setiap handler punya anotasi godoc. Setelah mengubah handler atau dto, jalankan `make swag`.
- **Route:** semua di bawah `/api/v1`. Middleware JWT dipasang per route (saat ini hanya `/checkout`), bukan di seluruh group.

## Sebelum menyatakan selesai

- `go vet ./...` lolos dan `go build ./...` berhasil.
- Kalau handler atau dto berubah: `make swag` sudah dijalankan.
- Kalau env baru ditambahkan: `.env.example` dan tabel env di `README.md` sudah diperbarui.
- Laporkan apa adanya. Kalau sesuatu belum diuji atau diukur, sebutkan.
