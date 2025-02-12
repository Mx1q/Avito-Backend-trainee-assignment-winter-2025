package service

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
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

func isValid(authInfo *entity.Auth) error { // FIXME: check if authInfo is nil
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
	err := isValid(authInfo)
	if err != nil {
		s.logger.Warnf("User %s sent invalid data: %v", authInfo.Username, err)
		return "", errs.InvalidData
	}

	userDb, err := s.authRepo.GetByUsername(ctx, authInfo.Username)
	if err != nil {
		s.logger.Warnf("User %s trying to login: %v", authInfo.Username, err)
		return "", errs.InternalError
	}
	if userDb == nil {
		s.logger.Infof("User %s not exists, trying to register", authInfo.Username)
		err = s.register(ctx, authInfo)
		if err != nil {
			s.logger.Warnf("User %s trying to register: %v", authInfo.Username, err)
			return "", err
		}
	} else {
		if !s.hasher.VerifyPassword(authInfo.Password, userDb.Password) {
			s.logger.Warnf("User %s trying to login with invalid pass", authInfo.Username)
			return "", errs.InvalidCredentials
		}
	}

	token, err := jwt.CreateToken(authInfo.Username, s.jwtKey)
	if err != nil {
		s.logger.Warnf("User %s trying to login: creating auth token error (%v)",
			authInfo.Username, err)
		return "", errs.InternalError
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
