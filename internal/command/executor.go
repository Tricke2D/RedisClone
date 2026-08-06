package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
)

// AOFLogger adalah interface untuk mencatat write command ke AOF.
type AOFLogger interface {
	LogCommand(cmd []string) error
}

// Snapshotter adalah interface untuk memicu RDB snapshot manual.
type Snapshotter interface {
	SaveSnapshot() error
}

// ReplicaPropagator adalah interface untuk mem-broadcast write command ke replica.
type ReplicaPropagator interface {
	Propagate(cmd []string)
}

// ReplicaOfHandler adalah interface untuk menjalankan command SLAVEOF.
type ReplicaOfHandler interface {
	StartReplication(masterAddr string) error
	StopReplication()
}

// Executor handles command parsing dan execution
type Executor struct {
	store       *storage.Store
	handlers    map[string]CommandHandler
	aof         AOFLogger
	snapshotter Snapshotter
	propagator  ReplicaPropagator
	replicaOf   ReplicaOfHandler
	writeCmds   map[string]bool
}

// CommandHandler adalah signature function untuk implementasi command
// Semua handler adalah method dari Executor dengan signature func([]string) (*resp.Value, error)
type CommandHandler func([]string) (*resp.Value, error)

// NewExecutor creates command executor instance dasar.
func NewExecutor(store *storage.Store) *Executor {
	ex := &Executor{
		store:    store,
		handlers: make(map[string]CommandHandler),
	}
	ex.registerHandlers()
	ex.registerWriteCommands()
	return ex
}

// NewExecutorWithAOF creates command executor dengan AOF logging.
func NewExecutorWithAOF(store *storage.Store, aof AOFLogger) *Executor {
	ex := NewExecutor(store)
	ex.aof = aof
	return ex
}

// WithSnapshotter mengaktifkan command SAVE.
func (ex *Executor) WithSnapshotter(s Snapshotter) *Executor {
	ex.snapshotter = s
	return ex
}

// WithPropagator memasang replica propagator.
func (ex *Executor) WithPropagator(p ReplicaPropagator) *Executor {
	ex.propagator = p
	return ex
}

// WithReplicaOfHandler mengaktifkan command SLAVEOF.
func (ex *Executor) WithReplicaOfHandler(h ReplicaOfHandler) *Executor {
	ex.replicaOf = h
	return ex
}

// registerHandlers mendaftarkan semua command yang didukung
func (ex *Executor) registerHandlers() {
	// String commands (Fase 1)
	ex.handlers["GET"] = ex.handleGET
	ex.handlers["SET"] = ex.handleSET
	ex.handlers["DEL"] = ex.handleDEL
	ex.handlers["EXISTS"] = ex.handleEXISTS
	ex.handlers["INCR"] = ex.handleINCR
	ex.handlers["DECR"] = ex.handleDECR
	ex.handlers["APPEND"] = ex.handleAPPEND
	ex.handlers["STRLEN"] = ex.handleSTRLEN
	ex.handlers["KEYS"] = ex.handleKEYS

	// Connection commands (Fase 1)
	ex.handlers["PING"] = ex.handlePING
	ex.handlers["ECHO"] = ex.handleECHO
	ex.handlers["SELECT"] = ex.handleSELECT

	// List commands (Fase 2)
	ex.handlers["LPUSH"] = ex.handleLPUSH
	ex.handlers["RPUSH"] = ex.handleRPUSH
	ex.handlers["LPOP"] = ex.handleLPOP
	ex.handlers["RPOP"] = ex.handleRPOP
	ex.handlers["LRANGE"] = ex.handleLRANGE
	ex.handlers["LLEN"] = ex.handleLLEN
	ex.handlers["LINDEX"] = ex.handleLINDEX

	// Hash commands (Fase 2)
	ex.handlers["HSET"] = ex.handleHSET
	ex.handlers["HGET"] = ex.handleHGET
	ex.handlers["HDEL"] = ex.handleHDEL
	ex.handlers["HEXISTS"] = ex.handleHEXISTS
	ex.handlers["HGETALL"] = ex.handleHGETALL
	ex.handlers["HKEYS"] = ex.handleHKEYS
	ex.handlers["HVALS"] = ex.handleHVALS
	ex.handlers["HLEN"] = ex.handleHLEN

	// Set commands (Fase 2)
	ex.handlers["SADD"] = ex.handleSADD
	ex.handlers["SREM"] = ex.handleSREM
	ex.handlers["SMEMBERS"] = ex.handleSMEMBERS
	ex.handlers["SCARD"] = ex.handleSCARD
	ex.handlers["SISMEMBER"] = ex.handleSISMEMBER

	// TTL/Expiry commands (Fase 2)
	ex.handlers["EXPIRE"] = ex.handleEXPIRE
	ex.handlers["PEXPIRE"] = ex.handlePEXPIRE
	ex.handlers["TTL"] = ex.handleTTL
	ex.handlers["PTTL"] = ex.handlePTTL
	ex.handlers["PERSIST"] = ex.handlePERSIST

	// Admin & Replication commands (Fase 3)
	ex.handlers["SAVE"] = ex.handleSAVE
	ex.handlers["SLAVEOF"] = ex.handleSLAVEOF
}

// registerWriteCommands mendaftarkan command write
func (ex *Executor) registerWriteCommands() {
	writeCommandNames := []string{
		"SET", "DEL", "INCR", "DECR", "APPEND",
		"LPUSH", "RPUSH", "LPOP", "RPOP",
		"HSET", "HDEL",
		"SADD", "SREM",
		"EXPIRE", "PEXPIRE", "PERSIST",
	}
	ex.writeCmds = make(map[string]bool, len(writeCommandNames))
	for _, name := range writeCommandNames {
		ex.writeCmds[name] = true
	}
}

// Execute parses command dan routes ke handler
func (ex *Executor) Execute(cmd []string) (*resp.Value, error) {
	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmdName := strings.ToUpper(cmd[0])

	handler, ok := ex.handlers[cmdName]
	if !ok {
		return resp.NewError(fmt.Sprintf("ERR unknown command '%s'", cmd[0])), nil
	}

	result, err := handler(cmd[1:])
	if err != nil {
		return resp.NewError(fmt.Sprintf("ERR %v", err)), nil
	}

	isSuccessfulWrite := ex.writeCmds[cmdName] && result.Type != resp.TypeError

	if isSuccessfulWrite && ex.aof != nil {
		if logErr := ex.aof.LogCommand(cmd); logErr != nil {
			return result, fmt.Errorf("AOF write failed: %w", logErr)
		}
	}

	if isSuccessfulWrite && ex.propagator != nil {
		ex.propagator.Propagate(cmd)
	}

	return result, nil
}

// === Connection Commands (Fase 1) ===

func (ex *Executor) handlePING(args []string) (*resp.Value, error) {
	if len(args) == 0 {
		return resp.NewSimpleString("PONG"), nil
	}
	if len(args) == 1 {
		return resp.NewBulkString(args[0]), nil
	}
	return nil, fmt.Errorf("wrong number of arguments for 'ping' command")
}

func (ex *Executor) handleECHO(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'echo' command")
	}
	return resp.NewBulkString(args[0]), nil
}

func (ex *Executor) handleSELECT(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'select' command")
	}
	return resp.NewSimpleString("OK"), nil
}

// === Basic Key Commands (Fase 1) ===

func (ex *Executor) handleGET(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'get' command")
	}
	val, err := ex.store.Get(args[0])
	if err != nil {
		return nil, err
	}
	if val == nil {
		return resp.NewNull(), nil
	}
	strVal, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return resp.NewBulkString(strVal), nil
}

func (ex *Executor) handleSET(args []string) (*resp.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'set' command")
	}
	_, err := ex.store.Set(args[0], args[1], nil)
	if err != nil {
		return nil, err
	}
	return resp.NewSimpleString("OK"), nil
}

func (ex *Executor) handleDEL(args []string) (*resp.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("wrong number of arguments for 'del' command")
	}
	count, err := ex.store.Delete(args...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(count)), nil
}

func (ex *Executor) handleEXISTS(args []string) (*resp.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("wrong number of arguments for 'exists' command")
	}
	count, err := ex.store.Exists(args...)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(count)), nil
}

func (ex *Executor) handleKEYS(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'keys' command")
	}
	keys, err := ex.store.Keys(args[0])
	if err != nil {
		return nil, err
	}
	values := make([]*resp.Value, len(keys))
	for i, k := range keys {
		values[i] = resp.NewBulkString(k)
	}
	return resp.NewArray(values...), nil
}

// === String Numeric Commands (Fase 1) ===

func (ex *Executor) handleINCR(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'incr' command")
	}
	key := args[0]
	val, err := ex.store.Get(key)
	if err != nil {
		return nil, err
	}
	var currentNum int64 = 0
	if val != nil {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		num, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value is not an integer or out of range")
		}
		currentNum = num
	}
	newNum := currentNum + 1
	_, err = ex.store.Set(key, strconv.FormatInt(newNum, 10), nil)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(newNum), nil
}

func (ex *Executor) handleDECR(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'decr' command")
	}
	key := args[0]
	val, err := ex.store.Get(key)
	if err != nil {
		return nil, err
	}
	var currentNum int64 = 0
	if val != nil {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		num, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("value is not an integer or out of range")
		}
		currentNum = num
	}
	newNum := currentNum - 1
	_, err = ex.store.Set(key, strconv.FormatInt(newNum, 10), nil)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(newNum), nil
}

func (ex *Executor) handleAPPEND(args []string) (*resp.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'append' command")
	}
	key := args[0]
	appendValue := args[1]
	val, err := ex.store.Get(key)
	if err != nil {
		return nil, err
	}
	var currentStr string
	if val != nil {
		strVal, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		currentStr = strVal
	}
	newStr := currentStr + appendValue
	_, err = ex.store.Set(key, newStr, nil)
	if err != nil {
		return nil, err
	}
	return resp.NewInteger(int64(len(newStr))), nil
}

func (ex *Executor) handleSTRLEN(args []string) (*resp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments for 'strlen' command")
	}
	key := args[0]
	val, err := ex.store.Get(key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return resp.NewInteger(0), nil
	}
	strVal, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return resp.NewInteger(int64(len(strVal))), nil
}