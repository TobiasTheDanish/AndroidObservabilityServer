package database

import (
	"ObservabilityServer/internal/model"
	"slices"
	"testing"
)

func TestGetLatestDatabaseLogs(t *testing.T) {
	srv := New(config)
	teamId, _ := srv.CreateTeam(model.NewTeamData{Name: "Test Team"})
	appId, err := srv.CreateApplication(model.NewApplicationData{
		Name:   "TestApp",
		TeamId: teamId,
	})
	if err != nil || appId == -1 {
		t.Fatalf("Creating app failed. %v\n", err)
	}

	sessionId := "TestSession123"

	err = srv.CreateLog(model.NewLogData{
		AppId:     appId,
		SessionId: sessionId,
		Message:   "Hello",
		Data:      map[string]any{},
		CreatedAt: 0,
	})
	if err != nil {
		t.Logf("Creating log failed with error: %v\n", err)
	}
	err = srv.CreateLog(model.NewLogData{
		AppId:     appId,
		SessionId: sessionId,
		Message:   "Hello again",
		Data:      map[string]any{},
		CreatedAt: 0,
	})
	if err != nil {
		t.Logf("Creating log failed with error: %v\n", err)
	}

	logs, err := srv.GetLatestLogsBySessionId(sessionId, 10, 0, nil)
	if err != nil {
		t.Fatalf("Getting latest logs without filters failed. %v\n", err)
	}

	if len(logs) != 2 || logs[0].Message != "Hello" || logs[1].Message != "Hello again" {
		t.Fatalf("Getting latest logs without filters did not return expected logs. Got: %v", logs)
	}

	filters := []LogFilter{
		{
			Key:      "message",
			Operator: "eq",
			Value:    "Hello",
		},
	}

	logs, err = srv.GetLatestLogsBySessionId(sessionId, 10, 0, filters)
	if err != nil {
		t.Fatalf("Getting latest logs failed. %v\n", err)
	}

	if len(logs) != 1 || logs[0].Message != "Hello" {
		t.Fatalf("Getting latest logs did not return expected logs. Got: %v", logs)
	}
}

func TestLogFilterToWhereClause(t *testing.T) {
	filters := []LogFilter{
		{
			Key:      "id",
			Operator: "eq",
			Value:    "123",
		},
	}

	expectedW := `id = $1`
	expectedV := make([]string, 0)
	expectedV = append(expectedV, "123")

	w, v := logFilterToWhereClause(filters, 1)
	if w != expectedW {
		t.Errorf("logFilterToWhereClaus failed = expected %s got %s", expectedW, w)
	}
	if slices.Compare(v, expectedV) != 0 {
		t.Errorf("logFilterToWhereClaus failed = expected %s got %s", expectedV, v)
	}
}
