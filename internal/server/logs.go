package server

import (
	"ObservabilityServer/internal/database"
	"ObservabilityServer/internal/model"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

var (
	logKeyMap map[string]string = map[string]string{
		"id":        "id",
		"appId":     "app_id",
		"sessionId": "session_id",
		"message":   "message",
		"data":      "data",
		"createdAt": "created_at",
	}
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
	} else if pageSize <= 0 {
		pageSize = 1
	}

	filters := parseLogQuerySearchParams(c.QueryParams())

	ents, err := s.db.GetLatestLogsBySessionId(sessionId, pageSize, pageSize*(page-1), filters)
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

/*
 * Expected format of query param: key=value[,op]
 *
 * Valid keys are found in the logKeyMap.
 *
 * Valid operations are:
 * eq (default)
 * contains
 * starts
 * ends
 * greater
 * less
 */
func parseLogQuerySearchParams(q url.Values) []database.LogFilter {
	filters := make([]database.LogFilter, 0)

	for k, values := range q {
		key, ok := logKeyMap[k]
		if !ok || len(values) == 0 {
			continue
		}

		val := values[0]
		value, op, ok := strings.Cut(val, ",")
		if !ok {
			op = "eq"
		}

		filter := database.LogFilter{
			Key:      key,
			Operator: op,
			Value:    sanitizeFilterValue(value, op),
		}
		filters = append(filters, filter)
	}

	return filters
}

func sanitizeFilterValue(val, op string) string {
	var sb strings.Builder

	for _, c := range val {
		switch c {
		case '\\':
			sb.WriteRune('\\')

		case '%', '_':
			switch op {
			case "contains", "starts", "ends":
				sb.WriteRune('\\')
			}
		}

		sb.WriteRune(c)
	}

	return sb.String()
}
