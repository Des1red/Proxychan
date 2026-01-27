package web

import (
	"database/sql"
	"net/http"
	"proxychan/internal/system"
)

func adminLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html, err := staticFS.ReadFile("static/auth.html")
		if err != nil {
			http.Error(w, "failed to load auth page", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(html)
	}
}

func adminLoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", 400)
			return
		}

		pwd := r.FormValue("password")
		if pwd == "" {
			http.Error(w, "missing password", 400)
			return
		}

		if err := system.VerifyAdminCredentials(db, pwd); err != nil {
			http.Error(w, "invalid password", http.StatusUnauthorized)
			return
		}

		token := issueAdminToken()

		http.SetCookie(w, &http.Cookie{
			Name:     adminCookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   0,
		})

		http.Redirect(w, r, "/connections", http.StatusSeeOther)
	}
}

func adminNotConfiguredHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Admin Access Disabled</title>
	<style>
		body {
			font-family: monospace;
			background: #0f1115;
			color: #e6e6e6;
			display: flex;
			justify-content: center;
			align-items: center;
			height: 100vh;
		}
		.box {
			text-align: center;
			opacity: 0.85;
		}
	</style>
</head>
<body>
	<div class="box">
		<h2>Admin Interface Disabled</h2>
		<p>Configure admin password to access page.</p>
	</div>
</body>
</html>
		`))
	}
}

func adminLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// remove token from memory
		if c, err := r.Cookie(adminCookieName); err == nil {
			adminTokensMu.Lock()
			delete(adminTokens, c.Value)
			adminTokensMu.Unlock()
		}

		// expire cookie in browser
		http.SetCookie(w, &http.Cookie{
			Name:     adminCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
