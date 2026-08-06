package storage

import (
	"encoding/gob"
	"io"
)

// init mendaftarkan seluruh tipe konkret yang mungkin disimpan di StoredValue.Value
func init() {
	gob.Register("")                    // string value
	gob.Register([]string{})            // list value
	gob.Register(map[string]string{})   // hash value
	gob.Register(map[string]struct{}{}) // set value
}

// EncodeSnapshot menulis seluruh isi store ke writer dalam format gob.
func (s *Store) EncodeSnapshot(w io.Writer) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	encoder := gob.NewEncoder(w)
	return encoder.Encode(s.data)
}

// DecodeSnapshot memuat data dari reader (format gob) ke dalam store.
func (s *Store) DecodeSnapshot(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	var data map[string]*StoredValue

	if err := decoder.Decode(&data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = data
	return nil
}