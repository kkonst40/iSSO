package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/dto"
	"github.com/kkonst40/isso/internal/middleware"
	"github.com/kkonst40/isso/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func New(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) All(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.All(r.Context())
	if err != nil {
		return
	}

	userDTOs := make([]dto.GetUser, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, dto.GetUser{
			ID:    user.ID,
			Login: user.Login,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(userDTOs); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Exist(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)
	idStr := r.PathValue("id")
	ID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid request url (id)", http.StatusBadRequest)
		return
	}

	exist, err := h.userService.Exist(r.Context(), ID, requesterID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !exist {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LRUUser
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		//
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "pechenye",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // false ТОЛЬКО для localhost без https
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(60 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)
	err := h.userService.Logout(r.Context(), requesterID)
	if err != nil {
		http.Error(w, "Error logging out", http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "pechenye",
		Value: "",
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.LRUUser
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userService.Create(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	//requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)

	var req dto.LRUUser
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//h.userService.Update(r.Context(), )

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)
	idStr := r.PathValue("id")
	ID, err := uuid.Parse(idStr)
	if err != nil {
		//
	}

	err = h.userService.Delete(r.Context(), ID, requesterID)
	if err != nil {
		//
	}

	w.WriteHeader(http.StatusNoContent)
}
