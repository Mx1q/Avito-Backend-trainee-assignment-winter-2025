package web

import (
	"context"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
)

func GetStringClaimFromJWT(ctx context.Context, claim string) (strVal string, err error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("getting claims from JWT: %w", err)
	}

	val, ok := claims[claim]
	if !ok {
		return "", fmt.Errorf("failed getting claim \"%s\" from JWT token", claim)
	}

	strVal, ok = val.(string)
	if !ok {
		return "", fmt.Errorf("converting interface to string")
	}

	return strVal, nil
}
