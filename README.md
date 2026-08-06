# 🚀 Redis Clone — Production-Grade In-Memory Database

**Redis-compatible KV Store | RESP Protocol | Go | 35 Commands | AOF + RDB Persistence | Master-Slave Replication**

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Protocol-DC382D?logo=redis&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green)

---

## 📋 Daftar Isi

- [📍 Studi Kasus](#-studi-kasus)
- [✨ Fitur Utama](#-fitur-utama)
- [🏗️ Arsitektur Sistem](#️-arsitektur-sistem)
- [🛠️ Tech Stack](#️-tech-stack)
- [💻 Requirements](#-requirements)
- [🚀 Instalasi & Menjalankan](#-instalasi--menjalankan)
- [📝 Referensi Command](#-referensi-command)
- [🧪 Testing](#-testing)
- [📊 Benchmarking](#-benchmarking)
- [📁 Struktur Project](#-struktur-project)
- [⚠️ Batasan & Roadmap](#️-batasan--roadmap)
- [📞 Kontribusi](#-kontribusi)

---

## 📍 Studi Kasus

Bayangkan kamu memiliki aplikasi dengan kebutuhan caching atau penyimpanan in-memory yang cepat. Redis adalah solusi standar, tapi bagaimana jika kamu ingin **memahami cara kerjanya dari dalam**?

### Masalah yang Dipecahkan

Banyak engineer menggunakan Redis tanpa memahami:

- ❌ Bagaimana protokol RESP bekerja di balik layar
- ❌ Bagaimana Redis menangani concurrency dan threading
- ❌ Bagaimana persistence (AOF/RDB) benar-benar bekerja
- ❌ Bagaimana replication master-slave diimplementasikan
- ❌ Bagaimana benchmarking dan optimisasi performa dilakukan

### Solusi: Redis Clone

Implementasi Redis dari nol menggunakan Go dengan **pendekatan production-grade** (11 minggu development):

✅ **RESP Protocol** — Parser & encoder dari nol (tanpa library)  
✅ **TCP Server** — Goroutine-per-connection dengan concurrency control  
✅ **35 Commands** — String, List, Hash, Set, TTL, Admin, Replication  
✅ **AOF Persistence** — Write-ahead log dengan fsync policy (always/everysec)  
✅ **RDB Persistence** — Binary snapshot dengan encoding/gob (6x lebih cepat)  
✅ **Master-Slave Replication** — Full sync + command streaming real-time  
✅ **0 Race Conditions** — Terbukti dengan `go test -race` + stress test 100 goroutine  
✅ **Benchmarking** — Go benchmark + redis-benchmark comparison  

**Hasil:** Database in-memory production-grade dengan 35 command, 2 persistence strategies, replication, dan 0 race conditions.

---

## ✨ Fitur Utama

### 🎯 Core Features

| Fitur | Deskripsi |
|-------|-----------|
| **RESP Protocol** | Implementasi lengkap parser + encoder (Simple String, Error, Integer, Bulk String, Array) |
| **TCP Server** | Concurrent connection handling dengan goroutine per client |
| **Thread-Safe Storage** | `sync.RWMutex` dengan fine-grained locking |
| **35 Commands** | Lengkap dari Fase 1 sampai Fase 3 |

### 📊 Data Types & Commands

| Data Type | Commands |
|-----------|----------|
| **String** | `SET`, `GET`, `DEL`, `EXISTS`, `INCR`, `DECR`, `APPEND`, `STRLEN`, `KEYS` |
| **List** | `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LRANGE`, `LLEN`, `LINDEX` |
| **Hash** | `HSET`, `HGET`, `HDEL`, `HEXISTS`, `HGETALL`, `HKEYS`, `HVALS`, `HLEN` |
| **Set** | `SADD`, `SREM`, `SMEMBERS`, `SCARD`, `SISMEMBER` |
| **TTL** | `EXPIRE`, `PEXPIRE`, `TTL`, `PTTL`, `PERSIST` |

### 💾 Persistence Strategy

| Strategy | Deskripsi | Use Case |
|----------|-----------|----------|
| **AOF (Append-Only File)** | Log semua command dalam RESP format | High durability, setiap command tercatat |
| **RDB (Snapshot)** | Binary snapshot dengan gob encoding | Fast recovery, compact size |
| **Hybrid** | RDB priority + AOF fallback | Best of both worlds |

### 🔄 Replication

- **Master-Slave Architecture** — Full sync + command streaming
- **SYNC Command** — Protocol internal untuk replica connect
- **SLAVEOF Command** — Client command untuk setup replication
- **Command Propagation** — Broadcast write commands ke semua replica otomatis

### 🎁 Bonus Features

- **Active Expiry** — Background goroutine scan setiap 100ms
- **Lazy Expiry** — Key expired otomatis hilang saat diakses
- **SAVE Command** — Trigger RDB snapshot manual
- **Backup Strategy** — RDB + AOF untuk recovery maksimal
- **Structured Logging** — Timestamp + contextual info
- **pprof Support** — CPU/memory profiling endpoint

---

## 🏗️ Arsitektur Sistem

```
┌────────────────────────────────────────────────────────────────┐
│         Redis Clone — Production-Grade KV Store               │
└────────────────────────────────────────────────────────────────┘

    Client (redis-cli)
           │
           ▼
    ┌─────────────────┐
    │   TCP Server    │
    │  (concurrent)   │
    └────────┬────────┘
             │
             ▼
    ┌─────────────────┐
    │  RESP Parser    │  → Parse command
    └────────┬────────┘
             │
             ▼
    ┌─────────────────────────────────────────┐
    │      Command Executor (Router)          │
    │                                         │
    │ ┌─────────────────────────────────────┐ │
    │ │  String │ List │ Hash │ Set │ TTL  │ │
    │ └─────────────────────────────────────┘ │
    └────────┬────────────────┬───────────────┘
             │                │
        ┌────▼────┐      ┌────▼──────┐
        │ Storage │      │ AOF Log   │
        │ (sync.) │      │ (append)  │
        └────┬────┘      └───────────┘
             │
        ┌────▼──────────────┐
        │ Replication Hub   │ → Propagate ke replicas
        └───────────────────┘
        
    ┌────────────────────────────┐
    │  Background Tasks:         │
    │  • RDB Snapshot Writer     │
    │  • TTL Expiry Manager      │
    └────────────────────────────┘
```

### Data Flow

1. **Client Connection** → TCP Server terima koneksi
2. **Parse Command** → RESP parser dekode command
3. **Route & Execute** → Command executor route ke handler spesifik
4. **Storage Operation** → Modifikasi data di storage
5. **Persist** → AOF log command (jika write)
6. **Propagate** → Broadcast ke replicas (jika master)
7. **Encode Response** → RESP encoder encode hasil
8. **Send Response** → Kirim balik ke client
9. **Background** → RDB snapshot + TTL cleanup (goroutine terpisah)

---

## 🛠️ Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| **Language** | Go 1.21+ |
| **Networking** | `net` package (raw TCP) |
| **Protocol** | Custom RESP implementation |
| **Concurrency** | Goroutines + `sync.RWMutex` |
| **Persistence (AOF)** | Append-only file, RESP format |
| **Persistence (RDB)** | `encoding/gob` binary snapshot |
| **Profiling** | `net/http/pprof` |
| **Testing** | `go test` + `-race` flag |
| **Containerization** | Docker |
| **CLI** | Flag parsing |

---

## 💻 Requirements

- **Go** v1.21 atau lebih baru
- **Docker** v20.x atau lebih baru (opsional)
- **Git** v2.x atau lebih baru
- **redis-cli** untuk testing (opsional, bisa pakai telnet)

---

## 🚀 Instalasi & Menjalankan

### 1️⃣ Clone Repository

```bash
git clone https://github.com/Tricke2D/RedisClone.git
cd RedisClone
```

### 2️⃣ Setup Go Module

```bash
go mod init github.com/Tricke2D/RedisClone
go get golang.org/x/sync@latest
go get github.com/stretchr/testify@latest
go mod tidy
```

### 3️⃣ Buat Struktur Project

```bash
# PowerShell
New-Item -ItemType Directory -Force -Path "cmd\server"
New-Item -ItemType Directory -Force -Path "internal\protocol"
New-Item -ItemType Directory -Force -Path "internal\server"
New-Item -ItemType Directory -Force -Path "internal\storage"
New-Item -ItemType Directory -Force -Path "internal\command"
New-Item -ItemType Directory -Force -Path "internal\logger"
New-Item -ItemType Directory -Force -Path "internal\persistence"
New-Item -ItemType Directory -Force -Path "internal\replication"
New-Item -ItemType Directory -Force -Path "pkg\resp"
New-Item -ItemType Directory -Force -Path "test"
New-Item -ItemType Directory -Force -Path "scripts"
New-Item -ItemType Directory -Force -Path "docs"
New-Item -ItemType Directory -Force -Path "bin"
```

### 4️⃣ Build & Run

#### Jalankan Default (AOF + RDB aktif, port 6379)

```bash
go build -o bin/RedisClone.exe cmd/server/main.go
.\bin\RedisClone.exe -addr localhost:6379 -aof appendonly.aof -rdb dump.rdb
```

#### Custom Port (jika Redis asli sudah berjalan)

```bash
.\bin\RedisClone.exe -addr localhost:6380 -aof appendonly.aof -rdb dump.rdb
```

#### Benchmark Mode (nonaktifkan persistence)

```bash
.\bin\RedisClone.exe -addr localhost:6379 -aof "" -rdb ""
```

#### Dengan AOF Sync Policy (lebih cepat)

```bash
.\bin\RedisClone.exe -addr localhost:6379 -aof appendonly.aof -aof-sync everysec
```

#### Dengan pprof Profiling

```bash
.\bin\RedisClone.exe -addr localhost:6379 -pprof-addr localhost:6060
```

### 5️⃣ Testing dengan redis-cli

**Terminal 1: Jalankan server**

```bash
.\bin\RedisClone.exe -addr localhost:6380 -aof appendonly.aof -rdb dump.rdb
```

**Terminal 2: Buka redis-cli**

```bash
redis-cli -h localhost -p 6380
```

#### Test String Commands

```redis
PING                    # → PONG
SET nama "Budi"         # → OK
GET nama                # → "Budi"
INCR counter            # → (integer) 1
STRLEN nama             # → (integer) 4
```

#### Test List Commands

```redis
LPUSH mylist a b c      # → (integer) 3
LRANGE mylist 0 -1      # → 1) "c" 2) "b" 3) "a"
LPOP mylist             # → "c"
LLEN mylist             # → (integer) 2
```

#### Test Hash Commands

```redis
HSET user:1 name "Alice" age "30"  # → (integer) 2
HGET user:1 name                   # → "Alice"
HGETALL user:1                     # → 1) "name" 2) "Alice" 3) "age" 4) "30"
HLEN user:1                        # → (integer) 2
```

#### Test Set Commands

```redis
SADD tags go redis docker          # → (integer) 3
SMEMBERS tags                      # → 1) "go" 2) "redis" 3) "docker"
SCARD tags                         # → (integer) 3
SISMEMBER tags go                  # → (integer) 1
```

#### Test TTL Commands

```redis
SET session "active"               # → OK
EXPIRE session 60                  # → (integer) 1
TTL session                        # → (integer) 58
PERSIST session                    # → (integer) 1
```

#### Trigger RDB Snapshot

```redis
SAVE                               # → OK
```

### 6️⃣ Test RDB Recovery

```bash
# Matikan server (Ctrl+C)

# Jalankan lagi
.\bin\RedisClone.exe -addr localhost:6380 -aof appendonly.aof -rdb dump.rdb

# Log akan muncul:
# [INFO] RDB snapshot loaded | map[path:dump.rdb keys:5]
# [INFO] Server started | map[address:localhost:6380]

redis-cli -h localhost -p 6380
> GET nama
"Budi"  # ✅ Data selamat!
```

### 7️⃣ Test Master-Slave Replication

**Terminal 1: Master (port 6379)**

```bash
.\bin\RedisClone.exe -addr localhost:6379 -aof master.aof -rdb master.rdb
```

**Terminal 2: Replica (port 6380)**

```bash
.\bin\RedisClone.exe -addr localhost:6380 -aof replica.aof -rdb replica.rdb
```

**Terminal 3: Hubungkan replica ke master**

```bash
redis-cli -h localhost -p 6380
> SLAVEOF localhost 6379
OK
```

**Terminal 4: Tulis data di master**

```bash
redis-cli -h localhost -p 6379
> SET produk "Laptop"
OK
```

**Terminal 5: Baca data di replica**

```bash
redis-cli -h localhost -p 6380
> GET produk
"Laptop"  # ✅ Ter-replikasi otomatis!
```

### 8️⃣ Akses pprof Profiling

```bash
# Jalankan server dengan pprof
.\bin\RedisClone.exe -addr localhost:6379 -pprof-addr localhost:6060

# CPU Profiling (30 detik)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory Profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine Profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Visualisasi (butuh graphviz)
(pprof) web
```

---

## 📝 Referensi Command

### Connection Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `PING` | Test koneksi | `PING` → `PONG` |
| `ECHO` | Echo message | `ECHO "hello"` → `"hello"` |
| `SELECT` | Select DB (stub) | `SELECT 0` → `OK` |

### String Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `SET` | Set key-value | `SET foo bar` → `OK` |
| `GET` | Get value by key | `GET foo` → `"bar"` |
| `DEL` | Delete key(s) | `DEL foo` → `(integer) 1` |
| `EXISTS` | Check key existence | `EXISTS foo` → `(integer) 1` |
| `INCR` | Increment integer | `INCR counter` → `(integer) 1` |
| `DECR` | Decrement integer | `DECR counter` → `(integer) 0` |
| `APPEND` | Append to string | `APPEND msg "Hello"` → `(integer) 5` |
| `STRLEN` | Get string length | `STRLEN msg` → `(integer) 5` |
| `KEYS` | Pattern matching | `KEYS *` → `["foo", "bar"]` |

### List Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `LPUSH` | Push to left | `LPUSH mylist a b` → `(integer) 2` |
| `RPUSH` | Push to right | `RPUSH mylist c` → `(integer) 3` |
| `LPOP` | Pop from left | `LPOP mylist` → `"b"` |
| `RPOP` | Pop from right | `RPOP mylist` → `"c"` |
| `LRANGE` | Get range | `LRANGE mylist 0 -1` → `["a"]` |
| `LLEN` | Get length | `LLEN mylist` → `(integer) 1` |
| `LINDEX` | Get by index | `LINDEX mylist 0` → `"a"` |

### Hash Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `HSET` | Set field-value | `HSET user name "Alice"` → `(integer) 1` |
| `HGET` | Get field value | `HGET user name` → `"Alice"` |
| `HDEL` | Delete field | `HDEL user age` → `(integer) 1` |
| `HEXISTS` | Check field | `HEXISTS user name` → `(integer) 1` |
| `HGETALL` | Get all fields | `HGETALL user` → `["name","Alice"]` |
| `HKEYS` | Get all keys | `HKEYS user` → `["name"]` |
| `HVALS` | Get all values | `HVALS user` → `["Alice"]` |
| `HLEN` | Get field count | `HLEN user` → `(integer) 1` |

### Set Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `SADD` | Add members | `SADD tags go redis` → `(integer) 2` |
| `SREM` | Remove members | `SREM tags go` → `(integer) 1` |
| `SMEMBERS` | Get all members | `SMEMBERS tags` → `["redis"]` |
| `SCARD` | Get cardinality | `SCARD tags` → `(integer) 1` |
| `SISMEMBER` | Check member | `SISMEMBER tags redis` → `(integer) 1` |

### TTL Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `EXPIRE` | Set TTL in seconds | `EXPIRE key 60` → `(integer) 1` |
| `PEXPIRE` | Set TTL in ms | `PEXPIRE key 60000` → `(integer) 1` |
| `TTL` | Get TTL in seconds | `TTL key` → `(integer) 58` |
| `PTTL` | Get TTL in ms | `PTTL key` → `(integer) 58000` |
| `PERSIST` | Remove TTL | `PERSIST key` → `(integer) 1` |

### Admin & Replication Commands

| Command | Deskripsi | Contoh |
|---------|-----------|--------|
| `SAVE` | Manual RDB snapshot | `SAVE` → `OK` |
| `SLAVEOF` | Setup replication | `SLAVEOF localhost 6379` → `OK` |

---

## 🧪 Testing

### Run All Tests

```bash
# Unit tests
go test -v ./...

# With race detection (PENTING!)
go test -v -race ./...

# Specific package
go test -v ./internal/storage

# Specific test
go test -v -run TestLPush ./internal/storage
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out
```

### Stress Test

```bash
# 100 goroutines, 10,000 operations
go test -v -race -run TestStress ./test

# Memory stability test
go test -v -run TestStressMemoryStable ./test
```

### Benchmark

```bash
# All benchmarks
go test -bench=. -benchmem ./internal/command

# Specific benchmark
go test -bench=BenchmarkGET -benchmem ./internal/command

# Concurrent benchmark
go test -bench=BenchmarkConcurrentGETSET -cpu=1,2,4,8 ./internal/command
```

### Integration Tests

```bash
# Full server test (requires running server)
go test -v ./test -run TestServer

# Replication test
go test -v ./test -run TestMasterSlaveReplication

# AOF persistence test
go test -v ./test -run TestAOFPersistence
```

---

## 📊 Benchmarking

### Go Benchmark Results

```bash
go test -bench=. -benchmem ./internal/command
```

Expected output:
```
BenchmarkGET-8                  5000000      250 ns/op      16 B/op      1 allocs/op
BenchmarkSET-8                  2000000      600 ns/op      64 B/op      3 allocs/op
BenchmarkLPUSH-8                  50000    25000 ns/op     512 B/op      5 allocs/op
BenchmarkHSET-8                1000000     1200 ns/op     128 B/op      4 allocs/op
BenchmarkConcurrentGETSET-8     200000     8000 ns/op      32 B/op      2 allocs/op
```

### Redis Benchmark Comparison

```bash
# 1. Jalankan Redis Clone tanpa persistence
.\bin\RedisClone.exe -addr localhost:6379 -aof "" -rdb ""

# 2. Jalankan redis-benchmark
redis-benchmark -h localhost -p 6379 -t get,set,lpush,hset,sadd -n 100000 -q

# 3. Bandingkan dengan Redis asli
redis-server --port 6380 --daemonize yes --save ""
redis-benchmark -h localhost -p 6380 -t get,set,lpush,hset,sadd -n 100000 -q
```

### Performance Metrics

| Operation | Redis Clone | Redis | Ratio |
|-----------|------------|-------|-------|
| `GET` | ~400,000 ops/sec | ~500,000 ops/sec | 80% |
| `SET` | ~150,000 ops/sec | ~200,000 ops/sec | 75% |
| `LPUSH` | ~5,000 ops/sec | ~100,000 ops/sec | 5%* |
| `HSET` | ~80,000 ops/sec | ~150,000 ops/sec | 53% |

> *LPUSH lambat karena implementasi menggunakan slice (O(n)), berbeda dengan Redis yang pakai linked list (O(1))

---

## 📁 Struktur Project

```
RedisClone/
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── Dockerfile
│
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
│
├── internal/
│   ├── protocol/
│   │   ├── parser.go                  # RESP parser
│   │   ├── encoder.go                 # RESP encoder
│   │   ├── parser_test.go
│   │   └── encoder_test.go
│   │
│   ├── server/
│   │   ├── server.go                  # TCP server lifecycle
│   │   ├── handler.go                 # Per-connection handler
│   │   ├── errors.go                  # Redis-style errors
│   │   ├── expiry.go                  # Active expiry manager
│   │   └── replication.go             # Master-side replication
│   │
│   ├── storage/
│   │   ├── store.go                   # Core storage (map + RWMutex)
│   │   ├── list.go                    # List operations
│   │   ├── hash.go                    # Hash operations
│   │   ├── set.go                     # Set operations
│   │   ├── ttl.go                     # TTL/Expiry operations
│   │   ├── snapshot.go                # Encode/Decode snapshot
│   │   └── *_test.go                  # Unit tests
│   │
│   ├── command/
│   │   ├── executor.go                # Command dispatcher
│   │   ├── string_commands.go         # SET/GET/DEL/etc
│   │   ├── list_commands.go           # LPUSH/RPUSH/etc
│   │   ├── hash_commands.go           # HSET/HGET/etc
│   │   ├── set_commands.go            # SADD/SREM/etc
│   │   ├── ttl_commands.go            # EXPIRE/TTL/etc
│   │   ├── admin_commands.go          # SAVE
│   │   ├── replication_commands.go    # SLAVEOF
│   │   ├── *_test.go                  # Unit tests
│   │   └── executor_bench_test.go     # Benchmarks
│   │
│   ├── persistence/
│   │   ├── aof.go                     # AOF writer/reader
│   │   ├── rdb.go                     # RDB writer/reader
│   │   └── *_test.go                  # Unit tests
│   │
│   ├── replication/
│   │   ├── hub.go                     # Replica registry (master side)
│   │   ├── replica.go                 # Replica connector (slave side)
│   │   └── *_test.go                  # Unit tests
│   │
│   └── logger/
│       └── logger.go                  # Structured logger
│
├── pkg/
│   └── resp/
│       └── types.go                   # RESP type definitions
│
├── test/
│   ├── integration_test.go            # E2E tests
│   ├── replication_integration_test.go
│   ├── aof_integration_test.go
│   └── stress_test.go                 # Concurrent stress tests
│
├── scripts/
│   └── benchmark.sh                   # Redis benchmark script
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── PROTOCOL.md
│   ├── REDIS_CLONE_FASE_1_LENGKAP.md
│   ├── REDIS_CLONE_FASE_2_LENGKAP.md
│   └── REDIS_CLONE_FASE_3_LENGKAP.md
│
└── bin/
    └── RedisClone.exe                 # Compiled binary
```

---

## ⚠️ Batasan & Roadmap

### Batasan Saat Ini

| Batasan | Penjelasan | Solusi Future |
|---------|-----------|----------------|
| **LPUSH O(n)** | Pakai slice + append, bukan linked list | Ganti ke `container/list` |
| **Replication Blocking** | Propagate synchronous dengan client | Buffered channel + async goroutine |
| **No Chained Replication** | Replica tidak punya sub-replica | Pasang propagator di replica |
| **No BGSAVE** | SAVE blocking, tanpa background version | Copy-on-write (fork) |
| **No SCAN** | KEYS * scan semua key sekaligus | Implementasi SCAN cursor |
| **No Auth** | Belum ada autentikasi | Tambah AUTH command |
| **Single Instance** | Tidak ada clustering | Cluster mode |

### Roadmap Pengembangan

- ☐ **BGSAVE** — Background snapshot non-blocking
- ☐ **SCAN Command** — Cursor-based iteration untuk dataset besar
- ☐ **Linked List** — Ganti list implementation ke doubly-linked list
- ☐ **Replication Backlog** — Buffer untuk non-blocking replication
- ☐ **AUTH Command** — Autentikasi untuk production
- ☐ **Cluster Mode** — Sharding dan distributed storage
- ☐ **Sentinel** — High availability and failover
- ☐ **Pub/Sub** — Publish-subscribe messaging
- ☐ **Sorted Set (ZSET)** — Data type baru
- ☐ **Lua Scripting** — EVAL command support
- ☐ **TLS/SSL** — Encrypted connections

---

## 📞 Kontribusi

**Repository:** https://github.com/Tricke2D/RedisClone

**Issues:** https://github.com/Tricke2D/RedisClone/issues

Contributions are welcome! 🎉 Silakan fork repository ini dan submit pull request.

---

## 📜 License

**MIT License** — Silakan digunakan untuk keperluan belajar dan pengembangan.

Made with ❤️ by Muhamad Syukron Zakka
