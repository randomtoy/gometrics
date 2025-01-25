package compress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GzipDecompress(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		if c.Request().Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Failed to decompress gzip")
			}
			defer reader.Close()

			c.Request().Body = io.NopCloser(reader)
		}
		return next(c)
	}
}
