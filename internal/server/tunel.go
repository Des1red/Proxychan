package server

import (
	"context"
	"net"
	"proxychan/internal/socks5"
	"sync"
	"time"
)

func (s *Server) handleTunnel(
	ctx context.Context,
	client net.Conn,
	username string,
	srcIP net.IP,
	req *socks5.Request,
) {
	id := s.nextConnID.Add(1)

	ac := &ActiveConn{
		ID:          id,
		Username:    username,
		SourceIP:    srcIP.String(),
		Destination: req.Address,
		StartedAt:   time.Now(),
	}

	s.connMu.Lock()
	s.conns[id] = ac
	s.connMu.Unlock()

	defer func() {
		s.connMu.Lock()
		delete(s.conns, id)
		s.connMu.Unlock()
	}()

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := s.cfg.Dialer.DialContext(dialCtx, "tcp", req.Address)
	if err != nil {
		_ = socks5.WriteReply(client, 0x05)
		s.cfg.Logger.Warnf(
			"dial fail %s -> %s: %v",
			client.RemoteAddr(),
			req.Address,
			err,
		)
		return
	}
	defer out.Close()

	_ = socks5.WriteReply(client, 0x00)

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
