package service

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/logger"
	"context"
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

func (s *ItemService) isValid(itemName string, username string) error {
	if itemName == "" {
		return fmt.Errorf("empty item name")
	}
	if username == "" {
		return fmt.Errorf("empty username")
	}
	return nil
}

func (s *ItemService) BuyItem(ctx context.Context, itemName string, username string) error {
	s.logger.Infof("User %s trying to buy item %s", username, itemName)
	err := s.isValid(itemName, username)
	if err != nil {
		s.logger.Warnf("User %s trying to buy item %s: %v", username, itemName, err)
		return errs.InvalidData
	}

	err = s.itemRepo.BuyItem(ctx, itemName, username)
	if err != nil {
		s.logger.Warnf("User %s trying to buy item %s: %v", username, itemName, err)
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
