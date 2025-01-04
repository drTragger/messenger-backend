package repository

import (
	"database/sql"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
	"time"
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
		INSERT INTO users (username, phone, password) 
		VALUES ($1, $2, $3)
	`
	_, err := repo.DB.Exec(query, user.Username, user.Phone, user.Password)
	return err
}

// GetUserByPhone fetches a user by phone
func (repo *UserRepository) GetUserByPhone(phone string) (*models.User, error) {
	query := `
		SELECT id, username, phone, password, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE phone = $1
	`

	row := repo.DB.QueryRow(query, phone)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.Password, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

// GetUserByID fetches a user by ID
func (repo *UserRepository) GetUserByID(userID uint) (*models.User, error) {
	query := `
		SELECT id, username, phone, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE id = $1
	`

	row := repo.DB.QueryRow(query, userID)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

func (repo *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, phone, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE username = $1
	`

	row := repo.DB.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

func (repo *UserRepository) VerifyPhone(phone string) error {
	query := `
		UPDATE users
		SET phone_verified_at = $1
		WHERE phone = $2;
	`

	_, err := repo.DB.Exec(query, time.Now(), phone)
	return err
}
