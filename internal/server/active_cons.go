package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"time"
)

type ActiveConn struct {
	ID          uint64
	Username    string
	SourceIP    string
	Destination string
	StartedAt   time.Time
}

func normalizeSourceIP(s string) string {
	// SourceIP might be "ip:port" or just "ip"
	host, _, err := net.SplitHostPort(s)
	if err == nil && host != "" {
		return host
	}
	return s
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

type ConnGroup struct {
	SourceIP string       `json:"source_ip"`
	Count    int          `json:"count"`
	Conns    []ActiveConn `json:"conns"` // newest -> oldest
}

// Groups by source IP and sorts:
// - each group's conns: newest -> oldest
// - groups: newest activity -> oldest activity
func groupConnectionsByIP(conns []ActiveConn) []ConnGroup {
	byIP := make(map[string][]ActiveConn, 16)

	for _, c := range conns {
		ip := normalizeSourceIP(c.SourceIP)
		c.SourceIP = ip // normalize in output
		byIP[ip] = append(byIP[ip], c)
	}

	groups := make([]ConnGroup, 0, len(byIP))
	for ip, list := range byIP {
		sort.Slice(list, func(i, j int) bool {
			return list[i].StartedAt.After(list[j].StartedAt)
		})
		groups = append(groups, ConnGroup{
			SourceIP: ip,
			Count:    len(list),
			Conns:    list,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		// group activity = newest conn in that group
		return groups[i].Conns[0].StartedAt.After(groups[j].Conns[0].StartedAt)
	})

	return groups
}

func (s *Server) runAdminEndpoint(ctx context.Context) {
	mux := http.NewServeMux()

	// HTML view (browser)
	mux.HandleFunc("/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		conns := s.SnapshotConnections()
		groups := groupConnectionsByIP(conns)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		fmt.Fprintln(w, "<!DOCTYPE html>")
		fmt.Fprintln(w, "<html><head><title>ProxyChan Connections</title>")
		fmt.Fprintln(w, `<style>
			body { font-family: monospace; padding: 20px; }
			details { margin-bottom: 12px; }
			summary { cursor: pointer; font-weight: bold; }
			.conn { padding-left: 20px; }
		</style>`)
		fmt.Fprintln(w, "</head><body>")
		fmt.Fprintln(w, "<h2>Active Connections</h2>")

		now := time.Now()

		for _, g := range groups {
			fmt.Fprintf(
				w,
				"<details><summary>%s (%d connections)</summary>",
				g.SourceIP,
				g.Count,
			)

			for _, c := range g.Conns {
				age := now.Sub(c.StartedAt).Truncate(time.Second)
				user := c.Username
				if user == "" {
					user = "-"
				}

				fmt.Fprintf(
					w,
					`<div class="conn">ID=%d USER=%s DST=%s AGE=%s</div>`,
					c.ID,
					user,
					c.Destination,
					age,
				)
			}

			fmt.Fprintln(w, "</details>")
		}

		fmt.Fprintln(w, "</body></html>")
	})

	// JSON (CLI / API)
	mux.HandleFunc("/connections/by-ip", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		conns := s.SnapshotConnections()
		groups := groupConnectionsByIP(conns)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(groups)
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

func ListActiveConnectionsByIP() ([]ConnGroup, error) {
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get("http://127.0.0.1:6060/connections/by-ip")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy admin endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin endpoint returned status %s", resp.Status)
	}

	var groups []ConnGroup
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return groups, nil
}

func GetActiveConnectionCount() (int, error) {
	groups, err := ListActiveConnectionsByIP()
	if err != nil {
		return 0, err
	}

	total := 0
	for _, g := range groups {
		total += g.Count
	}

	return total, nil
}
