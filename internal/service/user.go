package service

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"context"
	"errors"
	"fmt"
)

type UserService struct {
	logger   logger.ILogger
	userRepo entity.IUserRepository
}

func NewUserService(repo entity.IUserRepository, logger logger.ILogger) entity.IUserService {
	return &UserService{
		logger:   logger,
		userRepo: repo,
	}
}

func (s *UserService) isValid(fromUser string, toUser string, amount int32) error {
	if fromUser == "" {
		return fmt.Errorf("empty fromUser")
	}
	if toUser == "" {
		return fmt.Errorf("empty toUser")
	}
	if amount <= 0 {
		return fmt.Errorf("negative or zero amount of coins")
	}
	if fromUser == toUser {
		return fmt.Errorf("same user as reciever and sender")
	}

	return nil
}

func (s *UserService) SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error {
	err := s.isValid(fromUser, toUser, amount)
	if err != nil {
		s.logger.Warnf("User \"%s\" trying to transfer coins (%d) to \"%s\": %v",
			fromUser, amount, toUser, err)
		return errs.InvalidData
	}
	s.logger.Infof("User \"%s\" trying to transfer coins (%d) to \"%s\"",
		fromUser, amount, toUser)

	err = s.userRepo.SendCoins(ctx, fromUser, toUser, amount)
	if err != nil {
		s.logger.Warnf("User \"%s\" trying to transfer coins (%d) to \"%s\": %v",
			fromUser, amount, toUser, err)

		if errors.Is(err, errs.UserNotFound) || errors.Is(err, errs.NotEnoughCoins) {
			return err
		}
		return errs.InternalError
	}

	return nil
}

func (s *UserService) GetCoinsHistory(ctx context.Context, username string) (int32, *entity.CoinsHistory, error) {
	if username == "" {
		s.logger.Warnf("Getting coins history for empty username")
		return 0, nil, errs.InvalidData
	}
	s.logger.Infof("Getting coins history for user \"%s\"", username)

	coins, coinsHistory, err := s.userRepo.GetCoinsHistory(ctx, username)
	if err != nil {
		s.logger.Warnf("Getting coins history for user \"%s\": %v", username, err)
		return 0, nil, errs.InternalError
	}

	return coins, coinsHistory, nil
}
