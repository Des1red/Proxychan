package server

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"strings"
	"sync"
	"time"

	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/socks5"
	"proxychan/internal/system"

	"github.com/sirupsen/logrus"
)

type Config struct {
	ListenAddr  string
	Dialer      dialer.Dialer
	IdleTimeout time.Duration
	Logger      *logrus.Logger

	RequireAuth bool
	AuthFunc    func(username, password string) error
}

type Server struct {
	cfg Config
}

func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = logging.GetLogger()
	}
	return &Server{cfg: cfg}
}

func RequiresAuth(listenAddr string) bool {
	host, _, err := net.SplitHostPort(listenAddr)
	if err != nil {
		return true // fail closed
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return true
	}

	return !ip.IsLoopback()
}

func (s *Server) Run(ctx context.Context, db *sql.DB) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}

	s.cfg.Logger.Infof("listening on %s", s.cfg.ListenAddr)
	// Log public address if required
	if s.cfg.RequireAuth {
		if ip, err := detectPublicIP(); err == nil {
			s.cfg.Logger.Infof("public address: %s", net.JoinHostPort(ip, strings.Split(s.cfg.ListenAddr, ":")[1]))
		}
	}

	// Stop listener on context cancel.
	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		c, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.cfg.Logger.Errorf("accept error: %v", err)
			continue
		}
		go s.handleConn(ctx, c, db)
	}
}

func (s *Server) handleConn(ctx context.Context, client net.Conn, db *sql.DB) {
	defer client.Close()
	// Set deadline for handshake
	_ = client.SetDeadline(time.Now().Add(15 * time.Second)) // handshake deadline
	// Perform authentication handshake
	username, err := socks5.HandleHandshake(client, socks5.HandshakeOptions{
		RequireAuth: s.cfg.RequireAuth,
		AuthFunc:    s.cfg.AuthFunc,
	})
	if err != nil {
		s.cfg.Logger.Warnf("handshake error from %s: %v", client.RemoteAddr(), err)
		return
	}

	// If authentication is required, check the user status
	if s.cfg.RequireAuth {
		// Check if the user is active
		active, err := system.IsActive(db, username)
		if err != nil {
			s.cfg.Logger.Warnf("error checking if user %s is active: %v", username, err)
			return
		}
		// If the user is inactive, reject the connection
		if !active {
			s.cfg.Logger.Warnf("user %s is inactive, rejecting connection", username)
			_ = socks5.WriteReply(client, 0x05) // Connection refused
			return
		}
	}

	req, err := socks5.ReadRequest(client)
	if err != nil {
		// Unsupported command or parse failure.
		_ = socks5.WriteReply(client, 0x07) // Command not supported
		s.cfg.Logger.Warnf("request error from %s: %v", client.RemoteAddr(), err)
		return
	}

	// Dial outbound (via selected dialer).
	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := s.cfg.Dialer.DialContext(dialCtx, "tcp", req.Address)
	if err != nil {
		_ = socks5.WriteReply(client, 0x05) // Connection refused (generic-ish)
		s.cfg.Logger.Warnf("dial fail %s -> %s: %v", client.RemoteAddr(), req.Address, err)
		return
	}
	defer out.Close()

	// Handshake done: tunnel established.
	_ = socks5.WriteReply(client, 0x00)

	// Clear handshake deadline; apply idle timeout (optional).
	_ = client.SetDeadline(time.Time{})
	_ = out.SetDeadline(time.Time{})

	s.tunnel(client, out)
}

func (s *Server) tunnel(a, b net.Conn) {
	// Optional idle timeout: refreshed by traffic in either direction.
	var (
		idle = s.cfg.IdleTimeout
		mu   sync.Mutex
	)

	refreshDeadline := func() {
		if idle <= 0 {
			return
		}
		dl := time.Now().Add(idle)
		_ = a.SetDeadline(dl)
		_ = b.SetDeadline(dl)
	}

	refreshDeadline()

	copyWithRefresh := func(dst, src net.Conn) {
		buf := make([]byte, 32*1024)
		for {
			n, rerr := src.Read(buf)
			if n > 0 {
				mu.Lock()
				refreshDeadline()
				mu.Unlock()

				_, werr := dst.Write(buf[:n])
				if werr != nil {
					return
				}
			}
			if rerr != nil {
				halfCloseWrite(dst)
				return
			}
		}
	}

	done := make(chan struct{}, 2)
	go func() { copyWithRefresh(b, a); done <- struct{}{} }()
	go func() { copyWithRefresh(a, b); done <- struct{}{} }()

	<-done
	<-done
}

func halfCloseWrite(c net.Conn) {
	if tc, ok := c.(*net.TCPConn); ok {
		_ = tc.CloseWrite()
		return
	}
	_ = c.Close()
}
