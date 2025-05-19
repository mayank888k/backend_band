package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LoginRequest represents the request for user login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateEmployeeRequest represents the request for creating a new employee
type CreateEmployeeRequest struct {
	Name                     string  `json:"name" binding:"required"`
	MobileNumber             string  `json:"mobileNumber" binding:"required"`
	Email                    string  `json:"email" binding:"required,email"`
	Address                  string  `json:"address" binding:"required"`
	TotalAmountToBePaid      float64 `json:"totalAmountToBePaid"`
	TotalAmountPaidInAdvance float64 `json:"totalAmountPaidInAdvance"`
	Username                 string  `json:"username" binding:"required"`
	Password                 string  `json:"password" binding:"required"`
}

// CreatePaymentRequest represents the request for adding a payment
type CreatePaymentRequest struct {
	AmountPaid float64 `json:"amountPaid" binding:"required,gt=0"`
	Date       string  `json:"date" binding:"required"` // Will be parsed to time.Time
}

// CreateAdminUserRequest represents the request for creating a new admin user
type CreateAdminUserRequest struct {
	Name         string `json:"name" binding:"required"`
	MobileNumber string `json:"mobileNumber" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// DeletePaymentRequest represents the request for deleting a payment
type DeletePaymentRequest struct {
	PaymentID primitive.ObjectID `json:"paymentId" binding:"required"`
}
