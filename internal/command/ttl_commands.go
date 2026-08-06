package command

import (
	"fmt"
	"strconv"

	"github.com/company/redis-clone/pkg/resp"
)

// handleEXPIRE: EXPIRE key seconds
func (ex *Executor) handleEXPIRE(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'expire' command")
	}
	seconds, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	ok, err := ex.store.Expire(args[0], seconds)
	if err != nil {
		return nil, err
	}
	if ok {
		return resp.NewInteger(1), nil
	}
	return resp.NewInteger(0), nil
}

// handlePEXPIRE: PEXPIRE key milliseconds
func (ex *Executor) handlePEXPIRE(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'pexpire' command")
	}
	ms, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("value is not an integer or out of range")
	}

	ok, err := ex.store.PExpire(args[0], ms)
	if err != nil {
		return nil, err
	}
	if ok {
		return resp.NewInteger(1), nil
	}
	return resp.NewInteger(0), nil
}

// handleTTL: TTL key
func (ex *Executor) handleTTL(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'ttl' command")
	}
	seconds, err := ex.store.TTL(args[0])
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(seconds), nil
}

// handlePTTL: PTTL key
func (ex *Executor) handlePTTL(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'pttl' command")
	}
	ms, err := ex.store.PTTL(args[0])
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(ms), nil
}

// handlePERSIST: PERSIST key
func (ex *Executor) handlePERSIST(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'persist' command")
	}
	ok, err := ex.store.Persist(args[0])
	if err != nil {
		return nil, err
	}
	if ok {
		return resp.NewInteger(1), nil
	}
	return resp.NewInteger(0), nil
}