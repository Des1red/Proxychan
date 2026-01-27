package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"proxychan/internal/models"
	"proxychan/internal/system"
	"sort"
	"time"
)

func (s *Server) Warnf(format string, args ...any) {
	s.cfg.Logger.Warnf(format, args...)
}

func normalizeSourceIP(s string) string {
	// SourceIP might be "ip:port" or just "ip"
	host, _, err := net.SplitHostPort(s)
	if err == nil && host != "" {
		return host
	}
	return s
}
func (s *Server) SnapshotConnections() []models.ActiveConn {
	s.connMu.RLock()
	defer s.connMu.RUnlock()

	out := make([]models.ActiveConn, 0, len(s.conns))
	for _, c := range s.conns {
		out = append(out, *c)
	}
	return out
}

// Groups by source IP and sorts:
// - each group's conns: newest -> oldest
// - groups: newest activity -> oldest activity
func groupConnectionsByIP(conns []models.ActiveConn) []models.ConnGroup {
	byIP := make(map[string][]models.ActiveConn, 16)

	for _, c := range conns {
		ip := normalizeSourceIP(c.SourceIP)
		c.SourceIP = ip // normalize in output
		byIP[ip] = append(byIP[ip], c)
	}

	groups := make([]models.ConnGroup, 0, len(byIP))
	for ip, list := range byIP {
		sort.Slice(list, func(i, j int) bool {
			return list[i].StartedAt.After(list[j].StartedAt)
		})
		groups = append(groups, models.ConnGroup{
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
func (s *Server) GroupConnectionsByIP(
	conns []models.ActiveConn,
) []models.ConnGroup {
	return groupConnectionsByIP(conns)
}

func ListActiveConnectionsByIP() ([]models.ConnGroup, error) {
	client := &http.Client{Timeout: 3 * time.Second}

	req, _ := http.NewRequest(
		"GET",
		"http://127.0.0.1:6060/connections/by-ip",
		nil,
	)
	sec, err := system.InternalAdminSecret()
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-ProxyChan-Internal", sec)

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy admin endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin endpoint returned status %s", resp.Status)
	}

	var groups []models.ConnGroup
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
