package crypto

import (
	"bytes"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HMACSHA256Middleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method != http.MethodPost {
				return next(c)
			}

			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Ошибка чтения тела запроса"})
			}
			defer c.Request().Body.Close()

			receivedHash := c.Request().Header.Get("HashSHA256")
			expectedHash := ComputeHMACSHA256(string(body), secretKey)

			if receivedHash != expectedHash {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": "Хеш подписи не совпадает"})
			}

			c.Request().Body = io.NopCloser(bytes.NewReader(body))

			return next(c)
		}
	}
}
