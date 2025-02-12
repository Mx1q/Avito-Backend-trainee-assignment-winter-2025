package models

import "Avito-Backend-trainee-assignment-winter-2025/internal/entity"

type Item struct {
	Type     string `json:"type,omitempty"`
	Quantity int32  `json:"quantity,omitempty"`
}

func ToItemTransport(item *entity.Item) *Item {
	return &Item{
		Type:     item.Name,
		Quantity: item.Quantity,
	}
}

func ToInventoryTransport(items []*entity.Item) []*Item {
	inventory := make([]*Item, len(items))
	for i := 0; i < len(items); i++ {
		inventory[i] = ToItemTransport(items[i])
	}

	return inventory
}
