package models

import (
	"time"
)

// Booking represents a user's booking request data
type Booking struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	AdditionalPhone string    `json:"additionalPhone,omitempty"`
	PackageType     string    `json:"packageType"`
	EventDate       time.Time `json:"date"`
	Venue           string    `json:"venue"`
	City            string    `json:"city"`
	Customization   string    `json:"customization,omitempty"`
	BandTime        string    `json:"bandTime,omitempty"`
	CustomTimeSlot  string    `json:"customTimeSlot,omitempty"`
	NumberOfPeople  int       `json:"numberOfPeople,omitempty"`
	NumberOfLights  int       `json:"numberOfLights,omitempty"`
	NumberOfDhols   int       `json:"numberOfDhols,omitempty"`
	GhodaBaggi      int       `json:"ghodaBaggi,omitempty"`
	GhodiForBaraat  bool      `json:"ghodiForBaraat"`
	Fireworks       bool      `json:"fireworks"`
	FireworksAmount int       `json:"fireworksAmount,omitempty"`
	FlowerCanon     bool      `json:"flowerCanon"`
	DoliForVidai    bool      `json:"DoliForVidai"`
	Amount          int       `json:"amount"`
	AdvancePayment  int       `json:"advancePayment"`
	PhoneVerified   bool      `json:"phoneVerified"`
	CreatedAt       time.Time `json:"createdAt"`
}
