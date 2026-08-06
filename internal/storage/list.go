package storage

import (
	"fmt"
	"time"
)

// getListUnsafe mengambil list dari key.
func (s *Store) getListUnsafe(key string) ([]string, error) {
	stored, ok := s.data[key]
	if !ok {
		return []string{}, nil
	}
	if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
		return []string{}, nil
	}
	list, ok := stored.Value.([]string)
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return list, nil
}

// setListUnsafe menyimpan list ke key.
func (s *Store) setListUnsafe(key string, list []string) {
	if len(list) == 0 {
		delete(s.data, key)
		return
	}

	var expiresAt *time.Time
	if existing, ok := s.data[key]; ok {
		expiresAt = existing.ExpiresAt
	}

	s.data[key] = &StoredValue{
		Value:      list,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
}

// LPush menyisipkan value di awal list.
func (s *Store) LPush(key string, values ...string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("at least one value required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return 0, err
	}

	for _, v := range values {
		list = append([]string{v}, list...)
	}

	s.setListUnsafe(key, list)
	return len(list), nil
}

// RPush menyisipkan value di akhir list.
func (s *Store) RPush(key string, values ...string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}
	if len(values) == 0 {
		return 0, fmt.Errorf("at least one value required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return 0, err
	}

	list = append(list, values...)

	s.setListUnsafe(key, list)
	return len(list), nil
}

// LPop menghapus dan mengembalikan elemen pertama list.
func (s *Store) LPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return "", false, err
	}
	if len(list) == 0 {
		return "", false, nil
	}

	val := list[0]
	list = list[1:]
	s.setListUnsafe(key, list)

	return val, true, nil
}

// RPop menghapus dan mengembalikan elemen terakhir list.
func (s *Store) RPop(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return "", false, err
	}
	if len(list) == 0 {
		return "", false, nil
	}

	val := list[len(list)-1]
	list = list[:len(list)-1]
	s.setListUnsafe(key, list)

	return val, true, nil
}

// LRange mengembalikan elemen list dalam rentang.
func (s *Store) LRange(key string, start, stop int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return nil, err
	}

	length := len(list)
	if length == 0 {
		return []string{}, nil
	}

	start = normalizeIndex(start, length)
	stop = normalizeIndex(stop, length)

	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop || start >= length {
		return []string{}, nil
	}

	result := make([]string, stop-start+1)
	copy(result, list[start:stop+1])
	return result, nil
}

// LLen mengembalikan jumlah elemen pada list.
func (s *Store) LLen(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}

// LIndex mengembalikan elemen list pada index tertentu.
func (s *Store) LIndex(key string, index int) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, err := s.getListUnsafe(key)
	if err != nil {
		return "", false, err
	}

	index = normalizeIndex(index, len(list))
	if index < 0 || index >= len(list) {
		return "", false, nil
	}

	return list[index], true, nil
}

// normalizeIndex mengonversi index negatif.
func normalizeIndex(idx, length int) int {
	if idx < 0 {
		idx = length + idx
	}
	return idx
}