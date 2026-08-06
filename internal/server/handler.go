package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/company/redis-clone/internal/command"
	"github.com/company/redis-clone/internal/persistence"
	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
)

// handleConnection memproses satu koneksi client
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	s.log.Info("Client connected", map[string]interface{}{"remote_addr": clientAddr})

	parser := protocol.NewRespParser(conn)
	encoder := protocol.NewRespEncoder(conn)

	executor := command.NewExecutor(s.store)
	if s.aof != nil {
		executor = command.NewExecutorWithAOF(s.store, s.aof)
	}
	executor = executor.WithPropagator(s.replHub)
	if s.rdbPath != "" {
		executor = executor.WithSnapshotter(&rdbSnapshotter{path: s.rdbPath, store: s.store})
	}
	executor = executor.WithReplicaOfHandler(s)

	for {
		val, err := parser.Parse()
		if err != nil {
			if err.Error() == "EOF" {
				s.log.Info("Client disconnected", map[string]interface{}{"remote_addr": clientAddr})
			} else {
				s.log.Error("Parse error", map[string]interface{}{
					"remote_addr": clientAddr,
					"error":       err,
				})
				_ = encoder.Encode(resp.NewError(fmt.Sprintf("ERR %v", err)))
			}
			return
		}

		cmd, err := protocol.ParseCommand(val)
		if err != nil {
			s.log.Error("Invalid command format", map[string]interface{}{
				"remote_addr": clientAddr,
				"error":       err,
			})
			_ = encoder.Encode(resp.NewError("ERR Protocol error: invalid command format"))
			continue
		}

		if len(cmd) == 0 {
			continue
		}

		// SYNC ditangani SPESIAL: koneksi ini beralih jadi replica stream
		if strings.EqualFold(cmd[0], "SYNC") {
			s.serveReplica(conn, clientAddr, encoder)
			return
		}

		result, err := executor.Execute(cmd)
		if err != nil {
			s.log.Warn("Command execution warning", map[string]interface{}{
				"remote_addr": clientAddr,
				"command":     cmd[0],
				"error":       err,
			})
		}

		if err := encoder.Encode(result); err != nil {
			s.log.Error("Failed to encode response", map[string]interface{}{
				"remote_addr": clientAddr,
				"error":       err,
			})
			return
		}
	}
}

// rdbSnapshotter adalah adapter ke interface command.Snapshotter
type rdbSnapshotter struct {
	path  string
	store *storage.Store
}

func (rs *rdbSnapshotter) SaveSnapshot() error {
	return persistence.NewRDBWriter(rs.path).Save(rs.store)
}