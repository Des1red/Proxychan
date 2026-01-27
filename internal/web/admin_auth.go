package web

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/http"
	"proxychan/internal/system"
	"strings"
	"sync"
	"time"
)

const adminCookieName = "proxychan_admin"

var (
	adminTokens   = make(map[string]time.Time)
	adminTokensMu sync.RWMutex
)

// issueAdminToken creates a new in-memory admin auth token
func issueAdminToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	adminTokensMu.Lock()
	adminTokens[token] = time.Now()
	adminTokensMu.Unlock()

	return token
}

// isAdminAuthenticated checks whether request has a valid admin cookie
func isAdminAuthenticated(r *http.Request) bool {
	c, err := r.Cookie(adminCookieName)
	if err != nil {
		return false
	}

	adminTokensMu.RLock()
	_, ok := adminTokens[c.Value]
	adminTokensMu.RUnlock()

	return ok
}

func adminGate(db *sql.DB, app http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// allow static assets unconditionally
		if strings.HasPrefix(r.URL.Path, "/static/") {
			app.ServeHTTP(w, r)
			return
		}
		sec, err := system.InternalAdminSecret()
		if err == nil && r.Header.Get("X-ProxyChan-Internal") == sec {
			app.ServeHTTP(w, r)
			return
		}

		ok, err := system.AdminPasswordConfigured(db)
		if err != nil {
			http.Error(w, "admin auth error", http.StatusInternalServerError)
			return
		}

		if !ok {
			adminNotConfiguredHandler().ServeHTTP(w, r)
			return
		}

		// allow login endpoints without auth
		if r.URL.Path == "/login" ||
			r.URL.Path == "/login/submit" ||
			r.URL.Path == "/logout" {
			app.ServeHTTP(w, r)
			return
		}

		// browser auth path
		if isAdminAuthenticated(r) {
			app.ServeHTTP(w, r)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
}
