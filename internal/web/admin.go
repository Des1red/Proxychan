package web

import (
	"context"
	"database/sql"
	"net/http"
	"proxychan/internal/models"
)

type ConnectionProvider interface {
	SnapshotConnections() []models.ActiveConn
	GroupConnectionsByIP([]models.ActiveConn) []models.ConnGroup
	Warnf(format string, args ...any)
}

func RunAdminEndpoint(ctx context.Context, p ConnectionProvider, db *sql.DB) {
	app := http.NewServeMux()

	app.Handle("/static/", http.FileServer(http.FS(staticFS)))
	app.HandleFunc("/login", adminLoginPage())
	app.HandleFunc("/login/submit", adminLoginHandler(db))
	app.HandleFunc("/connections", connectionsHTMLHandler())
	app.HandleFunc("/connections/by-ip", connectionsJSONHandler(p))
	app.HandleFunc("/logout", adminLogoutHandler())

	handler := adminGate(db, app)

	srv := &http.Server{
		Addr:    "127.0.0.1:6060",
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		p.Warnf("admin endpoint error: %v", err)
	}
}
