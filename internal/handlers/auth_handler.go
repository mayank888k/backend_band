package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Login handles authentication for both employees and admin users
func Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	db := database.GetDB()

	// First try to find an admin user
	var adminID int
	var adminPassword string
	err := db.QueryRow("SELECT id, password FROM admin_users WHERE username = ?", request.Username).Scan(&adminID, &adminPassword)
	if err == nil {
		// Found an admin user, verify password
		err := bcrypt.CompareHashAndPassword([]byte(adminPassword), []byte(request.Password))
		if err == nil {
			// Password is correct
			c.JSON(http.StatusOK, models.LoginResponse{
				Username: request.Username,
				IsAdmin:  true,
				Message:  "Login successful",
			})
			return
		}
	} else if err != sql.ErrNoRows {
		// Some other database error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// If not found or password incorrect, try employee
	var employeeID int
	var employeePassword string
	err = db.QueryRow("SELECT id, password FROM employees WHERE username = ?", request.Username).Scan(&employeeID, &employeePassword)
	if err == nil {
		// Found an employee, verify password
		err := bcrypt.CompareHashAndPassword([]byte(employeePassword), []byte(request.Password))
		if err == nil {
			// Password is correct
			c.JSON(http.StatusOK, models.LoginResponse{
				Username: request.Username,
				IsAdmin:  false,
				Message:  "Login successful",
			})
			return
		}
	} else if err != sql.ErrNoRows {
		// Some other database error occurred
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	// If we get here, authentication failed
	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
}
