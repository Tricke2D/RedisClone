package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetGet verifies operasi dasar Set + Get
func TestSetGet(t *testing.T) {
	store := NewStore()

	_, err := store.Set("foo", "bar", nil)
	require.NoError(t, err)

	val, err := store.Get("foo")
	require.NoError(t, err)
	assert.Equal(t, "bar", val)
}

// TestGetNonExistent verifies Get pada key yang tidak ada
func TestGetNonExistent(t *testing.T) {
	store := NewStore()

	val, err := store.Get("nokey")
	require.NoError(t, err)
	assert.Nil(t, val)
}

// TestDelete verifies penghapusan key
func TestDelete(t *testing.T) {
	store := NewStore()
	store.Set("foo", "bar", nil)

	count, err := store.Delete("foo")
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	val, _ := store.Get("foo")
	assert.Nil(t, val)
}

// TestExists verifies pengecekan keberadaan key
func TestExists(t *testing.T) {
	store := NewStore()
	store.Set("key1", "value1", nil)

	count, err := store.Exists("key1", "key2")
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestExpiry verifies lazy expiry pada key dengan TTL
func TestExpiry(t *testing.T) {
	store := NewStore()
	ttl := 50 * time.Millisecond
	store.Set("temp", "value", &ttl)

	// Sebelum expired
	val, _ := store.Get("temp")
	assert.Equal(t, "value", val)

	// Setelah expired
	time.Sleep(100 * time.Millisecond)
	val, _ = store.Get("temp")
	assert.Nil(t, val)
}

// TestConcurrentAccess verifies thread-safety dengan banyak goroutine
func TestConcurrentAccess(t *testing.T) {
	store := NewStore()
	done := make(chan bool)

	// 50 goroutine menulis, 50 goroutine membaca secara bersamaan
	for i := 0; i < 50; i++ {
		go func(n int) {
			store.Set("key", n, nil)
			done <- true
		}(i)
	}
	for i := 0; i < 50; i++ {
		go func() {
			store.Get("key")
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}