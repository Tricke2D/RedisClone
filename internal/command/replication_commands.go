package command

import (
	"fmt"
	"strings"

	"github.com/company/redis-clone/pkg/resp"
)

// handleSLAVEOF: SLAVEOF host port | SLAVEOF NO ONE
func (ex *Executor) handleSLAVEOF(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'slaveof' command")
	}
	if ex.replicaOf == nil {
		return nil, fmt.Errorf("replication is not configured on this server")
	}

	if strings.EqualFold(args[0], "NO") && strings.EqualFold(args[1], "ONE") {
		ex.replicaOf.StopReplication()
		return resp.NewSimpleString("OK"), nil
	}

	masterAddr := args[0] + ":" + args[1]
	if err := ex.replicaOf.StartReplication(masterAddr); err != nil {
		return nil, err
	}
	return resp.NewSimpleString("OK"), nil
}