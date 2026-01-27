package web

import (
	"context"
	"net/http"
	"proxychan/internal/models"
)

type ConnectionProvider interface {
	SnapshotConnections() []models.ActiveConn
	GroupConnectionsByIP([]models.ActiveConn) []models.ConnGroup
	Warnf(format string, args ...any)
}

func RunAdminEndpoint(ctx context.Context, p ConnectionProvider) {
	mux := http.NewServeMux()

	// static assets
	mux.Handle(
		"/static/",
		http.FileServer(http.FS(staticFS)),
	)

	mux.HandleFunc("/connections", connectionsHTMLHandler())
	mux.HandleFunc("/connections/by-ip", connectionsJSONHandler(p))

	srv := &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.Warnf("admin endpoint error: %v", err)
	}
}
