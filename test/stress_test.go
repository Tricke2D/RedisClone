package test

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/internal/server"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStressConcurrentClients(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	addr := "localhost:16390"
	srv := server.NewServer(server.Config{Addr: addr})
	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Shutdown()

	const numClients = 100
	const opsPerClient = 100

	var wg sync.WaitGroup
	errCh := make(chan error, numClients)

	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", addr)
			if err != nil {
				errCh <- fmt.Errorf("client %d dial: %w", clientID, err)
				return
			}
			defer conn.Close()

			encoder := protocol.NewRespEncoder(conn)
			parser := protocol.NewRespParser(conn)

			for j := 0; j < opsPerClient; j++ {
				cmd := resp.NewArray(
					resp.NewBulkString("SADD"),
					resp.NewBulkString("shared_set"),
					resp.NewBulkString(fmt.Sprintf("client%d_item%d", clientID, j)),
				)
				if err := encoder.Encode(cmd); err != nil {
					errCh <- fmt.Errorf("client %d encode: %w", clientID, err)
					return
				}
				if _, err := parser.Parse(); err != nil {
					errCh <- fmt.Errorf("client %d parse: %w", clientID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Error(err)
	}

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	encoder := protocol.NewRespEncoder(conn)
	parser := protocol.NewRespParser(conn)
	encoder.Encode(resp.NewArray(resp.NewBulkString("SCARD"), resp.NewBulkString("shared_set")))
	result, err := parser.Parse()
	require.NoError(t, err)

	assert.Equal(t, int64(numClients*opsPerClient), result.Num)
}

func TestStressMemoryStable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	addr := "localhost:16391"
	srv := server.NewServer(server.Config{Addr: addr})
	go srv.Start()
	time.Sleep(100 * time.Millisecond)
	defer srv.Shutdown()

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	encoder := protocol.NewRespEncoder(conn)
	parser := protocol.NewRespParser(conn)

	for i := 0; i < 10000; i++ {
		encoder.Encode(resp.NewArray(
			resp.NewBulkString("SET"),
			resp.NewBulkString(fmt.Sprintf("stress_key_%d", i)),
			resp.NewBulkString("value"),
		))
		parser.Parse()
		encoder.Encode(resp.NewArray(
			resp.NewBulkString("EXPIRE"),
			resp.NewBulkString(fmt.Sprintf("stress_key_%d", i)),
			resp.NewBulkString("1"),
		))
		parser.Parse()
	}

	time.Sleep(2 * time.Second)

	encoder.Encode(resp.NewArray(resp.NewBulkString("KEYS"), resp.NewBulkString("*")))
	result, err := parser.Parse()
	require.NoError(t, err)

	assert.Equal(t, 0, len(result.Arr))
}