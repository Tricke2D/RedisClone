package command

import (
	"testing"

	"github.com/company/redis-clone/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReplicaOfHandler struct {
	started string
	stopped bool
}

func (f *fakeReplicaOfHandler) StartReplication(masterAddr string) error {
	f.started = masterAddr
	return nil
}

func (f *fakeReplicaOfHandler) StopReplication() {
	f.stopped = true
}

func TestSlaveOfStart(t *testing.T) {
	store := storage.NewStore()
	handler := &fakeReplicaOfHandler{}
	executor := NewExecutor(store).WithReplicaOfHandler(handler)

	result, err := executor.Execute([]string{"SLAVEOF", "localhost", "6379"})
	require.NoError(t, err)
	assert.Equal(t, "OK", result.Str)
	assert.Equal(t, "localhost:6379", handler.started)
}

func TestSlaveOfNoOne(t *testing.T) {
	store := storage.NewStore()
	handler := &fakeReplicaOfHandler{}
	executor := NewExecutor(store).WithReplicaOfHandler(handler)

	result, err := executor.Execute([]string{"SLAVEOF", "NO", "ONE"})
	require.NoError(t, err)
	assert.Equal(t, "OK", result.Str)
	assert.True(t, handler.stopped)
}