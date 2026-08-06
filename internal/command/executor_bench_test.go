package command

import (
	"fmt"
	"testing"

	"github.com/company/redis-clone/internal/storage"
)

func BenchmarkGET(b *testing.B) {
	store := storage.NewStore()
	executor := NewExecutor(store)
	executor.Execute([]string{"SET", "bench_key", "bench_value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute([]string{"GET", "bench_key"})
	}
}

func BenchmarkSET(b *testing.B) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute([]string{"SET", fmt.Sprintf("key%d", i), "value"})
	}
}

func BenchmarkLPUSH(b *testing.B) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute([]string{"LPUSH", "bench_list", "item"})
	}
}

func BenchmarkHSET(b *testing.B) {
	store := storage.NewStore()
	executor := NewExecutor(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute([]string{"HSET", "bench_hash", fmt.Sprintf("field%d", i), "value"})
	}
}

func BenchmarkConcurrentGETSET(b *testing.B) {
	store := storage.NewStore()
	executor := NewExecutor(store)
	executor.Execute([]string{"SET", "shared_key", "value"})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			executor.Execute([]string{"GET", "shared_key"})
		}
	})
}