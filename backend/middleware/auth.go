package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"solicitud-permisos/models"
	"solicitud-permisos/utils"
)

func JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
				Success: false,
				Message: "Token requerido",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
				Success: false,
				Message: "Formato de token inválido",
			})
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(models.LoginResponse{
				Success: false,
				Message: "Token inválido o expirado",
			})
		}

		c.Locals("user", claims)

		return c.Next()
	}
}