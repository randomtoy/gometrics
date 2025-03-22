package crypto

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func HMACSHA256Middleware(secretKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			body, err := io.ReadAll(c.Request().Body)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Ошибка чтения тела запроса"})
			}
			defer c.Request().Body.Close()

			// agentKey := ""
			if c.Request().Header.Get("HashSHA256") != "" {
				expectedHash := ComputeHMACSHA256(string(body), secretKey)
				receivedHash := ComputeHMACSHA256(string(body), c.Request().Header.Get("HashSHA256"))
				fmt.Printf("Server received: '%v'\nExpected hash: %s\nGot hash: %s\n",
					c.Request().Header, expectedHash, receivedHash)

				if receivedHash != expectedHash {
					return c.JSON(http.StatusBadRequest, echo.Map{"error": "Хеш подписи не совпадает"})
				}
			}
			c.Request().Body = io.NopCloser(bytes.NewReader(body))

			return next(c)
		}
	}
}
