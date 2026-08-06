package persistence

import (
	"fmt"
	"os"

	"github.com/company/redis-clone/internal/storage"
)

// RDBWriter menulis snapshot penuh dari store ke file dalam format binary (gob-encoded).
type RDBWriter struct {
	path string
}

// NewRDBWriter membuat writer untuk file RDB di path yang diberikan.
func NewRDBWriter(path string) *RDBWriter {
	return &RDBWriter{path: path}
}

// Save menulis snapshot store saat ini ke file RDB.
func (w *RDBWriter) Save(store *storage.Store) error {
	tmpPath := w.path + ".tmp"

	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp RDB file: %w", err)
	}

	if err := store.EncodeSnapshot(file); err != nil {
		file.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close temp RDB file: %w", err)
	}

	if err := os.Rename(tmpPath, w.path); err != nil {
		return fmt.Errorf("failed to finalize RDB file: %w", err)
	}

	return nil
}

// RDBReader membaca file RDB dan memuat isinya ke dalam store.
type RDBReader struct {
	path string
}

// NewRDBReader membuat reader untuk file RDB di path yang diberikan.
func NewRDBReader(path string) *RDBReader {
	return &RDBReader{path: path}
}

// Exists mengecek apakah file RDB ada.
func (r *RDBReader) Exists() bool {
	_, err := os.Stat(r.path)
	return err == nil
}

// Load membaca file RDB dan memuat isinya ke store.
func (r *RDBReader) Load(store *storage.Store) error {
	file, err := os.Open(r.path)
	if err != nil {
		return fmt.Errorf("failed to open RDB file: %w", err)
	}
	defer file.Close()

	if err := store.DecodeSnapshot(file); err != nil {
		return fmt.Errorf("failed to decode RDB snapshot: %w", err)
	}

	return nil
}