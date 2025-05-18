package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// CreateAdminUser creates a new admin user
func CreateAdminUser(c *gin.Context) {
	var request struct {
		Name         string `json:"name" binding:"required"`
		MobileNumber string `json:"mobileNumber" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Username     string `json:"username" binding:"required"`
		Password     string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	db := database.GetDB()

	// Check if username already exists in admin users
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users WHERE username = ?", request.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin username already exists"})
		return
	}

	// Check if username already exists in employees
	err = db.QueryRow("SELECT COUNT(*) FROM employees WHERE username = ?", request.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists as an employee"})
		return
	}

	// Create new admin user
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(`
		INSERT INTO admin_users (
			name, mobile_number, email, username, password, is_admin_user, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		request.Name, request.MobileNumber, request.Email, request.Username,
		string(hashedPassword), true, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user", "details": err.Error()})
		return
	}

	// Get the ID of the newly created admin
	adminID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get admin ID", "details": err.Error()})
		return
	}

	// Return success response without exposing the password
	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin user created successfully",
		"admin": gin.H{
			"id":           adminID,
			"name":         request.Name,
			"mobileNumber": request.MobileNumber,
			"email":        request.Email,
			"username":     request.Username,
			"isAdminUser":  true,
		},
	})
}
