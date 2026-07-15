package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	errs "github.com/kkonst40/sso-service/internal/domain/errors"
	"github.com/kkonst40/sso-service/internal/domain/model"
)

type Service struct {
	jwtProvider   JWTProvider
	pwdHasher     PasswordHasher
	credValidator CredValidator
	eventProducer Producer
	userRepo      UserRepo
	sessionRepo   SessionRepo
}

type UserRepo interface {
	Create(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, ID uuid.UUID) error
	Exist(ctx context.Context, IDs []uuid.UUID) ([]uuid.UUID, error)
	GetByID(ctx context.Context, ID uuid.UUID) (model.User, error)
	GetByLogin(ctx context.Context, login string) (model.User, error)
	GetIDsByLogins(ctx context.Context, logins []string) ([]model.UserInfo, error)
	GetLoginsByIDs(ctx context.Context, IDs []uuid.UUID) ([]model.UserInfo, error)
	Update(ctx context.Context, user *model.User) error
}

type SessionRepo interface {
	Create(ctx context.Context, session *model.Session) error
	Delete(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID) (uuid.UUID, error)
	DeleteAll(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type JWTProvider interface {
	Generate(user *model.User, session *model.Session) (string, error)
	GetTTLDays() int
}

type PasswordHasher interface {
	VerifyPwd(password string, passwordHash string) bool
	GeneratePwdHash(password string) (string, error)
}

type CredValidator interface {
	ValidateLogin(login string) bool
	ValidatePwd(pwd string) bool
}

type Producer interface {
	SendLoginUpdate(ctx context.Context, userID uuid.UUID, login string) error
	SendSessionInvalidation(ctx context.Context, sessionID uuid.UUID, ttlDays int) error
}

func New(
	jwtProvider JWTProvider,
	pwdHasher PasswordHasher,
	credValidator CredValidator,
	eventProducer Producer,
	userRepo UserRepo,
	sessionRepo SessionRepo,
) *Service {
	return &Service{
		jwtProvider:   jwtProvider,
		pwdHasher:     pwdHasher,
		credValidator: credValidator,
		eventProducer: eventProducer,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
	}
}

func (s *Service) GetLoginsByIDs(ctx context.Context, userIDs []uuid.UUID) ([]model.UserInfo, error) {
	userInfos, err := s.userRepo.GetLoginsByIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("get logins by IDs: %w", err)
	}

	return userInfos, nil
}

func (s *Service) GetIDsByLogins(ctx context.Context, userLogins []string) ([]model.UserInfo, error) {
	userInfos, err := s.userRepo.GetIDsByLogins(ctx, userLogins)
	if err != nil {
		return nil, fmt.Errorf("get IDs by logins: %w", err)
	}

	return userInfos, nil
}

func (s *Service) Exist(ctx context.Context, IDs []uuid.UUID) ([]uuid.UUID, error) {
	ids, err := s.userRepo.Exist(ctx, IDs)
	if err != nil {
		return nil, fmt.Errorf("check users existence: %w", err)
	}

	return ids, nil
}

func (s *Service) Login(ctx context.Context, login, password string, deviceID uuid.UUID) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return "", errs.ErrInvalidCredentials
		}
		return "", fmt.Errorf("get user by login '%v' to log in: %w", login, err)
	}

	if !s.pwdHasher.VerifyPwd(password, user.PasswordHash) {
		return "", errs.ErrInvalidCredentials
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("%w: session id: %w", errs.ErrGenerating, err)
	}

	session := &model.Session{
		ID:       sessionID,
		UserID:   user.ID,
		DeviceID: deviceID,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", fmt.Errorf("creating session: %w", err)
	}

	token, err := s.jwtProvider.Generate(&user, session)
	if err != nil {
		return "", fmt.Errorf("%w: jwt token: %w", errs.ErrGenerating, err)
	}

	return token, nil
}

func (s *Service) Create(ctx context.Context, login, password string) error {
	if !s.credValidator.ValidateLogin(login) {
		return errs.ErrInvalidLogin
	}
	if !s.credValidator.ValidatePwd(password) {
		return errs.ErrInvalidPwd
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("%w: user id: %w", errs.ErrGenerating, err)
	}
	pwdHash, err := s.pwdHasher.GeneratePwdHash(password)
	if err != nil {
		return fmt.Errorf("%w: password hash: %w", errs.ErrGenerating, err)
	}

	user := &model.User{
		ID:           userID,
		Login:        login,
		PasswordHash: pwdHash,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *Service) UpdateLogin(ctx context.Context, ID uuid.UUID, newLogin string) error {
	if !s.credValidator.ValidateLogin(newLogin) {
		return errs.ErrInvalidLogin
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("get user %v to update: %w", ID, err)
	}

	user.Login = newLogin

	if err := s.userRepo.Update(ctx, &user); err != nil {
		return fmt.Errorf("update user %v: %w", ID, err)
	}

	if err := s.eventProducer.SendLoginUpdate(ctx, user.ID, user.Login); err != nil {
		log.Println("sending update login event to event queue error: %w", err)
	}

	return nil
}

func (s *Service) UpdatePassword(ctx context.Context, ID uuid.UUID, newPwd string) error {
	if !s.credValidator.ValidatePwd(newPwd) {
		return errs.ErrInvalidPwd
	}

	user, err := s.userRepo.GetByID(ctx, ID)
	if err != nil {
		return fmt.Errorf("get user %v to update: %w", ID, err)
	}

	newPwdHash, err := s.pwdHasher.GeneratePwdHash(newPwd)
	if err != nil {
		return fmt.Errorf("%w: password hash: %w", errs.ErrGenerating, err)
	}

	user.PasswordHash = newPwdHash

	if err := s.userRepo.Update(ctx, &user); err != nil {
		return fmt.Errorf("update user %v: %w", ID, err)
	}

	return nil
}

func (s *Service) Delete(ctx context.Context, ID, requesterID uuid.UUID) error {
	if requesterID != ID {
		return errs.ErrForbidden
	}

	if err := s.userRepo.Delete(ctx, ID); err != nil {
		return fmt.Errorf("delete user %v: %w", ID, err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, ID uuid.UUID) error {
	sessionIDs, err := s.sessionRepo.DeleteAll(ctx, ID)
	if err != nil {
		return fmt.Errorf("session delete: %w", err)
	}

	for _, sessionID := range sessionIDs {
		if err := s.eventProducer.SendSessionInvalidation(ctx, sessionID, s.jwtProvider.GetTTLDays()); err != nil {
			log.Println("sending session invalidation event to event queue error: %w", err)
		}
	}

	return nil
}

func (s *Service) Logout(ctx context.Context, ID, deviceID uuid.UUID) error {
	sessionID, err := s.sessionRepo.Delete(ctx, ID, deviceID)
	if err != nil {
		return fmt.Errorf("session delete: %w", err)
	}

	if err := s.eventProducer.SendSessionInvalidation(ctx, sessionID, s.jwtProvider.GetTTLDays()); err != nil {
		log.Println("sending session invalidation event to event queue error: %w", err)
	}

	return nil
}
