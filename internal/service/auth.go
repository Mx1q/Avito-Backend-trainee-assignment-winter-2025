package service

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/jwt"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"context"
	"fmt"
)

type AuthService struct {
	logger   logger.ILogger
	authRepo entity.IAuthRepository
	hasher   jwt.IHashCrypto
	jwtKey   string
}

func NewAuthService(repo entity.IAuthRepository, logger logger.ILogger, hasher jwt.IHashCrypto, jwtKey string) entity.IAuthService {
	return &AuthService{
		logger:   logger,
		authRepo: repo,
		hasher:   hasher,
		jwtKey:   jwtKey,
	}
}

func isValid(authInfo entity.Auth) error {
	if authInfo.Username == "" {
		return fmt.Errorf("empty username")
	}
	if authInfo.Password.Password == "" {
		return fmt.Errorf("empty password")
	}
	return nil
}

func (s *AuthService) Auth(ctx context.Context, authInfo *entity.Auth) (string, error) {
	s.logger.Infof("User %s trying to login", authInfo.Username)
	err := isValid(*authInfo)
	if err != nil {
		s.logger.Warnf("User %s sent invalid data: %v", authInfo.Username, err)
		return "", err
	}

	hashedPass, err := s.hasher.HashPassword(authInfo.Password.Password)
	if err != nil {
		s.logger.Warnf("User %s hashing pass: %v", authInfo.Username, err)
		return "", err
	}
	authInfo.Password.HashedPassword = hashedPass

	err = s.authRepo.Auth(ctx, authInfo)
	if err != nil {
		s.logger.Warnf("User %s trying to login: %v", authInfo.Username, err)
		return "", err
	}

	token, err := jwt.CreateToken(authInfo.Username, s.jwtKey)
	if err != nil {
		s.logger.Warnf("User %s trying to login: geerating auth token error (%v)", err)
		return "", fmt.Errorf("generating token: %w", err)
	}

	return token, nil
}
