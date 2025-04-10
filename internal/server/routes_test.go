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

func TestMain(m *testing.M) {
	teardown, err := database.SetupTestDatabase()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
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
		db: database.New(),
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
	c := e.NewContext(req, resp)
	s := &Server{
		db: database.New(),
	}

	c.Set("ownerId", 1)

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
