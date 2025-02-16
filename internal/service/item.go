package service

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"context"
	"errors"
	"fmt"
)

type ItemService struct {
	logger   logger.ILogger
	itemRepo entity.IItemRepository
}

func NewItemService(repo entity.IItemRepository, logger logger.ILogger) entity.IItemService {
	return &ItemService{
		logger:   logger,
		itemRepo: repo,
	}
}

func (s *ItemService) isValid(purchase *entity.Purchase) error {
	if purchase == nil {
		return fmt.Errorf("pointer to struct is nil")
	}
	if purchase.ItemName == "" {
		return fmt.Errorf("empty item name")
	}
	if purchase.Username == "" {
		return fmt.Errorf("empty username")
	}
	return nil
}

func (s *ItemService) BuyItem(ctx context.Context, purchase *entity.Purchase) error {
	err := s.isValid(purchase)
	if err != nil {
		s.logger.Warnf("buying item invalid data: %v", err)
		return errs.InvalidData
	}
	s.logger.Infof("User %s trying to buy item %s", purchase.Username, purchase.ItemName)

	err = s.itemRepo.BuyItem(ctx, purchase)
	if err != nil {
		s.logger.Warnf("User %s trying to buy item %s: %v", purchase.Username, purchase.ItemName, err)
		if errors.Is(err, errs.ItemNotFound) ||
			errors.Is(err, errs.UserNotFound) || errors.Is(err, errs.NotEnoughCoins) {
			return err
		}
		return errs.InternalError
	}

	return nil
}

func (s *ItemService) GetInventory(ctx context.Context, username string) ([]*entity.Item, error) {
	s.logger.Infof("User \"%s\" getting his inventory", username)
	if username == "" {
		s.logger.Warnf("Getting inventory for empty username")
		return nil, errs.InvalidData
	}

	items, err := s.itemRepo.GetInventory(ctx, username)
	if err != nil {
		s.logger.Warnf("User \"%s\" getting his inventory: %v", username, err)
		return nil, errs.InternalError
	}

	return items, nil
}
