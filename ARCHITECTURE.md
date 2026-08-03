# Architecture Notes

## 1.1 Idempotency

State idempotency saya simpan di MySQL, tabel `idempotency_keys`, dengan `key` (UUID dari header `Idempotency-Key`) sebagai primary key. Saya memang tidak memakai Redis karena cache response dan insert disbursement harus atomic dan itu paling mudah dijamin kalau keduanya hidup di database yang sama, dalam satu transaksi.

Alurnya: usecase menyiapkan row idempotency berisi `user_id`, hash dari payload, dan response body yang sudah di serialize, lalu menginsert row itu bersama row disbursement dalam satu transaksi. Kalau key sudah pernah dipakai, insert pertama gagal dengan duplicate entry, transaksi rollback, response lama dibaca dari tabel dan dikembalikan apa adanya dengan header `X-Idempotent-Replayed: true`. Untuk dua request paralel dengan key sama, keduanya mencoba insert primary key yang sama, InnoDB membuat request kedua menunggu sampai transaksi pertama commit, baru gagal duplicate. Jadi disbursement tidak mungkin tercipta dua kali, tanpa lock eksplisit.

Hash payload ikut disimpan supaya key yang sama tapi body berbeda ditolak 409, bukan diam-diam mereplay response lama. TTL 24 jam disimpan di kolom `expires_at` row kadaluarsa dibersihkan lazily saat key yang sama dipakai lagi. Trade-off dibanding Redis: sedikit lebih lambat dan tabel perlu dibersihkan berkala, tapi konsisten, transaksional, dan tidak menambah infrastruktur.

## 1.2 Concurrency & Locking

Saya memakai conditional update compare and swap di level row:

```sql
UPDATE disbursements
SET status = ?, approved_by = ?, updated_at = ?
WHERE id = ? AND status = 'PENDING'
```

Dua admin menekan approve bersamaan berarti dua UPDATE ke row yang sama. InnoDB men-serialize keduanya lewat row lock: yang pertama commit mengubah status, yang kedua tidak lagi match kondisi `status = 'PENDING'` sehingga `RowsAffected = 0` API membalas 409 dengan pesan yang jelas bahwa disbursement baru saja diproses request lain.

Kenapa bukan pessimistic `SELECT ... FOR UPDATE`: pendekatan itu menahan row lock selama transaksi hidup, menambah round-trip dan peluang deadlock, padahal di kasus ini satu statement UPDATE sudah atomic, lock eksplisit tidak menambah jaminan apa pun. Kenapa bukan optimistic locking dengan kolom `version`: transisi status di sini one-way dan final (PENDING → APPROVED/REJECTED), jadi kolom `status` sendiri sudah berfungsi sebagai version. Trade-off-nya: kalau nanti ada update field lain yang boleh berjalan paralel (misalnya edit note oleh dua orang), pendekatan ini perlu di-upgrade ke kolom version sungguhan supaya lost update di field lain juga terdeteksi.

Sebelum UPDATE, status memang dibaca dulu tapi hanya untuk menghasilkan 404/pesan error yang informatif. Sumber kebenarannya tetap kondisi di UPDATE, bukan hasil pembacaan itu.

## Catatan: penyimpanan refresh token

Refresh token disimpan di database (tabel `refresh_tokens`), yang disimpan hash sha256-nya, bukan nilai mentah. Alasannya: token tetap valid setelah aplikasi restart, logout bisa merevoke token secara spesifik (kolom `revoked_at`), dan kalau database bocor, hash tidak bisa dipakai login. Trade off dibanding in-memory: satu query ekstra per refresh menurut saya sepadan karena refresh hanya terjadi tiap ~15 menit per user.
