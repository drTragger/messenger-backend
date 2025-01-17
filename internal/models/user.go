package models

import "time"

type User struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Phone           string     `json:"phone"`
	Password        string     `json:"-"` // Omit password in JSON responses
	LastSeen        *time.Time `json:"lastSeen"`
	IsOnline        *bool      `json:"isOnline,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	PhoneVerifiedAt *time.Time `json:"phoneVerifiedAt,omitempty"`
}

type OnlineUser struct {
	IsOnline bool      `json:"isOnline"`
	LastSeen time.Time `json:"lastSeen"`
}
