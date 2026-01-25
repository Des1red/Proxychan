package server

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"proxychan/internal/system"
	"time"
)

func (s *Server) whitelistPoller(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v, err := system.GetWhitelistVersion(db)
			if err != nil {
				s.cfg.Logger.Warnf("whitelist version check failed: %v", err)
				continue
			}

			s.mu.RLock()
			cur := s.whitelistVersion
			s.mu.RUnlock()

			if v != cur {
				wl, err := system.LoadWhitelist(db)
				if err != nil {
					s.cfg.Logger.Warnf("whitelist reload failed: %v", err)
					continue
				}

				s.mu.Lock()
				s.whitelist = wl
				s.whitelistVersion = v
				s.mu.Unlock()

				s.cfg.Logger.Infof("whitelist reloaded (%d entries)", len(wl))
			}
		}
	}
}

func (s *Server) ipAllowed(ip net.IP) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, n := range s.whitelist {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *Server) checkSource(client net.Conn) (net.IP, error) {
	host, _, err := net.SplitHostPort(client.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	s.cfg.Logger.Infof(
		"incoming remote host=%q full=%q",
		host,
		client.RemoteAddr().String(),
	)

	ip := net.ParseIP(host)
	if ip == nil || !s.ipAllowed(ip) {
		s.cfg.Logger.Warnf("connection from %s blocked by whitelist", host)
		return nil, errors.New("source not allowed")
	}

	return ip, nil
}
