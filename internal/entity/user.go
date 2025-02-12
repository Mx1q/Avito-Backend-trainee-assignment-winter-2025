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

type IUserRepository interface {
	SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error
	GetCoinsHistory(ctx context.Context, username string) (int32, *CoinsHistory, error)
}

type IUserService interface {
	SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error
	GetCoinsHistory(ctx context.Context, username string) (int32, *CoinsHistory, error)
}
