package db

import (
	"database/sql"
	"fmt"
	"log"

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
