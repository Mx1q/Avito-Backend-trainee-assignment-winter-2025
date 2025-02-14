package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type ITokenManager interface {
	CreateToken(username string) (string, error)
	VerifyToken(tokenString string) (*jwt.Token, error)
}

type TokenManager struct {
	jwtKey string
}

func NewTokenManager(jwtKey string) ITokenManager {
	return &TokenManager{
		jwtKey: jwtKey,
	}
}

func (m *TokenManager) CreateToken(username string) (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
			"iss": "AvitoShop",
			"exp": time.Now().Add(time.Hour * 24).Unix(),
			"iat": time.Now().Unix(),
		})

	tokenString, err := token.SignedString([]byte(m.jwtKey))
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return tokenString, nil
}

func (m *TokenManager) VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.jwtKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}
