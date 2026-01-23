package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported SOCKS version")
	ErrNoAcceptableMethod = errors.New("no acceptable auth method")
	ErrUnsupportedCommand = errors.New("unsupported command")
)

type Request struct {
	Cmd     byte   // 0x01 CONNECT
	Address string // host:port (domain or IP)
}

// HandleHandshake supports: SOCKS5 + NO AUTH only.
func HandleHandshake(rw io.ReadWriter) error {
	// Client greeting: VER, NMETHODS, METHODS...
	var hdr [2]byte
	if _, err := io.ReadFull(rw, hdr[:]); err != nil {
		return err
	}
	if hdr[0] != 0x05 {
		return ErrUnsupportedVersion
	}
	nMethods := int(hdr[1])
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(rw, methods); err != nil {
		return err
	}

	// We accept only 0x00 (no auth)
	chosen := byte(0xFF)
	for _, m := range methods {
		if m == 0x00 {
			chosen = 0x00
			break
		}
	}

	// Server method selection: VER, METHOD
	if _, err := rw.Write([]byte{0x05, chosen}); err != nil {
		return err
	}
	if chosen == 0xFF {
		return ErrNoAcceptableMethod
	}
	return nil
}

// ReadRequest parses a SOCKS5 CONNECT request (CMD=0x01).
func ReadRequest(r io.Reader) (Request, error) {
	// VER CMD RSV ATYP
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Request{}, err
	}
	if hdr[0] != 0x05 {
		return Request{}, ErrUnsupportedVersion
	}
	cmd := hdr[1]
	atyp := hdr[3]

	if cmd != 0x01 {
		return Request{Cmd: cmd}, ErrUnsupportedCommand
	}

	host, err := readAddr(r, atyp)
	if err != nil {
		return Request{}, err
	}

	var portBuf [2]byte
	if _, err := io.ReadFull(r, portBuf[:]); err != nil {
		return Request{}, err
	}
	port := binary.BigEndian.Uint16(portBuf[:])

	return Request{
		Cmd:     cmd,
		Address: net.JoinHostPort(host, fmt.Sprintf("%d", port)),
	}, nil
}

func readAddr(r io.Reader, atyp byte) (string, error) {
	switch atyp {
	case 0x01: // IPv4
		var b [4]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", err
		}
		return net.IP(b[:]).String(), nil
	case 0x04: // IPv6
		var b [16]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return "", err
		}
		return net.IP(b[:]).String(), nil
	case 0x03: // DOMAIN
		var l [1]byte
		if _, err := io.ReadFull(r, l[:]); err != nil {
			return "", err
		}
		if l[0] == 0 {
			return "", errors.New("empty domain")
		}
		d := make([]byte, int(l[0]))
		if _, err := io.ReadFull(r, d); err != nil {
			return "", err
		}
		return string(d), nil
	default:
		return "", fmt.Errorf("unknown ATYP 0x%02x", atyp)
	}
}

// WriteReply writes a SOCKS5 reply.
// REP: 0x00 success, else error code.
// BND.ADDR/BND.PORT can be zeros; clients generally accept that.
func WriteReply(w io.Writer, rep byte) error {
	// VER REP RSV ATYP BND.ADDR BND.PORT
	// We'll send ATYP=IPv4 and 0.0.0.0:0
	_, err := w.Write([]byte{
		0x05, rep, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	})
	return err
}

// socks5ConnectOverConn performs a SOCKS5 handshake + CONNECT
// over an already-established TCP connection.
func ConnectOverConn(c net.Conn, address string) error {
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
