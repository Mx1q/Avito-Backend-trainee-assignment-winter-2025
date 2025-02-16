package jwt

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func FGetStringClaimFromJWT(ctx *fiber.Ctx, claim string) (string, error) {
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	strVal := claims[claim].(string)
	var err error
	if strVal == "" {
		err = fmt.Errorf("empty claim")
	}
	return strVal, err
}
