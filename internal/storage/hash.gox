package storage

import (
	"fmt"
	"time"
)

// getHashUnsafe mengambil hash map dari key.
func (s *Store) getHashUnsafe(key string) (map[string]string, error) {
	stored, ok := s.data[key]
	if !ok {
		return map[string]string{}, nil
	}
	if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
		return map[string]string{}, nil
	}
	hash, ok := stored.Value.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return hash, nil
}

// setHashUnsafe menyimpan hash map ke key.
func (s *Store) setHashUnsafe(key string, hash map[string]string) {
	if len(hash) == 0 {
		delete(s.data, key)
		return
	}

	var expiresAt *time.Time
	if existing, ok := s.data[key]; ok {
		expiresAt = existing.ExpiresAt
	}

	s.data[key] = &StoredValue{
		Value:      hash,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
}

// HSet mengeset satu atau lebih pasangan field-value pada hash.
func (s *Store) HSet(key string, fieldValues map[string]string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.getHashUnsafe(key)
	if err != nil {
		return 0, err
	}

	hash := make(map[string]string, len(existing)+len(fieldValues))
	for k, v := range existing {
		hash[k] = v
	}

	newFields := 0
	for field, value := range fieldValues {
		if _, exists := hash[field]; !exists {
			newFields++
		}
		hash[field] = value
	}

	s.setHashUnsafe(key, hash)
	return newFields, nil
}

// HGet mengambil value dari field tertentu pada hash.
func (s *Store) HGet(key, field string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, err := s.getHashUnsafe(key)
	if err != nil {
		return "", false, err
	}
	val, ok := hash[field]
	return val, ok, nil
}

// HDel menghapus satu atau lebih field dari hash.
func (s *Store) HDel(key string, fields ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.getHashUnsafe(key)
	if err != nil {
		return 0, err
	}

	hash := make(map[string]string, len(existing))
	for k, v := range existing {
		hash[k] = v
	}

	deleted := 0
	for _, f := range fields {
		if _, ok := hash[f]; ok {
			delete(hash, f)
			deleted++
		}
	}

	s.setHashUnsafe(key, hash)
	return deleted, nil
}

// HExists mengecek apakah field tertentu ada pada hash.
func (s *Store) HExists(key, field string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, err := s.getHashUnsafe(key)
	if err != nil {
		return false, err
	}
	_, ok := hash[field]
	return ok, nil
}

// HGetAll mengembalikan seluruh field-value pada hash.
func (s *Store) HGetAll(key string) (map[string]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getHashUnsafe(key)
}

// HKeys mengembalikan semua nama field pada hash.
func (s *Store) HKeys(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, err := s.getHashUnsafe(key)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(hash))
	for k := range hash {
		keys = append(keys, k)
	}
	return keys, nil
}

// HVals mengembalikan semua value pada hash.
func (s *Store) HVals(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, err := s.getHashUnsafe(key)
	if err != nil {
		return nil, err
	}
	vals := make([]string, 0, len(hash))
	for _, v := range hash {
		vals = append(vals, v)
	}
	return vals, nil
}

// HLen mengembalikan jumlah field pada hash.
func (s *Store) HLen(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hash, err := s.getHashUnsafe(key)
	if err != nil {
		return 0, err
	}
	return len(hash), nil
}