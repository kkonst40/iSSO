package apperror

import (
	"errors"
	"net/http"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPwd         = errors.New("invalid password")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrInternalDB         = errors.New("internal db error")
	ErrUserNotFound       = errors.New("user not found")
	ErrLoginTaken         = errors.New("user already exists")
	ErrNoPermission       = errors.New("no permission")
	ErrGeneratingError    = errors.New("generating error")
)

func GetMsgCode(err error) (string, int) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return "Invalid login or password", http.StatusUnauthorized

	case errors.Is(err, ErrInvalidLogin):
		return "Invalid login", http.StatusUnprocessableEntity

	case errors.Is(err, ErrInvalidPwd):
		return "Invalid password", http.StatusUnprocessableEntity

	case errors.Is(err, ErrUserNotFound):
		return "User not found", http.StatusNotFound

	case errors.Is(err, ErrLoginTaken):
		return "Login already taken", http.StatusConflict

	case errors.Is(err, ErrNoPermission):
		return "User has no permission", http.StatusForbidden

	default:
		if err != nil {
			return "Internal server error", http.StatusInternalServerError
		}
		return "", http.StatusOK
	}
}
