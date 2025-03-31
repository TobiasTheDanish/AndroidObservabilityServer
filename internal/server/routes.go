package server

import (
	"ObservabilityServer/internal/auth"
	"ObservabilityServer/internal/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Validator = NewValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", s.HelloWorldHandler)

	e.GET("/health", s.healthHandler)

	authV1 := e.Group("/api/auth")

	authV1.POST("/owners", s.createOwnerHandler)
	authV1.POST("/owners/:id/keys", s.createKeyHandler)

	apiV1 := e.Group("/api/v1", s.APIKeyMiddleware)

	apiV1.POST("/sessions", s.createSessionHandler)
	apiV1.POST("/events", s.createEventHandler)
	apiV1.POST("/traces", s.createTraceHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) createOwnerHandler(c echo.Context) error {
	var ownerDTO model.OwnerDTO
	if err := c.Bind(&ownerDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&ownerDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	id, err := s.db.CreateOwner(model.NewOwnerData{
		Name: ownerDTO.Name,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Owner could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Owner created",
		"id":      id,
	})
}

func (s *Server) createKeyHandler(c echo.Context) error {
	var apiKeyDTO model.NewApiKeyDTO
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiKeyDTO.OwnerId = id
	if err := c.Validate(&apiKeyDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	key, err := auth.GenerateApiKey()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = s.db.CreateApiKey(model.NewApiKeyData{
		Key:     auth.HashApiKey(key),
		OwnerId: apiKeyDTO.OwnerId,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("ApiKey could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Api Key created",
		"key":     key,
	})
}

func (s *Server) createSessionHandler(c echo.Context) error {
	ownerId := c.Get("ownerId")
	if ownerId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing owner id")
	}

	var sessionData model.SessionDTO
	if err := c.Bind(&sessionData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body could not be parsed: %v", err),
		})
	}
	if err := c.Validate(&sessionData); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := s.db.CreateSession(model.NewSessionData{
		Id:             sessionData.Id,
		InstallationId: sessionData.InstallationId,
		OwnerId:        ownerId.(int),
		CreatedAt:      sessionData.CreatedAt,
		Crashed:        sessionData.Crashed,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Session could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Session created",
	})
}

func (s *Server) createEventHandler(c echo.Context) error {
	ownerId := c.Get("ownerId")
	if ownerId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing owner id")
	}

	var dto model.EventDTO
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body could not be parsed: %v", err),
		})
	}
	if err := c.Validate(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := s.db.CreateEvent(model.NewEventData{
		Id:             dto.Id,
		SessionId:      dto.SessionId,
		OwnerId:        ownerId.(int),
		Type:           dto.Type,
		SerializedData: dto.SerializedData,
		CreatedAt:      dto.CreatedAt,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Event could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Event created",
	})
}

func (s *Server) createTraceHandler(c echo.Context) error {
	ownerId := c.Get("ownerId")
	if ownerId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing owner id")
	}

	data := new(model.TraceDTO)
	if err := c.Bind(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := s.db.CreateTrace(model.NewTraceData{
		TraceId:      data.TraceId,
		SessionId:    data.SessionId,
		GroupId:      data.GroupId,
		ParentId:     data.ParentId,
		OwnerId:      ownerId.(int),
		Name:         data.Name,
		Status:       data.Status,
		ErrorMessage: data.ErrorMessage,
		StartedAt:    data.StartedAt,
		EndedAt:      data.EndedAt,
		HasEnded:     data.HasEnded,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Trace could not be created: %v", err))
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Trace created",
	})
}
