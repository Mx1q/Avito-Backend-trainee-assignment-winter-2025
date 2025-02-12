package postgres

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool) entity.IUserRepository {
	return &userRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *userRepository) SendCoins(ctx context.Context, fromUser string, toUser string, amount int32) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}
	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				err = fmt.Errorf("%v: %w", rollbackErr, err)
			}
		}
	}()

	query, args, err := r.builder.Select("coins").
		From("users").
		Where(squirrel.Or{
			squirrel.Eq{"username": fromUser},
			squirrel.Eq{"username": toUser},
		}).
		OrderBy("username").
		Suffix("for update").
		ToSql()
	if err != nil {
		return fmt.Errorf("building getting user coins query: %w", err)
	}

	usersCoins := make([]int32, 2) // количество монет каждого двух пользователей (order by username)
	index := 0
	rows, err := tx.Query(
		ctx,
		query,
		args...,
	)
	for rows.Next() {
		var tmp int32
		err = rows.Scan(&tmp)
		if err != nil {
			return fmt.Errorf("getting user coins: %w", err)
		}
		usersCoins[index] = tmp
		index++
	}
	if index != 2 {
		err = errs.UserNotFound
		return err
	}

	if fromUser > toUser && usersCoins[1] < amount ||
		fromUser < toUser && usersCoins[0] < amount {
		err = errs.NotEnoughCoins
		return err
	}

	query, args, err = r.builder.Update("users").
		Set("coins", squirrel.Expr("coins - ?", amount)).
		Where(squirrel.Eq{"username": fromUser}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building getting user coins query: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("decrementing user \"%s\" coins: %w", fromUser, err)
	}

	query, args, err = r.builder.Update("users").
		Set("coins", squirrel.Expr("coins + ?", amount)).
		Where(squirrel.Eq{"username": toUser}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building getting user coins query: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("incrementing user \"%s\" coins: %w", toUser, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commiting transaction error: %w", err)
	}
	return nil
}
