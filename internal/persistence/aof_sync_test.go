package persistence

import (
	"os"
	"testing"
	"time"

	"github.com/company/redis-clone/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeStore untuk testing AOF
type fakeStoreForAOFTest struct {
	store *storage.Store
}

func newFakeStoreForAOFTest() *fakeStoreForAOFTest {
	return &fakeStoreForAOFTest{store: storage.NewStore()}
}

// fakeExecutor mengimplementasikan Replayer untuk testing
type fakeExecutorForAOFTest struct {
	store *storage.Store
}

func (f *fakeExecutorForAOFTest) Execute(cmd []string) (*storage.StoredValue, error) {
	switch cmd[0] {
	case "SET":
		f.store.Set(cmd[1], cmd[2], nil)
	case "DEL":
		f.store.Delete(cmd[1])
	}
	return nil, nil
}

// TestAOFEverySecondPolicyStillPersists verifies data tetap persist meski fsync tidak segera
func TestAOFEverySecondPolicyStillPersists(t *testing.T) {
	path := "test_everysec.aof"
	defer os.Remove(path)

	writer, err := NewAOFWriterWithPolicy(path, SyncEverySecond)
	require.NoError(t, err)

	require.NoError(t, writer.LogCommand([]string{"SET", "foo", "bar"}))
	require.NoError(t, writer.Close())

	store := storage.NewStore()
	executor := &fakeExecutorForAOFTest{store: store}
	reader := NewAOFReader(path)

	count, err := reader.Recover(executor)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	val, _ := store.Get("foo")
	assert.Equal(t, "bar", val)
}

// TestAOFBackgroundSyncDoesNotPanic verifies background syncer berjalan tanpa error
func TestAOFBackgroundSyncDoesNotPanic(t *testing.T) {
	path := "test_bgsync.aof"
	defer os.Remove(path)

	writer, err := NewAOFWriterWithPolicy(path, SyncEverySecond)
	require.NoError(t, err)

	writer.LogCommand([]string{"SET", "a", "1"})
	time.Sleep(1200 * time.Millisecond)
	writer.LogCommand([]string{"SET", "b", "2"})

	require.NoError(t, writer.Close())
}