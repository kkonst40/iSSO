package eventbus

import (
	"time"

	"github.com/google/uuid"
)

const (
	eventTypeSessionInvalidation = "SESSION_INVALIDATION"
	eventTypeLoginUpdate         = "LOGIN_UPDATE"
)

type eventMessage struct {
	Type      string    `json:"type"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type sessionInvalidationPayload struct {
	SessionID uuid.UUID `json:"session_id"`
	TTLDays   int       `json:"ttl_days"`
}

type loginUpdatePayload struct {
	UserID uuid.UUID `json:"user_id"`
	Login  string    `json:"login"`
}
