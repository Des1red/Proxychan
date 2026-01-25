package server

import (
	"context"
	"database/sql"
	"net"
	"proxychan/internal/system"
	"strings"
	"time"
)

func (s *Server) denylistPoller(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			v, err := system.GetDenylistVersion(db)
			if err != nil {
				s.cfg.Logger.Warnf("denylist version check failed: %v", err)
				continue
			}

			s.denyMu.RLock()
			cur := s.denyVersion
			s.denyMu.RUnlock()

			if v != cur {
				rt, err := system.LoadDenylist(db)
				if err != nil {
					s.cfg.Logger.Warnf("denylist reload failed: %v", err)
					continue
				}

				s.denyMu.Lock()
				s.denyIPNets = rt.IPNets
				s.denyDomainExact = rt.DomainExact
				s.denyDomainSuffix = rt.DomainSuffix
				s.denyVersion = v
				s.denyMu.Unlock()

				s.cfg.Logger.Infof("denylist reloaded (ip/cidr=%d, exact=%d, suffix=%d)",
					len(rt.IPNets), len(rt.DomainExact), len(rt.DomainSuffix))
			}
		}
	}
}

func normalizeDestDomain(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimSuffix(host, ".")
	return host
}

func (s *Server) destDenied(host string) (hitType, hitPattern string, denied bool) {
	// IP?
	if ip := net.ParseIP(host); ip != nil {
		s.denyMu.RLock()
		nets := s.denyIPNets
		s.denyMu.RUnlock()

		for _, n := range nets {
			if n.Contains(ip) {
				return "ip/cidr", n.String(), true
			}
		}
		return "", "", false
	}

	// Domain
	d := normalizeDestDomain(host)
	if d == "" {
		return "", "", false
	}

	s.denyMu.RLock()
	_, exact := s.denyDomainExact[d]
	suffixes := s.denyDomainSuffix
	s.denyMu.RUnlock()

	if exact {
		return "domain_exact", d, true
	}

	for _, suf := range suffixes {
		if strings.HasSuffix(d, suf) {
			return "domain_suffix", suf, true
		}
	}

	return "", "", false
}
