package http

import (
	"encoding/json"
	"net/http"

	"github.com/jhony-samosir/SS-NotificationService/internal/delivery/http/middleware"
)

func NewRouter(hmacSecret string, handler *NotificationHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Setup protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("/api/notifications/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})
	protected.HandleFunc("/api/notifications/devices", handler.RegisterDevice)
	protected.HandleFunc("/api/notifications/history", handler.GetHistory)
	protected.HandleFunc("/api/notifications/", handler.MarkAsRead) // Matches /api/notifications/{id}/read in go 1.22+ it can be more specific, but we'll use base path and handle in handler

	// Wrap protected routes with HMAC middleware
	mux.Handle("/api/notifications/", middleware.HMACValidator(hmacSecret)(protected))

	return mux
}
