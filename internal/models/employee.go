package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Employee represents an employee in the system
type Employee struct {
	ID                       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                     string             `json:"name" bson:"name"`
	MobileNumber             string             `json:"mobileNumber" bson:"mobile_number"`
	Email                    string             `json:"email" bson:"email"`
	Address                  string             `json:"address" bson:"address"`
	IsEmployee               bool               `json:"isEmployee" bson:"is_employee"`
	TotalAmountToBePaid      float64            `json:"totalAmountToBePaid" bson:"total_amount_to_be_paid"`
	TotalAmountPaidInAdvance float64            `json:"totalAmountPaidInAdvance" bson:"total_amount_paid_in_advance"`
	Username                 string             `json:"username" bson:"username"`
	Password                 string             `json:"-" bson:"password"` // Password is not exposed in JSON responses
	CreatedAt                time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt                time.Time          `json:"updatedAt" bson:"updated_at"`
	Payments                 []Payment          `json:"payments,omitempty" bson:"payments,omitempty"`
}
