package postgres

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type authRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewAuthRepository(db *pgxpool.Pool) entity.IAuthRepository {
	return &authRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r authRepository) GetByUsername(ctx context.Context, username string) (*entity.Auth, error) {
	query, args, err := r.builder.Select("password").
		From("users").
		Where(squirrel.Eq{"username": username}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}

	authDb := new(entity.Auth)
	err = r.db.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&authDb.Password,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user by username: %w", err)
	}
	return authDb, nil
}

func (r authRepository) Register(ctx context.Context, authInfo *entity.Auth) error {
	query, args, err := r.builder.Insert("users").
		Columns("username", "password").
		Values(authInfo.Username, authInfo.Password).
		ToSql()
	if err != nil {
		return fmt.Errorf("building query: %w", err)
	}

	_, err = r.db.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}
