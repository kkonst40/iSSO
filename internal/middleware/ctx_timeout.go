package middleware

import (
	"net/http"
	"time"
)

func Timeout(next http.Handler, timeoutDuration time.Duration) http.Handler {
	return http.TimeoutHandler(next, timeoutDuration, "Internal Server Error (Timeout)")
}
