package jwt

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = bcrypt.MinCost

type IHashCrypto interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) bool
}

type HashCrypto struct {
}

func NewHashCrypto() IHashCrypto {
	return HashCrypto{}
}

func (c HashCrypto) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	if err != nil {
		return "", fmt.Errorf("generating hash: %w", err)
	}
	return string(bytes), nil
}

func (c HashCrypto) VerifyPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
