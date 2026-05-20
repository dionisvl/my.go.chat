package root

import (
	"net/http"

	"mygochat/web"
)

// ServeHTTP serves the embedded chat client HTML.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(web.IndexHTML)
}
