package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Payment represents a payment made to an employee
type Payment struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AmountPaid float64            `json:"amountPaid" bson:"amount_paid"`
	Date       time.Time          `json:"date" bson:"date"`
	EmployeeID primitive.ObjectID `json:"employeeId" bson:"employee_id"`
	CreatedAt  time.Time          `json:"createdAt" bson:"created_at"`
}
