package command

import (
	"fmt"

	"github.com/company/redis-clone/pkg/resp"
)

// handleSAVE: SAVE
// Memicu RDB snapshot secara manual (blocking/synchronous)
func (ex *Executor) handleSAVE(args []string) (*resp.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("wrong number of arguments for 'save' command")
	}
	if ex.snapshotter == nil {
		return nil, fmt.Errorf("RDB persistence is not enabled on this server")
	}
	if err := ex.snapshotter.SaveSnapshot(); err != nil {
		return nil, fmt.Errorf("snapshot failed: %w", err)
	}
	return resp.NewSimpleString("OK"), nil
}