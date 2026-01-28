package server

import (
	"database/sql"
	"errors"
	"net"
	"proxychan/internal/socks5"
	"proxychan/internal/system"
	"time"
)

func (s *Server) authenticate(client net.Conn, db *sql.DB) (string, error) {
	_ = client.SetDeadline(time.Now().Add(15 * time.Second))

	username, err := socks5.HandleHandshake(client, socks5.HandshakeOptions{
		RequireAuth: s.cfg.RequireAuth,
		AuthFunc:    s.cfg.AuthFunc,
	})
	if err != nil {
		s.cfg.Logger.Warnf(
			"handshake error from %s: %v",
			client.RemoteAddr(),
			err,
		)
		return "", err
	}

	if s.cfg.RequireAuth {
		active, err := system.IsActive(db, username)
		if err != nil {
			s.cfg.Logger.Warnf(
				"error checking if user %s is active: %v",
				username,
				err,
			)
			return "", err
		}

		if !active {
			s.cfg.Logger.Warnf(
				"user %s is inactive, rejecting connection",
				username,
			)
			_ = socks5.WriteReply(client, 0x05)
			return "", errors.New("user inactive")
		}
	}

	return username, nil
}

func (s *Server) readAndAuthorizeRequest(
	client net.Conn,
	username string,
) (*socks5.Request, error) {

	req, err := socks5.ReadRequest(client)
	if err != nil {
		_ = socks5.WriteReply(client, 0x07)
		s.cfg.Logger.Warnf(
			"request error from %s: %v",
			client.RemoteAddr(),
			err,
		)
		return nil, err
	}

	destHost, _, err := net.SplitHostPort(req.Address)
	if err != nil {
		_ = socks5.WriteReply(client, 0x01)
		return nil, err
	}

	if typ, pat, denied := s.destDenied(destHost); denied {
		_ = socks5.WriteReply(client, 0x02)
		s.cfg.Logger.Warnf(
			"egress denied user=%q src=%s dst=%s ruleType=%s rule=%s",
			username,
			client.RemoteAddr().String(),
			req.Address,
			typ,
			pat,
		)
		return nil, errors.New("destination denied")
	}

	return &req, nil
}
