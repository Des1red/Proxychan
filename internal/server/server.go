package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"proxychan/internal/dialer"
	"proxychan/internal/socks5"
)

type Config struct {
	ListenAddr  string
	Dialer      dialer.Dialer
	IdleTimeout time.Duration
	Logger      *log.Logger
}

type Server struct {
	cfg Config
}

func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = log.New(io.Discard, "", 0)
	}
	return &Server{cfg: cfg}
}

func (s *Server) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}
	s.cfg.Logger.Printf("listening on %s", s.cfg.ListenAddr)

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
			s.cfg.Logger.Printf("accept error: %v", err)
			continue
		}
		go s.handleConn(ctx, c)
	}
}

func (s *Server) handleConn(ctx context.Context, client net.Conn) {
	defer client.Close()

	_ = client.SetDeadline(time.Now().Add(15 * time.Second)) // handshake deadline
	if err := socks5.HandleHandshake(client); err != nil {
		s.cfg.Logger.Printf("handshake error from %s: %v", client.RemoteAddr(), err)
		return
	}

	req, err := socks5.ReadRequest(client)
	if err != nil {
		// Unsupported command or parse failure.
		_ = socks5.WriteReply(client, 0x07) // Command not supported
		s.cfg.Logger.Printf("request error from %s: %v", client.RemoteAddr(), err)
		return
	}

	// Dial outbound (via selected dialer).
	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := s.cfg.Dialer.DialContext(dialCtx, "tcp", req.Address)
	if err != nil {
		_ = socks5.WriteReply(client, 0x05) // Connection refused (generic-ish)
		s.cfg.Logger.Printf("dial fail %s -> %s: %v", client.RemoteAddr(), req.Address, err)
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
