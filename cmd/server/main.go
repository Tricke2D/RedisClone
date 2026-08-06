package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/redis-clone/internal/persistence"
	"github.com/company/redis-clone/internal/server"
)

func main() {
	addr := flag.String("addr", "localhost:6379", "Server address (host:port)")
	aofPath := flag.String("aof", "appendonly.aof", "Path file AOF (kosongkan untuk menonaktifkan)")
	aofSync := flag.String("aof-sync", "always", "Strategi fsync AOF: 'always' atau 'everysec'")
	rdbPath := flag.String("rdb", "dump.rdb", "Path file RDB snapshot (kosongkan untuk menonaktifkan)")
	snapshotInterval := flag.Duration("snapshot-interval", 5*time.Minute, "Interval snapshot RDB otomatis")
	replicaOf := flag.String("replicaof", "", "Alamat master 'host:port' untuk langsung jadi replica saat startup")
	pprofAddr := flag.String("pprof-addr", "", "Alamat pprof profiling endpoint (mis. 'localhost:6060')")
	flag.Parse()

	if *addr == "" {
		fmt.Fprintln(os.Stderr, "Error: address cannot be empty")
		os.Exit(1)
	}

	var syncPolicy persistence.SyncPolicy
	switch *aofSync {
	case "always":
		syncPolicy = persistence.SyncAlways
	case "everysec":
		syncPolicy = persistence.SyncEverySecond
	default:
		fmt.Fprintf(os.Stderr, "Error: -aof-sync harus 'always' atau 'everysec', dapat '%s'\n", *aofSync)
		os.Exit(1)
	}

	if *pprofAddr != "" {
		go func() {
			log.Printf("pprof profiling endpoint listening on %s\n", *pprofAddr)
			log.Println(http.ListenAndServe(*pprofAddr, nil))
		}()
	}

	cfg := server.Config{
		Addr:             *addr,
		AOFPath:          *aofPath,
		AOFSyncPolicy:    syncPolicy,
		RDBPath:          *rdbPath,
		SnapshotInterval: *snapshotInterval,
		ReplicaOf:        *replicaOf,
	}
	srv := server.NewServer(cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal: %v\n", sig)
		if err := srv.Shutdown(); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
			os.Exit(1)
		}
	}
}