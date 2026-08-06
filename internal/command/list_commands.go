package command

import (
	"fmt"
	"strconv"

	"github.com/company/redis-clone/pkg/resp"
)

// handleLPUSH: LPUSH key value [value ...]
func (ex *Executor) handleLPUSH(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'lpush' command")
	}
	length, err := ex.store.LPush(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(length)), nil
}

// handleRPUSH: RPUSH key value [value ...]
func (ex *Executor) handleRPUSH(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'rpush' command")
	}
	length, err := ex.store.RPush(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(length)), nil
}

// handleLPOP: LPOP key
func (ex *Executor) handleLPOP(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'lpop' command")
	}
	val, found, err := ex.store.LPop(args[0])
	if err != nil {
		return nil, err
	}
	if !found {
		return resp.NewNull(), nil
	}
	return resp.NewBulkString(val), nil
}

// handleRPOP: RPOP key
func (ex *Executor) handleRPOP(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'rpop' command")
	}
	val, found, err := ex.store.RPop(args[0])
	if err != nil {
		return nil, err
	}
	if !found {
		return resp.NewNull(), nil
	}
	return resp.NewBulkString(val), nil
}

// handleLRANGE: LRANGE key start stop
func (ex *Executor) handleLRANGE(args []string) (*resp.Value, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("wrong number of arguments for 'lrange' command")
	}
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	items, err := ex.store.LRange(args[0], start, stop)
	if err != nil {
		return nil, err
	}

	values := make([]*resp.Value, len(items))
	for i, item := range items {
		values[i] = resp.NewBulkString(item)
	}
	return resp.NewArray(values...), nil
}

// handleLLEN: LLEN key
func (ex *Executor) handleLLEN(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'llen' command")
	}
	length, err := ex.store.LLen(args[0])
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(length)), nil
}

// handleLINDEX: LINDEX key index
func (ex *Executor) handleLINDEX(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'lindex' command")
	}
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	val, found, err := ex.store.LIndex(args[0], index)
	if err != nil {
		return nil, err
	}
	if !found {
		return resp.NewNull(), nil
	}
	return resp.NewBulkString(val), nil
}