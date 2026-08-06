package storage

import (
	"fmt"
	"time"
)

// Expire mengeset TTL dalam detik.
func (s *Store) Expire(key string, seconds int64) (bool, error) {
	return s.pexpire(key, time.Duration(seconds)*time.Second)
}

// PExpire mengeset TTL dalam milidetik.
func (s *Store) PExpire(key string, milliseconds int64) (bool, error) {
	return s.pexpire(key, time.Duration(milliseconds)*time.Millisecond)
}

func (s *Store) pexpire(key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.data[key]
	if !ok {
		return false, nil
	}
	if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
		delete(s.data, key)
		return false, nil
	}

	expiresAt := time.Now().Add(ttl)
	stored.ExpiresAt = &expiresAt
	return true, nil
}

// TTL mengembalikan sisa waktu dalam detik.
func (s *Store) TTL(key string) (int64, error) {
	ms, err := s.PTTL(key)
	if err != nil {
		return 0, err
	}
	if ms < 0 {
		return ms, nil
	}
	return (ms + 999) / 1000, nil
}

// PTTL mengembalikan sisa waktu dalam milidetik.
func (s *Store) PTTL(key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stored, ok := s.data[key]
	if !ok {
		return -2, nil
	}
	if stored.ExpiresAt == nil {
		return -1, nil
	}

	remaining := time.Until(*stored.ExpiresAt)
	if remaining <= 0 {
		return -2, nil
	}
	return remaining.Milliseconds(), nil
}

// Persist menghapus TTL dari key.
func (s *Store) Persist(key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.data[key]
	if !ok || stored.ExpiresAt == nil {
		return false, nil
	}

	stored.ExpiresAt = nil
	return true, nil
}