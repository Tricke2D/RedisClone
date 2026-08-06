package persistence

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/company/redis-clone/internal/protocol"
	"github.com/company/redis-clone/pkg/resp"
)

// SyncPolicy menentukan seberapa sering AOFWriter memanggil fsync ke disk.
type SyncPolicy int

const (
	SyncAlways SyncPolicy = iota
	SyncEverySecond
)

// AOFWriter menulis setiap write command ke file append-only log.
type AOFWriter struct {
	mu     sync.Mutex
	file   *os.File
	policy SyncPolicy

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewAOFWriter membuka file AOF dengan SyncPolicy default (SyncAlways).
func NewAOFWriter(path string) (*AOFWriter, error) {
	return NewAOFWriterWithPolicy(path, SyncAlways)
}

// NewAOFWriterWithPolicy membuka file AOF dengan SyncPolicy tertentu.
func NewAOFWriterWithPolicy(path string, policy SyncPolicy) (*AOFWriter, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	w := &AOFWriter{file: file, policy: policy, stopCh: make(chan struct{})}

	if policy == SyncEverySecond {
		w.startBackgroundSync()
	}

	return w, nil
}

// startBackgroundSync menjalankan goroutine yang memanggil fsync setiap 1 detik.
func (w *AOFWriter) startBackgroundSync() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-w.stopCh:
				return
			case <-ticker.C:
				w.mu.Lock()
				_ = w.file.Sync()
				w.mu.Unlock()
			}
		}
	}()
}

// LogCommand menulis satu command ke file AOF.
func (w *AOFWriter) LogCommand(cmd []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	values := make([]*resp.Value, len(cmd))
	for i, c := range cmd {
		values[i] = resp.NewBulkString(c)
	}

	encoder := protocol.NewRespEncoder(w.file)
	if err := encoder.Encode(resp.NewArray(values...)); err != nil {
		return fmt.Errorf("failed to write AOF entry: %w", err)
	}

	if w.policy == SyncAlways {
		return w.file.Sync()
	}
	return nil
}

// Close menutup file AOF.
func (w *AOFWriter) Close() error {
	if w.policy == SyncEverySecond {
		close(w.stopCh)
		w.wg.Wait()
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.file.Sync()
	return w.file.Close()
}

// AOFReader membaca file AOF dan mereplay setiap command.
type AOFReader struct {
	path string
}

// NewAOFReader membuat reader untuk file AOF.
func NewAOFReader(path string) *AOFReader {
	return &AOFReader{path: path}
}

// Replayer adalah interface untuk mereplay command.
type Replayer interface {
	Execute(cmd []string) (*resp.Value, error)
}

// Recover membaca seluruh isi file AOF dan mereplay setiap command.
func (r *AOFReader) Recover(executor Replayer) (int, error) {
	file, err := os.Open(r.path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to open AOF file for recovery: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	parser := protocol.NewRespParser(reader)

	replayed := 0
	for {
		val, err := parser.Parse()
		if err != nil {
			break
		}

		cmd, err := protocol.ParseCommand(val)
		if err != nil {
			return replayed, fmt.Errorf("corrupt AOF entry at position %d: %w", replayed, err)
		}

		if _, err := executor.Execute(cmd); err != nil {
			return replayed, fmt.Errorf("failed to replay command %v: %w", cmd, err)
		}
		replayed++
	}

	return replayed, nil
}