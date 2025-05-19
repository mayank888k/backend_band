package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminUser represents an admin user in the system
type AdminUser struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`
	MobileNumber string             `json:"mobileNumber" bson:"mobile_number"`
	Email        string             `json:"email" bson:"email"`
	Username     string             `json:"username" bson:"username"`
	Password     string             `json:"-" bson:"password"` // Password is not exposed in JSON responses
	IsAdminUser  bool               `json:"isAdminUser" bson:"is_admin_user"`
	CreatedAt    time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updated_at"`
}
