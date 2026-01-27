package server

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"proxychan/internal/dialer"
	"proxychan/internal/logging"
	"proxychan/internal/models"
	"proxychan/internal/web"

	"github.com/sirupsen/logrus"
)

type Config struct {
	ListenAddr     string // SOCKS5
	HTTPListenAddr string // HTTP CONNECT (optional)

	Dialer      dialer.Dialer
	IdleTimeout time.Duration
	Logger      *logrus.Logger

	RequireAuth bool
	AuthFunc    func(username, password string) error
}

type Server struct {
	cfg Config

	// ip whitelist
	mu               sync.RWMutex
	whitelist        []net.IPNet
	whitelistVersion int64

	//destination blacklist
	denyMu           sync.RWMutex
	denyIPNets       []net.IPNet
	denyDomainExact  map[string]struct{}
	denyDomainSuffix []string
	denyVersion      int64

	//active connections
	connMu     sync.RWMutex
	conns      map[uint64]*models.ActiveConn
	nextConnID atomic.Uint64
}

func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = logging.GetLogger()
	}
	return &Server{
		cfg:   cfg,
		conns: make(map[uint64]*models.ActiveConn),
	}
}

func (s *Server) logStartupInfo() {
	s.cfg.Logger.Infof("listening on %s", s.cfg.ListenAddr)

	if s.cfg.RequireAuth {
		if ip, err := detectPublicIP(); err == nil {
			_, port, _ := net.SplitHostPort(s.cfg.ListenAddr)
			s.cfg.Logger.Infof(
				"public address: %s",
				net.JoinHostPort(ip, port),
			)
		}
	}
}

func (s *Server) acceptLoop(ctx context.Context, ln net.Listener, db *sql.DB) error {
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

func (s *Server) startListener(ctx context.Context) (net.Listener, error) {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	return ln, nil
}

func (s *Server) startHTTPProxy(ctx context.Context, db *sql.DB) {
	ln, err := net.Listen("tcp", s.cfg.HTTPListenAddr)
	if err != nil {
		s.cfg.Logger.Fatalf("http proxy listen failed: %v", err)
	}

	s.cfg.Logger.Infof("HTTP CONNECT proxy listening on %s", s.cfg.HTTPListenAddr)

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handleHTTPConn(ctx, c, db)
	}
}

func (s *Server) Run(ctx context.Context, db *sql.DB) error {
	if err := s.initPolicies(ctx, db); err != nil {
		return err
	}

	ln, err := s.startListener(ctx)
	if err != nil {
		return err
	}

	s.logStartupInfo()

	if s.cfg.HTTPListenAddr != "" {
		go s.startHTTPProxy(ctx, db)
	}

	go web.RunAdminEndpoint(ctx, s, db)

	return s.acceptLoop(ctx, ln, db)
}

func (s *Server) handleConn(ctx context.Context, client net.Conn, db *sql.DB) {
	defer client.Close()

	srcIP, err := s.checkSource(client)
	if err != nil {
		return
	}

	username, err := s.authenticate(client, db)
	if err != nil {
		return
	}

	req, err := s.readAndAuthorizeRequest(client, username)
	if err != nil {
		return
	}

	s.handleTunnel(ctx, client, username, srcIP, req)
}
