package persistence

import (
	"os"
	"testing"

	"github.com/company/redis-clone/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRDBSaveAndLoad(t *testing.T) {
	path := "test_dump.rdb"
	defer os.Remove(path)

	store := storage.NewStore()
	store.Set("foo", "bar", nil)
	store.RPush("mylist", "x", "y")

	writer := NewRDBWriter(path)
	require.NoError(t, writer.Save(store))

	reader := NewRDBReader(path)
	assert.True(t, reader.Exists())

	restored := storage.NewStore()
	require.NoError(t, reader.Load(restored))

	val, _ := restored.Get("foo")
	assert.Equal(t, "bar", val)

	list, _ := restored.LRange("mylist", 0, -1)
	assert.Equal(t, []string{"x", "y"}, list)
}

func TestRDBReaderExistsFalseWhenNoFile(t *testing.T) {
	reader := NewRDBReader("does_not_exist.rdb")
	assert.False(t, reader.Exists())
}

func TestRDBOverwriteKeepsOldFileOnFailure(t *testing.T) {
	path := "test_overwrite.rdb"
	defer os.Remove(path)
	defer os.Remove(path + ".tmp")

	store1 := storage.NewStore()
	store1.Set("version", "1", nil)
	require.NoError(t, NewRDBWriter(path).Save(store1))

	store2 := storage.NewStore()
	store2.Set("version", "2", nil)
	require.NoError(t, NewRDBWriter(path).Save(store2))

	restored := storage.NewStore()
	require.NoError(t, NewRDBReader(path).Load(restored))

	val, _ := restored.Get("version")
	assert.Equal(t, "2", val)
}