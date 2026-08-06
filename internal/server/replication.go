package server

import (
	"bytes"
	"net"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/pkg/resp"
)

// serveReplica menangani koneksi client yang mengirim command SYNC.
func (s *Server) serveReplica(conn net.Conn, clientAddr string, encoder *protocol.RespEncoder) {
	s.log.Info("Replica connected, starting full sync", map[string]interface{}{"remote_addr": clientAddr})

	var buf bytes.Buffer
	if err := s.store.EncodeSnapshot(&buf); err != nil {
		s.log.Error("Failed to encode snapshot for replica", map[string]interface{}{"error": err})
		return
	}

	if err := encoder.Encode(resp.NewBulkString(buf.String())); err != nil {
		s.log.Error("Failed to send snapshot to replica", map[string]interface{}{"error": err})
		return
	}

	s.replHub.Register(clientAddr, encoder)
	defer s.replHub.Unregister(clientAddr)

	s.log.Info("Full sync complete, streaming writes", map[string]interface{}{
		"remote_addr":     clientAddr,
		"total_replicas": s.replHub.ReplicaCount(),
	})

	parser := protocol.NewRespParser(conn)
	for {
		if _, err := parser.Parse(); err != nil {
			s.log.Info("Replica disconnected", map[string]interface{}{"remote_addr": clientAddr})
			return
		}
	}
}