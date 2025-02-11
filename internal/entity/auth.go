package entity

import "context"

type Pass struct {
	Password string
	//IsHashed bool
	HashedPassword string
}

type Auth struct {
	Username string
	//Password string
	Password string
}

type IAuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*Auth, error)
	Register(ctx context.Context, authInfo *Auth) error
}

type IAuthService interface {
	Auth(ctx context.Context, authInfo *Auth) (string, error) // sing up if not exists
}
