package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"proxychan/internal/logging"
	"time"
)

var (
	ErrUnsupportedVersion = errors.New("unsupported SOCKS version")
	ErrNoAcceptableMethod = errors.New("no acceptable auth method")
	ErrUnsupportedCommand = errors.New("unsupported command")

	ErrAuthFailed = errors.New("authentication failed")
)

type HandshakeOptions struct {
	RequireAuth bool
	AuthFunc    func(username, password string) error
}

type Request struct {
	Cmd     byte   // 0x01 CONNECT
	Address string // host:port (domain or IP)
}

const (
	methodNoAuth   = 0x00
	methodUserPass = 0x02
	methodNoAccept = 0xFF
)

const (
	authStatusSuccess = 0x00
	authStatusFailure = 0x01
)

const (
	socksVersion5 = 0x05

	cmdConnect = 0x01

	atypIPv4   = 0x01
	atypDomain = 0x03
	atypIPv6   = 0x04
)

// HandleHandshake negotiates SOCKS5 auth.
func HandleHandshake(rw io.ReadWriter, opt HandshakeOptions) (string, error) {
	var hdr [2]byte
	if _, err := io.ReadFull(rw, hdr[:]); err != nil {
		logging.GetLogger().Errorf("Failed to read handshake header: %v", err)
		return "", err
	}
	if hdr[0] != socksVersion5 {
		logging.GetLogger().Warnf("Unsupported SOCKS version: %d", hdr[0])
		return "", ErrUnsupportedVersion
	}

	nMethods := int(hdr[1])
	if nMethods <= 0 {
		_, _ = rw.Write([]byte{socksVersion5, methodNoAccept})
		logging.GetLogger().Warn("No acceptable methods in handshake")
		return "", ErrNoAcceptableMethod
	}

	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(rw, methods); err != nil {
		logging.GetLogger().Errorf("Failed to read SOCKS methods: %v", err)
		return "", err
	}

	required := byte(methodNoAuth)
	if opt.RequireAuth {
		required = methodUserPass
		if opt.AuthFunc == nil {
			_, _ = rw.Write([]byte{socksVersion5, methodNoAccept})
			logging.GetLogger().Error("RequireAuth set but AuthFunc is nil")
			return "", fmt.Errorf("socks5: RequireAuth set but AuthFunc is nil")
		}
	}

	chosen := byte(methodNoAccept)
	for _, m := range methods {
		if m == required {
			chosen = required
			break
		}
	}

	if _, err := rw.Write([]byte{socksVersion5, chosen}); err != nil {
		logging.GetLogger().Errorf("Failed to write SOCKS method response: %v", err)
		return "", err
	}
	if chosen == methodNoAccept {
		logging.GetLogger().Warn("No acceptable authentication method selected")
		return "", ErrNoAcceptableMethod
	}

	var username string
	if chosen == methodUserPass {
		u, p, err := readUserPassAuth(rw)
		if err != nil {
			_ = writeUserPassStatus(rw, authStatusFailure)
			logging.GetLogger().Errorf("Failed to read user/password: %v", err)
			return "", err
		}

		if err := opt.AuthFunc(u, p); err != nil {
			_ = writeUserPassStatus(rw, authStatusFailure)
			logging.GetLogger().Errorf("Authentication failed for user: %v", err)
			return "", ErrAuthFailed
		}

		if err := writeUserPassStatus(rw, authStatusSuccess); err != nil {
			logging.GetLogger().Errorf("Failed to write user pass status: %v", err)
			return "", err
		}
		// Return the username upon successful authentication
		username = u
		if err := writeUserPassStatus(rw, authStatusSuccess); err != nil {
			logging.GetLogger().Errorf("Failed to write user pass status: %v", err)
			return "", err
		}
	}

	return username, nil // Return the username after successful authentication
}

func ReadRequest(r io.Reader) (Request, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		logging.GetLogger().Errorf("Failed to read request header: %v", err)
		return Request{}, err
	}
	if hdr[0] != socksVersion5 {
		logging.GetLogger().Warnf("Unsupported SOCKS version: %d", hdr[0])
		return Request{}, ErrUnsupportedVersion
	}

	cmd := hdr[1]
	atyp := hdr[3]

	if cmd != cmdConnect {
		logging.GetLogger().Warnf("Unsupported command: 0x%02x", cmd)
		return Request{Cmd: cmd}, ErrUnsupportedCommand
	}

	host, err := readAddr(r, atyp)
	if err != nil {
		logging.GetLogger().Errorf("Failed to read address: %v", err)
		return Request{}, err
	}

	var portBuf [2]byte
	if _, err := io.ReadFull(r, portBuf[:]); err != nil {
		logging.GetLogger().Errorf("Failed to read port: %v", err)
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
	case atypIPv4:
		var b [4]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			logging.GetLogger().Errorf("Failed to read IPv4 address: %v", err)
			return "", err
		}
		return net.IP(b[:]).String(), nil
	case atypIPv6:
		var b [16]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			logging.GetLogger().Errorf("Failed to read IPv6 address: %v", err)
			return "", err
		}
		return net.IP(b[:]).String(), nil
	case atypDomain:
		var l [1]byte
		if _, err := io.ReadFull(r, l[:]); err != nil {
			logging.GetLogger().Errorf("Failed to read domain length: %v", err)
			return "", err
		}
		if l[0] == 0 {
			logging.GetLogger().Error("Received empty domain")
			return "", errors.New("empty domain")
		}
		d := make([]byte, int(l[0]))
		if _, err := io.ReadFull(r, d); err != nil {
			logging.GetLogger().Errorf("Failed to read domain: %v", err)
			return "", err
		}
		return string(d), nil
	default:
		logging.GetLogger().Errorf("Unknown ATYP 0x%02x", atyp)
		return "", fmt.Errorf("unknown ATYP 0x%02x", atyp)
	}
}

func WriteReply(w io.Writer, rep byte) error {
	_, err := w.Write([]byte{
		socksVersion5, rep, 0x00, atypIPv4,
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	})
	if err != nil {
		logging.GetLogger().Errorf("Failed to write SOCKS reply: %v", err)
	}
	return err
}

func ConnectOverConn(c net.Conn, address string) error {
	if _, err := c.Write([]byte{socksVersion5, 0x01, methodNoAuth}); err != nil {
		logging.GetLogger().Errorf("Failed to write SOCKS greeting: %v", err)
		return fmt.Errorf("socks5 greeting write: %w", err)
	}

	var resp [2]byte
	if _, err := io.ReadFull(c, resp[:]); err != nil {
		logging.GetLogger().Errorf("Failed to read SOCKS greeting response: %v", err)
		return fmt.Errorf("socks5 greeting read: %w", err)
	}
	if resp[0] != socksVersion5 || resp[1] != methodNoAuth {
		logging.GetLogger().Errorf("SOCKS5 auth not accepted: %v", resp)
		return fmt.Errorf("socks5 auth not accepted")
	}

	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		logging.GetLogger().Errorf("Failed to split host:port: %v", err)
		return err
	}

	port, err := parsePort(portStr)
	if err != nil {
		logging.GetLogger().Errorf("Failed to parse port: %v", err)
		return err
	}

	req := []byte{socksVersion5, cmdConnect, 0x00}

	ip := net.ParseIP(host)
	switch {
	case ip == nil:
		req = append(req, atypDomain, byte(len(host)))
		req = append(req, []byte(host)...)
	case ip.To4() != nil:
		req = append(req, atypIPv4)
		req = append(req, ip.To4()...)
	default:
		req = append(req, atypIPv6)
		req = append(req, ip.To16()...)
	}

	var p [2]byte
	binary.BigEndian.PutUint16(p[:], uint16(port))
	req = append(req, p[:]...)

	_ = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.Write(req); err != nil {
		logging.GetLogger().Errorf("Failed to write SOCKS connect request: %v", err)
		return fmt.Errorf("socks5 connect write: %w", err)
	}
	_ = c.SetWriteDeadline(time.Time{})

	var hdr [4]byte
	if _, err := io.ReadFull(c, hdr[:]); err != nil {
		logging.GetLogger().Errorf("Failed to read SOCKS connect response: %v", err)
		return fmt.Errorf("socks5 reply read: %w", err)
	}
	if hdr[1] != authStatusSuccess {
		logging.GetLogger().Errorf("SOCKS5 connect failed, REP=0x%02x", hdr[1])
		return fmt.Errorf("socks5 connect failed, REP=0x%02x", hdr[1])
	}

	return drainSocksBind(c, hdr[3])
}
