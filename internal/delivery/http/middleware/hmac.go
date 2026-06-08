package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

// HMACValidator validates requests using HMAC-SHA256
func HMACValidator(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signature := r.Header.Get("X-Internal-Signature")
			if signature == "" {
				http.Error(w, "Unauthorized: Missing Signature", http.StatusUnauthorized)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Restore body for next handlers
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			expectedMAC := hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
				http.Error(w, "Unauthorized: Invalid Signature", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
