package command

import (
	"fmt"

	"github.com/company/redis-clone/pkg/resp"
)

// handleSADD: SADD key member [member ...]
func (ex *Executor) handleSADD(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'sadd' command")
	}
	added, err := ex.store.SAdd(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(added)), nil
}

// handleSREM: SREM key member [member ...]
func (ex *Executor) handleSREM(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'srem' command")
	}
	removed, err := ex.store.SRem(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(removed)), nil
}

// handleSMEMBERS: SMEMBERS key
func (ex *Executor) handleSMEMBERS(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'smembers' command")
	}
	members, err := ex.store.SMembers(args[0])
	if err != nil {
		return nil, err
	}
	values := make([]*resp.Value, len(members))
	for i, m := range members {
		values[i] = resp.NewBulkString(m)
	}
	return resp.NewArray(values...), nil
}

// handleSCARD: SCARD key
func (ex *Executor) handleSCARD(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'scard' command")
	}
	count, err := ex.store.SCard(args[0])
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(count)), nil
}

// handleSISMEMBER: SISMEMBER key member
func (ex *Executor) handleSISMEMBER(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'sismember' command")
	}
	isMember, err := ex.store.SIsMember(args[0], args[1])
	if err != nil {
		return nil, err
	}
	if isMember {
		return resp.NewInteger(1), nil
	}
	return resp.NewInteger(0), nil
}