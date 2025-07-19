package server

import (
	"ObservabilityServer/internal/model"
	"log"
	"net/http"
	"strconv"

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

func (s *Server) getLatestSessionLogsHandler(c echo.Context) error {
	sessionId := c.Param("id")
	session, err := s.db.GetSession(sessionId)
	if err != nil {
		log.Printf("Getting session failed: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown session id")
	}
	app, err := s.db.GetApplication(session.AppId)
	if err != nil {
		log.Printf("Getting app failed: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown session id")
	}
	authSession := c.Get("session").(model.AuthSessionEntity)
	if !s.db.ValidateTeamUserLink(app.TeamId, authSession.UserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Access denied to this app")
	}

	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.QueryParam("pageSize"))
	if err != nil || pageSize > 100 {
		pageSize = 100
	}
	if pageSize <= 0 {
		pageSize = 1
	}

	ents, err := s.db.GetLatestLogsBySessionId(sessionId, pageSize, page-1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": err.Error(),
		})
	}

	DTOS := make([]model.GetLogDTO, len(ents))
	for i, ent := range ents {
		DTOS[i] = model.GetLogDTO{
			Id:        ent.Id,
			SessionId: ent.SessionId,
			Message:   ent.Message,
			Data:      ent.Data,
			CreatedAt: ent.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Success",
		"logs":    DTOS,
	})
}
