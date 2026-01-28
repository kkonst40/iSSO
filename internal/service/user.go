package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/apperror"
	"github.com/kkonst40/isso/internal/model"
	"github.com/kkonst40/isso/internal/repo"
	"github.com/kkonst40/isso/internal/utils"
)

type UserService struct {
	jwtProvider   *utils.JWTProvider
	pwdHandler    *utils.PasswordHandler
	credValidator *utils.CredValidator
	userRepo      *repo.UserRepo
	specialID     uuid.UUID
}

func New(
	jwtProvider *utils.JWTProvider,
	pwdHandler *utils.PasswordHandler,
	credValidator *utils.CredValidator,
	userRepo *repo.UserRepo,
	specialID uuid.UUID,
) *UserService {
	return &UserService{
		jwtProvider:   jwtProvider,
		pwdHandler:    pwdHandler,
		credValidator: credValidator,
		userRepo:      userRepo,
		specialID:     specialID,
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
		if !errors.Is(err, apperror.ErrInternalDB) {
			return "", apperror.ErrInvalidCredentials
		}
		return "", err
	}

	if !s.pwdHandler.VerifyPwd(password, user.PasswordHash) {
		return "", apperror.ErrInvalidCredentials
	}

	return s.jwtProvider.Generate(user)
}

func (s *UserService) Create(ctx context.Context, login, password string) error {
	if !s.credValidator.ValidateLogin(login) {
		return apperror.ErrInvalidLogin
	}
	if !s.credValidator.ValidatePwd(password) {
		return apperror.ErrInvalidPwd
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("%w: user id", apperror.ErrGeneratingError)
	}
	pwdHash, err := s.pwdHandler.GeneratePwdHash(password)
	if err != nil {
		return fmt.Errorf("%w: password hash", apperror.ErrGeneratingError)
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
	if !s.credValidator.ValidateLogin(newLogin) {
		return apperror.ErrInvalidLogin
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	user.Login = newLogin

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdatePassword(ctx context.Context, ID uuid.UUID, newPwd string) error {
	if !s.credValidator.ValidatePwd(newPwd) {
		return apperror.ErrInvalidPwd
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	newPwdHash, err := s.pwdHandler.GeneratePwdHash(newPwd)
	if err != nil {
		return fmt.Errorf("%w: password hash", apperror.ErrGeneratingError)
	}

	user.PasswordHash = newPwdHash

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, ID, requesterID uuid.UUID) error {
	if requesterID != ID && requesterID != s.specialID {
		return apperror.ErrNoPermission
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
