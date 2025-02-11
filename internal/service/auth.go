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
	if authInfo.Password == "" {
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

	userDb, err := s.authRepo.GetByUsername(ctx, authInfo.Username)
	if err != nil {
		s.logger.Warnf("User %s trying to login: %v", authInfo.Username, err)
		return "", err
	}
	if userDb == nil {
		err = s.register(ctx, authInfo)
		if err != nil {
			return "", err
		}
	} else {
		if !s.hasher.VerifyPassword(authInfo.Password, userDb.Password) {
			s.logger.Warnf("login user: invalid password")
			return "", fmt.Errorf("invalid password")
		}
	}

	token, err := jwt.CreateToken(authInfo.Username, s.jwtKey)
	if err != nil {
		s.logger.Warnf("User %s trying to login: generating auth token error (%v)", err)
		return "", fmt.Errorf("generating token: %w", err)
	}

	return token, nil
}

func (s *AuthService) register(ctx context.Context, authInfo *entity.Auth) error {
	hashedPass, err := s.hasher.HashPassword(authInfo.Password)
	if err != nil {
		s.logger.Warnf("User %s hashing pass: %v", authInfo.Username, err)
		return err
	}
	authInfo.Password = hashedPass

	err = s.authRepo.Register(ctx, authInfo)
	if err != nil {
		s.logger.Warnf("User %s trying to login: %v", authInfo.Username, err)
		return err
	}

	return nil
}
