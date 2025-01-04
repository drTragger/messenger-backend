package models

import "time"

type User struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Phone           string     `json:"phone"`
	Password        string     `json:"-"` // Omit password in JSON responses
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	PhoneVerifiedAt *time.Time `json:"phoneVerifiedAt"`
}
