package errs

import "fmt"

const (
	UniqueConstraintSQLState = "23505"
)

var (
	InvalidData        = fmt.Errorf("invalid data")
	InternalError      = fmt.Errorf("internal error")
	InvalidCredentials = fmt.Errorf("invalid credentials")
	NotEnoughCoins     = fmt.Errorf("not enough coins")
	UserNotFound       = fmt.Errorf("user not found")
	ItemNotFound       = fmt.Errorf("item not found")
	UserAlreadyExists  = fmt.Errorf("user already exists")
)
