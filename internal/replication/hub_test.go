package replication

import (
	"net"
	"testing"
	"time"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubPropagateToRegisteredReplica(t *testing.T) {
	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	hub := NewHub()
	encoder := protocol.NewRespEncoder(serverSide)
	hub.Register("fake-replica-addr", encoder)

	resultCh := make(chan *resp.Value, 1)
	go func() {
		parser := protocol.NewRespParser(clientSide)
		val, err := parser.Parse()
		require.NoError(t, err)
		resultCh <- val
	}()

	hub.Propagate([]string{"SET", "foo", "bar"})

	select {
	case val := <-resultCh:
		assert.Equal(t, resp.TypeArray, val.Type)
		assert.Equal(t, "SET", val.Arr[0].Str)
		assert.Equal(t, "foo", val.Arr[1].Str)
		assert.Equal(t, "bar", val.Arr[2].Str)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for propagated command")
	}
}

func TestHubUnregisterStopsPropagation(t *testing.T) {
	hub := NewHub()
	assert.Equal(t, 0, hub.ReplicaCount())

	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	encoder := protocol.NewRespEncoder(serverSide)
	hub.Register("addr1", encoder)
	assert.Equal(t, 1, hub.ReplicaCount())

	hub.Unregister("addr1")
	assert.Equal(t, 0, hub.ReplicaCount())
}