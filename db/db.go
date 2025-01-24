package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/drTragger/messenger-backend/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func InitDB(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to the database")
	return db, nil
}

// LoadQuery loads an SQL query from the specified file path.
// `fileName` is the name of the SQL file, and `subDir` is an optional list of subdirectories.
func LoadQuery(fileName string, subDir ...string) (string, error) {
	parts := append([]string{"db", "queries"}, subDir...)
	parts = append(parts, fileName)
	path := filepath.Join(parts...)

	queryBytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(queryBytes)), nil
}
