package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// generateBookingID creates a unique 6-character alphanumeric booking ID
func generateBookingID() (string, error) {
	// Define character set (alphanumeric)
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	const charsetLen = byte(len(charset))

	// Create a byte slice for the result
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		// We need a random number from 0 to len(charset)-1
		// To avoid modulo bias, we get a random byte and check if it's in the valid range
		max := 256 - (256 % int(charsetLen))
		b := make([]byte, 1)
		for {
			_, err := rand.Read(b)
			if err != nil {
				return "", err
			}
			// Reject values that would create modulo bias
			if int(b[0]) < max {
				// Use modulo now that we've eliminated the bias
				result[i] = charset[b[0]%charsetLen]
				break
			}
		}
	}

	return string(result), nil
}

// checkBookingIDExists checks if a booking ID already exists in the database
func checkBookingIDExists(ctx context.Context, coll *mongo.Collection, id string) (bool, error) {
	count, err := coll.CountDocuments(ctx, bson.M{"booking_id": id})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateBooking handles the creation of a new booking
func CreateBooking(c *gin.Context) {
	var booking models.Booking
	if err := c.ShouldBindJSON(&booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Get MongoDB collection
	coll := database.GetCollection("bookings")
	ctx := context.Background()

	// Try to generate a unique booking ID up to 5 times
	var bookingID string
	var err error
	var exists bool

	for attempt := 0; attempt < 5; attempt++ {
		// Generate a booking ID
		bookingID, err = generateBookingID()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate booking ID"})
			return
		}

		// Check if ID already exists
		exists, err = checkBookingIDExists(ctx, coll, bookingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
			return
		}

		// If ID doesn't exist, we can use it
		if !exists {
			break
		}

		// If we've reached the last attempt and still found duplicates
		if attempt == 4 && exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate unique booking ID after multiple attempts. Please try again later."})
			return
		}
	}

	// Set booking fields
	booking.BookingID = bookingID
	booking.CreatedAt = time.Now()
	booking.PhoneVerified = true

	// Insert booking into MongoDB
	result, err := coll.InsertOne(ctx, booking)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
		return
	}

	// Set the MongoDB ID
	booking.ID = result.InsertedID.(primitive.ObjectID)

	// Send booking confirmation (as a log, since we're simulating)
	log.Printf("BOOKING CONFIRMATION: Dear %s, your booking with Modern Band (ID: %s) has been confirmed! We look forward to making your event special.", booking.Name, booking.BookingID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Booking created successfully",
		"booking": booking,
	})
}

// GetBooking retrieves booking details by ID or phone number
func GetBooking(c *gin.Context) {
	bookingID := c.Query("booking_id")
	contactNumber := c.Query("contact_number")

	if bookingID == "" && contactNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either booking_id or contact_number is required"})
		return
	}

	coll := database.GetCollection("bookings")
	ctx := context.Background()

	var filter bson.M
	if bookingID != "" {
		filter = bson.M{"booking_id": bookingID}
	} else {
		filter = bson.M{"phone": contactNumber}
	}

	// Set options for sorting by created_at in descending order
	opts := options.Find()
	if contactNumber != "" {
		opts.SetSort(bson.D{{"created_at", -1}})
	}

	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve booking"})
		return
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse booking data"})
		return
	}

	if len(bookings) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No bookings found"})
		return
	}

	if bookingID != "" && len(bookings) == 1 {
		// If querying by ID, return single booking
		c.JSON(http.StatusOK, gin.H{"booking": bookings[0]})
	} else {
		// If querying by phone, return multiple bookings
		c.JSON(http.StatusOK, gin.H{"bookings": bookings})
	}
}

// GetAllBookings retrieves all bookings from the database
func GetAllBookings(c *gin.Context) {
	coll := database.GetCollection("bookings")
	ctx := context.Background()

	// Set options for sorting by created_at in descending order
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bookings", "details": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking
	if err = cursor.All(ctx, &bookings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse booking data", "details": err.Error()})
		return
	}

	// Check if any bookings were found
	if len(bookings) == 0 {
		c.JSON(http.StatusOK, gin.H{"bookings": []models.Booking{}, "count": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"count":    len(bookings),
	})
}

// DeleteBooking deletes a booking by ID
func DeleteBooking(c *gin.Context) {
	bookingID := c.Param("id")

	// Add more debug logging
	fmt.Printf("DeleteBooking handler called with ID: '%s'\n", bookingID)

	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking ID is required"})
		return
	}

	coll := database.GetCollection("bookings")
	ctx := context.Background()

	// Check if booking exists
	count, err := coll.CountDocuments(ctx, bson.M{"booking_id": bookingID})
	if err != nil {
		fmt.Printf("Database error when checking if booking exists: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	fmt.Printf("Found %d bookings with ID: '%s'\n", count, bookingID)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// Delete the booking
	result, err := coll.DeleteOne(ctx, bson.M{"booking_id": bookingID})
	if err != nil {
		fmt.Printf("Database error when deleting booking: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete booking", "details": err.Error()})
		return
	}

	fmt.Printf("Successfully deleted booking with ID: '%s'\n", bookingID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Booking deleted successfully",
		"id":      bookingID,
		"count":   result.DeletedCount,
	})
}

// DeletePastBookings deletes all bookings with event dates in the past
func DeletePastBookings(c *gin.Context) {
	coll := database.GetCollection("bookings")
	ctx := context.Background()

	// Calculate the start of today (midnight)
	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Delete bookings with event dates from yesterday or older
	result, err := coll.DeleteMany(ctx, bson.M{
		"event_date": bson.M{
			"$lt": startOfToday, // Only delete bookings before today's start (midnight)
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete past bookings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Past bookings (before today) deleted successfully",
		"count":   result.DeletedCount,
	})
}
