package config

import (
	"database/sql"
	"fmt"
	"log" // Using log for consistency
	"os"
	"strconv"
	"sync"
	"time"

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
	DB            *db.Queries
	Port          string
	DB_URL        string // Storing for reference, actual connection is globalDBConn
	JWTSecret     string
	JWTExpiration time.Duration
}

// initializeGlobalDB is responsible for setting up the database connection pool and queries object.
// It is intended to be called only once.
func initializeGlobalDB() {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env is not found when running in an environment where vars are already set.
		log.Println("No .env file found or error loading it, relying on environment variables.")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		initConfigErr = fmt.Errorf("DB_URL environment variable not set")
		log.Println(initConfigErr) // Log the error
		return
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		initConfigErr = fmt.Errorf("cannot open database connection: %w", err)
		log.Println(initConfigErr) // Log the error
		return
	}

	// Configure connection pool settings.
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)
	conn.SetConnMaxLifetime(5 * time.Minute)
	conn.SetConnMaxIdleTime(1 * time.Minute)

	err = conn.Ping()
	if err != nil {
		conn.Close() // Close the connection if ping fails
		initConfigErr = fmt.Errorf("cannot connect to database (ping failed): %w", err)
		log.Println(initConfigErr) // Log the error
		return
	}

	globalDBConn = conn
	globalQueries = db.New(globalDBConn)
	log.Println("Database connection pool initialized successfully.")
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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, relying on environment variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set for the API server
		log.Printf("PORT environment variable not set, defaulting to %s", port)
	}

	dbURL := os.Getenv("DB_URL") // Get for reference in APIConfig

	queries, err := GetDBQueries()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB queries: %w", err)
	}

	// Load JWT Configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable not set")
	}

	jwtExpMinutesStr := os.Getenv("JWT_EXPIRATION_MINUTES")
	jwtExpMinutes, err := strconv.Atoi(jwtExpMinutesStr)
	if err != nil || jwtExpMinutes <= 0 {
		log.Printf("JWT_EXPIRATION_MINUTES not set or invalid ('%s'), defaulting to 60 minutes. Error: %v", jwtExpMinutesStr, err)
		jwtExpMinutes = 60 // Default to 60 minutes
	}
	jwtExpiration := time.Duration(jwtExpMinutes) * time.Minute

	return &APIConfig{
		DB:            queries,
		Port:          port,
		DB_URL:        dbURL,
		JWTSecret:     jwtSecret,
		JWTExpiration: jwtExpiration,
	}, nil
}

// CloseDB closes the global database connection pool.
func CloseDB() {
	if globalDBConn != nil {
		log.Println("Closing database connection pool.")
		err := globalDBConn.Close()
		if err != nil {
			log.Printf("Error closing database connection pool: %v\n", err)
		}
	}
}
