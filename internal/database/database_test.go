package database

import (
	"ObservabilityServer/internal/auth"
	"ObservabilityServer/internal/model"
	"context"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	var (
		dbName = "database"
		dbPwd  = "password"
		dbUser = "user"
	)

	dbContainer, err := postgres.Run(
		context.Background(),
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	database = dbName
	password = dbPwd
	username = dbUser

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "5432/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	host = dbHost
	port = dbPort.Port()

	return dbContainer.Terminate, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartPostgresContainer()
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

	ownerId, err := srv.CreateOwner(model.NewOwnerData{Name: "TestOwner"})
	if err != nil || ownerId == -1 {
		t.Fatalf("Creating owner failed. %v\n", err)
	}
}

func TestValidateApiKey(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateOwner(model.NewOwnerData{Name: "TestOwner"})

	key, err := auth.GenerateApiKey()
	if err != nil {
		t.Fatalf("Could not generate api key: %v\n", err)
	}

	srv.CreateApiKey(model.NewApiKeyData{
		Key:     key,
		OwnerID: ownerId,
	})

	if !srv.ValidateApiKey(key) {
		t.Fatalf("ValidateApiKey returned false with key: %s\n", key)
	}
}

func TestCreateSession(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateOwner(model.NewOwnerData{Name: "TestOwner"})

	data := model.NewSessionData{
		Id:             "TestSession123",
		InstallationId: "InstallationIdForTestSession123",
		OwnerId:        ownerId,
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

func TestCreateEvent(t *testing.T) {
	srv := New()

	ownerId, _ := srv.CreateOwner(model.NewOwnerData{Name: "TestOwner"})

	sessionData := model.NewSessionData{
		Id:             "TestSession1234",
		InstallationId: "InstallationIdForTestSession123",
		OwnerId:        ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	eventData := model.NewEventData{
		Id:             "TestEvent",
		SessionId:      sessionData.Id,
		OwnerId:        ownerId,
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

	ownerId, _ := srv.CreateOwner(model.NewOwnerData{Name: "TestOwner"})

	sessionData := model.NewSessionData{
		Id:             "TestSession12345",
		InstallationId: "InstallationIdForTestSession123",
		OwnerId:        ownerId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	traceData := model.NewTraceData{
		TraceId:      "TestTrace",
		SessionId:    sessionData.Id,
		GroupId:      "TestGroup",
		ParentId:     "",
		OwnerId:      ownerId,
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
