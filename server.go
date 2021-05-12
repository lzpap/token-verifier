package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// IndexRequest returns INDEX
func IndexRequest(c echo.Context) error {
	return c.String(http.StatusOK, "INDEX")
}

// Server gets the server instance.
func Server() *echo.Echo {
	serverOnce.Do(func() {
		server = echo.New()
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			Skipper:      middleware.DefaultSkipper,
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))

		server.HTTPErrorHandler = func(err error, c echo.Context) {
			log.Warnf("Request failed: %s", err)

			var statusCode int
			var message string

			switch errors.Unwrap(err) {
			case echo.ErrUnauthorized:
				statusCode = http.StatusUnauthorized
				message = "unauthorized"

			case echo.ErrForbidden:
				statusCode = http.StatusForbidden
				message = "access forbidden"

			case echo.ErrInternalServerError:
				statusCode = http.StatusInternalServerError
				message = "internal server error"

			case echo.ErrNotFound:
				statusCode = http.StatusNotFound
				message = "not found"

			case echo.ErrBadRequest:
				statusCode = http.StatusBadRequest
				message = "bad request"

			default:
				statusCode = http.StatusInternalServerError
				message = "internal server error"
			}

			message = fmt.Sprintf("%s, error: %+v", message, err)
			c.String(statusCode, message)
		}
	})
	return server
}
