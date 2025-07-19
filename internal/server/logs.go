package server

import (
	"ObservabilityServer/internal/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) createLogHandler(c echo.Context) error {
	maybeAppId := c.Get("appId")
	if maybeAppId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
	}
	appId, ok := maybeAppId.(int)
	if !ok {
		return echo.NewHTTPError(http.StatusInternalServerError, "Invalid app id")
	}

	var dto model.CreateLogDTO
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}
	if err := c.Validate(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
	}

	err := s.db.CreateLog(model.NewLogData{
		AppId:     appId,
		SessionId: dto.SessionId,
		Message:   dto.Message,
		Data:      dto.Data,
		CreatedAt: dto.CreatedAt,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Log created",
	})
}
