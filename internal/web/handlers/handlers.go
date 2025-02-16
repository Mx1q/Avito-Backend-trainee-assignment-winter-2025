package handlers

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/jwt"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/models"
	"errors"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func errorMap(errText string) *fiber.Map {
	return &fiber.Map{
		"errors": errText,
	}
}

func AuthHandler(app *app.App) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const prompt = "Authorization"
		var req models.Auth
		err := ctx.BodyParser(&req)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, errs.InvalidData.Error())))
		}

		ua := models.ToAuthEntity(&req)
		token, err := app.AuthService.Auth(ctx.Context(), ua)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			} else if errors.Is(err, errs.InvalidCredentials) {
				return ctx.Status(fiber.StatusUnauthorized).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			} else {
				return ctx.Status(fiber.StatusInternalServerError).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			}
		}

		return ctx.Status(fiber.StatusOK).JSON(models.AuthResponse{Token: token})
	}
}

func BuyItemHandler(app *app.App) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const prompt = "Buying item"

		itemName := ctx.Params("item")
		username, err := jwt.FGetStringClaimFromJWT(ctx, "sub")
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, "invalid token")))
		}

		purchase := &entity.Purchase{
			Username: username,
			ItemName: itemName,
		}
		err = app.ItemService.BuyItem(ctx.Context(), purchase)
		if err != nil {
			if errors.Is(err, errs.ItemNotFound) || errors.Is(err, errs.UserNotFound) ||
				errors.Is(err, errs.NotEnoughCoins) {
				return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

func SendCoinsHandler(app *app.App) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const prompt = "Sending coins"

		fromUser, err := jwt.FGetStringClaimFromJWT(ctx, "sub")
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, "invalid token")))
		}

		var req models.CoinsTransfer
		err = ctx.BodyParser(&req)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, errs.InvalidData.Error())))
		}

		transfer := &entity.TransferCoins{
			FromUser: fromUser,
			ToUser:   req.ToUser,
			Amount:   req.Amount,
		}
		err = app.UserService.SendCoins(ctx.Context(), transfer)
		if err != nil {
			log.Println(transfer, err)

			if errors.Is(err, errs.InvalidData) || errors.Is(err, errs.NotEnoughCoins) ||
				errors.Is(err, errs.UserNotFound) {
				return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
		}

		return ctx.SendStatus(fiber.StatusOK)
	}
}

func GetUserInfoHandler(app *app.App) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		const prompt = "Getting user info"

		username, err := jwt.FGetStringClaimFromJWT(ctx, "sub")
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, "invalid token")))
		}

		items, err := app.ItemService.GetInventory(ctx.Context(), username)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
		}

		coins, coinHistory, err := app.UserService.GetCoinsHistory(ctx.Context(), username)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				return ctx.Status(fiber.StatusBadRequest).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
			}
			return ctx.Status(fiber.StatusInternalServerError).JSON(errorMap(fmt.Sprintf("%s: %s", prompt, err.Error())))
		}

		return ctx.Status(fiber.StatusOK).JSON(models.InfoResponse{
			Coins:       coins,
			Inventory:   models.ToInventoryTransport(items),
			CoinHistory: models.ToCoinsHistoryTransport(coinHistory),
		})
	}
}
