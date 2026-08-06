#!/usr/bin/env bash
set -e

CLONE_PORT=6379
REAL_PORT=6380
REQUESTS=100000

echo "=== Building Redis Clone ==="
make build

echo "=== Starting Redis Clone on port $CLONE_PORT ==="
./bin/redis-clone -addr localhost:$CLONE_PORT -aof "" -rdb "" &
CLONE_PID=$!
sleep 1

echo ""
echo "=== Benchmark: Redis Clone ==="
redis-benchmark -h localhost -p $CLONE_PORT -t get,set,lpush,hset,sadd -n $REQUESTS -q

kill $CLONE_PID
wait $CLONE_PID 2>/dev/null || true

if command -v redis-server &> /dev/null; then
    echo ""
    echo "=== Starting real Redis on port $REAL_PORT ==="
    redis-server --port $REAL_PORT --daemonize yes --save ""
    sleep 1

    echo ""
    echo "=== Benchmark: Redis Asli ==="
    redis-benchmark -h localhost -p $REAL_PORT -t get,set,lpush,hset,sadd -n $REQUESTS -q

    redis-cli -p $REAL_PORT shutdown nosave 2>/dev/null || true

    echo ""
    echo "=== Bandingkan angka ops/sec di atas: Redis Clone vs Redis Asli ==="
else
    echo "redis-server tidak ditemukan, lewati perbandingan dengan Redis asli."
fi