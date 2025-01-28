package logger

import (
	"bytes"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type responseWriterWithBody struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (w *responseWriterWithBody) Write(data []byte) (int, error) {
	_, _ = w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func ResponseLogger(l zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			request := c.Request()

			response := c.Response()
			writer := &responseWriterWithBody{
				ResponseWriter: response.Writer,
				status:         http.StatusOK,
				body:           &bytes.Buffer{},
			}
			response.Writer = writer

			err := next(c)

			duration := time.Since(start)
			l.Sugar().Infoln(
				"uri", request.RequestURI,
				"method", request.Method,
				"status", response.Status,
				"duration", duration,
				"size", response.Size,
				"body", writer.body.String(),
			)
			return err
		}
	}

}
