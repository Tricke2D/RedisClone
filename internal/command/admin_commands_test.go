package command

import (
	"testing"

	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSnapshotter struct {
	called bool
	err    error
}

func (f *fakeSnapshotter) SaveSnapshot() error {
	f.called = true
	return f.err
}

func TestSaveCommand(t *testing.T) {
	store := storage.NewStore()
	snap := &fakeSnapshotter{}
	executor := NewExecutor(store).WithSnapshotter(snap)

	result, err := executor.Execute([]string{"SAVE"})
	require.NoError(t, err)
	assert.Equal(t, "OK", result.Str)
	assert.True(t, snap.called)
}

func TestSaveCommandWithoutSnapshotter(t *testing.T) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	result, err := executor.Execute([]string{"SAVE"})
	require.NoError(t, err)
	assert.Equal(t, resp.TypeError, result.Type)
}