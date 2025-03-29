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

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	CreateSession(data model.NewSessionData) error
	CreateEvent(data model.NewEventData) error
	CreateTrace(data model.NewTraceData) error

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
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	version    = os.Getenv("BLUEPRINT_DB_VERSION")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}

	err = dbInstance.init()
	if err != nil {
		log.Fatalf("Could not init db: %v\n", err)
	}

	return dbInstance
}

func (s *service) CreateSession(data model.NewSessionData) error {
	crashed := 0
	if data.Crashed {
		crashed = 1
	}

	res, err := s.db.Exec("INSERT INTO public.ob_sessions (id, installation_id, created_at, crashed) VALUES ($1, $2, $3, $4)", data.Id, data.InstallationId, data.CreatedAt, crashed)
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
	sql := " INSERT INTO public.ob_events( id, session_id, created_at, type, serialized_data) VALUES ($1, $2, $3, $4, $5)"

	res, err := s.db.Exec(sql, data.ID, data.SessionID, data.CreatedAt, data.Type, data.SerializedData)
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
	sql := "INSERT INTO public.ob_trace( trace_id, session_id, group_id, parent_id, name, status, error_message, started_at, ended_at, has_ended) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"

	hasEnded := 0
	if data.HasEnded {
		hasEnded = 1
	}

	res, err := s.db.Exec(sql, data.TraceId, data.SessionId, data.GroupId, data.ParentId, data.Name, data.Status, data.ErrorMessage, data.StartedAt, data.EndedAt, hasEnded)
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

func (s *service) init() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS public.__db_version__ ( current_version INTEGER )")

	if err != nil {
		return err
	}

	row := tx.QueryRow("SELECT current_version FROM public.__db_version__ LIMIT 1")

	var version int32
	if row.Scan(&version) != nil {
		err = createTables(tx)
		if err != nil {
			log.Fatalf("Could not create database tables: %v\n", err)
			return err
		}

		version = 1
		_, err = tx.Exec("INSERT INTO public.__db_version__ VALUES (1)")
		if err != nil {
			log.Fatalf("Could not insert new version number into version table: %v\n", err)
			return err
		}
	}

	return tx.Commit()
}

func createTables(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE IF NOT EXISTS public.ob_sessions (id TEXT PRIMARY KEY, installation_id TEXT NOT NULL, created_at INTEGER NOT NULL, crashed SMALLINT NOT NULL DEFAULT 0)")

	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS public.ob_events (id TEXT PRIMARY KEY, session_id TEXT NOT NULL, created_at INTEGER NOT NULL, type TEXT NOT NULL, serialized_data TEXT DEFAULT '', FOREIGN KEY (session_id) REFERENCES public.ob_sessions (id) ON DELETE NO ACTION ON UPDATE NO ACTION)")

	if err != nil {
		return err
	}

	_, err = tx.Exec("CREATE TABLE IF NOT EXISTS public.ob_trace (trace_id TEXT PRIMARY KEY, session_id TEXT NOT NULL, group_id TEXT NOT NULL, parent_id TEXT DEFAULT '', name TEXT NOT NULL, status TEXT NOT NULL, error_message TEXT DEFAULT '', started_at BIGINT NOT NULL, ended_at BIGINT NOT NULL DEFAULT 0, has_ended INTEGER NOT NULL DEFAULT 0, FOREIGN KEY (session_id) REFERENCES public.ob_sessions (id) ON DELETE NO ACTION ON UPDATE NO ACTION)")

	return err
}
