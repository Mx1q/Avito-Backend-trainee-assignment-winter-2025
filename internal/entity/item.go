package entity

import "context"

type Item struct {
	Name     string
	Price    int32
	Quantity int32
}

type Purchase struct {
	Username string
	ItemName string
}

type IItemRepository interface {
	GetInventory(ctx context.Context, username string) ([]*Item, error)
	BuyItem(ctx context.Context, purchase *Purchase) error
}

type IItemService interface {
	GetInventory(ctx context.Context, username string) ([]*Item, error)
	BuyItem(ctx context.Context, purchase *Purchase) error
}
