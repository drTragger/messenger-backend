package repository

import (
	"database/sql"
	"errors"
	"github.com/drTragger/messenger-backend/internal/models"
)

const (
	UserSearchLimit = 10
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser inserts a new user into the database
func (ur *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, phone, password) 
		VALUES ($1, $2, $3)
	`
	_, err := ur.DB.Exec(query, user.Username, user.Phone, user.Password)
	return err
}

// GetUserByPhone fetches a user by phone
func (ur *UserRepository) GetUserByPhone(phone string) (*models.User, error) {
	query := `
		SELECT id, username, phone, password, last_seen, profile_picture, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE phone = $1
	`

	row := ur.DB.QueryRow(query, phone)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.Password, &user.LastSeen, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

// GetUserByID fetches a user by ID
func (ur *UserRepository) GetUserByID(userID uint) (*models.User, error) {
	query := `
		SELECT id, username, phone, last_seen, profile_picture, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE id = $1
	`

	row := ur.DB.QueryRow(query, userID)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.LastSeen, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

func (ur *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, phone, last_seen, profile_picture, created_at, updated_at, phone_verified_at 
		FROM users 
		WHERE username = $1
	`

	row := ur.DB.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Phone, &user.LastSeen, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt, &user.PhoneVerifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // User not found
	}
	return user, err
}

func (ur *UserRepository) GetUsersBySearch(search string) ([]*models.User, error) {
	query := `
		SELECT id, username, phone, last_seen, profile_picture, created_at, updated_at
		FROM users
		WHERE phone ILIKE $1 OR username ILIKE $1
		LIMIT $2
	`

	searchTerm := "%" + search + "%"

	rows, err := ur.DB.Query(query, searchTerm, UserSearchLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User

		err := rows.Scan(&user.ID, &user.Username, &user.Phone, &user.LastSeen, &user.ProfilePicture, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *UserRepository) VerifyPhone(phone string) error {
	query := `
		UPDATE users
		SET phone_verified_at = NOW(), updated_at = NOW()
		WHERE phone = $1;
	`

	_, err := ur.DB.Exec(query, phone)
	return err
}

func (ur *UserRepository) UpdateLastSeen(userID uint) error {
	query := `
		UPDATE users
		SET last_seen = NOW(), updated_at = NOW()
		WHERE id = $1;
	`

	_, err := ur.DB.Exec(query, userID)
	return err
}

func (ur *UserRepository) UpdateProfilePicture(userID uint, picturePath *string) error {
	query := `
		UPDATE users
		SET profile_picture = $1, updated_at = NOW()
		WHERE id = $2;
	`

	_, err := ur.DB.Exec(query, picturePath, userID)
	return err
}
