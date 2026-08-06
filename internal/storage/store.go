package storage

import (
	"fmt"
	"sync"
	"time"
)

// Store represents thread-safe in-memory key-value storage
type Store struct {
	mu   sync.RWMutex
	data map[string]*StoredValue
}

// StoredValue membungkus value asli beserta metadata
type StoredValue struct {
	Value      interface{}
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	LastAccess time.Time
}

// NewStore creates new in-memory store instance
func NewStore() *Store {
	return &Store{data: make(map[string]*StoredValue)}
}

// Set stores value untuk given key dengan optional TTL
func (s *Store) Set(key string, value interface{}, ttl *time.Duration) (interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var oldVal interface{}
	if existing, ok := s.data[key]; ok {
		oldVal = existing.Value
	}

	var expiresAt *time.Time
	if ttl != nil {
		t := time.Now().Add(*ttl)
		expiresAt = &t
	}

	s.data[key] = &StoredValue{
		Value:      value,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}

	return oldVal, nil
}

// Get retrieves value untuk given key
func (s *Store) Get(key string) (interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	stored, ok := s.data[key]
	if !ok {
		return nil, nil
	}

	// Lazy expiry: cek saat diakses
	if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
		return nil, nil
	}

	return stored.Value, nil
}

// Delete removes key dari store
func (s *Store) Delete(keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, fmt.Errorf("at least one key required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0
	for _, key := range keys {
		if _, ok := s.data[key]; ok {
			delete(s.data, key)
			deleted++
		}
	}

	return deleted, nil
}

// Exists checks jika key ada (dan belum expired)
func (s *Store) Exists(keys ...string) (int, error) {
	if len(keys) == 0 {
		return 0, fmt.Errorf("at least one key required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, key := range keys {
		if stored, ok := s.data[key]; ok {
			if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
				continue
			}
			count++
		}
	}

	return count, nil
}

// Size returns total jumlah key (approx, tidak cek expiry)
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Clear removes all keys dari store
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*StoredValue)
	return nil
}

// Keys returns semua key yang match pattern (for KEYS command)
func (s *Store) Keys(pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0)
	for k, stored := range s.data {
		if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
			continue
		}
		if pattern == "*" || pattern == k {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

// RemoveExpired melakukan ACTIVE EXPIRY: scan seluruh key dan hapus yang sudah expired.
// Dipanggil secara berkala oleh background goroutine (expiry manager).
// Returns jumlah key yang dihapus.
func (s *Store) RemoveExpired() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	removed := 0
	for key, stored := range s.data {
		if stored.ExpiresAt != nil && now.After(*stored.ExpiresAt) {
			delete(s.data, key)
			removed++
		}
	}
	return removed
}