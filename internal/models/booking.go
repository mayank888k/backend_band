package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Booking represents a user's booking request data
type Booking struct {
	ID              primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BookingID       string             `json:"bookingId" bson:"booking_id"`
	Name            string             `json:"name" bson:"name"`
	Email           string             `json:"email" bson:"email"`
	Phone           string             `json:"phone" bson:"phone"`
	AdditionalPhone string             `json:"additionalPhone,omitempty" bson:"additional_phone,omitempty"`
	PackageType     string             `json:"packageType" bson:"package_type"`
	EventDate       time.Time          `json:"date" bson:"event_date"`
	Venue           string             `json:"venue" bson:"venue"`
	City            string             `json:"city" bson:"city"`
	Customization   string             `json:"customization,omitempty" bson:"customization,omitempty"`
	BandTime        string             `json:"bandTime,omitempty" bson:"band_time,omitempty"`
	CustomTimeSlot  string             `json:"customTimeSlot,omitempty" bson:"custom_time_slot,omitempty"`
	NumberOfPeople  int                `json:"numberOfPeople,omitempty" bson:"number_of_people,omitempty"`
	NumberOfLights  int                `json:"numberOfLights,omitempty" bson:"number_of_lights,omitempty"`
	NumberOfDhols   int                `json:"numberOfDhols,omitempty" bson:"number_of_dhols,omitempty"`
	GhodaBaggi      int                `json:"ghodaBaggi,omitempty" bson:"ghoda_baggi,omitempty"`
	GhodiForBaraat  bool               `json:"ghodiForBaraat" bson:"ghodi_for_baraat"`
	Fireworks       bool               `json:"fireworks" bson:"fireworks"`
	FireworksAmount int                `json:"fireworksAmount,omitempty" bson:"fireworks_amount,omitempty"`
	FlowerCanon     bool               `json:"flowerCanon" bson:"flower_canon"`
	DoliForVidai    bool               `json:"DoliForVidai" bson:"doli_for_vidai"`
	Amount          int                `json:"amount" bson:"amount"`
	AdvancePayment  int                `json:"advancePayment" bson:"advance_payment"`
	PhoneVerified   bool               `json:"phoneVerified" bson:"phone_verified"`
	CreatedAt       time.Time          `json:"createdAt" bson:"created_at"`
}
