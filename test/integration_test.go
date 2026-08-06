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

// TestServerBasicCommands tests basic GET/SET/DEL flow end-to-end
func TestServerBasicCommands(t *testing.T) {
	cfg := server.Config{Addr: "localhost:16379"} // Port berbeda untuk testing
	srv := server.NewServer(cfg)

	go srv.Start()
	time.Sleep(100 * time.Millisecond) // Beri waktu server untuk start

	defer srv.Shutdown()

	conn, err := net.Dial("tcp", "localhost:16379")
	require.NoError(t, err)
	defer conn.Close()

	parser := protocol.NewRespParser(conn)
	encoder := protocol.NewRespEncoder(conn)

	// Test SET
	cmd := resp.NewArray(
		resp.NewBulkString("SET"),
		resp.NewBulkString("testkey"),
		resp.NewBulkString("testvalue"),
	)
	require.NoError(t, encoder.Encode(cmd))

	result, err := parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeSimpleString, result.Type)
	assert.Equal(t, "OK", result.Str)

	// Test GET
	cmd = resp.NewArray(resp.NewBulkString("GET"), resp.NewBulkString("testkey"))
	require.NoError(t, encoder.Encode(cmd))

	result, err = parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeBulkString, result.Type)
	assert.Equal(t, "testvalue", result.Str)

	// Test DEL
	cmd = resp.NewArray(resp.NewBulkString("DEL"), resp.NewBulkString("testkey"))
	require.NoError(t, encoder.Encode(cmd))

	result, err = parser.Parse()
	require.NoError(t, err)
	assert.Equal(t, resp.TypeInteger, result.Type)
	assert.Equal(t, int64(1), result.Num)

	// Test GET setelah DEL (harus null)
	cmd = resp.NewArray(resp.NewBulkString("GET"), resp.NewBulkString("testkey"))
	require.NoError(t, encoder.Encode(cmd))

	result, err = parser.Parse()
	require.NoError(t, err)
	assert.True(t, result.IsNull())
}

// TestConcurrentConnections verifies multiple clients bisa connect simultaneously
func TestConcurrentConnections(t *testing.T) {
	cfg := server.Config{Addr: "localhost:16380"}
	srv := server.NewServer(cfg)

	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Shutdown()

	numClients := 10
	done := make(chan bool, numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer func() { done <- true }()

			conn, err := net.Dial("tcp", "localhost:16380")
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", clientID, err)
				return
			}
			defer conn.Close()

			encoder := protocol.NewRespEncoder(conn)
			parser := protocol.NewRespParser(conn)

			cmd := resp.NewArray(resp.NewBulkString("PING"))
			if err := encoder.Encode(cmd); err != nil {
				t.Errorf("Client %d failed to send: %v", clientID, err)
				return
			}

			result, err := parser.Parse()
			if err != nil {
				t.Errorf("Client %d failed to parse: %v", clientID, err)
				return
			}

			if result.Type != resp.TypeSimpleString || result.Str != "PONG" {
				t.Errorf("Client %d got unexpected response: %v", clientID, result)
			}
		}(i)
	}

	for i := 0; i < numClients; i++ {
		<-done
	}
}

// TestStringCommandsFlow tests INCR/APPEND/STRLEN via TCP connection asli
func TestStringCommandsFlow(t *testing.T) {
	cfg := server.Config{Addr: "localhost:16381"}
	srv := server.NewServer(cfg)

	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Shutdown()

	conn, err := net.Dial("tcp", "localhost:16381")
	require.NoError(t, err)
	defer conn.Close()

	parser := protocol.NewRespParser(conn)
	encoder := protocol.NewRespEncoder(conn)

	sendCmd := func(parts ...string) *resp.Value {
		values := make([]*resp.Value, len(parts))
		for i, p := range parts {
			values[i] = resp.NewBulkString(p)
		}
		require.NoError(t, encoder.Encode(resp.NewArray(values...)))
		result, err := parser.Parse()
		require.NoError(t, err)
		return result
	}

	// INCR dua kali
	result := sendCmd("INCR", "counter")
	assert.Equal(t, int64(1), result.Num)
	result = sendCmd("INCR", "counter")
	assert.Equal(t, int64(2), result.Num)

	// APPEND
	result = sendCmd("APPEND", "msg", "Hello")
	assert.Equal(t, int64(5), result.Num)

	// STRLEN
	result = sendCmd("STRLEN", "msg")
	assert.Equal(t, int64(5), result.Num)
}