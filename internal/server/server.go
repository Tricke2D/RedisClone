package server

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/company/redis-clone/internal/command"
	"github.com/company/redis-clone/internal/logger"
	"github.com/company/redis-clone/internal/persistence"
	"github.com/company/redis-clone/internal/replication"
	"github.com/company/redis-clone/internal/storage"
)

// Server represents the Redis-like server instance
type Server struct {
	addr             string
	aofPath          string
	aofSyncPolicy    persistence.SyncPolicy
	rdbPath          string
	snapshotInterval time.Duration
	initialReplicaOf string

	listener net.Listener
	store    *storage.Store
	log      logger.Logger
	aof      *persistence.AOFWriter
	replHub  *replication.Hub

	mu               sync.RWMutex
	running          bool
	wg               sync.WaitGroup
	shutdownOnce     sync.Once
	connectedClients int32

	masterAddr  string
	replicaConn net.Conn

	shutdownCh chan struct{}
}

// Config holds server configuration
type Config struct {
	Addr             string
	AOFPath          string
	AOFSyncPolicy    persistence.SyncPolicy
	RDBPath          string
	SnapshotInterval time.Duration
	ReplicaOf        string
}

// NewServer creates dan initializes new server instance
func NewServer(config Config) *Server {
	if config.Addr == "" {
		config.Addr = "localhost:6379"
	}
	if config.RDBPath != "" && config.SnapshotInterval <= 0 {
		config.SnapshotInterval = 5 * time.Minute
	}

	return &Server{
		addr:             config.Addr,
		aofPath:          config.AOFPath,
		aofSyncPolicy:    config.AOFSyncPolicy,
		rdbPath:          config.RDBPath,
		snapshotInterval: config.SnapshotInterval,
		initialReplicaOf: config.ReplicaOf,
		store:            storage.NewStore(),
		log:              logger.NewStdoutLogger(),
		replHub:          replication.NewHub(),
		shutdownCh:       make(chan struct{}),
	}
}

// Start begins listening untuk TCP connections (blocking)
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.mu.Unlock()

	// 1. Load persistence: RDB diprioritaskan, AOF sebagai fallback
	if err := s.loadPersistence(); err != nil {
		return fmt.Errorf("persistence load failed: %w", err)
	}

	// 2. Buka AOF writer
	if s.aofPath != "" {
		writer, err := persistence.NewAOFWriterWithPolicy(s.aofPath, s.aofSyncPolicy)
		if err != nil {
			return fmt.Errorf("failed to open AOF writer: %w", err)
		}
		s.aof = writer
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	s.mu.Lock()
	s.listener = listener
	s.running = true
	s.mu.Unlock()

	s.log.Info("Server started", map[string]interface{}{
		"address": s.addr, "aof_path": s.aofPath, "rdb_path": s.rdbPath,
	})

	// 3. Background managers
	s.startExpiryManager()
	s.startSnapshotManager()

	// 4. Jika dikonfigurasi sebagai replica sejak startup
	if s.initialReplicaOf != "" {
		if err := s.StartReplication(s.initialReplicaOf); err != nil {
			s.log.Error("Initial replication failed", map[string]interface{}{"error": err})
		}
	}

	s.acceptConnections()
	return nil
}

// loadPersistence memuat data awal: RDB diprioritaskan, AOF sebagai fallback
func (s *Server) loadPersistence() error {
	if s.rdbPath != "" {
		reader := persistence.NewRDBReader(s.rdbPath)
		if reader.Exists() {
			if err := reader.Load(s.store); err != nil {
				return fmt.Errorf("RDB load failed: %w", err)
			}
			s.log.Info("RDB snapshot loaded", map[string]interface{}{
				"path": s.rdbPath, "keys": s.store.Size(),
			})
			return nil
		}
	}

	if s.aofPath != "" {
		reader := persistence.NewAOFReader(s.aofPath)
		replayExecutor := command.NewExecutor(s.store)

		count, err := reader.Recover(replayExecutor)
		if err != nil {
			return err
		}
		s.log.Info("AOF recovery complete", map[string]interface{}{"commands_replayed": count})
	}

	return nil
}

// startSnapshotManager menjalankan background goroutine untuk RDB snapshot
func (s *Server) startSnapshotManager() {
	if s.rdbPath == "" {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.snapshotInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.shutdownCh:
				return
			case <-ticker.C:
				s.saveSnapshot()
			}
		}
	}()
}

// saveSnapshot menulis snapshot RDB saat ini ke disk
func (s *Server) saveSnapshot() {
	writer := persistence.NewRDBWriter(s.rdbPath)
	if err := writer.Save(s.store); err != nil {
		s.log.Error("RDB snapshot failed", map[string]interface{}{"error": err})
		return
	}
	s.log.Info("RDB snapshot saved", map[string]interface{}{"path": s.rdbPath, "keys": s.store.Size()})
}

// StartReplication membuat server ini menjadi REPLICA
func (s *Server) StartReplication(masterAddr string) error {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to master %s: %w", masterAddr, err)
	}

	s.mu.Lock()
	s.masterAddr = masterAddr
	s.replicaConn = conn
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer conn.Close()

		s.log.Info("Starting replication from master", map[string]interface{}{"master": masterAddr})

		replayExecutor := command.NewExecutor(s.store)
		if err := replication.RunSync(conn, s.store, replayExecutor); err != nil {
			s.log.Error("Replication connection ended", map[string]interface{}{"error": err})
		}

		s.mu.Lock()
		s.replicaConn = nil
		s.mu.Unlock()
	}()

	return nil
}

// StopReplication menghentikan mode replica
func (s *Server) StopReplication() {
	s.mu.Lock()
	masterAddr := s.masterAddr
	conn := s.replicaConn
	s.masterAddr = ""
	s.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
	if masterAddr != "" {
		s.log.Info("Replication stopped", map[string]interface{}{"previous_master": masterAddr})
	}
}

// acceptConnections terus menerima koneksi TCP masuk
func (s *Server) acceptConnections() {
	defer s.listener.Close()

	for {
		select {
		case <-s.shutdownCh:
			s.log.Info("Stop accepting new connections", nil)
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownCh:
				return
			default:
				s.log.Error("Failed to accept connection", map[string]interface{}{"error": err})
				continue
			}
		}

		atomic.AddInt32(&s.connectedClients, 1)

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer atomic.AddInt32(&s.connectedClients, -1)
			s.handleConnection(conn)
		}()
	}
}

// Shutdown menghentikan server secara graceful
func (s *Server) Shutdown() error {
	s.shutdownOnce.Do(func() {
		s.mu.Lock()
		if !s.running {
			s.mu.Unlock()
			return
		}
		s.running = false
		s.mu.Unlock()

		s.log.Info("Shutting down server", nil)
		close(s.shutdownCh)

		s.StopReplication()

		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()
		<-done

		if s.aof != nil {
			if err := s.aof.Close(); err != nil {
				s.log.Error("Failed to close AOF file", map[string]interface{}{"error": err})
			}
		}

		if s.rdbPath != "" {
			s.saveSnapshot()
		}

		s.log.Info("Server stopped successfully", nil)
	})
	return nil
}

// GetConnectedClients returns number of currently connected clients
func (s *Server) GetConnectedClients() int {
	return int(atomic.LoadInt32(&s.connectedClients))
}

// GetRunning returns whether server is currently running
func (s *Server) GetRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}