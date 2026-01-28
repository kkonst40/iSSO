package utils

import "github.com/kkonst40/isso/internal/config"

type CredValidator struct {
	loginChars     map[rune]struct{}
	maxLoginLength int
	minLoginLength int
	pwdChars       map[rune]struct{}
	maxPwdLength   int
	minPwdLength   int
}

func NewValidator(cfg *config.Config) *CredValidator {
	loginCharsMap := make(map[rune]struct{})
	pwdCharsMap := make(map[rune]struct{})

	for _, c := range cfg.Cred.LoginChars {
		loginCharsMap[c] = struct{}{}
	}
	for _, c := range cfg.Cred.PwdChars {
		pwdCharsMap[c] = struct{}{}
	}

	return &CredValidator{
		loginChars:     loginCharsMap,
		maxLoginLength: cfg.Cred.MaxLoginLength,
		minLoginLength: cfg.Cred.MinLoginLength,
		pwdChars:       pwdCharsMap,
		maxPwdLength:   cfg.Cred.MaxPwdLength,
		minPwdLength:   cfg.Cred.MinPwdLength,
	}
}

func (v *CredValidator) ValidateLogin(login string) bool {
	length := 0
	for _, c := range login {
		length++
		if length > v.maxLoginLength {
			return false
		}
		if _, ok := v.loginChars[c]; !ok {
			return false
		}
	}

	if length < v.minLoginLength {
		return false
	}

	return true
}

func (v *CredValidator) ValidatePwd(pwd string) bool {
	length := 0
	for _, c := range pwd {
		length++
		if length > v.maxPwdLength {
			return false
		}
		if _, ok := v.pwdChars[c]; !ok {
			return false
		}
	}

	if length < v.minPwdLength {
		return false
	}

	return true
}
