package middleware

import (
	"context"
	"net/http"

	"github.com/kkonst40/isso/internal/utils"
)

type contextKey string

const RequesterIDKey contextKey = "requesterID"

func Auth(next http.HandlerFunc, jwtProvider *utils.JWTProvider) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("pechenye")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := jwtProvider.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), RequesterIDKey, claims.ID)

		next(w, r.WithContext(ctx))
	})
}
