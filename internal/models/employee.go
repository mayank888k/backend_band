package models

import (
	"time"
)

// Employee represents an employee in the system
type Employee struct {
	ID                       uint      `json:"id"`
	Name                     string    `json:"name"`
	MobileNumber             string    `json:"mobileNumber"`
	Email                    string    `json:"email"`
	Address                  string    `json:"address"`
	IsEmployee               bool      `json:"isEmployee"`
	TotalAmountToBePaid      float64   `json:"totalAmountToBePaid"`
	TotalAmountPaidInAdvance float64   `json:"totalAmountPaidInAdvance"`
	Username                 string    `json:"username"`
	Password                 string    `json:"-"` // Password is not exposed in JSON responses
	CreatedAt                time.Time `json:"createdAt"`
	UpdatedAt                time.Time `json:"updatedAt"`
	Payments                 []Payment `json:"payments,omitempty"`
}
