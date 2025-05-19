// filepath: internal/config/config.go
package config

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time" // Added for connection pool settings

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	db "github.com/twomotive/dropwise/internal/database/sqlc"
)

var (
	dbOnce        sync.Once
	globalDBConn  *sql.DB     // Holds the global connection pool
	globalQueries *db.Queries // Holds the global sqlc Queries instance
	initConfigErr error       // To store any error during one-time initialization
)

// APIConfig holds application-wide configurations.
type APIConfig struct {
	DB     *db.Queries
	Port   string
	DB_URL string // Storing for reference, actual connection is globalDBConn
}

// initializeGlobalDB is responsible for setting up the database connection pool and queries object.
// It is intended to be called only once.
func initializeGlobalDB() {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env is not found when running in an environment where vars are already set.
		fmt.Println("No .env file found or error loading it, relying on environment variables.")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		initConfigErr = fmt.Errorf("DB_URL environment variable not set")
		return
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		initConfigErr = fmt.Errorf("cannot open database connection: %w", err)
		return
	}

	// Configure connection pool settings.
	// These are example values; adjust them based on your Cloud SQL instance capabilities
	// and expected number of concurrent Cloud Function instances.
	// For Cloud Functions, it's often better to have fewer connections per instance
	// as you can have many instances.
	conn.SetMaxOpenConns(5)                  // Max number of open connections to the database
	conn.SetMaxIdleConns(2)                  // Max number of connections in the idle connection pool
	conn.SetConnMaxLifetime(5 * time.Minute) // Max amount of time a connection may be reused
	conn.SetConnMaxIdleTime(1 * time.Minute) // Max amount of time a connection may be idle before being closed

	err = conn.Ping()
	if err != nil {
		conn.Close() // Close the connection if ping fails
		initConfigErr = fmt.Errorf("cannot connect to database (ping failed): %w", err)
		return
	}

	globalDBConn = conn
	globalQueries = db.New(globalDBConn)
	fmt.Println("Database connection pool initialized successfully.")
}

// GetDBQueries returns the initialized sqlc Queries object, ensuring one-time initialization.
func GetDBQueries() (*db.Queries, error) {
	dbOnce.Do(func() {
		initializeGlobalDB()
	})
	if initConfigErr != nil {
		return nil, initConfigErr
	}
	if globalQueries == nil { // Should be caught by initConfigErr, but as a safeguard
		return nil, fmt.Errorf("database queries not initialized and no error was reported")
	}
	return globalQueries, nil
}

// LoadConfig loads configuration from environment variables and returns an APIConfig.
// It now uses the globally initialized database connection.
func LoadConfig() (*APIConfig, error) {
	// Ensure environment variables are loaded (godotenv.Load is idempotent if called multiple times,
	// but initializeGlobalDB also calls it).
	// If there are other non-DB configs to load from .env, ensure it's accessible here.
	// For simplicity, we assume initializeGlobalDB handles .env loading sufficiently for DB_URL.

	port := os.Getenv("PORT")
	// PORT might not be set or relevant for worker, handle as needed.
	// For API server, it's critical.
	// if port == "" && /* isAPIServerContext */ {
	//     return nil, fmt.Errorf("PORT environment variable not set")
	// }

	dbURL := os.Getenv("DB_URL") // Get for reference in APIConfig

	queries, err := GetDBQueries()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB queries: %w", err)
	}

	return &APIConfig{
		DB:     queries,
		Port:   port,
		DB_URL: dbURL,
	}, nil
}

// CloseDB closes the global database connection pool.
// Useful for graceful shutdown in long-running applications (like the API server).
// Cloud Functions typically manage instance lifecycle, so explicit closing there is less critical
// but doesn't hurt if called (e.g. via a deferred call in main if it were a standalone app).
func CloseDB() {
	if globalDBConn != nil {
		fmt.Println("Closing database connection pool.")
		err := globalDBConn.Close()
		if err != nil {
			fmt.Printf("Error closing database connection pool: %v\n", err)
		}
	}
}
