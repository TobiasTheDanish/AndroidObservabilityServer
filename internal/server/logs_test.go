package server

import (
	"ObservabilityServer/internal/database"
	"ObservabilityServer/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestServer_createLogHandler(t *testing.T) {
	teardown, config, err := database.SetupTestDatabase("public")
	if err != nil {
		t.Fatalf("could not start postgres container: %v", err)
	}

	db := database.New(config)
	teamId, err := db.CreateTeam(model.NewTeamData{Name: "Test team"})
	if err != nil {
		t.Fatalf("Could not create test team: %v", err)
	}
	appId, err = db.CreateApplication(model.NewApplicationData{
		Name:   "Test owner",
		TeamId: teamId,
	})
	if err != nil {
		t.Fatalf("Could not create test application: %v", err)
	}
	sessionId := "6d40d812-7888-4fd1-98bf-ee92c9be1894"
	err = db.CreateSession(model.NewSessionData{
		Id:             sessionId,
		InstallationId: "1234",
		AppId:          appId,
		CreatedAt:      171234503,
		Crashed:        false,
	})

	s := &Server{
		db: db,
	}
	e := echo.New()
	e.Validator = NewValidator()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cb       func() (echo.Context, *httptest.ResponseRecorder)
		code     int
		expected map[string]any
	}{
		{
			name: "Test happy path",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				logData := model.CreateLogDTO{
					SessionId: sessionId,
					Message:   "Hello world",
					Data:      map[string]any{},
					CreatedAt: 171234553,
				}

				body, err := json.Marshal(logData)
				if err != nil {
					t.Fatalf("Could not marshal collectionDTO: %v", err)
				}
				reader := bytes.NewReader(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android", reader)
				resp := httptest.NewRecorder()
				req.Header.Set("Content-type", "application/json")

				c := e.NewContext(req, resp)

				c.Set("appId", appId)
				return c, resp
			},
			code:     http.StatusCreated,
			expected: map[string]any{"message": "Log created"},
		},
		{
			name: "Test unknown session id",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				logData := model.CreateLogDTO{
					SessionId: "7d40d812-7888-4fd1-98bf-ee92c9be1894",
					Message:   "Hello world",
					Data:      map[string]any{},
					CreatedAt: 171234553,
				}

				body, err := json.Marshal(logData)
				if err != nil {
					t.Fatalf("Could not marshal collectionDTO: %v", err)
				}
				reader := bytes.NewReader(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android", reader)
				resp := httptest.NewRecorder()
				req.Header.Set("Content-type", "application/json")

				c := e.NewContext(req, resp)

				c.Set("appId", appId)
				return c, resp
			},
			code:     http.StatusInternalServerError,
			expected: map[string]any{"message": "ERROR: insert or update on table \"ob_logs\" violates foreign key constraint \"ob_logs_session_id_fkey\" (SQLSTATE 23503)"},
		},
		{
			name: "Test non uuid session id",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				logData := model.CreateLogDTO{
					SessionId: "7d40d812-4fd1-98bf-ee92c9be1894",
					Message:   "Hello world",
					Data:      map[string]any{},
					CreatedAt: 171234553,
				}

				body, err := json.Marshal(logData)
				if err != nil {
					t.Fatalf("Could not marshal collectionDTO: %v", err)
				}
				reader := bytes.NewReader(body)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android", reader)
				resp := httptest.NewRecorder()
				req.Header.Set("Content-type", "application/json")

				c := e.NewContext(req, resp)

				c.Set("appId", appId)
				return c, resp
			},
			code:     http.StatusBadRequest,
			expected: map[string]any{"message": "Key: 'CreateLogDTO.SessionId' Error:Field validation for 'SessionId' failed on the 'uuid' tag"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, resp := tt.cb()
			gotErr := s.createLogHandler(c)
			if gotErr != nil {
				t.Errorf("createLogHandler() failed: %v", gotErr)
				return
			}

			var actual map[string]any
			// Decode the response body into the actual map
			if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
				t.Fatalf("createLogHandler() error decoding response body: %v", err)
			}
			if resp.Code != tt.code {
				t.Fatalf("createLogHandler() wrong status code = %v, expected = %v, body = %#v", resp.Code, tt.code, actual)
			}

			// Compare the decoded response with the expected value
			if !reflect.DeepEqual(tt.expected, actual) {
				t.Fatalf("createLogHandler() wrong response body. expected = %v, actual = %v", tt.expected, actual)
			}
		})
	}

	if teardown != nil && teardown(context.Background()) != nil {
		t.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestServer_getLatestSessionLogsHandler(t *testing.T) {
	teardown, config, err := database.SetupTestDatabase("public")
	if err != nil {
		t.Fatalf("could not start postgres container: %v", err)
	}

	db := database.New(config)
	userId, err := db.CreateUser(model.NewUserData{Name: "Test user"})
	if err != nil {
		t.Fatalf("Could not create test user: %v", err)
	}
	teamId, err := db.CreateTeam(model.NewTeamData{Name: "Test team"})
	if err != nil {
		t.Fatalf("Could not create test team: %v", err)
	}
	db.CreateTeamUserLink(model.NewTeamUserLinkData{
		TeamId: teamId,
		UserId: userId,
		Role:   "owner",
	})
	appId, err = db.CreateApplication(model.NewApplicationData{
		Name:   "Test app",
		TeamId: teamId,
	})
	if err != nil {
		t.Fatalf("Could not create test application: %v", err)
	}
	sessionId := "6d40d812-7888-4fd1-98bf-ee92c9be1894"
	err = db.CreateSession(model.NewSessionData{
		Id:             sessionId,
		InstallationId: "1234",
		AppId:          appId,
		CreatedAt:      171234503,
		Crashed:        false,
	})
	if err != nil {
		t.Fatalf("Could not create test application: %v", err)
	}

	for range 1000 {
		data := model.NewLogData{
			AppId:     appId,
			SessionId: sessionId,
			Message:   "Hello world",
			Data:      map[string]any{},
			CreatedAt: 12345567755,
		}
		if err = db.CreateLog(data); err != nil {
			t.Fatalf("Could not create test application: %v", err)
		}
	}

	s := &Server{
		db: db,
	}
	e := echo.New()
	e.Validator = NewValidator()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		cb               func() (echo.Context, *httptest.ResponseRecorder)
		code             int
		expectedMessage  string
		expectedLogCount int
	}{
		{
			name: "Test happy path",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android?pageSize=5", nil)
				resp := httptest.NewRecorder()

				c := e.NewContext(req, resp)

				authSession := model.AuthSessionEntity{
					Id:     "secret",
					UserId: userId,
					Expiry: 20000000000,
				}
				c.Set("session", authSession)
				c.SetParamNames("id")
				c.SetParamValues(sessionId)
				return c, resp
			},
			code:             http.StatusOK,
			expectedMessage:  "Success",
			expectedLogCount: 5,
		},
		{
			name: "Test second page",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android?pageSize=5&page=2", nil)
				resp := httptest.NewRecorder()

				c := e.NewContext(req, resp)

				authSession := model.AuthSessionEntity{
					Id:     "secret",
					UserId: userId,
					Expiry: 20000000000,
				}
				c.Set("session", authSession)
				c.SetParamNames("id")
				c.SetParamValues(sessionId)
				return c, resp
			},
			code:             http.StatusOK,
			expectedMessage:  "Success",
			expectedLogCount: 5,
		},
		{
			name: "Test page < 0",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android?pageSize=5&page=-1", nil)
				resp := httptest.NewRecorder()

				c := e.NewContext(req, resp)

				authSession := model.AuthSessionEntity{
					Id:     "secret",
					UserId: userId,
					Expiry: 20000000000,
				}
				c.Set("session", authSession)
				c.SetParamNames("id")
				c.SetParamValues(sessionId)
				return c, resp
			},
			code:             http.StatusOK,
			expectedMessage:  "Success",
			expectedLogCount: 5,
		},
		{
			name: "Test page * size > len",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android?pageSize=10&page=101", nil)
				resp := httptest.NewRecorder()

				c := e.NewContext(req, resp)

				authSession := model.AuthSessionEntity{
					Id:     "secret",
					UserId: userId,
					Expiry: 20000000000,
				}
				c.Set("session", authSession)
				c.SetParamNames("id")
				c.SetParamValues(sessionId)
				return c, resp
			},
			code:             http.StatusOK,
			expectedMessage:  "Success",
			expectedLogCount: 0,
		},
		{
			name: "Test pageSize > 100",
			cb: func() (echo.Context, *httptest.ResponseRecorder) {
				req := httptest.NewRequest(http.MethodPost, "/api/v1/installations/android?pageSize=1000", nil)
				resp := httptest.NewRecorder()

				c := e.NewContext(req, resp)

				authSession := model.AuthSessionEntity{
					Id:     "secret",
					UserId: userId,
					Expiry: 20000000000,
				}
				c.Set("session", authSession)
				c.SetParamNames("id")
				c.SetParamValues(sessionId)
				return c, resp
			},
			code:             http.StatusOK,
			expectedMessage:  "Success",
			expectedLogCount: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, resp := tt.cb()
			gotErr := s.getLatestSessionLogsHandler(c)
			if gotErr != nil {
				t.Errorf("getLatestSessionLogsHandler() failed: %v", gotErr)
				return
			}

			var actual map[string]any
			// Decode the response body into the actual map
			if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
				t.Fatalf("getLatestSessionLogsHandler() error decoding response body: %v", err)
			}
			if resp.Code != tt.code {
				t.Fatalf("getLatestSessionLogsHandler() wrong status code = %v, expected = %v, body = %#v", resp.Code, tt.code, actual)
			}

			// Compare the decoded response with the expected value
			if !reflect.DeepEqual(tt.expectedMessage, actual["message"]) {
				t.Fatalf("getLatestSessionLogsHandler() wrong response body. expected = %v, actual = %v", tt.expectedMessage, actual)
			}
			if resp.Code == http.StatusOK {
				logsJson := actual["logs"]
				logs := logsJson.([]interface{})

				if len(logs) != tt.expectedLogCount {
					t.Fatalf("getLatestSessionLogsHandler() = Expected %v logs, but got %v logs", tt.expectedLogCount, len(logs))
				}
			}
		})
	}
	s.db.Close()

	if teardown != nil && teardown(context.Background()) != nil {
		t.Fatalf("could not teardown postgres container: %v", err)
	}
}
