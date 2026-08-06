package command

import (
	"fmt"

	"github.com/company/redis-clone/pkg/resp"
)

// handleHSET: HSET key field value [field value ...]
func (ex *Executor) handleHSET(args []string) (*resp.Value, error) {
	if len(args) < 3 || len(args)%2 != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'hset' command")
	}

	key := args[0]
	fieldValues := make(map[string]string)
	for i := 1; i < len(args); i += 2 {
		fieldValues[args[i]] = args[i+1]
	}

	newFields, err := ex.store.HSet(key, fieldValues)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(newFields)), nil
}

// handleHGET: HGET key field
func (ex *Executor) handleHGET(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'hget' command")
	}
	val, found, err := ex.store.HGet(args[0], args[1])
	if err != nil {
		return nil, err
	}
	if !found {
		return resp.NewNull(), nil
	}
	return resp.NewBulkString(val), nil
}

// handleHDEL: HDEL key field [field ...]
func (ex *Executor) handleHDEL(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'hdel' command")
	}
	deleted, err := ex.store.HDel(args[0], args[1:]...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(deleted)), nil
}

// handleHEXISTS: HEXISTS key field
func (ex *Executor) handleHEXISTS(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'hexists' command")
	}
	exists, err := ex.store.HExists(args[0], args[1])
	if err != nil {
		return nil, err
	}
	if exists {
		return resp.NewInteger(1), nil
	}
	return resp.NewInteger(0), nil
}

// handleHGETALL: HGETALL key
func (ex *Executor) handleHGETALL(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'hgetall' command")
	}
	hash, err := ex.store.HGetAll(args[0])
	if err != nil {
		return nil, err
	}

	values := make([]*resp.Value, 0, len(hash)*2)
	for field, val := range hash {
		values = append(values, resp.NewBulkString(field), resp.NewBulkString(val))
	}
	return resp.NewArray(values...), nil
}

// handleHKEYS: HKEYS key
func (ex *Executor) handleHKEYS(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'hkeys' command")
	}
	keys, err := ex.store.HKeys(args[0])
	if err != nil {
		return nil, err
	}
	values := make([]*resp.Value, len(keys))
	for i, k := range keys {
		values[i] = resp.NewBulkString(k)
	}
	return resp.NewArray(values...), nil
}

// handleHVALS: HVALS key
func (ex *Executor) handleHVALS(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'hvals' command")
	}
	vals, err := ex.store.HVals(args[0])
	if err != nil {
		return nil, err
	}
	values := make([]*resp.Value, len(vals))
	for i, v := range vals {
		values[i] = resp.NewBulkString(v)
	}
	return resp.NewArray(values...), nil
}

// handleHLEN: HLEN key
func (ex *Executor) handleHLEN(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'hlen' command")
	}
	length, err := ex.store.HLen(args[0])
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(length)), nil
}