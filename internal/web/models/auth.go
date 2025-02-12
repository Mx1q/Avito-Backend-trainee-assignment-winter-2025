package models

import "Avito-Backend-trainee-assignment-winter-2025/internal/entity"

type Auth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token,omitempty"`
}

func ToAuthEntity(auth *Auth) *entity.Auth {
	return &entity.Auth{
		Username: auth.Username,
		Password: auth.Password,
	}
}
