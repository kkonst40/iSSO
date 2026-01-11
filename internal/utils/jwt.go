package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kkonst40/isso/internal/config"
	"github.com/kkonst40/isso/internal/model"
)

type UserClaims struct {
	ID       uuid.UUID `json:"id"`
	UserName string    `json:"userName"`
	TokenID  uuid.UUID `json:"tokenId"`
	jwt.RegisteredClaims
}

type JWTProvider struct {
	Cfg *config.JWTConfig
}

func NewJWTProvider(cfg *config.JWTConfig) *JWTProvider {
	return &JWTProvider{
		Cfg: cfg,
	}
}

func (p *JWTProvider) Generate(user *model.User) (string, error) {
	claims := UserClaims{
		ID:       user.ID,
		TokenID:  user.TokenID,
		UserName: user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   p.Cfg.Issuer,
			Audience: []string{p.Cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Duration(p.Cfg.ExpireDays) * 24 * time.Hour),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(p.Cfg.SecretKey))
}

func (p *JWTProvider) ValidateToken(tokenString string) (*UserClaims, error) {
	claims := &UserClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(p.Cfg.SecretKey), nil
		},
		jwt.WithIssuer(p.Cfg.Issuer),
		jwt.WithAudience(p.Cfg.Audience),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
