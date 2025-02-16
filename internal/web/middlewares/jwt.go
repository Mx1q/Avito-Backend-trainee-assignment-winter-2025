package middlewares

import (
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func JwtMiddleware(jwtKey string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return jwtware.New(jwtware.Config{
			SigningKey: jwtware.SigningKey{Key: []byte(jwtKey)},
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"errors": err.Error(),
				})
			},
		})(ctx)
	}
}
