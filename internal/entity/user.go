package entity

import "context"

type User struct {
	Username string
	coins    int32
}

type IUserRepository interface {
	SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error
	//GetCoinsHistory(username string) ([]User, []User, error) // received, sent
}

type IUserService interface {
	SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error
	//GetCoinsHistory(username string) ([]User, []User, error) // received, sent
}
