package server

import (
	"ObservabilityServer/internal/auth"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var apiAuthSecret = os.Getenv("OBSERVE_API_SECRET")

// Validates the ApiKey passed via Authorization header(if any)
// and sets the appId of the key on the echo Context
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

		appId, err := s.db.GetAppId(hashedApiKey)
		if err != nil {
			log.Println(err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Could not get info based on api key")
		}

		c.Set("appId", appId)

		return next(c)
	}
}

func (s *Server) AppAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authSecret := c.Request().Header.Get("Authorization")
		if authSecret == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		if !strings.HasPrefix(authSecret, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header. Expected prefix \"Bearer\"")
		}

		authSecret = strings.TrimPrefix(authSecret, "Bearer ")

		session, err := s.db.GetAuthSession(authSecret)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		if authSecret != session.Id {
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		nowMillis := time.Now().UnixMilli()
		if nowMillis >= session.Expiry {
			s.db.DeleteAuthSession(session.Id)
			return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
		}

		c.Set("session", session)

		return next(c)
	}
}
