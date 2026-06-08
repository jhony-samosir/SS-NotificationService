package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jhony-samosir/SS-NotificationService/internal/domain"
	"github.com/jhony-samosir/SS-NotificationService/internal/usecase"
)

type NotificationHandler struct {
	apiUsecase *usecase.NotificationAPIUsecase
}

func NewNotificationHandler(apiUsecase *usecase.NotificationAPIUsecase) *NotificationHandler {
	return &NotificationHandler{apiUsecase: apiUsecase}
}

func (h *NotificationHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized: Missing X-User-Id", http.StatusUnauthorized)
		return
	}

	var payload struct {
		DeviceToken string `json:"device_token"`
		DeviceType  string `json:"device_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.apiUsecase.RegisterDevice(r.Context(), userID, payload.DeviceToken, payload.DeviceType); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Device registered"})
}

func (h *NotificationHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized: Missing X-User-Id", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 20
	}

	history, err := h.apiUsecase.GetHistory(r.Context(), userID, page, limit)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if len(history) == 0 {
		history = []domain.Notification{} // Avoid null in JSON array response
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized: Missing X-User-Id", http.StatusUnauthorized)
		return
	}

	// Extract ID from /api/notifications/{id}/read
	// path looks like /api/notifications/123/read -> ["", "api", "notifications", "123", "read"]
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	notificationID := parts[3]

	if err := h.apiUsecase.MarkAsRead(r.Context(), userID, notificationID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Marked as read"})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized: Missing X-User-Id", http.StatusUnauthorized)
		return
	}

	if err := h.apiUsecase.MarkAllAsRead(r.Context(), userID); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "All marked as read"})
}
