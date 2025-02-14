package handlers

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/jwt"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

const (
	TokenExpirationMinutes = 60 * 24
)

type ErrorResponse struct {
	Error string `json:"errors,omitempty"`
}

func errorResponse(w http.ResponseWriter, err string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: err})
}

func AuthHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prompt = "Authorization"
		var req models.Auth
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %s", prompt, "invalid data").Error(), http.StatusBadRequest)
			return
		}

		ua := models.ToAuthEntity(&req)
		token, err := app.AuthService.Auth(r.Context(), ua)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusBadRequest)
			} else if errors.Is(err, errs.InvalidCredentials) {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
			} else {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusInternalServerError)
			}
			return
		}

		cookie := http.Cookie{
			Name:    "access_token",
			Value:   token,
			Path:    "/",
			Secure:  true,
			Expires: time.Now().Add(TokenExpirationMinutes * time.Minute),
		}
		http.SetCookie(w, &cookie)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.AuthResponse{Token: token})
	}
}

func BuyItemHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prompt = "Buying item"

		itemName := chi.URLParam(r, "item")
		username, err := jwt.GetStringClaimFromJWT(r.Context(), "sub")
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
		}

		purchase := entity.Purchase{
			Username: username,
			ItemName: itemName,
		}
		err = app.ItemService.BuyItem(r.Context(), &purchase)
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %s", prompt,
				http.StatusText(http.StatusInternalServerError)).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func SendCoinsHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prompt = "Sending coins"

		fromUser, err := jwt.GetStringClaimFromJWT(r.Context(), "sub")
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
		}

		var req models.CoinsTransfer
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %s", prompt, "invalid data").Error(), http.StatusBadRequest)
			return
		}

		transfer := &entity.TransferCoins{
			FromUser: fromUser,
			ToUser:   req.ToUser,
			Amount:   req.Amount,
		}
		err = app.UserService.SendCoins(r.Context(), transfer)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusBadRequest)
			} else {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func GetUserInfoHandler(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const prompt = "Getting user info"

		username, err := jwt.GetStringClaimFromJWT(r.Context(), "sub")
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
		}

		items, err := app.ItemService.GetInventory(r.Context(), username)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusBadRequest)
			} else {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusInternalServerError)
			}
			return
		}

		coins, coinHistory, err := app.UserService.GetCoinsHistory(r.Context(), username)
		if err != nil {
			if errors.Is(err, errs.InvalidData) {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusBadRequest)
			} else {
				errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusInternalServerError)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.Info{
			Coins:       coins,
			Inventory:   models.ToInventoryTransport(items),
			CoinHistory: models.ToCoinsHistoryTransport(coinHistory),
		})
	}
}
