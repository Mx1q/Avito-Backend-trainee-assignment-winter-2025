package entity

import "context"

type Pass struct {
	Password string
	IsHashed bool
}

type Auth struct {
	Username string
	//Password string
	Password Pass
}

type IAuthRepository interface {
	Auth(ctx context.Context, authInfo *Auth) error
}

type IAuthService interface {
	Auth(ctx context.Context, authInfo *Auth) error
}
