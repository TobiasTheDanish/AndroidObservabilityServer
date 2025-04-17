package server

import (
	"ObservabilityServer/internal/auth"
	"ObservabilityServer/internal/model"
	"fmt"
	"log"
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

	authV1 := e.Group("/api/auth", s.AuthMiddleware)

	authV1.POST("/apps", s.createAppHandler)
	authV1.POST("/apps/:id/keys", s.createKeyHandler)

	apiV1 := e.Group("/api/v1", s.APIKeyMiddleware)

	apiV1.POST("/installations", s.createInstallationHandler)
	apiV1.POST("/collection", s.createCollectionHandler)
	apiV1.POST("/sessions", s.createSessionHandler)
	apiV1.POST("/sessions/:id/crash", s.sessionCrashHandler)
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

func (s *Server) createAppHandler(c echo.Context) error {
	var appDTO model.ApplicationDTO
	if err := c.Bind(&appDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&appDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	id, err := s.db.CreateApplication(model.NewApplicationData{
		Name: appDTO.Name,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("app could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "App created",
		"id":      id,
	})
}

func (s *Server) createKeyHandler(c echo.Context) error {
	var apiKeyDTO model.NewApiKeyDTO
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiKeyDTO.AppId = id
	if err := c.Validate(&apiKeyDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	key, err := auth.GenerateApiKey()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = s.db.CreateApiKey(model.NewApiKeyData{
		Key:   auth.HashApiKey(key),
		AppId: apiKeyDTO.AppId,
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

func (s *Server) createInstallationHandler(c echo.Context) error {
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
	}

	var installationDTO model.InstallationDTO
	if err := c.Bind(&installationDTO); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body could not be parsed: %v", err),
		})
	}
	if err := c.Validate(&installationDTO); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body validation failed: %v", err),
		})
	}

	err := s.db.CreateInstallation(model.NewInstallationData{
		Id:         installationDTO.Id,
		OwnerId:    appId.(int),
		SdkVersion: installationDTO.SdkVersion,
		Model:      installationDTO.Model,
		Brand:      installationDTO.Brand,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Installation could not be created: %v", err),
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Installation created",
	})
}

func (s *Server) createCollectionHandler(c echo.Context) error {
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
	}

	var collectionData model.CollectionDTO
	if err := c.Bind(&collectionData); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body could not be parsed: %v", err),
		})
	}
	if err := c.Validate(collectionData); err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Body validation failed: %v", err),
			"path":    "session",
		})
	}

	for i, e := range collectionData.Events {
		if err := c.Validate(&e); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Body validation failed: %v", err),
				"path":    fmt.Sprintf("events[%d]", i),
			})
		}
	}

	for i, e := range collectionData.Traces {
		if err := c.Validate(&e); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Body validation failed: %v", err),
				"path":    fmt.Sprintf("traces[%d]", i),
			})
		}
	}

	go func() {
		if collectionData.Session != nil {
			sessionDTO := collectionData.Session

			err := s.db.CreateSession(model.NewSessionData{
				Id:             sessionDTO.Id,
				InstallationId: sessionDTO.InstallationId,
				AppId:          appId.(int),
				CreatedAt:      sessionDTO.CreatedAt,
				Crashed:        sessionDTO.Crashed,
			})
			if err != nil {
				log.Printf("Error creating session (%v): %v\n", sessionDTO, err)
			}
		}

		for _, e := range collectionData.Events {
			err := s.db.CreateEvent(model.NewEventData{
				Id:             e.Id,
				SessionId:      e.SessionId,
				AppId:          appId.(int),
				SerializedData: e.SerializedData,
				Type:           e.Type,
				CreatedAt:      e.CreatedAt,
			})
			if err != nil {
				log.Printf("Error creating event (%v): %v\n", e, err)
			}
		}

		for _, t := range collectionData.Traces {
			err := s.db.CreateTrace(model.NewTraceData{
				TraceId:      t.TraceId,
				SessionId:    t.SessionId,
				ParentId:     t.ParentId,
				AppId:        appId.(int),
				Name:         t.Name,
				Status:       t.Status,
				ErrorMessage: t.ErrorMessage,
				StartedAt:    t.StartedAt,
				EndedAt:      t.EndedAt,
				HasEnded:     t.HasEnded,
			})
			if err != nil {
				log.Printf("Error creating trace (%v): %v\n", t, err)
			}
		}
	}()

	return c.JSON(http.StatusAccepted, map[string]string{
		"message": "Creation of collection have been started",
	})
}

func (s *Server) createSessionHandler(c echo.Context) error {
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
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
		AppId:          appId.(int),
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

func (s *Server) sessionCrashHandler(c echo.Context) error {
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
	}

	sessionId := c.Param("id")
	err := s.db.MarkSessionCrashed(sessionId, appId.(int))
	if err != nil {
		log.Printf("Error marking session with id %s as crashed: %v\n", sessionId, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not mark session as crashed")
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Session marked as crashed"})
}

func (s *Server) createEventHandler(c echo.Context) error {
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
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
		AppId:          appId.(int),
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
	appId := c.Get("appId")
	if appId == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Missing app id")
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
		AppId:        appId.(int),
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
