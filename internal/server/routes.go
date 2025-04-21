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
	e.Use(middleware.CORS())

	e.GET("/", s.HelloWorldHandler)
	e.GET("/health", s.healthHandler)

	// AUTH endpoints
	e.POST("/auth/register", s.createUserHandler)
	e.POST("/auth/sign-in", s.signInHandler)

	// APP v1 endpoints
	appV1 := e.Group("/app/v1", s.AppAuthMiddleware)
	appV1.GET("/teams", s.getTeamsHandler)
	appV1.POST("/teams", s.createTeamHandler)
	appV1.POST("/teams/:id/users", s.createTeamUserLinkHandler)
	appV1.POST("/apps", s.createAppHandler)
	appV1.POST("/apps/:id/keys", s.createKeyHandler)

	// Api v1 endpoints
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

func (s *Server) getTeamsHandler(c echo.Context) error {
	session := c.Get("session").(model.AuthSessionEntity)

	teams, err := s.db.GetTeamsForUser(session.UserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	teamDTOs := make([]model.GetTeamDTO, len(teams), len(teams))
	for i, team := range teams {
		teamDTOs[i] = model.GetTeamDTO{
			Id:   team.Id,
			Name: team.Name,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"message": "Success",
		"teams":   teamDTOs,
	})
}

func (s *Server) createTeamHandler(c echo.Context) error {
	var teamDTO model.CreateTeamDTO
	if err := c.Bind(&teamDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&teamDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	id, err := s.db.CreateTeam(model.NewTeamData{
		Name: teamDTO.Name,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Team created",
		"id":      id,
	})
}

func (s *Server) createTeamUserLinkHandler(c echo.Context) error {
	var linkDTO model.TeamUserLinkDTO
	if err := c.Bind(&linkDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	idParam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	linkDTO.TeamId = idParam
	if err := c.Validate(&linkDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = s.db.CreateTeamUserLink(model.NewTeamUserLinkData{
		TeamId: linkDTO.TeamId,
		UserId: linkDTO.UserId,
		Role:   linkDTO.Role,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "Link created",
	})

}

func (s *Server) createUserHandler(c echo.Context) error {
	var userDTO model.UserDTO
	if err := c.Bind(&userDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&userDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	pwHash, err := auth.HashPassword(userDTO.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	id, err := s.db.CreateUser(model.NewUserData{
		Name:         userDTO.Name,
		PasswordHash: pwHash,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"message": "User created",
		"id":      id,
	})
}

func (s *Server) signInHandler(c echo.Context) error {
	var dto model.SignInDTO
	if err := c.Bind(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&dto); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	userEntity, err := s.db.GetUserByName(dto.Username)
	if err != nil || !auth.ValidatePassword(
		dto.Password,
		userEntity.PasswordHash,
	) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
	}

	sessionId, err := auth.GenerateSessionToken()
	if err != nil {
		log.Printf("Error generating session id: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create session id")
	}
	sessionExpiry := auth.GetExpiryForSession()

	err = s.db.CreateAuthSession(model.NewAuthSessionData{
		Id:     sessionId,
		UserId: userEntity.Id,
		Expiry: sessionExpiry,
	})
	if err != nil {
		log.Printf("Error storing session id: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create session id")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message":   "Sign in successful",
		"sessionId": sessionId,
	})
}

func (s *Server) createAppHandler(c echo.Context) error {
	var appDTO model.ApplicationDTO
	if err := c.Bind(&appDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := c.Validate(&appDTO); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	session := c.Get("session").(model.AuthSessionEntity)
	if !s.db.ValidateTeamUserLink(appDTO.TeamId, session.UserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Access denied to this team")
	}

	id, err := s.db.CreateApplication(model.NewApplicationData{
		Name:   appDTO.Name,
		TeamId: appDTO.TeamId,
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

	session := c.Get("session").(model.AuthSessionEntity)
	app, err := s.db.GetApplication(apiKeyDTO.AppId)
	if err != nil {
		log.Printf("Error getting application: %v\n", err)
		return echo.NewHTTPError(http.StatusNotFound, "No application found with provided id")
	}

	if !s.db.ValidateTeamUserLink(app.TeamId, session.UserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Access denied to this app")
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
		AppId:      appId.(int),
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
