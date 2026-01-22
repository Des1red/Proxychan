package dialer

import (
	"context"
	"net"
	"time"
)

type directDialer struct {
	timeout time.Duration
}

func NewDirect(connectTimeout time.Duration) Dialer {
	return &directDialer{timeout: connectTimeout}
}

func (d *directDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	nd := net.Dialer{Timeout: d.timeout}
	return nd.DialContext(ctx, network, address)
}
