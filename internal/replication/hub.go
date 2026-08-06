package replication

import (
	"sync"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/pkg/resp"
)

// Hub adalah registry thread-safe untuk semua koneksi replica yang terhubung.
type Hub struct {
	mu       sync.RWMutex
	replicas map[string]*replicaConn
}

type replicaConn struct {
	mu      sync.Mutex
	encoder *protocol.RespEncoder
}

// NewHub membuat replication hub kosong.
func NewHub() *Hub {
	return &Hub{replicas: make(map[string]*replicaConn)}
}

// Register mendaftarkan koneksi replica baru.
func (h *Hub) Register(addr string, encoder *protocol.RespEncoder) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.replicas[addr] = &replicaConn{encoder: encoder}
}

// Unregister menghapus replica dari hub.
func (h *Hub) Unregister(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.replicas, addr)
}

// Propagate mem-broadcast satu write command ke SEMUA replica.
func (h *Hub) Propagate(cmd []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	values := make([]*resp.Value, len(cmd))
	for i, c := range cmd {
		values[i] = resp.NewBulkString(c)
	}
	message := resp.NewArray(values...)

	for _, r := range h.replicas {
		r.mu.Lock()
		_ = r.encoder.Encode(message)
		r.mu.Unlock()
	}
}

// ReplicaCount mengembalikan jumlah replica yang saat ini terhubung.
func (h *Hub) ReplicaCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.replicas)
}