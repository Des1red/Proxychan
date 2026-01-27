package server

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"proxychan/internal/models"
	"proxychan/internal/system"
	"strings"
	"time"
)

func (s *Server) handleHTTPConn(ctx context.Context, client net.Conn, db *sql.DB) {

	defer client.Close()

	br := bufio.NewReader(client)
	// 1. source whitelist
	srcIP, err := s.checkSource(client)
	if err != nil {
		return
	}
	srcIPStr := srcIP.String()

	// 2. parse CONNECT line + headers
	target, hdr, err := readHTTPConnect(br)
	if err != nil {
		writeHTTPError(client, 405, "Method Not Allowed")
		return
	}

	// 3. auth (respects --no-auth)
	username, err := s.httpAuthFromHeaders(hdr, client, db)
	if err != nil {
		return
	}

	// 4. dest denylist
	host, _, _ := net.SplitHostPort(target)
	if typ, pat, denied := s.destDenied(host); denied {
		writeHTTPError(client, 403, "Forbidden")
		s.cfg.Logger.Warnf(
			"http egress denied user=%q src=%s dst=%s ruleType=%s rule=%s",
			username, srcIP, target, typ, pat,
		)
		return
	}

	// 5. track connection
	id := s.registerConn(username, srcIPStr, target)
	defer s.unregisterConn(id)

	// 6. dial outbound
	out, err := s.cfg.Dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		writeHTTPError(client, 502, "Bad Gateway")
		return
	}
	defer out.Close()

	// 7. acknowledge tunnel
	_, _ = client.Write([]byte(
		"HTTP/1.1 200 Connection Established\r\n" +
			"Proxy-Agent: ProxyChan\r\n\r\n",
	))

	// 8. tunnel (important: use raw conn, not reader)
	s.tunnel(client, out)
}

func readHTTPConnect(br *bufio.Reader) (target string, hdr textproto.MIMEHeader, err error) {
	line, err := br.ReadString('\n')
	if err != nil {
		return "", nil, err
	}

	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) < 3 || parts[0] != "CONNECT" {
		return "", nil, errors.New("not CONNECT")
	}

	tp := textproto.NewReader(br)
	hdr, err = tp.ReadMIMEHeader()
	if err != nil {
		return "", nil, err
	}

	return parts[1], hdr, nil
}

func (s *Server) httpAuthFromHeaders(hdr textproto.MIMEHeader, conn net.Conn, db *sql.DB) (string, error) {
	if !s.cfg.RequireAuth {
		return "", nil
	}

	pa := hdr.Get("Proxy-Authorization")
	u, p, ok := parseBasicProxyAuth(pa)
	if !ok {
		writeHTTPError(conn, 407, "Proxy Authentication Required")
		_, _ = conn.Write([]byte("Proxy-Authenticate: Basic realm=\"ProxyChan\"\r\n\r\n"))
		return "", errors.New("missing proxy auth")
	}

	if err := s.cfg.AuthFunc(u, p); err != nil {
		_, _ = fmt.Fprintf(
			conn,
			"HTTP/1.1 407 Proxy Authentication Required\r\n"+
				"Proxy-Authenticate: Basic realm=\"ProxyChan\"\r\n"+
				"Content-Length: 0\r\n\r\n",
		)
		return "", errors.New("bad proxy auth")
	}

	// optional: active check (same as socks)
	active, err := system.IsActive(db, u)
	if err != nil {
		return "", err
	}
	if !active {
		writeHTTPError(conn, 403, "Forbidden")
		return "", errors.New("user inactive")
	}

	return u, nil
}

func parseBasicProxyAuth(v string) (user, pass string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(v, prefix) {
		return "", "", false
	}
	b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(v[len(prefix):]))
	if err != nil {
		return "", "", false
	}
	creds := string(b)
	i := strings.IndexByte(creds, ':')
	if i < 0 {
		return "", "", false
	}
	return creds[:i], creds[i+1:], true
}

func writeHTTPError(w io.Writer, code int, msg string) {
	_, _ = fmt.Fprintf(
		w,
		"HTTP/1.1 %d %s\r\n"+
			"Content-Length: 0\r\n"+
			"Proxy-Agent: ProxyChan\r\n\r\n",
		code,
		msg,
	)
}

func (s *Server) unregisterConn(id uint64) {
	s.connMu.Lock()
	delete(s.conns, id)
	s.connMu.Unlock()
}

func (s *Server) registerConn(username, srcIP, dst string) uint64 {
	id := s.nextConnID.Add(1)

	ac := &models.ActiveConn{
		ID:          id,
		Username:    username,
		SourceIP:    srcIP,
		Destination: dst,
		StartedAt:   time.Now(),
	}

	s.connMu.Lock()
	s.conns[id] = ac
	s.connMu.Unlock()

	return id
}
