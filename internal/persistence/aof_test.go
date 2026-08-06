package persistence

import (
	"os"
	"testing"

	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeExecutor struct {
	store *storage.Store
}

func (f *fakeExecutor) Execute(cmd []string) (*resp.Value, error) {
	switch cmd[0] {
	case "SET":
		f.store.Set(cmd[1], cmd[2], nil)
	case "DEL":
		f.store.Delete(cmd[1])
	}
	return resp.NewSimpleString("OK"), nil
}

func TestAOFWriteAndRecover(t *testing.T) {
	path := "test_aof.log"
	defer os.Remove(path)

	writer, err := NewAOFWriter(path)
	require.NoError(t, err)

	require.NoError(t, writer.LogCommand([]string{"SET", "foo", "bar"}))
	require.NoError(t, writer.LogCommand([]string{"SET", "baz", "qux"}))
	require.NoError(t, writer.Close())

	store := storage.NewStore()
	executor := &fakeExecutor{store: store}
	reader := NewAOFReader(path)

	count, err := reader.Recover(executor)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	val, _ := store.Get("foo")
	assert.Equal(t, "bar", val)
	val, _ = store.Get("baz")
	assert.Equal(t, "qux", val)
}

func TestAOFRecoverNonExistentFile(t *testing.T) {
	store := storage.NewStore()
	executor := &fakeExecutor{store: store}
	reader := NewAOFReader("does_not_exist.log")

	count, err := reader.Recover(executor)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestAOFDeleteCommandReplay(t *testing.T) {
	path := "test_aof_del.log"
	defer os.Remove(path)

	writer, err := NewAOFWriter(path)
	require.NoError(t, err)
	writer.LogCommand([]string{"SET", "foo", "bar"})
	writer.LogCommand([]string{"DEL", "foo"})
	writer.Close()

	store := storage.NewStore()
	executor := &fakeExecutor{store: store}
	reader := NewAOFReader(path)

	count, err := reader.Recover(executor)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	val, _ := store.Get("foo")
	assert.Nil(t, val)
}