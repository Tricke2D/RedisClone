package replication

import (
	"bytes"
	"fmt"
	"net"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
)

// Replayer adalah interface untuk mereplay command dari master.
type Replayer interface {
	Execute(cmd []string) (*resp.Value, error)
}

// RunSync menjalankan seluruh siklus hidup koneksi REPLICA ke MASTER.
func RunSync(conn net.Conn, store *storage.Store, executor Replayer) error {
	encoder := protocol.NewRespEncoder(conn)
	parser := protocol.NewRespParser(conn)

	syncCmd := resp.NewArray(resp.NewBulkString("SYNC"))
	if err := encoder.Encode(syncCmd); err != nil {
		return fmt.Errorf("failed to send SYNC: %w", err)
	}

	snapshotVal, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to receive snapshot: %w", err)
	}
	if snapshotVal.Type != resp.TypeBulkString {
		return fmt.Errorf("expected bulk string snapshot, got type %c", snapshotVal.Type)
	}

	if err := store.DecodeSnapshot(bytes.NewReader([]byte(snapshotVal.Str))); err != nil {
		return fmt.Errorf("failed to decode snapshot: %w", err)
	}

	for {
		val, err := parser.Parse()
		if err != nil {
			return fmt.Errorf("connection to master lost: %w", err)
		}

		cmd, err := protocol.ParseCommand(val)
		if err != nil {
			return fmt.Errorf("invalid command from master: %w", err)
		}

		if _, err := executor.Execute(cmd); err != nil {
			return fmt.Errorf("failed to apply command from master: %w", err)
		}
	}
}