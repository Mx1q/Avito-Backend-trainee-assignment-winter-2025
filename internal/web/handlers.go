package web

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	errs "Avito-Backend-trainee-assignment-winter-2025/internal/pkg/errors"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/jwt"
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
		const prompt = "authorization"
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

		_, err = jwt.VerifyToken(token, app.Config.Jwt.Key)
		if err != nil {
			errorResponse(w, "Invalid token", http.StatusUnauthorized)
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
		prompt := "Buying item"

		itemName := chi.URLParam(r, "item")
		username, err := GetStringClaimFromJWT(r.Context(), "sub")
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
		}

		err = app.ItemService.BuyItem(r.Context(), itemName, username)
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %s", prompt,
				http.StatusText(http.StatusInternalServerError)).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
