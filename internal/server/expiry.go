package server

import "time"

// startExpiryManager menjalankan background goroutine untuk active expiry.
// Dipanggil oleh Server.Start() setelah server siap menerima koneksi.
func (s *Server) startExpiryManager() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.shutdownCh:
				return
			case <-ticker.C:
				removed := s.store.RemoveExpired()
				if removed > 0 {
					s.log.Info("Active expiry cycle", map[string]interface{}{
						"keys_removed": removed,
					})
				}
			}
		}
	}()
}