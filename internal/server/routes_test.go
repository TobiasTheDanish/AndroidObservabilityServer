package server

import (
	"ObservabilityServer/internal/database"
	"ObservabilityServer/internal/model"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
)

var (
	db      database.Service
	ownerId int
)

func TestMain(m *testing.M) {
	teardown, err := database.SetupTestDatabase()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	db = database.New()

	ownerId, err = db.CreateOwner(model.NewOwnerData{
		Name: "Test owner",
	})
	if err != nil {
		log.Fatalf("Could not create test owner: %v", err)
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

	c.Set("ownerId", ownerId)

	err = s.createCollectionHandler(c)
	if err != nil {
		t.Errorf("createCollectionHandler() error = %v", err)
		return
	}
	if resp.Code != http.StatusAccepted {
		t.Errorf("createCollectionHandler() wrong status code = %v", resp.Code)
		return
	}

	expected := map[string]string{"message": "Creation of collection have been started"}
	var actual map[string]string
	// Decode the response body into the actual map
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("createCollectionHandler() error decoding response body: %v", err)
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

	c.Set("ownerId", ownerId)
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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/collection", reader)
	resp := httptest.NewRecorder()
	req.Header.Set("Content-type", "application/json")

	e.Validator = NewValidator()
	c := e.NewContext(req, resp)
	s := &Server{
		db: db,
	}

	c.Set("ownerId", ownerId)
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
