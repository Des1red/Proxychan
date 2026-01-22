package dialer

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// socks5ConnectOverConn performs a SOCKS5 handshake + CONNECT
// over an already-established TCP connection.
func socks5ConnectOverConn(c net.Conn, address string) error {
	// --- greeting ---
	if _, err := c.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		return fmt.Errorf("socks5 greeting write: %w", err)
	}

	var resp [2]byte
	if _, err := io.ReadFull(c, resp[:]); err != nil {
		return fmt.Errorf("socks5 greeting read: %w", err)
	}
	if resp[0] != 0x05 || resp[1] != 0x00 {
		return fmt.Errorf("socks5 auth not accepted")
	}

	// --- build CONNECT ---
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	port, err := parsePort(portStr)
	if err != nil {
		return err
	}

	req := []byte{0x05, 0x01, 0x00} // VER, CONNECT, RSV

	ip := net.ParseIP(host)
	switch {
	case ip == nil:
		if len(host) > 255 {
			return fmt.Errorf("domain too long")
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

	var p [2]byte
	binary.BigEndian.PutUint16(p[:], uint16(port))
	req = append(req, p[:]...)

	_ = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.Write(req); err != nil {
		return fmt.Errorf("socks5 connect write: %w", err)
	}
	_ = c.SetWriteDeadline(time.Time{})

	// --- reply ---
	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		return fmt.Errorf("socks5 reply read: %w", err)
	}
	if hdr[1] != 0x00 {
		return fmt.Errorf("socks5 connect failed, REP=0x%02x", hdr[1])
	}

	return drainSocksBind(c, hdr[3])
}
