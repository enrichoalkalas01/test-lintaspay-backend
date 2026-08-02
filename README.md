# test-lintaspay

API disbursement untuk coding test Senior Backend Developer. Dibangun dengan Go (Fiber), GORM, MySQL, dan Uber Fx untuk dependency injection.

Jawaban Bagian 1 (idempotency & concurrency) ada di [ARCHITECTURE.md](ARCHITECTURE.md).

## Menjalankan

### Docker (paling cepat)

```bash
docker compose up --build
```

Schema + seed user otomatis ter-apply saat MySQL pertama kali boot (mount `migrations/0001_init.up.sql` ke `docker-entrypoint-initdb.d`). API tersedia di `http://localhost:3000/api/v1`.

### Manual

Butuh Go 1.25+ dan MySQL 8.

```bash
cp .env.example .env      # sesuaikan kredensial MySQL & JWT_SECRET
make migrate              # apply migrations/*.up.sql (idempotent, aman di-rerun)
make run                  # build + jalankan
# atau: make dev          # hot-reload via air
```

Rollback schema: `make migrate-down`. Unit test: `make test`.

## Seed Users

| Username   | Password      | Role       |
|------------|---------------|------------|
| superadmin | superadmin123 | superadmin |
| admin      | admin123      | admin      |
| operator   | operator123   | operator   |

Role: `operator` buat + lihat disbursement; `admin` + ubah status; `superadmin` + hapus dan lihat audit log.

## Endpoints

Base path: `/api/v1`. Semua endpoint selain `/auth/*` dan `/health` butuh header `Authorization: Bearer <access_token>`. Swagger UI: `http://localhost:3000/swagger/` (non-production).

| Method | Path                        | Akses            | Keterangan |
|--------|-----------------------------|------------------|------------|
| POST   | /auth/login                 | public           | access token (15m) + refresh token (7d) |
| POST   | /auth/refresh               | public           | tukar refresh token dengan access token baru |
| POST   | /auth/logout                | public           | revoke refresh token |
| GET    | /health                     | public           | status DB, 503 kalau DB down |
| GET    | /disbursements              | semua role       | filter: `page, limit, search, status, date_from, date_to, sort_by, sort_order` |
| GET    | /disbursements/:id          | semua role       | |
| POST   | /disbursements              | semua role       | support header `Idempotency-Key` (uuid v4) |
| POST   | /disbursements/batch        | semua role       | maks 100 item, partial success (201 semua sukses / 207 sebagian / 400 semua gagal) |
| PATCH  | /disbursements/:id/status   | admin, superadmin| APPROVED / REJECTED, concurrency-safe |
| DELETE | /disbursements/:id          | superadmin       | soft delete, hanya status PENDING |
| GET    | /audit-logs                 | superadmin       | filter: `entity_id, action, date_from, date_to` |

Perilaku idempotency: request kedua dengan key sama dalam 24 jam mengembalikan response identik + header `X-Idempotent-Replayed: true`, tanpa memproses ulang. Key sama dengan payload berbeda ditolak 409.

## Schema Database

Migration file: [`migrations/0001_init.up.sql`](migrations/0001_init.up.sql).

- `users` — seed 3 user, password bcrypt, unique index di `username`
- `refresh_tokens` — hash sha256 token (unique), `revoked_at` untuk logout, FK ke users
- `disbursements` — soft delete via `deleted_at`; index di `status`, `created_at`, `recipient_name`, `deleted_at` (kolom yang dipakai filter/sort); CHECK constraint `amount >= 10000` dan whitelist status
- `idempotency_keys` — PK di `key` (reservasi atomic), simpan request hash + cached response (JSON) + `expires_at`
- `audit_logs` — `before`/`after` sebagai kolom JSON; index di `entity_id`, `action`, `created_at`

## Struktur Project

```
cmd/api           entry point (fx modules)
cmd/migrate       SQL migration runner sederhana
configs           viper (.env) + fx module wiring
internal/domain/entity    struct + interface (kontrak antar layer)
internal/repository       akses data (GORM)
internal/usecase          business logic
internal/handler          HTTP handler (tanpa logika bisnis)
internal/routes           registrasi route + RBAC middleware
pkg               shared: httpresponse, middlewares, logger, reqctx, database
migrations        SQL migration files
```

## Catatan Keputusan Teknis

- **Context propagation** — semua layer menerima `context.Context`. `request_id` (middleware) dan identitas user (JWT middleware) masuk ke context via `pkg/common/reqctx`, dan `logger.WithCtx(ctx, log)` menempelkan keduanya ke setiap log line. Response selalu menyertakan `X-Request-ID`.
- **Idempotency & concurrency** — detail di ARCHITECTURE.md. Ringkas: reservasi key + insert disbursement satu transaksi; approval pakai conditional update (`WHERE status = 'PENDING'`).
- **Audit log** — ditulis sinkron tapi kegagalannya hanya dicatat ke server log, tidak pernah menggagalkan operasi utama. Pakai `context.WithoutCancel` supaya penulisan tidak ikut batal saat request context selesai.
- **Structured logging** — zap. Development default console (enak dibaca), production / `LOG_FORMAT=json` menghasilkan JSON per request: `request_id`, `method`, `path`, `status_code`, `latency_ms`, `user`.
- **`/health` sengaja public** — menyimpang dari "semua endpoint selain /auth/* wajib JWT" karena load balancer / readiness probe tidak bisa membawa token.
- **Rate limiting (bonus)** — per user (fallback IP): POST /disbursements 30 req/menit, endpoint terproteksi lain 120 req/menit. Storage in-memory, cukup untuk single instance; untuk multi-instance perlu diganti backing Redis.
- **Library tambahan** — `golang-jwt/jwt/v5` (JWT standar de-facto untuk Go) dan `golang.org/x/crypto/bcrypt` (hash password). Sisanya sudah ada di stack: Fiber, GORM, Fx, Viper, Zap.

## Unit Test

```bash
make test
```

Fokus di logika bisnis kritis (`internal/usecase/disbursements_test.go`): kalkulasi `admin_fee` (termasuk boundary 5 juta), idempotency handler (replay, key reuse dengan payload beda, key expired, format key invalid), validasi perubahan status (sudah diproses, kalah race/rows affected 0, status invalid, not found), soft delete non-PENDING, batch create partial success, dan memastikan kegagalan audit log tidak menggagalkan operasi.
