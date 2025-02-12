package entity

import "context"

type Item struct {
	Name  string
	Price int32
}

type IItemRepository interface {
	//GetInventory(username string) ([]Item, error)
	BuyItem(ctx context.Context, itemName string, username string) error
}

type IItemService interface {
	//GetInventory(username string) ([]Item, error)
	BuyItem(ctx context.Context, itemName string, username string) error
}
