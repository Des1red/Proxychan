package web

import (
	"encoding/json"
	"net/http"
)

func connectionsHTMLHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		html, err := staticFS.ReadFile("static/connections.html")

		if err != nil {
			http.Error(w, "failed to load html", http.StatusInternalServerError)
			return
		}

		w.Write(html)
	}
}

func connectionsJSONHandler(p ConnectionProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		conns := p.SnapshotConnections()
		groups := p.GroupConnectionsByIP(conns)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(groups)
	}
}
