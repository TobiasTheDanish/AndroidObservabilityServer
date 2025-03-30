package server

import (
	"ObservabilityServer/internal/auth"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Validates the ApiKey passed via Authorization header(if any)
// and sets the ownerId of the key on the echo Context
func (s *Server) APIKeyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		apiKey := c.Request().Header.Get("Authorization")
		if apiKey == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "No authorization")
		}

		apiKey = strings.TrimPrefix(apiKey, "Bearer ")

		hashedApiKey := auth.HashApiKey(apiKey)

		if ok := s.db.ValidateApiKey(hashedApiKey); !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}

		ownerId, err := s.db.GetOwnerId(hashedApiKey)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not get info based on api key")
		}

		c.Set("ownerId", ownerId)

		return next(c)
	}
}
