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

func (r *itemRepository) BuyItem(ctx context.Context, purchase *entity.Purchase) error {
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
		Where(squirrel.Eq{"username": purchase.Username}).
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
		return errs.UserNotFound
	}

	query, args, err = r.builder.Select("price").
		From("items").
		Where(squirrel.Eq{"name": purchase.ItemName}).
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
		return errs.ItemNotFound
	}

	if userCoins < itemPrice {
		err = errs.NotEnoughCoins
		return err
	}

	query, args, err = r.builder.Update("users").
		Set("coins", userCoins-itemPrice).
		Where(squirrel.Eq{"username": purchase.Username}).
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
		return fmt.Errorf("updating user \"%s\" coins: %w", purchase.Username, err)
	}

	query, args, err = r.builder.Insert("purchases").
		Columns("username", "item").
		Values(purchase.Username, purchase.ItemName).
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
			purchase.Username, purchase.ItemName, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("user \"%s\" buing item \"%s\" (commiting transaction error): %w",
			purchase.Username, purchase.ItemName, err)
	}
	return nil
}

func (r *itemRepository) GetInventory(ctx context.Context, username string) ([]*entity.Item, error) {
	query, args, err := r.builder.Select("item", "count(*)").
		From("purchases").
		Where(squirrel.Eq{"username": username}).
		GroupBy("item").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("building getting user coins query: %w", err)
	}

	rows, err := r.db.Query(
		ctx,
		query,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("getting items owned by user: %w", err)
	}
	items := make([]*entity.Item, 0)
	for rows.Next() {
		tmp := new(entity.Item)
		err = rows.Scan(
			&tmp.Name,
			&tmp.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, tmp)
	}

	return items, nil
}
