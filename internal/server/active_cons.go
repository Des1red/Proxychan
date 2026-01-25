package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ActiveConn struct {
	ID          uint64
	Username    string
	SourceIP    string
	Destination string
	StartedAt   time.Time
}

func (s *Server) SnapshotConnections() []ActiveConn {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	out := make([]ActiveConn, 0, len(s.conns))
	for _, c := range s.conns {
		out = append(out, *c)
	}
	return out
}

func (s *Server) runAdminEndpoint(ctx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		conns := s.SnapshotConnections()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(conns)
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.cfg.Logger.Warnf("admin endpoint error: %v", err)
	}
}

func ListActiveConnections() ([]ActiveConn, error) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get("http://127.0.0.1:6060/connections")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy admin endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin endpoint returned status %s", resp.Status)
	}

	var conns []ActiveConn
	if err := json.NewDecoder(resp.Body).Decode(&conns); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return conns, nil
}
