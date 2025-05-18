package models

import (
	"time"
)

// AdminUser represents an admin user in the system
type AdminUser struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	MobileNumber string    `json:"mobileNumber"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Password     string    `json:"-"` // Password is not exposed in JSON responses
	IsAdminUser  bool      `json:"isAdminUser"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
