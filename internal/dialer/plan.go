package dialer

import (
	"context"
	"errors"
	"net"
	"time"
)

type Plan struct {
	base Dialer
	hops []ChainHop
}

func NewPlan(base Dialer) (*Plan, error) {
	if base == nil {
		return nil, errors.New("nil base dialer")
	}
	return &Plan{base: base}, nil
}

func NewChainedPlan(hops []ChainHop, base Dialer) (*Plan, error) {
	if base == nil {
		return nil, errors.New("nil base dialer")
	}
	if len(hops) == 0 {
		return nil, errors.New("empty chain")
	}
	return &Plan{
		base: base,
		hops: hops,
	}, nil
}

func (p *Plan) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	// No chain => preserve existing behavior.
	if len(p.hops) == 0 {
		return p.base.DialContext(ctx, network, address)
	}

	// SOCKS5 hops are TCP only.
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, errors.New("chained plan supports tcp only")
	}

	// 1) Reach first hop using the base dialer (direct OR tor).
	c, err := p.base.DialContext(ctx, "tcp", p.hops[0].Addr)
	if err != nil {
		return nil, err
	}

	// If anything fails after this point, ensure the conn is closed.
	ok := false
	defer func() {
		if !ok {
			_ = c.Close()
		}
	}()

	// If ctx has a deadline, apply it during the chain setup.
	// (Cleared after chain is established so server tunnel deadlines apply normally.)
	if dl, has := ctx.Deadline(); has {
		_ = c.SetDeadline(dl)
	}

	// 2) For each subsequent hop, CONNECT to it over the existing conn.
	for i := 1; i < len(p.hops); i++ {
		if err := socks5ConnectOverConn(c, p.hops[i].Addr); err != nil {
			return nil, err
		}
	}

	// 3) Final CONNECT to the destination over the last hop.
	if err := socks5ConnectOverConn(c, address); err != nil {
		return nil, err
	}

	// Clear setup deadline.
	_ = c.SetDeadline(time.Time{})

	ok = true
	return c, nil
}
