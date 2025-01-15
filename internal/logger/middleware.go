package logger

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ResponseLogger(l zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			request := c.Request()

			response := c.Response()
			err := next(c)

			duration := time.Since(start)
			l.Sugar().Infoln(
				"uri", request.RequestURI,
				"method", request.Method,
				"status", response.Status,
				"duration", duration,
				"size", response.Size,
			)
			return err
		}
	}

}
