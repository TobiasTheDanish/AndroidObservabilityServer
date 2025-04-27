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
	CreateTeam(data model.NewTeamData) (int, error)
	GetTeamsForUser(userId int) ([]model.TeamEntity, error)

	CreateUser(data model.NewUserData) (int, error)
	GetUserByName(username string) (model.UserEntity, error)
	GetUserById(id int) (model.UserEntity, error)

	CreateTeamUserLink(data model.NewTeamUserLinkData) error
	ValidateTeamUserLink(teamId, userId int) bool

	CreateAuthSession(data model.NewAuthSessionData) error
	GetAuthSession(sessionId string) (model.AuthSessionEntity, error)
	ExtendAuthSession(sessionId string, newExpiry int64) (string, error)
	DeleteAuthSession(sessionId string) error

	CreateApplication(data model.NewApplicationData) (int, error)
	GetApplication(id int) (model.ApplicationEntity, error)
	GetApplicationData(id int) (model.ApplicationDataEntity, error)
	GetTeamApplications(teamId int) ([]model.ApplicationEntity, error)

	CreateApiKey(data model.NewApiKeyData) error
	// Validates that the given apiKey exists in the database and is active
	ValidateApiKey(string) bool
	// Returns the id of the owner of the ApiKey
	GetAppId(apiKey string) (int, error)

	CreateInstallation(data model.NewInstallationData) error
	CreateSession(data model.NewSessionData) error
	MarkSessionCrashed(id string, ownerId int) error
	CreateEvent(data model.NewEventData) error
	CreateTrace(data model.NewTraceData) error
	CreateMemoryUsage(data model.NewMemoryUsageData) error

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

func (s *service) CreateTeam(data model.NewTeamData) (int, error) {
	query := "INSERT INTO public.ob_teams(name) VALUES ($1) RETURNING id"

	var id int
	err := s.db.QueryRow(query, data.Name).Scan(&id)

	return id, err
}

func (s *service) GetTeamsForUser(userId int) ([]model.TeamEntity, error) {
	query := "SELECT t.id, t.name FROM public.ob_teams AS t INNER JOIN public.ob_team_users AS tu ON tu.team_id = t.id WHERE tu.user_id = $1"

	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]model.TeamEntity, 0)
	for rows.Next() {
		var team model.TeamEntity
		err := rows.Scan(&team.Id, &team.Name)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	return teams, nil
}

func (s *service) CreateUser(data model.NewUserData) (int, error) {
	query := "INSERT INTO public.ob_users(name, pw_hash) VALUES ($1, $2) RETURNING id"

	var id int
	err := s.db.QueryRow(query, data.Name, data.PasswordHash).Scan(&id)

	return id, err
}

func (s *service) GetUserByName(username string) (model.UserEntity, error) {
	query := "SELECT id, name, pw_hash FROM public.ob_users WHERE name = $1"

	var entity model.UserEntity
	err := s.db.QueryRow(query, username).Scan(&entity.Id, &entity.Name, &entity.PasswordHash)

	return entity, err
}

func (s *service) GetUserById(id int) (model.UserEntity, error) {
	query := "SELECT id, name, pw_hash FROM public.ob_users WHERE id = $1"

	var entity model.UserEntity
	err := s.db.QueryRow(query, id).Scan(&entity.Id, &entity.Name, &entity.PasswordHash)

	return entity, err
}

func (s *service) CreateTeamUserLink(data model.NewTeamUserLinkData) error {
	query := "INSERT INTO public.ob_team_users(team_id, user_id, role) VALUES ($1, $2, $3)"

	_, err := s.db.Exec(query, data.TeamId, data.UserId, data.Role)

	return err
}

func (s *service) ValidateTeamUserLink(teamId, userId int) bool {
	query := "SELECT EXISTS(SELECT 1 FROM public.ob_team_users WHERE team_id = $1 AND user_id = $2)"

	var exists bool
	if err := s.db.QueryRow(query, teamId, userId).Scan(&exists); err != nil {
		log.Printf("Error validating team user link: %v\n", err)
		return false
	}

	return exists
}

func (s *service) CreateAuthSession(data model.NewAuthSessionData) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	// Delete previous sessions for user
	query := "DELETE FROM public.ob_auth_sessions WHERE user_id = $1"
	_, err = tx.Exec(query, data.UserId)
	if err != nil {
		return err
	}

	// Insert new session
	query = "INSERT INTO public.ob_auth_sessions(id, user_id, expiry) VALUES ($1, $2, $3)"
	_, err = tx.Exec(query, data.Id, data.UserId, data.Expiry)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *service) GetAuthSession(sessionId string) (model.AuthSessionEntity, error) {
	query := "SELECT id, user_id, expiry FROM public.ob_auth_sessions WHERE id = $1"

	var res model.AuthSessionEntity
	err := s.db.QueryRow(query, sessionId).Scan(
		&res.Id,
		&res.UserId,
		&res.Expiry,
	)

	return res, err
}

func (s *service) ExtendAuthSession(sessionId string, newExpiry int64) (string, error) {
	query := "UPDATE public.ob_auth_sessions SET expiry=$1 WHERE id = $2 RETURNING id"

	var res string
	err := s.db.QueryRow(query, newExpiry, sessionId).Scan(&res)

	return res, err
}

func (s *service) DeleteAuthSession(sessionId string) error {
	query := "DELETE FROM public.ob_auth_sessions WHERE id = $1"

	_, err := s.db.Exec(query, sessionId)

	return err
}

func (s *service) CreateApplication(data model.NewApplicationData) (int, error) {
	query := "INSERT INTO public.ob_applications(name, team_id) VALUES ($1, $2) RETURNING id"

	var id int
	err := s.db.QueryRow(query, data.Name, data.TeamId).Scan(&id)

	return id, err
}

func (s *service) GetApplication(id int) (model.ApplicationEntity, error) {
	query := "SELECT id, name, team_id FROM public.ob_applications WHERE id = $1"

	var res model.ApplicationEntity
	err := s.db.QueryRow(query, id).Scan(
		&res.Id,
		&res.Name,
		&res.TeamId,
	)

	return res, err
}

func (s *service) GetApplicationData(id int) (model.ApplicationDataEntity, error) {
	installationQuery := "SELECT id, sdk_version, model, brand, app_id FROM public.ob_installations WHERE app_id = $1"

	rows, err := s.db.Query(installationQuery, id)
	if err != nil {
		return model.ApplicationDataEntity{}, err
	}

	installations := make([]model.InstallationEntity, 0)
	for rows.Next() {
		var entity model.InstallationEntity
		err := rows.Scan(&entity.Id, &entity.SDKVersion, &entity.Model, &entity.Brand, &entity.AppId)
		if err != nil {
			log.Printf("Error scanning installation entity: %v\n", err)
			return model.ApplicationDataEntity{}, err
		}
		installations = append(installations, entity)
	}

	sessionQuery := "SELECT id, installation_id, created_at, crashed, app_id FROM public.ob_sessions WHERE app_id = $1"

	rows, err = s.db.Query(sessionQuery, id)
	if err != nil {
		return model.ApplicationDataEntity{}, err
	}

	sessions := make([]model.SessionEntity, 0)
	for rows.Next() {
		var entity model.SessionEntity
		err := rows.Scan(&entity.Id, &entity.InstallationId, &entity.CreatedAt, &entity.Crashed, &entity.AppId)
		if err != nil {
			log.Printf("Error scanning installation entity: %v\n", err)
			return model.ApplicationDataEntity{}, err
		}
		sessions = append(sessions, entity)
	}

	return model.ApplicationDataEntity{
		Installations: installations,
		Sessions:      sessions,
	}, nil
}

func (s *service) GetTeamApplications(teamId int) ([]model.ApplicationEntity, error) {
	query := "SELECT id, name, team_id FROM public.ob_applications WHERE team_id = $1"

	rows, err := s.db.Query(query, teamId)
	if err != nil {
		return nil, err
	}

	apps := make([]model.ApplicationEntity, 0)
	for rows.Next() {
		var app model.ApplicationEntity
		err = rows.Scan(
			&app.Id,
			&app.Name,
			&app.TeamId,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, nil
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
	res, err := s.db.Exec("INSERT INTO public.ob_installations (id, app_id, sdk_version, model, brand) VALUES ($1, $2, $3, $4, $5)", data.Id, data.AppId, data.SdkVersion, data.Model, data.Brand)
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

func (s *service) CreateMemoryUsage(data model.NewMemoryUsageData) error {
	query := "INSERT INTO public.ob_memory_usage (id, session_id, installation_id, free_memory, used_memory, max_memory, total_memory, available_heap_space) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err := s.db.Exec(query, data.Id, data.SessionId, data.InstallationId, data.FreeMemory, data.UsedMemory, data.MaxMemory, data.TotalMemory, data.AvailableHeapSpace)
	if err != nil {
		return err
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
