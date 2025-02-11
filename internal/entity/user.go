package entity

type User struct {
	Username string
	coins    int32
}

type IUserRepository interface {
	SendCoins(username string, amount int32) error
	GetCoinsHistory(username string) ([]User, []User, error) // received, sent
}

type IUserService interface {
	SendCoins(username string, amount int32) error
	GetCoinsHistory(username string) ([]User, []User, error) // received, sent
}
