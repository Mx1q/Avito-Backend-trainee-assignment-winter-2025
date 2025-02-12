package errs

import "fmt"

var (
	ErrNotFound        = fmt.Errorf("not found")
	InvalidData        = fmt.Errorf("invalid data")
	InternalError      = fmt.Errorf("internal error")
	InvalidCredentials = fmt.Errorf("invalid credentials")
	NotEnoughCoins     = fmt.Errorf("not enough coins")
	UserNotFound       = fmt.Errorf("user not found")
)
