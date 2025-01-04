package repository

import (
	"database/sql"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser inserts a new user into the database
func (repo *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password) 
		VALUES ($1, $2, $3)
	`
	_, err := repo.DB.Exec(query, user.Username, user.Email, user.Password)
	return err
}

// GetUserByEmail fetches a user by email
func (repo *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`

	row := repo.DB.QueryRow(query, email)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

// GetUserByID fetches a user by ID
func (repo *UserRepository) GetUserByID(userID int) (*models.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	row := repo.DB.QueryRow(query, userID)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

func (repo *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		WHERE username = $1
	`

	row := repo.DB.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}
