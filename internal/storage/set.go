package storage

import (
	"fmt"
	"time"
)

// getSetUnsafe mengambil set dari key.
func (s *Store) getSetUnsafe(key string) (map[string]struct{}, error) {
	stored, ok := s.data[key]
	if !ok {
		return map[string]struct{}{}, nil
	}
	if stored.ExpiresAt != nil && time.Now().After(*stored.ExpiresAt) {
		return map[string]struct{}{}, nil
	}
	set, ok := stored.Value.(map[string]struct{})
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return set, nil
}

// setSetUnsafe menyimpan set ke key.
func (s *Store) setSetUnsafe(key string, set map[string]struct{}) {
	if len(set) == 0 {
		delete(s.data, key)
		return
	}

	var expiresAt *time.Time
	if existing, ok := s.data[key]; ok {
		expiresAt = existing.ExpiresAt
	}

	s.data[key] = &StoredValue{
		Value:      set,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}
}

// SAdd menambahkan member ke set.
func (s *Store) SAdd(key string, members ...string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.getSetUnsafe(key)
	if err != nil {
		return 0, err
	}

	set := make(map[string]struct{}, len(existing)+len(members))
	for k := range existing {
		set[k] = struct{}{}
	}

	added := 0
	for _, m := range members {
		if _, exists := set[m]; !exists {
			set[m] = struct{}{}
			added++
		}
	}

	s.setSetUnsafe(key, set)
	return added, nil
}

// SRem menghapus member dari set.
func (s *Store) SRem(key string, members ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.getSetUnsafe(key)
	if err != nil {
		return 0, err
	}

	set := make(map[string]struct{}, len(existing))
	for k := range existing {
		set[k] = struct{}{}
	}

	removed := 0
	for _, m := range members {
		if _, exists := set[m]; exists {
			delete(set, m)
			removed++
		}
	}

	s.setSetUnsafe(key, set)
	return removed, nil
}

// SMembers mengembalikan semua member pada set.
func (s *Store) SMembers(key string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set, err := s.getSetUnsafe(key)
	if err != nil {
		return nil, err
	}
	members := make([]string, 0, len(set))
	for m := range set {
		members = append(members, m)
	}
	return members, nil
}

// SCard mengembalikan jumlah member pada set.
func (s *Store) SCard(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set, err := s.getSetUnsafe(key)
	if err != nil {
		return 0, err
	}
	return len(set), nil
}

// SIsMember mengecek apakah member ada pada set.
func (s *Store) SIsMember(key, member string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	set, err := s.getSetUnsafe(key)
	if err != nil {
		return false, err
	}
	_, ok := set[member]
	return ok, nil
}