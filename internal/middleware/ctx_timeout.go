package middleware

import (
	"net/http"
	"time"
)

func Timeout(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, 2*time.Second, "Internal Server Error (Timeout)")
}
