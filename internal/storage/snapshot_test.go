package storage

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotRoundTrip(t *testing.T) {
	store := NewStore()
	store.Set("str_key", "hello", nil)
	store.RPush("list_key", "a", "b", "c")
	store.HSet("hash_key", map[string]string{"field1": "value1"})
	store.SAdd("set_key", "member1", "member2")

	ttl := 100 * time.Second
	store.Set("ttl_key", "expiring", &ttl)

	var buf bytes.Buffer
	require.NoError(t, store.EncodeSnapshot(&buf))

	restored := NewStore()
	require.NoError(t, restored.DecodeSnapshot(&buf))

	strVal, _ := restored.Get("str_key")
	assert.Equal(t, "hello", strVal)

	listVal, _ := restored.LRange("list_key", 0, -1)
	assert.Equal(t, []string{"a", "b", "c"}, listVal)

	hashVal, _ := restored.HGetAll("hash_key")
	assert.Equal(t, "value1", hashVal["field1"])

	setVal, _ := restored.SMembers("set_key")
	assert.ElementsMatch(t, []string{"member1", "member2"}, setVal)

	ttlRemaining, _ := restored.TTL("ttl_key")
	assert.True(t, ttlRemaining > 0 && ttlRemaining <= 100)
}

func TestSnapshotEmptyStore(t *testing.T) {
	store := NewStore()

	var buf bytes.Buffer
	require.NoError(t, store.EncodeSnapshot(&buf))

	restored := NewStore()
	require.NoError(t, restored.DecodeSnapshot(&buf))
	assert.Equal(t, 0, restored.Size())
}