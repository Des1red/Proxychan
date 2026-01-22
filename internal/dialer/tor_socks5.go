package dialer

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type torSocks5Dialer struct {
	torAddr string
	timeout time.Duration
}

func NewTorSOCKS5(torSocksAddr string, connectTimeout time.Duration) Dialer {
	return &torSocks5Dialer{torAddr: torSocksAddr, timeout: connectTimeout}
}

func (t *torSocks5Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("tor dialer supports tcp only, got %q", network)
	}

	nd := net.Dialer{Timeout: t.timeout}
	c, err := nd.DialContext(ctx, "tcp", t.torAddr)
	if err != nil {
		return nil, fmt.Errorf("dial tor socks5 %s: %w", t.torAddr, err)
	}

	// If anything fails, close.
	if err := t.socks5Handshake(c); err != nil {
		_ = c.Close()
		return nil, err
	}
	if err := t.socks5Connect(c, address); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

func (t *torSocks5Dialer) socks5Handshake(c net.Conn) error {
	// Client greeting: VER=5, NMETHODS=1, METHODS={0x00 no-auth}
	if _, err := c.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return fmt.Errorf("tor socks5 greeting write: %w", err)
	}
	// Server choice: VER, METHOD
	var resp [2]byte
	if _, err := io.ReadFull(c, resp[:]); err != nil {
		return fmt.Errorf("tor socks5 greeting read: %w", err)
	}
	if resp[0] != 0x05 {
		return fmt.Errorf("tor socks5 bad version: %d", resp[0])
	}
	if resp[1] != 0x00 {
		return fmt.Errorf("tor socks5 auth method not accepted: 0x%02x", resp[1])
	}
	return nil
}

func (t *torSocks5Dialer) socks5Connect(c net.Conn, address string) error {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("split host:port %q: %w", address, err)
	}
	port, err := parsePort(portStr)
	if err != nil {
		return err
	}

	// Build CONNECT request:
	// VER=5 CMD=1 RSV=0 ATYP + DST.ADDR + DST.PORT
	req := make([]byte, 0, 6+len(host))
	req = append(req, 0x05, 0x01, 0x00) // ver, connect, rsv

	ip := net.ParseIP(host)
	switch {
	case ip == nil:
		// domain name
		if len(host) > 255 {
			return errors.New("domain too long for socks5")
		}
		req = append(req, 0x03, byte(len(host)))
		req = append(req, []byte(host)...)
	case ip.To4() != nil:
		req = append(req, 0x01)
		req = append(req, ip.To4()...)
	default:
		req = append(req, 0x04)
		req = append(req, ip.To16()...)
	}

	var portBuf [2]byte
	binary.BigEndian.PutUint16(portBuf[:], uint16(port))
	req = append(req, portBuf[:]...)

	_ = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.Write(req); err != nil {
		return fmt.Errorf("tor socks5 connect write: %w", err)
	}
	_ = c.SetWriteDeadline(time.Time{})

	// Read reply: VER REP RSV ATYP BND.ADDR BND.PORT
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return fmt.Errorf("tor socks5 connect read hdr: %w", err)
	}
	if hdr[0] != 0x05 {
		return fmt.Errorf("tor socks5 reply bad version: %d", hdr[0])
	}
	if hdr[1] != 0x00 {
		return fmt.Errorf("tor socks5 connect failed, REP=0x%02x", hdr[1])
	}

	// Drain BND.ADDR based on ATYP (not used)
	if err := drainSocksBind(c, hdr[3]); err != nil {
		return fmt.Errorf("tor socks5 drain bind: %w", err)
	}
	return nil
}

func drainSocksBind(r io.Reader, atyp byte) error {
	switch atyp {
	case 0x01: // IPv4
		if _, err := io.CopyN(io.Discard, r, 4+2); err != nil {
			return err
		}
	case 0x04: // IPv6
		if _, err := io.CopyN(io.Discard, r, 16+2); err != nil {
			return err
		}
	case 0x03: // domain
		var l [1]byte
		if _, err := io.ReadFull(r, l[:]); err != nil {
			return err
		}
		if _, err := io.CopyN(io.Discard, r, int64(l[0])+2); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown ATYP: 0x%02x", atyp)
	}
	return nil
}

func parsePort(s string) (uint16, error) {
	p, err := net.LookupPort("tcp", s)
	if err == nil {
		if p < 0 || p > 65535 {
			return 0, fmt.Errorf("invalid port: %s", s)
		}
		return uint16(p), nil
	}
	// fallback numeric parse
	var n int
	_, scanErr := fmt.Sscanf(s, "%d", &n)
	if scanErr != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("invalid port: %s", s)
	}
	return uint16(n), nil
}
