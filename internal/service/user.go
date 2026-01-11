package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/model"
	"github.com/kkonst40/isso/internal/repo"
	"github.com/kkonst40/isso/internal/utils"
)

type UserService struct {
	jwtProvider *utils.JWTProvider
	pwdHandler  *utils.PasswordHandler
	userRepo    *repo.UserRepo
	specialID   uuid.UUID
}

func New(
	jwtProvider *utils.JWTProvider,
	pwdHandler *utils.PasswordHandler,
	userRepo *repo.UserRepo,
	specialID uuid.UUID,
) *UserService {
	return &UserService{
		jwtProvider: jwtProvider,
		pwdHandler:  pwdHandler,
		userRepo:    userRepo,
		specialID:   specialID,
	}
}

func (s *UserService) All(ctx context.Context) ([]model.User, error) {
	return s.userRepo.GetAll(ctx)
}

func (s *UserService) Exist(ctx context.Context, IDs []uuid.UUID) ([]uuid.UUID, error) {
	//if requesterID != s.specialID {
	//	return false, fmt.Errorf("no permission")
	//}
	return s.userRepo.Exist(ctx, IDs)
}

func (s *UserService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if !s.pwdHandler.VerifyPwd(password, user.PasswordHash) {
		return "", fmt.Errorf("invalid password")
	}

	return s.jwtProvider.Generate(user)
}

func (s *UserService) Create(ctx context.Context, login, password string) error {
	if !isValidLogin(login) {
		return fmt.Errorf("invalid login")
	}
	if !isValidPassword(password) {
		return fmt.Errorf("invalid password")
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generating user id error")
	}
	pwdHash, err := s.pwdHandler.GeneratePwdHash(password)
	if err != nil {
		return fmt.Errorf("generating password hash error")
	}

	user := &model.User{
		ID:           userID,
		Login:        login,
		PasswordHash: pwdHash,
		TokenID:      uuid.New(),
	}

	return s.userRepo.Create(ctx, user)
}

func (s *UserService) UpdateLogin(ctx context.Context, ID uuid.UUID, newLogin string) error {
	if !isValidLogin(newLogin) {
		return fmt.Errorf("invalid login")
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	user.Login = newLogin

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdatePassword(ctx context.Context, ID uuid.UUID, newPwd string) error {
	if !isValidPassword(newPwd) {
		return fmt.Errorf("invalid password")
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	newPwdHash, err := s.pwdHandler.GeneratePwdHash(newPwd)
	if err != nil {
		return fmt.Errorf("generating password hash error")
	}

	user.PasswordHash = newPwdHash

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, ID, requesterID uuid.UUID) error {
	if requesterID != ID && requesterID != s.specialID {
		return fmt.Errorf("no permission")
	}

	return s.userRepo.Delete(ctx, ID)
}

func (s *UserService) Logout(ctx context.Context, ID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("logging out error")
	}

	user.TokenID = uuid.New()
	return s.userRepo.Update(ctx, user)
}

func isValidLogin(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !(r >= 'a' && r <= 'z' ||
			r >= 'A' && r <= 'Z' ||
			r >= '0' && r <= '9' ||
			r == '_') {
			return false
		}
	}
	return true
}

func isValidPassword(s string) bool {
	if len(s) < 8 || len(s) > 64 {
		return false
	}

	for _, r := range s {
		if r < 33 || r > 126 {
			return false
		}
	}
	return true
}
