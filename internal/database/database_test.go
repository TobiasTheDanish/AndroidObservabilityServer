package database

import (
	"ObservabilityServer/internal/auth"
	"ObservabilityServer/internal/model"
	"context"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	teardown, err := SetupTestDatabase()
	if err != nil {
		log.Fatalf("could not start postgres container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown postgres container: %v", err)
	}
}

func TestNew(t *testing.T) {
	srv := New()
	if srv == nil {
		t.Fatal("New() returned nil")
	}
}

func TestCreateOwner(t *testing.T) {
	srv := New()

	ownerId, err := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})
	if err != nil || ownerId == -1 {
		t.Fatalf("Creating owner failed. %v\n", err)
	}
}

func TestValidateApiKey(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	key, err := auth.GenerateApiKey()
	if err != nil {
		t.Fatalf("Could not generate api key: %v\n", err)
	}

	hash := auth.HashApiKey(key)

	srv.CreateApiKey(model.NewApiKeyData{
		Key:   hash,
		AppId: ownerId,
	})

	if !srv.ValidateApiKey(hash) {
		t.Fatalf("ValidateApiKey returned false with key: %s\n", key)
	}

	id, err := srv.GetAppId(hash)
	if err != nil {
		t.Fatalf("GetOwnerId failed with error: %v\n", err)
	}
	if id != ownerId {
		t.Fatalf("Owner ids did not match. expected %d, but got %d\n", ownerId, id)
	}
}

func TestCreateInstallation(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	data := model.NewInstallationData{
		Id:         "TestInstallation123",
		OwnerId:    ownerId,
		SdkVersion: 31,
		Model:      "MT9556",
		Brand:      "Newland",
	}

	err := srv.CreateInstallation(data)
	if err != nil {
		t.Fatalf("CreateInstallation failed: %v\n", err)
	}

	err = srv.CreateInstallation(data)
	if err == nil {
		t.Fatalf("CreateInstallation was expected to fail, but didnt!")
	}
}

func TestCreateSession(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	data := model.NewSessionData{
		Id:             "TestSession123",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	err := srv.CreateSession(data)
	if err != nil {
		t.Fatalf("CreateSession failed: %v\n", err)
	}

	err = srv.CreateSession(data)
	if err == nil {
		t.Fatalf("CreateSession was expected to fail, but didnt!")
	}
}

func TestMarkSessionCrashed(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	data := model.NewSessionData{
		Id:             "TestSession223",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	err := srv.CreateSession(data)
	if err != nil {
		t.Fatalf("CreateSession failed: %v\n", err)
	}

	err = srv.MarkSessionCrashed(data.Id, ownerId)
	if err != nil {
		t.Fatalf("MarkSessionCrashed failed: %v\n", err)
	}
}

func TestCreateEvent(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	sessionData := model.NewSessionData{
		Id:             "TestSession1234",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	eventData := model.NewEventData{
		Id:             "TestEvent",
		SessionId:      sessionData.Id,
		AppId:          ownerId,
		Type:           "TestEvent",
		SerializedData: "{}",
		CreatedAt:      2,
	}

	err := srv.CreateEvent(eventData)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v\n", err)
	}
}

func TestCreateTrace(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateApplication(model.NewApplicationData{Name: "TestOwner"})

	sessionData := model.NewSessionData{
		Id:             "TestSession12345",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	traceData := model.NewTraceData{
		TraceId:      "TestTrace",
		SessionId:    sessionData.Id,
		GroupId:      "TestGroup",
		ParentId:     "",
		AppId:        ownerId,
		Name:         "TraceTest",
		Status:       "Ok",
		ErrorMessage: "",
		StartedAt:    2,
		EndedAt:      4,
		HasEnded:     true,
	}

	err := srv.CreateTrace(traceData)
	if err != nil {
		t.Fatalf("CreateTrace failed: %v\n", err)
	}
}

func TestHealth(t *testing.T) {
	srv := New()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present")
	}

	if stats["message"] != "It's healthy" {
		t.Fatalf("expected message to be 'It's healthy', got %s", stats["message"])
	}
}

func TestClose(t *testing.T) {
	srv := New()

	if srv.Close() != nil {
		t.Fatalf("expected Close() to return nil")
	}
}
