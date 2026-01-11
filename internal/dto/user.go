package dto

import "github.com/google/uuid"

type GetUser struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}

// Login, register, and update user DTO
type LRUUser struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
