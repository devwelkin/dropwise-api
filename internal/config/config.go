// filepath: internal/config/config.go
package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	db "github.com/twomotive/dropwise/internal/database/sqlc"
)

// APIConfig holds application-wide configurations.
type APIConfig struct {
	DB     *db.Queries
	Port   string
	DB_URL string
}

// LoadConfig loads configuration from environment variables and initializes necessary services.
func LoadConfig() (*APIConfig, error) {
	err := godotenv.Load()
	if err != nil {
		// It's okay if .env is not found, might be set in environment
		fmt.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT environment variable not set")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DB_URL environment variable not set")
	}

	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("cannot open database connection: %w", err)
	}

	err = dbConn.Ping()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}

	queries := db.New(dbConn)

	return &APIConfig{
		DB:     queries,
		Port:   port,
		DB_URL: dbURL,
	}, nil
}
