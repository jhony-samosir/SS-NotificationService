package http

import (
	"encoding/json"
	"net/http"

	"github.com/jhony-samosir/SS-NotificationService/internal/delivery/http/middleware"
)

func NewRouter(hmacSecret string) *http.ServeMux {
	mux := http.NewServeMux()

	// Setup protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("/api/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Wrap protected routes with HMAC middleware
	mux.Handle("/api/notifications/", middleware.HMACValidator(hmacSecret)(protected))

	return mux
}
