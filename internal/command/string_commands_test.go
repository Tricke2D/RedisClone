package command

import (
	"testing"

	"github.com/company/redis-clone/internal/storage"
	"github.com/company/redis-clone/pkg/resp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestINCR verifies INCR command functionality
func TestINCR(t *testing.T) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	// INCR pada key baru (mulai dari 0, jadi 1)
	result, err := executor.Execute([]string{"INCR", "counter"})
	require.NoError(t, err)
	assert.Equal(t, resp.TypeInteger, result.Type)
	assert.Equal(t, int64(1), result.Num)

	// INCR lagi (jadi 2)
	result, err = executor.Execute([]string{"INCR", "counter"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Num)

	// Set ke string non-numeric, INCR harus error
	executor.Execute([]string{"SET", "notnum", "abc"})
	result, err = executor.Execute([]string{"INCR", "notnum"})
	assert.NoError(t, err) // Execute mengembalikan error via response, bukan Go error
	assert.Equal(t, resp.TypeError, result.Type)
}

// TestDECR verifies DECR command
func TestDECR(t *testing.T) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	result, err := executor.Execute([]string{"DECR", "counter"})
	require.NoError(t, err)
	assert.Equal(t, int64(-1), result.Num)
}

// TestAPPEND verifies APPEND command
func TestAPPEND(t *testing.T) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	result, err := executor.Execute([]string{"APPEND", "msg", "Hello"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), result.Num)

	result, err = executor.Execute([]string{"APPEND", "msg", " World"})
	require.NoError(t, err)
	assert.Equal(t, int64(11), result.Num)

	result, err = executor.Execute([]string{"GET", "msg"})
	require.NoError(t, err)
	assert.Equal(t, "Hello World", result.Str)
}

// TestSTRLEN verifies STRLEN command
func TestSTRLEN(t *testing.T) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	// Key tidak ada
	result, err := executor.Execute([]string{"STRLEN", "nokey"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), result.Num)

	// Key ada
	executor.Execute([]string{"SET", "mykey", "hello"})
	result, err = executor.Execute([]string{"STRLEN", "mykey"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), result.Num)
}