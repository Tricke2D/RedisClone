package test

import (
	"net"
	"testing"
	"time"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/internal/server"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sendCommand(t *testing.T, conn net.Conn, parts ...string) *resp.Value {
	t.Helper()
	values := make([]*resp.Value, len(parts))
	for i, p := range parts {
		values[i] = resp.NewBulkString(p)
	}
	encoder := protocol.NewRespEncoder(conn)
	require.NoError(t, encoder.Encode(resp.NewArray(values...)))

	parser := protocol.NewRespParser(conn)
	val, err := parser.Parse()
	require.NoError(t, err)
	return val
}

func TestMasterSlaveReplication(t *testing.T) {
	masterAddr := "localhost:16400"
	replicaAddr := "localhost:16401"

	master := server.NewServer(server.Config{Addr: masterAddr})
	go master.Start()
	time.Sleep(100 * time.Millisecond)
	defer master.Shutdown()

	masterConn, err := net.Dial("tcp", masterAddr)
	require.NoError(t, err)
	sendCommand(t, masterConn, "SET", "existing_key", "existing_value")

	replica := server.NewServer(server.Config{Addr: replicaAddr})
	go replica.Start()
	time.Sleep(100 * time.Millisecond)
	defer replica.Shutdown()

	replicaConn, err := net.Dial("tcp", replicaAddr)
	require.NoError(t, err)
	defer replicaConn.Close()

	result := sendCommand(t, replicaConn, "SLAVEOF", "localhost", "16400")
	assert.Equal(t, "OK", result.Str)

	time.Sleep(300 * time.Millisecond)

	result = sendCommand(t, replicaConn, "GET", "existing_key")
	assert.Equal(t, "existing_value", result.Str)

	sendCommand(t, masterConn, "SET", "new_key", "new_value")
	time.Sleep(300 * time.Millisecond)

	result = sendCommand(t, replicaConn, "GET", "new_key")
	assert.Equal(t, "new_value", result.Str)
}