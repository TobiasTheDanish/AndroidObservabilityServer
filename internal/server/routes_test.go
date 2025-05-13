package server

import (
	"ObservabilityServer/internal/database"
	"ObservabilityServer/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
)

var (
	db    database.Service
	appId int
)

func TestMain(m *testing.M) {
	teardown, err := database.SetupTestDatabase()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	db = database.New()

	teamId, err := db.CreateTeam(model.NewTeamData{Name: "Test team"})
	if err != nil {
		log.Fatalf("Could not create test team: %v", err)
	}
	appId, err = db.CreateApplication(model.NewApplicationData{
		Name:   "Test owner",
		TeamId: teamId,
	})
	if err != nil {
		log.Fatalf("Could not create test application: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}
	// Assertions
	if err := s.HelloWorldHandler(c); err != nil {
		t.Errorf("handler() error = %v", err)
		return
	}
	if resp.Code != http.StatusOK {
		t.Errorf("handler() wrong status code = %v", resp.Code)
		return
	}
	expected := map[string]string{"message": "Hello World"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("handler() error decoding response body: %v", err)
		return
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("handler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}

func TestTeamUserAuth(t *testing.T) {
	s := &Server{
		db: db,
	}
	e := echo.New()
	e.Validator = NewValidator()

	teamData := model.CreateTeamDTO{
		Name: "Test Team",
	}
	body, err := json.Marshal(teamData)
	if err != nil {
		t.Fatalf("Could not marshal Team: %v", err)
	}
	reader := bytes.NewReader(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	c := e.NewContext(req, resp)

	err = s.createTeamHandler(c)
	if err != nil {
		t.Errorf("createTeamHandler() error = %v", err)
		return
	}
	expected := map[string]string{
		"message": "Team created",
	}
	var teamActual map[string]any
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&teamActual); err != nil {
		t.Errorf("createTeamHandler() error decoding response body: %v", err)
		return
	}
	if resp.Code != http.StatusCreated {
		t.Errorf("createTeamHandler() wrong status code = %v", resp.Code)
		return
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected["message"], teamActual["message"]) {
		t.Errorf("createTeamHandler() wrong response body. expected = %v, actual = %v", expected, teamActual)
		return
	}

	userData := model.UserDTO{
		Name:     "Test user",
		Password: "abc1234",
	}
	body, err = json.Marshal(userData)
	if err != nil {
		t.Fatalf("Could not marshal User: %v", err)
	}
	reader = bytes.NewReader(body)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", reader)
	resp = httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	c = e.NewContext(req, resp)

	err = s.createUserHandler(c)
	if err != nil {
		t.Errorf("createUserHandler() error = %v", err)
		return
	}
	expected = map[string]string{"message": "User created", "id": "1"}
	var userActual map[string]any
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&userActual); err != nil {
		t.Errorf("createUserHandler() error decoding response body: %v", err)
		return
	}
	if resp.Code != http.StatusCreated {
		t.Errorf("createUserHandler() wrong status code = %v", resp.Code)
		return
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected["message"], userActual["message"]) {
		t.Errorf("createUserHandler() wrong response body. expected = %v, actual = %v", expected, userActual)
		return
	}

	linkData := model.TeamUserLinkDTO{
		UserId: int(userActual["id"].(float64)),
		Role:   "Owner",
	}
	body, err = json.Marshal(linkData)
	if err != nil {
		t.Fatalf("Could not marshal Link: %v", err)
	}
	reader = bytes.NewReader(body)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/teams/%d/users", int(teamActual["id"].(float64))), reader)
	resp = httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	c = e.NewContext(req, resp)
	c.SetParamNames("id")
	c.SetParamValues(strconv.Itoa(int(teamActual["id"].(float64))))

	err = s.createTeamUserLinkHandler(c)
	if err != nil {
		t.Errorf("createTeamUserLinkHandler() error = %v", err)
		return
	}
	linkExpected := map[string]string{"message": "Link created"}
	var linkActual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&linkActual); err != nil {
		t.Errorf("createTeamUserLinkHandler() error decoding response body: %v", err)
		return
	}
	if resp.Code != http.StatusCreated {
		t.Errorf("createTeamUserLinkHandler() wrong status code = %v", resp.Code)
		return
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(linkExpected, linkActual) {
		t.Errorf("createTeamUserLinkHandler() wrong response body. expected = %v, actual = %v", linkExpected, linkActual)
		return
	}
}

func TestCreateCollection(t *testing.T) {
	collection := model.CollectionDTO{
		Session: nil,
		Events:  make([]model.EventDTO, 0, 0),
		Traces:  make([]model.TraceDTO, 0, 0),
	}
	body, err := json.Marshal(collection)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)

	err = s.createCollectionHandler(c)
	if err != nil {
		t.Errorf("createCollectionHandler() error = %v", err)
		return
	}
	expected := map[string]string{"message": "Creation of collection have been started"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
		return
	}

	t.Logf("Response body = %v\n", actual)

	if resp.Code != http.StatusAccepted {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("createCollectionHandler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}

func TestCreateCollectionInvalidSession(t *testing.T) {
	collection := model.CollectionDTO{
		Session: &model.SessionDTO{
			Id:             "1234",
			InstallationId: "1234",
			CreatedAt:      17000000,
			Crashed:        false,
		},
		Events: make([]model.EventDTO, 0, 0),
		Traces: make([]model.TraceDTO, 0, 0),
	}
	body, err := json.Marshal(collection)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)

	err = s.createCollectionHandler(c)
	if err != nil {
		t.Errorf("createCollectionHandler() error = %v", err)
		return
	}
	expected := map[string]string{
		"message": "Body validation failed: Key: 'CollectionDTO.Session.Id' Error:Field validation for 'Id' failed on the 'uuid' tag",
		"path":    "session",
	}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
		return
	}

	if resp.Code != http.StatusBadRequest {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("createCollectionHandler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}

func TestCreateCollectionInvalidEvent(t *testing.T) {
	collection := model.CollectionDTO{
		Session: &model.SessionDTO{
			Id:             "f9bc714c-bb1a-4c47-b2a5-9b46b8568259",
			InstallationId: "1234",
			CreatedAt:      17000000,
			Crashed:        false,
		},
		Events: []model.EventDTO{
			{
				Id:             "55a7a1f2-8caa-4d13-b792-0d9e533f8705",
				SessionId:      "f9bc714c-bb1a-4c47-b2a5-9b46b8568259",
				Type:           "lifecycle_event",
				SerializedData: "{\"message\":\"hello\"}",
				CreatedAt:      1743194265,
			},
			{
				Id:             "55a7a1f2",
				SessionId:      "f9bc714c-bb1a-4c47-b2a5-9b46b8568259",
				Type:           "lifecycle_event",
				SerializedData: "{\"message\":\"hello\"}",
				CreatedAt:      1743194265,
			},
		},
		Traces: make([]model.TraceDTO, 0, 0),
	}
	body, err := json.Marshal(collection)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)

	err = s.createCollectionHandler(c)
	if err != nil {
		t.Errorf("createCollectionHandler() error = %v", err)
		return
	}
	expected := map[string]string{
		"message": "Body validation failed: Key: 'EventDTO.Id' Error:Field validation for 'Id' failed on the 'uuid' tag",
		"path":    "events[1]",
	}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
		return
	}

	t.Logf("body = %v\n", actual)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("createCollectionHandler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}

func TestCreateCollectionInvalidTrace(t *testing.T) {
	collection := model.CollectionDTO{
		Session: nil,
		Events:  []model.EventDTO{},
		Traces: []model.TraceDTO{
			{
				TraceId:   "2bea6cb8-b784-470f-969f-ef45d38c6aa7",
				SessionId: "f9bc714c-bb1a-4c47-b2a5-9b46b8568259",
				GroupId:   "a829bff3-7d17-4e88-b360-4279fbe4fe4c",
				Name:      "TestParentTrace",
				Status:    "Ok",
				StartedAt: 1743194275,
				EndedAt:   1743195275,
				HasEnded:  true,
			},
			{
				TraceId:   "2bea6cb8-b784-470f-969f-ef45d38c6aa7",
				SessionId: "f9bc714c-bb1a-4c47-b2a5-9b46b8568259",
				GroupId:   "a829bff3-7d17-4e88-b360-4279fbe4fe4c",
				Name:      "TestParentTrace",
				Status:    "Error",
				StartedAt: 1743194275,
				EndedAt:   1743195275,
				HasEnded:  true,
			},
		},
	}
	body, err := json.Marshal(collection)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)

	err = s.createCollectionHandler(c)
	if err != nil {
		t.Errorf("createCollectionHandler() error = %v", err)
		return
	}
	expected := map[string]string{
		"message": "Body validation failed: Key: 'TraceDTO.ErrorMessage' Error:Field validation for 'ErrorMessage' failed on the 'required_if' tag",
		"path":    "traces[1]",
	}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
		return
	}

	t.Logf("body = %v\n", actual)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("createCollectionHandler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}

func TestCreateInstallation(t *testing.T) {
	installation := model.InstallationDTO{
		Id:         "6d40d812-7888-4fd1-98bf-ee92c9be1891",
		SdkVersion: 32,
		Model:      "s32",
		Brand:      "Samsung",
	}
	body, err := json.Marshal(installation)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)
	err = s.createInstallationHandler(c)
	if err != nil {
		t.Fatalf("createInstallationHandler failed: %v\n", err)
	}
	if resp.Code != http.StatusCreated {
		t.Fatalf("createInstallationHandler() wrong status code = %v", resp.Code)
	}

	expected := map[string]string{"message": "Installation created"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Fatalf("createInstallationHandler() error decoding response body: %v", err)
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("createInstallationHandler() wrong response body. expected = %v, actual = %v", expected, actual)
	}
}

func TestCreateInstallationNonUUID(t *testing.T) {
	installation := model.InstallationDTO{
		Id:         "Test1234",
		SdkVersion: 32,
		Model:      "s32",
		Brand:      "Samsung",
	}
	body, err := json.Marshal(installation)
	if err != nil {
		t.Fatalf("Could not marshal collectionDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/installations", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)
	err = s.createInstallationHandler(c)
	if err != nil {
		t.Fatalf("createInstallationHandler failed: %v\n", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("createInstallationHandler() wrong status code = %v", resp.Code)
	}
	expected := map[string]string{"message": "Body validation failed: Key: 'InstallationDTO.Id' Error:Field validation for 'Id' failed on the 'uuid' tag"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Fatalf("createInstallationHandler() error decoding response body: %v", err)
	}
	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("createInstallationHandler() wrong response body. expected = %v, actual = %v", expected, actual)
	}
}

func TestCreateMemoryUsage(t *testing.T) {
	err := db.CreateSession(model.NewSessionData{
		Id:    "c40def38-6bf6-488e-905d-45ebacecf3e2",
		AppId: appId,
	})
	if err != nil {
		t.Fatalf("Could not create session: %v\n", err)
	}

	usages := make([]model.NewMemoryUsageDTO, 0, 0)
	usages = append(usages, model.NewMemoryUsageDTO{
		Id:                 "d581fbcc-6a0e-4fad-ae34-bdffe882629b",
		SessionId:          "c40def38-6bf6-488e-905d-45ebacecf3e2",
		InstallationId:     "dd72f2d8-c679-4e7c-bf6b-56f6ec78391b",
		FreeMemory:         12,
		UsedMemory:         12,
		TotalMemory:        24,
		MaxMemory:          36,
		AvailableHeapSpace: 24,
		CreatedAt:          12345678,
	})
	body, err := json.Marshal(usages)
	if err != nil {
		t.Fatalf("Could not marshal []NewMemoryUsageDTO: %v", err)
	}
	reader := bytes.NewReader(body)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resources/memory", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("appId", appId)

	err = s.createMemoryUsageHandler(c)
	if err != nil {
		t.Errorf("createMemoryUsageHandler() error = %v", err)
		return
	}
	expected := map[string]string{"message": "Memory usage created"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
		return
	}

	t.Logf("Response body = %v\n", actual)

	if resp.Code != http.StatusCreated {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	// Compare the decoded response with the expected value
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("createCollectionHandler() wrong response body. expected = %v, actual = %v", expected, actual)
		return
	}
}
