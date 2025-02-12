package postgres

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"context"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type itemRepository struct {
	db      *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

func NewItemRepository(db *pgxpool.Pool) entity.IItemRepository {
	return &itemRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r itemRepository) BuyItem(ctx context.Context, itemName string, username string) error {
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
		Where(squirrel.Eq{"username": username}).
		Suffix("for update").
		ToSql()
	if err != nil {
		return fmt.Errorf("building getting user coins query: %w", err)
	}

	var userCoins int32
	err = tx.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&userCoins,
	)
	if err != nil {
		return fmt.Errorf("getting user \"%s\" coins: %w", username, err)
	}

	query, args, err = r.builder.Select("price").
		From("items").
		Where(squirrel.Eq{"name": itemName}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building getting item price query: %w", err)
	}

	var itemPrice int32
	err = tx.QueryRow(
		ctx,
		query,
		args...,
	).Scan(
		&itemPrice,
	)
	if err != nil {
		return fmt.Errorf("getting item \"%s\" price: %w", itemName, err)
	}

	if userCoins < itemPrice {
		err = errs.NotEnoughCoins
		return err
	}

	query, args, err = r.builder.Update("users").
		Set("coins", userCoins-itemPrice).
		Where(squirrel.Eq{"username": username}).
		ToSql()
	if err != nil {
		return fmt.Errorf("building updating user coins query: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("updating user \"%s\" coins: %w", username, err)
	}

	query, args, err = r.builder.Insert("purchases").
		Columns("username", "item").
		Values(username, itemName).
		ToSql()
	if err != nil {
		return fmt.Errorf("building creating purchase query: %w", err)
	}

	_, err = tx.Exec(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return fmt.Errorf("creating user \"%s\" item \"%s\" purchase: %w",
			username, itemName, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("user \"%s\" buing item \"%s\" (commiting transaction error): %w",
			username, itemName, err)
	}
	return nil
}
