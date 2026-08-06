# Redis Clone - Production-Grade In-Memory KV Store

Implementasi Redis dari nol menggunakan Go:
- RESP Protocol (custom implementation)
- TCP Server (minimal, performant)
- Concurrency via goroutines + sync.RWMutex
- Persistence AOF + RDB (Fase 2-3)
- Replication master-slave (Fase 3)

## Getting Started

```bash
make setup
make build
make run