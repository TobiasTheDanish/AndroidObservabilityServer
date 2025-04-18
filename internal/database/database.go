package database

import (
	"ObservabilityServer/internal/model"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Service represents a service that interacts with a database.
type Service interface {
	CreateApplication(data model.NewApplicationData) (int, error)
	CreateApiKey(data model.NewApiKeyData) error

	CreateInstallation(data model.NewInstallationData) error
	CreateSession(data model.NewSessionData) error
	MarkSessionCrashed(id string, ownerId int) error
	CreateEvent(data model.NewEventData) error
	CreateTrace(data model.NewTraceData) error

	// Validates that the given apiKey exists in the database and is active
	ValidateApiKey(string) bool

	// Returns the id of the owner of the ApiKey
	GetAppId(apiKey string) (int, error)

	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("OBSERVE_DB_DATABASE")
	password   = os.Getenv("OBSERVE_DB_PASSWORD")
	username   = os.Getenv("OBSERVE_DB_USERNAME")
	port       = os.Getenv("OBSERVE_DB_PORT")
	host       = os.Getenv("OBSERVE_DB_HOST")
	schema     = os.Getenv("OBSERVE_DB_SCHEMA")
	migrations = os.Getenv("MIGRATIONS_PATH")
	dbInstance *service
)

func SetupTestDatabase() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	var (
		dbName = "routes_database"
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

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	sourcePath := fmt.Sprintf("file://../../migrations")

	m, err := migrate.New(sourcePath, connStr)
	if err != nil {
		return dbContainer.Terminate, fmt.Errorf("Error creating migrate instance: %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return dbContainer.Terminate, fmt.Errorf("Error migrating up: %v", err)
	}

	return dbContainer.Terminate, nil
}

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	dbInstance = &service{
		db: db,
	}

	return dbInstance
}

func (s *service) CreateApplication(data model.NewApplicationData) (int, error) {
	query := "INSERT INTO public.ob_applications(name) VALUES ($1) RETURNING id"

	var id int
	err := s.db.QueryRow(query, data.Name).Scan(&id)

	return id, err
}

func (s *service) CreateApiKey(data model.NewApiKeyData) error {
	query := "INSERT INTO public.ob_api_keys(key, app_id) VALUES ($1, $2)"

	res, err := s.db.Exec(query, data.Key, data.AppId)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("Expected 1 api key to be inserted but was %d", rowsAffected)
	}

	return nil
}

func (s *service) MarkSessionCrashed(id string, ownerId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return nil
	}
	query := "UPDATE public.ob_sessions SET crashed=1 WHERE id=$1 AND app_id=$2"

	res, err := tx.Exec(query, id, ownerId)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		tx.Rollback()
		return fmt.Errorf("Expected 1 session to be updated but was %d. Rolling back", rowsAffected)
	}

	return tx.Commit()
}

func (s *service) CreateInstallation(data model.NewInstallationData) error {
	res, err := s.db.Exec("INSERT INTO public.ob_installations (id, app_id, sdk_version, model, brand) VALUES ($1, $2, $3, $4, $5)", data.Id, data.OwnerId, data.SdkVersion, data.Model, data.Brand)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("Expected 1 session to be inserted but was %d", rowsAffected)
	}

	return nil
}

func (s *service) CreateSession(data model.NewSessionData) error {
	crashed := 0
	if data.Crashed {
		crashed = 1
	}

	res, err := s.db.Exec("INSERT INTO public.ob_sessions (id, installation_id, app_id, created_at, crashed) VALUES ($1, $2, $3, $4, $5)", data.Id, data.InstallationId, data.AppId, data.CreatedAt, crashed)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("Expected 1 session to be inserted but was %d", rowsAffected)
	}

	return nil
}

func (s *service) CreateEvent(data model.NewEventData) error {
	sql := " INSERT INTO public.ob_events( id, session_id, app_id, created_at, type, serialized_data) VALUES ($1, $2, $3, $4, $5, $6)"

	res, err := s.db.Exec(sql, data.Id, data.SessionId, data.AppId, data.CreatedAt, data.Type, data.SerializedData)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("Expected 1 event to be inserted but was %d", rowsAffected)
	}

	return nil
}

func (s *service) CreateTrace(data model.NewTraceData) error {
	sql := "INSERT INTO public.ob_trace( trace_id, session_id, group_id, parent_id, app_id, name, status, error_message, started_at, ended_at, has_ended) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)"

	hasEnded := 0
	if data.HasEnded {
		hasEnded = 1
	}

	res, err := s.db.Exec(sql, data.TraceId, data.SessionId, data.GroupId, data.ParentId, data.AppId, data.Name, data.Status, data.ErrorMessage, data.StartedAt, data.EndedAt, hasEnded)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("Expected 1 trace to be inserted but was %d", rowsAffected)
	}

	return nil
}

func (s *service) ValidateApiKey(apiKey string) bool {
	query := "SELECT EXISTS(SELECT 1 FROM public.ob_api_keys WHERE key = $1)"

	var exists bool
	if err := s.db.QueryRow(query, apiKey).Scan(&exists); err != nil {
		log.Printf("Error validating api key: %v\n", err)
		return false
	}

	return exists
}

func (s *service) GetAppId(apiKey string) (int, error) {
	query := "SELECT app_id FROM public.ob_api_keys WHERE key = $1"

	var ownerId int
	if err := s.db.QueryRow(query, apiKey).Scan(&ownerId); err != nil {
		return -1, err
	}
	return ownerId, nil
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatal(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
