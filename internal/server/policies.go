package server

import (
	"context"
	"database/sql"
	"proxychan/internal/system"
)

func (s *Server) initPolicies(ctx context.Context, db *sql.DB) error {
	// whitelist
	wl, err := system.LoadWhitelist(db)
	if err != nil {
		return err
	}

	v, err := system.GetWhitelistVersion(db)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.whitelist = wl
	s.whitelistVersion = v
	s.mu.Unlock()

	go s.whitelistPoller(ctx, db)

	// denylist
	rt, err := system.LoadDenylist(db)
	if err != nil {
		return err
	}

	dv, err := system.GetDenylistVersion(db)
	if err != nil {
		return err
	}

	s.denyMu.Lock()
	s.denyIPNets = rt.IPNets
	s.denyDomainExact = rt.DomainExact
	s.denyDomainSuffix = rt.DomainSuffix
	s.denyVersion = dv
	s.denyMu.Unlock()

	go s.denylistPoller(ctx, db)

	return nil
}
