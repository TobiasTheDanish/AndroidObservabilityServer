package database

import (
	"ObservabilityServer/internal/auth"
	"ObservabilityServer/internal/model"
	"context"
	"log"
	"slices"
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

func TestCreateTeam(t *testing.T) {
	srv := New()
	teamId, err := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	if err != nil || teamId == -1 {
		t.Fatalf("Creating team failed. %v\n", err)
	}
}

func TestCreateUser(t *testing.T) {
	srv := New()
	pwHash, err := auth.HashPassword("Abc123")
	if err != nil {
		t.Fatalf("Error hashing pw: %v\n", err)
	}

	userId, err := srv.CreateUser(model.NewUserData{
		Name:         "Test User",
		PasswordHash: pwHash,
	})
	if err != nil || userId == -1 {
		t.Fatalf("Creating user failed. %v\n", err)
	}
}

func TestGetUserByName(t *testing.T) {
	srv := New()
	pwHash, _ := auth.HashPassword("Abc123")

	_, err := srv.CreateUser(model.NewUserData{
		Name:         "Test User 2",
		PasswordHash: pwHash,
	})
	if err != nil {
		t.Fatalf("Creating user failed: %v\n", err)
	}

	_, err = srv.GetUserByName("Test User 2")
	if err != nil {
		t.Errorf("Getting user by name failed: %v\n", err)
	}

	user, err := srv.GetUserByName("Unknown User")
	if err == nil {
		t.Errorf("Getting user by unknown name was successful, but it should NOT! Fetched user: %v", user)
	}
}

func TestGetUserById(t *testing.T) {
	srv := New()
	pwHash, _ := auth.HashPassword("Abc123")

	id, err := srv.CreateUser(model.NewUserData{
		Name:         "Test User 3",
		PasswordHash: pwHash,
	})
	if err != nil {
		t.Fatalf("Creating user failed: %v\n", err)
	}

	_, err = srv.GetUserById(id)
	if err != nil {
		t.Errorf("Getting user by id failed: %v\n", err)
	}

	user, err := srv.GetUserById(-1)
	if err == nil {
		t.Errorf("Getting user by unknown id was successful, but it should NOT! Fetched user: %v", user)
	}
}

func TestCreateTeamUserLink(t *testing.T) {
	srv := New()

	teamId, err := srv.CreateTeam(model.NewTeamData{Name: "Test Team 2"})
	if err != nil {
		t.Fatalf("Creating team failed: %v\n", err)
	}
	pwHash, err := auth.HashPassword("Abc123")
	if err != nil {
		t.Fatalf("Error hashing pw: %v\n", err)
	}

	userId, err := srv.CreateUser(model.NewUserData{
		Name:         "Test User 4",
		PasswordHash: pwHash,
	})
	if err != nil {
		t.Fatalf("Creating user failed: %v\n", err)
	}

	err = srv.CreateTeamUserLink(model.NewTeamUserLinkData{
		TeamId: teamId,
		UserId: userId,
		Role:   "User",
	})

	if err != nil || userId == -1 {
		t.Fatalf("Creating team-user link failed. %v\n", err)
	}

	if !srv.ValidateTeamUserLink(teamId, userId) {
		t.Fatalf("Team-User link was not valid, when it should!")
	}
}

func TestCreateAuthSession(t *testing.T) {
	srv := New()
	pwHash, _ := auth.HashPassword("Abc123")

	userId, err := srv.CreateUser(model.NewUserData{
		Name:         "Test User 5",
		PasswordHash: pwHash,
	})
	if err != nil {
		t.Fatalf("Creating user failed: %v\n", err)
	}

	sessionId, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("Generating session token failed: %v\n", err)
	}

	err = srv.CreateAuthSession(model.NewAuthSessionData{
		Id:     sessionId,
		UserId: userId,
		Expiry: auth.GetExpiryForSession(),
	})
	if err != nil {
		t.Fatalf("Creating auth session failed: %v\n", err)
	}

	_, err = srv.GetAuthSession(sessionId)
	if err != nil {
		t.Fatalf("Getting auth session failed: %v\n", err)
	}
}

func TestCreateApp(t *testing.T) {
	srv := New()

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, err := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})
	if err != nil || appId == -1 {
		t.Fatalf("Creating app failed. %v\n", err)
	}

	_, err = srv.GetApplication(appId)
	if err != nil || appId == -1 {
		t.Fatalf("Getting app failed. %v\n", err)
	}
}

func TestGetApplicationData(t *testing.T) {
	srv := New()

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	installationId := "TestInstallation321"
	srv.CreateInstallation(model.NewInstallationData{
		Id:         installationId,
		AppId:      appId,
		SdkVersion: 33,
		Model:      "MT9556",
		Brand:      "Newland",
	})

	sessionId1 := "TestSession321"
	srv.CreateSession(model.NewSessionData{
		Id:             sessionId1,
		InstallationId: installationId,
		AppId:          appId,
		CreatedAt:      1,
		Crashed:        false,
	})
	sessionId2 := "TestSession1234"
	srv.CreateSession(model.NewSessionData{
		Id:             sessionId2,
		InstallationId: installationId,
		AppId:          appId,
		CreatedAt:      1,
		Crashed:        true,
	})

	appData, err := srv.GetApplicationData(appId)
	if err != nil {
		t.Fatalf("Getting app data failed with err: %v\n", err)
	}

	if !slices.ContainsFunc(appData.Installations, func(i model.InstallationEntity) bool {
		return i.Id == installationId
	}) {
		t.Error("The expected installation id was not returned!")
	}

	if !slices.ContainsFunc(appData.Sessions, func(s model.SessionEntity) bool {
		return s.Id == sessionId1 && !s.Crashed
	}) {
		t.Errorf("The session id '%s' was not returned!", sessionId1)
	}
	if !slices.ContainsFunc(appData.Sessions, func(s model.SessionEntity) bool {
		return s.Id == sessionId2 && s.Crashed
	}) {
		t.Errorf("The session id '%s' was not returned!", sessionId2)
	}
}

func TestValidateApiKey(t *testing.T) {
	srv := New()

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, err := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	key, err := auth.GenerateApiKey()
	if err != nil {
		t.Fatalf("Could not generate api key: %v\n", err)
	}

	hash := auth.HashApiKey(key)

	srv.CreateApiKey(model.NewApiKeyData{
		Key:   hash,
		AppId: appId,
	})

	if !srv.ValidateApiKey(hash) {
		t.Fatalf("ValidateApiKey returned false with key: %s\n", key)
	}

	id, err := srv.GetAppId(hash)
	if err != nil {
		t.Fatalf("GetOwnerId failed with error: %v\n", err)
	}
	if id != appId {
		t.Fatalf("Owner ids did not match. expected %d, but got %d\n", appId, id)
	}
}

func TestCreateInstallation(t *testing.T) {
	srv := New()

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	data := model.NewInstallationData{
		Id:         "TestInstallation123",
		AppId:      appId,
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

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	data := model.NewSessionData{
		Id:             "TestSession123",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          appId,
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

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	data := model.NewSessionData{
		Id:             "TestSession223",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          appId,
		CreatedAt:      1,
		Crashed:        false,
	}

	err := srv.CreateSession(data)
	if err != nil {
		t.Fatalf("CreateSession failed: %v\n", err)
	}

	err = srv.MarkSessionCrashed(data.Id, appId)
	if err != nil {
		t.Fatalf("MarkSessionCrashed failed: %v\n", err)
	}
}

func TestCreateEvent(t *testing.T) {
	srv := New()

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	sessionData := model.NewSessionData{
		Id:             "TestSession1234",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          appId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	eventData := model.NewEventData{
		Id:             "TestEvent",
		SessionId:      sessionData.Id,
		AppId:          appId,
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

	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, _ := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})

	sessionData := model.NewSessionData{
		Id:             "TestSession12345",
		InstallationId: "InstallationIdForTestSession123",
		AppId:          appId,
		CreatedAt:      1,
		Crashed:        false,
	}

	_ = srv.CreateSession(sessionData)

	traceData := model.NewTraceData{
		TraceId:      "TestTrace",
		SessionId:    sessionData.Id,
		GroupId:      "TestGroup",
		ParentId:     "",
		AppId:        appId,
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
