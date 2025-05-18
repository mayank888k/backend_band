package models

import (
	"time"
)

// Payment represents a payment made to an employee
type Payment struct {
	ID         uint      `json:"id"`
	AmountPaid float64   `json:"amountPaid"`
	Date       time.Time `json:"date"`
	EmployeeID uint      `json:"employeeId"`
	CreatedAt  time.Time `json:"createdAt"`
}
