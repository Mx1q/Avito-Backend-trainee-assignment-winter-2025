package entity

import "context"

type User struct {
	Username string
	Coins    int32
}

type CoinsHistory struct {
	Received []*User
	Sent     []*User
}

type TransferCoins struct {
	FromUser string
	ToUser   string
	Amount   int32
}

type IUserRepository interface {
	SendCoins(ctx context.Context, transfer *TransferCoins) error
	GetCoinsHistory(ctx context.Context, username string) (int32, *CoinsHistory, error)
}

type IUserService interface {
	SendCoins(ctx context.Context, transfer *TransferCoins) error
	GetCoinsHistory(ctx context.Context, username string) (int32, *CoinsHistory, error)
}
