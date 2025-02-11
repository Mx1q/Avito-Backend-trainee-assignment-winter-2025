package web

import (
	"Avito-Backend-trainee-assignment-winter-2025/internal/app"
	"Avito-Backend-trainee-assignment-winter-2025/internal/entity"
	"Avito-Backend-trainee-assignment-winter-2025/internal/pkg/jwt"
	"Avito-Backend-trainee-assignment-winter-2025/internal/web/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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

		ua := &entity.Auth{Username: req.Username, Password: req.Password}
		token, err := app.AuthService.Auth(r.Context(), ua)
		if err != nil {
			errorResponse(w, fmt.Errorf("%s: %w", prompt, err).Error(), http.StatusUnauthorized)
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
			Expires: time.Now().Add(3600 * 24 * time.Second),
		}
		http.SetCookie(w, &cookie)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.AuthResponse{Token: token})
	}
}
