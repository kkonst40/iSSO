package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/apperror"
	"github.com/kkonst40/isso/internal/config"
	"github.com/kkonst40/isso/internal/dto"
	"github.com/kkonst40/isso/internal/middleware"
	"github.com/kkonst40/isso/internal/service"
)

type UserHandler struct {
	userService *service.UserService
	cfg         *config.Config
}

func New(userService *service.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userService: userService,
		cfg:         cfg,
	}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
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
	var inputIDs []uuid.UUID
	err := json.NewDecoder(r.Body).Decode(&inputIDs)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingIDs, err := h.userService.Exist(r.Context(), inputIDs)
	if err != nil {
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(existingIDs); err != nil {
		http.Error(w, "Encoding response body error", http.StatusInternalServerError)
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
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cfg.JWT.CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // false только для localhost без https
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(60 * 24 * time.Hour),
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)
	err := h.userService.Logout(r.Context(), requesterID)
	if err != nil {
		//
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "pechenye",
		Value:  "",
		MaxAge: -1,
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
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdateLogin(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)

	var req dto.LRUUser
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userService.UpdateLogin(r.Context(), requesterID, req.Login)
	if err != nil {
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)

	var req dto.LRUUser
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userService.UpdatePassword(r.Context(), requesterID, req.Password)
	if err != nil {
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requesterID := r.Context().Value(middleware.RequesterIDKey).(uuid.UUID)
	idStr := r.PathValue("id")
	ID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid request parameter 'id'", http.StatusBadRequest)
	}

	err = h.userService.Delete(r.Context(), ID, requesterID)
	if err != nil {
		errMsg, errCode := apperror.GetMsgCode(err)
		http.Error(w, errMsg, errCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
