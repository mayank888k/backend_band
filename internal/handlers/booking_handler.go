package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
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
func checkBookingIDExists(db *sql.DB, id string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM bookings WHERE id = ?", id).Scan(&count)
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

	// Get database connection
	db := database.GetDB()

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
		exists, err = checkBookingIDExists(db, bookingID)
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

	booking.ID = bookingID
	booking.CreatedAt = time.Now()

	// Set phone verified to true (OTP verification removed)
	booking.PhoneVerified = true

	// Save booking to database
	_, err = db.Exec(`
		INSERT INTO bookings (
			id, name, email, phone, additional_phone, package_type, 
			event_date, venue, city, customization, band_time, 
			custom_time_slot, number_of_people, number_of_lights, 
			number_of_dhols, ghoda_baggi, ghodi_for_baraat, 
			fireworks, fireworks_amount, flower_canon, doli_for_vidai, 
			amount, advance_payment, phone_verified, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		booking.ID, booking.Name, booking.Email, booking.Phone, booking.AdditionalPhone, booking.PackageType,
		booking.EventDate.Format(time.RFC3339), booking.Venue, booking.City, booking.Customization, booking.BandTime,
		booking.CustomTimeSlot, booking.NumberOfPeople, booking.NumberOfLights,
		booking.NumberOfDhols, booking.GhodaBaggi, booking.GhodiForBaraat,
		booking.Fireworks, booking.FireworksAmount, booking.FlowerCanon, booking.DoliForVidai,
		booking.Amount, booking.AdvancePayment, booking.PhoneVerified, booking.CreatedAt.Format(time.RFC3339))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking", "details": err.Error()})
		return
	}

	// Send booking confirmation (as a log, since we're simulating)
	log.Printf("BOOKING CONFIRMATION: Dear %s, your booking with Modern Band (ID: %s) has been confirmed! We look forward to making your event special.", booking.Name, booking.ID)

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

	db := database.GetDB()
	var query string
	var arg interface{}

	if bookingID != "" {
		query = `
			SELECT id, name, email, phone, additional_phone, package_type, 
			event_date, venue, city, customization, band_time, 
			custom_time_slot, number_of_people, number_of_lights, 
			number_of_dhols, ghoda_baggi, ghodi_for_baraat, 
			fireworks, fireworks_amount, flower_canon, doli_for_vidai, 
			amount, advance_payment, phone_verified, created_at
			FROM bookings WHERE id = ?
		`
		arg = bookingID
	} else {
		query = `
			SELECT id, name, email, phone, additional_phone, package_type, 
			event_date, venue, city, customization, band_time, 
			custom_time_slot, number_of_people, number_of_lights, 
			number_of_dhols, ghoda_baggi, ghodi_for_baraat, 
			fireworks, fireworks_amount, flower_canon, doli_for_vidai, 
			amount, advance_payment, phone_verified, created_at
			FROM bookings WHERE phone = ?
			ORDER BY created_at DESC
		`
		arg = contactNumber
	}

	rows, err := db.Query(query, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve booking"})
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		var eventDateStr string
		var createdAtStr string

		err := rows.Scan(
			&booking.ID, &booking.Name, &booking.Email, &booking.Phone, &booking.AdditionalPhone, &booking.PackageType,
			&eventDateStr, &booking.Venue, &booking.City, &booking.Customization, &booking.BandTime,
			&booking.CustomTimeSlot, &booking.NumberOfPeople, &booking.NumberOfLights,
			&booking.NumberOfDhols, &booking.GhodaBaggi, &booking.GhodiForBaraat,
			&booking.Fireworks, &booking.FireworksAmount, &booking.FlowerCanon, &booking.DoliForVidai,
			&booking.Amount, &booking.AdvancePayment, &booking.PhoneVerified, &createdAtStr,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse booking data"})
			return
		}

		// Parse the dates
		booking.EventDate, _ = time.Parse(time.RFC3339, eventDateStr)
		booking.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		bookings = append(bookings, booking)
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
	db := database.GetDB()

	query := `
		SELECT id, name, email, phone, additional_phone, package_type, 
		event_date, venue, city, customization, band_time, 
		custom_time_slot, number_of_people, number_of_lights, 
		number_of_dhols, ghoda_baggi, ghodi_for_baraat, 
		fireworks, fireworks_amount, flower_canon, doli_for_vidai, 
		amount, advance_payment, phone_verified, created_at
		FROM bookings
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bookings", "details": err.Error()})
		return
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		var eventDateStr string
		var createdAtStr string

		err := rows.Scan(
			&booking.ID, &booking.Name, &booking.Email, &booking.Phone, &booking.AdditionalPhone, &booking.PackageType,
			&eventDateStr, &booking.Venue, &booking.City, &booking.Customization, &booking.BandTime,
			&booking.CustomTimeSlot, &booking.NumberOfPeople, &booking.NumberOfLights,
			&booking.NumberOfDhols, &booking.GhodaBaggi, &booking.GhodiForBaraat,
			&booking.Fireworks, &booking.FireworksAmount, &booking.FlowerCanon, &booking.DoliForVidai,
			&booking.Amount, &booking.AdvancePayment, &booking.PhoneVerified, &createdAtStr,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse booking data", "details": err.Error()})
			return
		}

		// Parse the dates
		booking.EventDate, _ = time.Parse(time.RFC3339, eventDateStr)
		booking.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		bookings = append(bookings, booking)
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

// DeletePastBookings deletes all bookings whose event date is in the past
func DeletePastBookings(c *gin.Context) {
	db := database.GetDB()

	// Get current date in RFC3339 format (used by the database)
	currentDate := time.Now().Format(time.RFC3339)

	// Delete bookings with event_date earlier than current date
	result, err := db.Exec("DELETE FROM bookings WHERE event_date < ?", currentDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete past bookings", "details": err.Error()})
		return
	}

	// Get number of affected rows (deleted bookings)
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get number of deleted bookings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Past bookings deleted successfully",
		"deleted_count": rowsAffected,
	})
}

// DeleteBooking deletes a specific booking by its ID
func DeleteBooking(c *gin.Context) {
	bookingID := c.Param("id")

	// Add more debug logging
	fmt.Printf("DeleteBooking handler called with ID: '%s'\n", bookingID)

	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking ID is required"})
		return
	}

	db := database.GetDB()

	// Check if booking exists
	var count int
	query := "SELECT COUNT(*) FROM bookings WHERE id = ?"
	fmt.Printf("Running query: %s with ID: '%s'\n", query, bookingID)

	err := db.QueryRow(query, bookingID).Scan(&count)
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
	deleteQuery := "DELETE FROM bookings WHERE id = ?"
	fmt.Printf("Running delete query: %s with ID: '%s'\n", deleteQuery, bookingID)

	_, err = db.Exec(deleteQuery, bookingID)
	if err != nil {
		fmt.Printf("Database error when deleting booking: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete booking", "details": err.Error()})
		return
	}

	fmt.Printf("Successfully deleted booking with ID: '%s'\n", bookingID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Booking deleted successfully",
		"id":      bookingID,
	})
}
