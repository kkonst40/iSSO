package utils

import "golang.org/x/crypto/bcrypt"

type PasswordHandler struct{}

func NewPasswordHandler() *PasswordHandler {
	return &PasswordHandler{}
}

func (h *PasswordHandler) GeneratePwdHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *PasswordHandler) VerifyPwd(password string, passwordHash string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(password),
	)
	return err == nil
}
